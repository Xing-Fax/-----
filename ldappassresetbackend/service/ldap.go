package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"golang.org/x/text/encoding/unicode"
)

type LDAPService struct {
	host          string
	port          string
	disableTLS    bool
	baseDn        string
	adminUser     string
	adminPassword string
	conn          *ldap.Conn
}

var (
	ldapServiceInstance *LDAPService
)

func NewLDAPService() (*LDAPService, error) {

	var initErr error = nil // 用来记录初始化过程中的错误

	cfg := g.Cfg().MustGet(context.TODO(), "ldap").Map()
	host, _ := cfg["host"].(string)
	port, _ := cfg["port"].(string)
	disableTLS, _ := cfg["disableTLS"].(bool)
	baseDn, _ := cfg["baseDn"].(string)
	adminUser, _ := cfg["adminUser"].(string)
	adminPassword, _ := cfg["adminPassword"].(string)

	ldapServiceInstance = &LDAPService{
		host:          host,
		port:          port,
		disableTLS:    disableTLS,
		baseDn:        baseDn,
		adminUser:     adminUser,
		adminPassword: adminPassword,
	}

	if err := ldapServiceInstance.connect(); err != nil {
		initErr = fmt.Errorf("failed to connect to LDAP server: %w", err)
	}

	// 如果初始化过程中出现错误，返回错误信息
	if initErr != nil {
		return nil, initErr
	}

	// 返回初始化后的实例
	return ldapServiceInstance, nil
}

func (s *LDAPService) connect() error {
	ldapURL := fmt.Sprintf("%s:%s", s.host, s.port)
	conn, err := ldap.DialURL(ldapURL, ldap.DialWithTLSConfig(&tls.Config{
		InsecureSkipVerify: s.disableTLS,
	}))
	if err != nil {
		return err
	}
	if err := conn.Bind(s.adminUser, s.adminPassword); err != nil {
		return err
	}
	s.conn = conn
	return nil
}

// Validate password with multiple checks
func validatePassword(password string) error {
	// Check minimum length
	if len(password) < 8 {
		return fmt.Errorf("verification failed")
	}

	// Check if contains at least one letter
	if !regexp.MustCompile(`[A-Za-z]`).MatchString(password) {
		return fmt.Errorf("verification failed")
	}

	// Check if contains at least one digit
	if !regexp.MustCompile(`\d`).MatchString(password) {
		return fmt.Errorf("verification failed")
	}

	// Check for allowed characters
	if strings.ContainsAny(password, " ") || !regexp.MustCompile(`^[A-Za-z\d!@#$%^&*(),.?":{}|<>]+$`).MatchString(password) {
		return fmt.Errorf("verification failed")
	}
	return nil
}

func ResetPassword(r *ghttp.Request) {
	username := r.Get("username").String()
	codeType := r.Get("type").String()
	ldapService, err := NewLDAPService()
	if err != nil {
		r.Response.WriteJsonExit(g.Map{"code": 10007, "message": err.Error()})
		return
	}
	// 获取用户的手机号码和邮箱
	mobile, mail, _, err := ldapService.GetUser(username)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{"code": 10008, "message": err.Error()})
		return
	}
	// 判断认证方式
	var identifier string

	if codeType == "mail" {
		identifier = mail
	}

	if codeType == "mobile" {
		identifier = mobile
	}

	if identifier == "" {
		r.Response.WriteJsonExit(g.Map{"code": 1005, "message": "Invalid code"})
		return
	}
	// 校验验证码
	code := r.Get("verifyCode").String()
	if !VerifyCode(identifier, code) {
		r.Response.WriteJsonExit(g.Map{"code": 1006, "message": "Invalid code"})
		return
	}

	newPassword := r.Get("newPassword").String()

	// 解密密码
	decryptedPassword, err := DecryptPassword(newPassword)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    10002,
			"message": "Failed to decrypt password",
		})
	}
	// 二次校验密码
	if err := validatePassword(decryptedPassword); err != nil {
		r.Response.WriteJson(g.Map{
			"code":    10003,
			"message": err.Error(),
		})
		return
	}
	// 发起重置密码请求
	if err := ldapService.Reset(username, decryptedPassword); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    10004,
			"message": err.Error(),
		})
		return
	}
	DeleteCode(identifier) // 删除验证码
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "Success",
	})
}

