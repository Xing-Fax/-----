package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/dchest/captcha"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

var captchaStore = struct {
	sync.RWMutex
	store map[string]captchaData
}{store: make(map[string]captchaData)}

type captchaData struct {
	answer  string
	created time.Time
}

const captchaExpiryDuration = 5 * time.Minute

func GenerateCaptcha(r *ghttp.Request) {
	id := captcha.New()

	captchaPath := filepath.Join("images", id+".png")

	answer := captcha.RandomDigits(6)

	var result string
	for _, b := range answer {
		result += strconv.Itoa(int(b))
	}

	captchaStore.Lock()
	captchaStore.store[id] = captchaData{
		answer:  result,
		created: time.Now(),
	}
	captchaStore.Unlock()

	img := captcha.NewImage(id, answer, 240, 80)

	file, err := os.Create(captchaPath)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    10001,
			"message": "Failed to generate verification code.",
		})
		return
	}
	defer file.Close()

	_, err = img.WriteTo(file)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    10001,
			"message": "Failed to generate verification code.",
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"code": 200,
		"id":   id,
		"url":  "/captcha/" + id + ".png",
	})
}

func VerifyCaptcha(r *ghttp.Request) bool {
	id := r.Get("verifyID").String()
	answer := r.Get("verifyCode").String()

	captchaStore.RLock()
	captchaData, exists := captchaStore.store[id]
	captchaStore.RUnlock()

	if !exists {
		// 没有此验证码
		return false
	}

	if time.Since(captchaData.created) > captchaExpiryDuration {
		DelectVerify(id)
		// 验证码超时
		return false
	}

	if answer == captchaData.answer {
		DelectVerify(id)
		// 验证成功
		return true
	} else {
		DelectVerify(id)
		// 无效验证码
		return false
	}
}

func DelectVerify(id string) {
	// 删除文件
	captchaPath := filepath.Join("images", id+".png")
	if err := os.Remove(captchaPath); err != nil {
		fmt.Println("Error deleting captcha image:", err)
	}
	// 删除验证码
	captchaStore.Lock()
	delete(captchaStore.store, id)
	captchaStore.Unlock()
}
