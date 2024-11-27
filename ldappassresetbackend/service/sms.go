package service

import (
	"context"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/gogf/gf/v2/frame/g"
)

type SmsService struct {
	accessKeyID     string
	accessKeySecret string
	signName        string
	templateCode    string
}

func NewSmsService() *SmsService {
	cfg := g.Cfg().MustGet(context.TODO(), "sms").Map()

	accessKeyID, _ := cfg["accessKeyID"].(string)
	accessKeySecret, _ := cfg["accessKeySecret"].(string)
	signName, _ := cfg["signName"].(string)
	templateCode, _ := cfg["templateCode"].(string)

	return &SmsService{
		accessKeyID:     accessKeyID,
		accessKeySecret: accessKeySecret,
		signName:        signName,
		templateCode:    templateCode,
	}
}

var SignName string
var TemplateCode string

func SendSms(phone, code string) error {
	client, _err := NewSmsService().CreateClient()
	if _err != nil {
		return _err
	}

	sendSmsRequest := &dysmsapi20170525.SendSmsRequest{
		PhoneNumbers:  tea.String(phone),
		SignName:      tea.String(SignName),
		TemplateCode:  tea.String(TemplateCode),
		TemplateParam: tea.String("{\"code\":\"" + code + "\"}"),
	}
	runtime := &util.RuntimeOptions{}

	tryErr := func() (_e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		_, _err := client.SendSmsWithOptions(sendSmsRequest, runtime)
		// fmt.Println(response)
		if _err != nil {
			return _err
		}

		return nil
	}()

	if tryErr != nil {
		var error = &tea.SDKError{}
		if _t, ok := tryErr.(*tea.SDKError); ok {
			error = _t
		} else {
			error.Message = tea.String(tryErr.Error())
		}
		// 错误 message
		fmt.Println(tea.StringValue(error.Message))
	}
	return _err
}

func (s *SmsService) CreateClient() (_result *dysmsapi20170525.Client, _err error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(s.accessKeyID),
		AccessKeySecret: tea.String(s.accessKeySecret),
	}
	SignName = s.signName
	TemplateCode = s.templateCode

	config.Endpoint = tea.String("dysmsapi.aliyuncs.com")
	_result = &dysmsapi20170525.Client{}
	_result, _err = dysmsapi20170525.NewClient(config)
	return _result, _err
}
