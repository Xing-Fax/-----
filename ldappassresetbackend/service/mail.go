package service

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"html/template"
	"regexp"
	"strconv"

	"gopkg.in/gomail.v2"

	"github.com/gogf/gf/v2/frame/g"
)

type EmailService struct {
	address       string
	port          string
	disableTLS    bool
	username      string
	password      string
	sender        string
	subject       string
	emailTemplate string
}

func NewEmailService() *EmailService {
	cfg := g.Cfg().MustGet(context.TODO(), "smtp").Map()

	address, _ := cfg["address"].(string)
	port, _ := cfg["port"].(string)
	disableTLS, _ := cfg["disableTLS"].(bool)
	username, _ := cfg["username"].(string)
	password, _ := cfg["password"].(string)
	sender, _ := cfg["sender"].(string)
	subject, _ := cfg["subject"].(string)
	emailTemplate, _ := cfg["emailTemplate"].(string)

	return &EmailService{
		address:       address,
		port:          port,
		disableTLS:    disableTLS,
		username:      username,
		password:      password,
		sender:        sender,
		subject:       subject,
		emailTemplate: emailTemplate,
	}
}

func (s *EmailService) SendEmail(user, email, code string) error {
	// 发件人UTF-8编码
	parsedSender, err := ParseSender(s.sender)
	if err != nil {
		return fmt.Errorf("failed to parse sender: %v", err)
	}
	m := gomail.NewMessage()

	m.SetHeader("From", parsedSender)
	m.SetHeader("To", email)
	m.SetHeader("Subject", s.subject)
	// 创建模板，进行内容替换
	tmpl, err := template.New("email").Parse(s.emailTemplate)
	if err != nil {
		return fmt.Errorf("template parsing failed: %v", err)
	}
	var emailContent bytes.Buffer
	err = tmpl.Execute(&emailContent, struct {
		User string
		Code string
	}{
		User: user,
		Code: code,
	})
	if err != nil {
		return fmt.Errorf("template replacement failed: %v", err)
	}

	m.SetBody("text/html", emailContent.String())

	port, err := strconv.Atoi(s.port)
	if err != nil {
		fmt.Println("Error converting port:", err)
		return nil
	}
	d := gomail.NewDialer(s.address, port, s.username, s.password)

	d.TLSConfig = &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         s.address,
	}

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

func ParseSender(sender string) (string, error) {
	re := regexp.MustCompile(`^(.+?) <(.+)>$`)
	matches := re.FindStringSubmatch(sender)

	if len(matches) != 3 {
		return "", fmt.Errorf("invalid sender format")
	}

	name := matches[1]
	email := matches[2]

	encodedName := "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(name)) + "?="

	parsedSender := fmt.Sprintf("%s <%s>", encodedName, email)

	return parsedSender, nil
}