func GetUserInfo(r *ghttp.Request) {
	username := r.Get("username").String()
	// 查找用户信息
	ldapService, err := NewLDAPService()
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    10007,
			"message": "Failed to connect to LDAP",
		})
		return
	}

	// 获取用户的手机号码和邮箱
	mobile, mail, _, err := ldapService.GetUser(username)
	if err != nil {
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

	// 对手机号进行打码
	maskedMobile := maskMobile(mobile)

	// 对邮箱进行打码
	maskedMail := maskMail(mail)

	// 返回打码后的信息
	r.Response.WriteJsonExit(g.Map{
		"code":   200,
		"mobile": maskedMobile,
		"mail":   maskedMail,
	})
}

// 打码手机号
func maskMobile(mobile string) string {
	if len(mobile) < 7 {
		return mobile // 如果手机号长度小于7，无法打码，直接返回原值
	}
	// 保留前3位，后1位，中间用星号替代
	return fmt.Sprintf("%s****%s", mobile[:3], mobile[len(mobile)-1:])
}

// 打码邮箱
func maskMail(mail string) string {
	// 获取邮箱的@符号位置
	atIndex := strings.Index(mail, "@")
	if atIndex == -1 || atIndex < 3 {
		return mail // 如果没有@符号或者邮箱长度不够，返回原值
	}

	// 分割邮箱前缀和域名
	localPart := mail[:atIndex]
	domain := mail[atIndex+1:]

	// 对前缀打码（保留前三个字符，后面用星号代替）
	maskedLocal := localPart[:3] + strings.Repeat("*", len(localPart)-3)

	// 对域名打码（保留前两位和后缀，其余用星号代替）
	domainParts := strings.Split(domain, ".")
	if len(domainParts) < 2 {
		return mail // 如果域名格式不正确，返回原值
	}

	domainPrefix := domainParts[0]
	domainSuffix := strings.Join(domainParts[1:], ".")

	maskedDomain := domainPrefix[:2] + strings.Repeat("*", len(domainPrefix)-2)

	return fmt.Sprintf("%s@%s.%s", maskedLocal, maskedDomain, domainSuffix)
}

func (s *LDAPService) Reset(username, newPassword string) error {
	if s.conn == nil {
		if err := s.connect(); err != nil {
			return fmt.Errorf("failed to reconnect to LDAP server: %v", err)
		}
	}

	searchRequest := ldap.NewSearchRequest(
		s.baseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(|(sAMAccountName=%s)(mobile=%s)(mail=%s))", username, username, username),
		[]string{"dn"},
		nil,
	)
	sr, err := s.conn.Search(searchRequest)
	if err != nil || len(sr.Entries) == 0 {
		return fmt.Errorf("user not found")
	}

	userDN := sr.Entries[0].DN

	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	pwdEncoded, err := utf16.NewEncoder().String("\"" + newPassword + "\"")
	if err != nil {
		return fmt.Errorf("failed to modify password: %v", err)
	}

	passwordModify := ldap.NewModifyRequest(userDN, nil)
	passwordModify.Replace("unicodePwd", []string{pwdEncoded})
	passwordModify.Replace("userAccountControl", []string{"512"})
	err = s.conn.Modify(passwordModify)
	if err != nil {
		return fmt.Errorf("failed to modify password: %v", err)
	}
	return nil
}

func (s *LDAPService) GetUser(username string) (string, string, string, error) {
	if s.conn == nil {
		if err := s.connect(); err != nil {
			return "", "", "", fmt.Errorf("failed to reconnect to LDAP server:  %v", err)
		}
	}

	// 创建搜索请求，查找用户名匹配的用户
	searchRequest := ldap.NewSearchRequest(
		s.baseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(|(sAMAccountName=%s)(mobile=%s)(mail=%s))", username, username, username),
		[]string{"sAMAccountName", "mobile", "mail", "name"}, // 请求返回这些属性
		nil,
	)

	// 执行搜索
	sr, err := s.conn.Search(searchRequest)
	if err != nil || len(sr.Entries) == 0 {
		return "", "", "", fmt.Errorf("user not found")
	}

	// 获取第一个匹配的条目的属性值
	entry := sr.Entries[0]

	mobile := entry.GetAttributeValue("mobile")
	mail := entry.GetAttributeValue("mail")
	name := entry.GetAttributeValue("name")

	return mobile, mail, name, nil
}
