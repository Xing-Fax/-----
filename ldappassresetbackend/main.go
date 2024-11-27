package main

import (
	"context"
	"fmt"
	"ldap-password-reset/service"

	"github.com/gogf/gf/os/gctx"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func main() {

	// 获取上下文
	// ctx := gctx.New()

	path, _ := g.Cfg().Get(context.Background(), "logger.path")
	level, _ := g.Cfg().Get(context.Background(), "logger.level")
	stdout, _ := g.Cfg().Get(context.Background(), "logger.stdout")
	// 配置日志组件
	g.Log().SetConfigWithMap(g.Map{
		"path":   path,   // 日志文件存储路径
		"level":  level,  // 日志级别（all, debug, info, notice, warning, error, critical）
		"stdout": stdout, // 是否输出到控制台
	})
	g.Log().Info(gctx.New(), "程序启动...")

	portVar, err := g.Cfg().Get(context.Background(), "server.port")
	if err != nil {
		fmt.Println("Error reading config:", err)
		return
	}

	port := portVar.Int()
	s := g.Server()
	s.SetAddr(fmt.Sprintf(":%d", port))

	// 验证码路径绑定
	s.AddStaticPath("/captcha", "images")
	s.AddStaticPath("/static", "public")
	s.SetServerRoot("public") // 静态文件目录为 public

	// 查找用户
	s.BindHandler("/api/get-user-info", func(r *ghttp.Request) {
		if r.Method != "POST" {
			r.Response.WriteJsonExit(g.Map{"success": false, "error": "Invalid request method"})
			return
		}
		username := r.Get("username").String()
		if username == "" {
			r.Response.WriteJsonExit(g.Map{"success": false, "error": "Username is required"})
			return
		}
		service.GetUserInfo(r)
	})

	// 创建图形验证码
	s.BindHandler("/api/generate-captcha", func(r *ghttp.Request) {
		if r.Method != "GET" {
			r.Response.WriteJsonExit(g.Map{"success": false, "error": "Invalid request method"})
			return
		}
		service.GenerateCaptcha(r)
	})

	// 公钥
	s.BindHandler("/api/public-key", func(r *ghttp.Request) {
		if r.Method != "GET" {
			r.Response.WriteJsonExit(g.Map{"success": false, "error": "Invalid request method"})
			return
		}
		service.GetPublicKey(r)
	})

	// 发送验证码
	s.BindHandler("/api/send-code", func(r *ghttp.Request) {
		if r.Method != "POST" {
			r.Response.WriteJsonExit(g.Map{"success": false, "error": "invalid request method"})
			return
		}
		// 图形验证码ID
		verifyID := r.Get("verifyID").String()
		if verifyID == "" {
			r.Response.WriteJsonExit(g.Map{"success": false, "error": "verifyID is required"})
			return
		}
		// 图形验证码答案
		verifyCode := r.Get("verifyCode").String()
		if verifyCode == "" {
			r.Response.WriteJsonExit(g.Map{"success": false, "error": "verifyCode is required"})
			return
		}
		// 用户名称
		username := r.Get("username").String()
		if username == "" {
			r.Response.WriteJsonExit(g.Map{"success": false, "error": "username is required"})
			return
		}
		// 发送类型
		codetype := r.Get("type").String()
		if codetype == "" {
			r.Response.WriteJsonExit(g.Map{"success": false, "error": "type is required"})
			return
		}

		service.SendVerificationCode(r)
	})

	// 验证验证码
	s.BindHandler("/api/verification-code", func(r *ghttp.Request) {
		if r.Method != "POST" {
			r.Response.WriteJsonExit(g.Map{"success": false, "error": "invalid request method"})
			return
		}
		// 图形验证码答案
		verifyCode := r.Get("verifyCode").String()
		if verifyCode == "" {
			r.Response.WriteJsonExit(g.Map{"success": false, "error": "verifyCode is required"})
			return
		}
		// 用户名称
		username := r.Get("username").String()
		if username == "" {
			r.Response.WriteJsonExit(g.Map{"success": false, "error": "username is required"})
			return
		}
		// 验证类型
		codetype := r.Get("type").String()
		if codetype == "" {
			r.Response.WriteJsonExit(g.Map{"success": false, "error": "type is required"})
			return
		}
		service.VerificationCode(r)
	})

	s.BindHandler("/api/reset-password", func(r *ghttp.Request) {
		if r.Method != "POST" {
			r.Response.WriteJsonExit(g.Map{"success": false, "error": "Invalid request method"})
			return
		}
		// 验证类型
		identifier := r.Get("type").String()
		if identifier == "" {
			r.Response.WriteJsonExit(g.Map{"success": false, "error": "Identifier is required"})
			return
		}
		// 收到的验证码
		code := r.Get("verifyCode").String()
		if code == "" {
			r.Response.WriteJsonExit(g.Map{"success": false, "error": "Code is required"})
			return
		}
		// 用户名称
		username := r.Get("username").String()
		if username == "" {
			r.Response.WriteJsonExit(g.Map{"success": false, "error": "Username is required"})
			return
		}
		// 新的密码
		newPassword := r.Get("newPassword").String()
		if newPassword == "" {
			r.Response.WriteJsonExit(g.Map{"success": false, "error": "New password is required"})
			return
		}
		service.ResetPassword(r)
	})

	s.Run()
}
