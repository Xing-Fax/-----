package service

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/gogf/gf/os/gctx"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

var (
	codeStorage = make(map[string]codeData)
	mu          sync.Mutex // 确保并发安全
)

type codeData struct {
	code     string
	created  time.Time
	tryCount int       // 跟踪尝试次数
	lastSend time.Time // 记录上次发送时间
}

// 设定超时时间
const codeExpiryDuration = 5 * time.Minute
const maxTryCount = 5                    // 最大试错次数
const minSendInterval = 60 * time.Second // 最小发送间隔时间（60秒）

func GenerateCode() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000)) // 生成 0 到 999999 的随机数
	if err != nil {
		panic(err) // 如果随机数生成失败，直接终止程序
	}
	return fmt.Sprintf("%06d", n.Int64()) // 格式化为6位数字，不足补0
}

// 存储验证码和生成时间
func StoreCode(identifier, code string) {
	mu.Lock()
	defer mu.Unlock()

	codeStorage[identifier] = codeData{
		code:     code,
		created:  time.Now(),
		tryCount: 0,          // 初始化尝试次数为 0
		lastSend: time.Now(), // 记录发送时间
	}
}

// 删除验证码
func DeleteCode(identifier string) {
	mu.Lock()
	defer mu.Unlock()

	delete(codeStorage, identifier)
}

// 验证验证码并检查是否超时和尝试次数
func VerifyCode(identifier, code string) bool {
	mu.Lock()
	defer mu.Unlock()

	if storedCodeData, ok := codeStorage[identifier]; ok {
		// 检查验证码是否过期
		if time.Since(storedCodeData.created) > codeExpiryDuration {
			// 超过超时时间，删除验证码
			delete(codeStorage, identifier)
			return false
		}

		// 检查尝试次数是否超出限制
		if storedCodeData.tryCount >= maxTryCount {
			// 超出最大尝试次数，删除验证码
			delete(codeStorage, identifier)
			return false
		}

		// 验证验证码是否匹配
		if storedCodeData.code == code {
			return true
		}

		// 验证失败，增加尝试次数
		storedCodeData.tryCount++
		codeStorage[identifier] = storedCodeData
	}

	return false
}

// 最小发送间隔检查
func isAllowedToSend(identifier string) bool {
	mu.Lock()
	defer mu.Unlock()

	if storedCodeData, ok := codeStorage[identifier]; ok {
		if time.Since(storedCodeData.lastSend) < minSendInterval {
			return false // 如果时间间隔小于最小发送间隔，返回 false
		}
	}

	return true // 允许发送验证码
}

func SendVerificationCode(r *ghttp.Request) {
	// 校验验证码
	if !VerifyCaptcha(r) {
		r.Response.WriteJson(g.Map{
			"code":    10006,
			"message": "Invalid code",
		})
		return
	}

	codeType := r.Get("type").String()
	username := r.Get("username").String()
	ldapService, err := NewLDAPService()
	if err != nil {
		g.Log().Info(gctx.New(), err.Error())
		r.Response.WriteJsonExit(g.Map{
			"code":    10007,
			"message": err.Error(),
		})
		return
	}

	// 获取用户的手机号码和邮箱
	mobile, mail, name, err := ldapService.GetUser(username)
	if err != nil {
		g.Log().Info(gctx.New(), err.Error())
		if err.Error() == "user not found" {
			r.Response.WriteJsonExit(g.Map{
				"code":    10012,
				"message": err.Error(),
			})
		}
		r.Response.WriteJsonExit(g.Map{
			"code":    10008,
			"message": err.Error(),
		})
		return
	}

	// 判断验证方式
	var identifier string
	if codeType == "mail" {
		identifier = mail
	} else if codeType == "mobile" {
		identifier = mobile
	} else {
		r.Response.WriteJsonExit(g.Map{
			"code":    10005,
			"message": "Invalid data",
		})
		return
	}

	// 检查是否可以发送验证码
	if !isAllowedToSend(identifier) {
		r.Response.WriteJson(g.Map{
			"code":    10009,
			"message": "Please wait 60 seconds before requesting a new verification code.",
		})
		return
	}

	// 生成随机验证码并绑定
	code := GenerateCode()
	StoreCode(identifier, code)

	// 更新验证码发送时间
	mu.Lock()
	if storedCodeData, ok := codeStorage[identifier]; ok {
		storedCodeData.lastSend = time.Now()     // 更新 lastSend 时间
		codeStorage[identifier] = storedCodeData // 重新存储修改后的数据
	}
	mu.Unlock()

	// 发送验证码
	if codeType == "mail" {
		if err := NewEmailService().SendEmail(name, mail, code); err != nil {
			r.Response.WriteJsonExit(g.Map{
				"code":    10010,
				"message": err.Error(),
			})
		}
		r.Response.WriteJsonExit(g.Map{
			"code":    200,
			"message": "Success",
		})
	} else if codeType == "mobile" {
		if err := SendSms(mobile, code); err != nil {
			r.Response.WriteJsonExit(g.Map{
				"code":    10011,
				"message": err.Error(),
			})
		}
		r.Response.WriteJsonExit(g.Map{
			"code":    200,
			"message": "Success",
		})
	}
}

// 验证发送的验证码
func VerificationCode(r *ghttp.Request) {
	username := r.Get("username").String()
	codeType := r.Get("type").String()
	ldapService, err := NewLDAPService()
	if err != nil {
		r.Response.WriteJsonExit(g.Map{"code": 10004, "message": "Failed to connect to LDAP"})
		return
	}

	// 获取用户的手机号码和邮箱
	mobile, mail, _, err := ldapService.GetUser(username)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{"code": 10008, "message": err.Error()})
		return
	}

	var identifier string
	if codeType == "mail" {
		identifier = mail
	} else if codeType == "mobile" {
		identifier = mobile
	} else {
		r.Response.WriteJsonExit(g.Map{"code": 10005, "message": "Invalid data"})
		return
	}

	code := r.Get("verifyCode").String()
	g.Log().Info(gctx.New(), username+codeType+code)
	if VerifyCode(identifier, code) {
		r.Response.WriteJsonExit(g.Map{"code": 200, "message": "Success"})
		return
	}
	r.Response.WriteJsonExit(g.Map{"code": 10006, "message": "Invalid code"})
}
