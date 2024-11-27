package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"log"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

var (
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	once       sync.Once
)

func initKeys() {
	once.Do(func() {
		var err error
		privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			log.Fatalf("Failed to generate RSA keys: %v", err)
		}
		publicKey = &privateKey.PublicKey
	})
}

// GetPublicKey 返回 base64 编码的公钥
func GetPublicKeyBase() string {
	initKeys()
	pubASN1, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		log.Fatalf("Failed to marshal public key: %v", err)
	}
	pubPem := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	})
	return string(pubPem)

}

// DecryptPassword 解密前端传来的加密密码
func DecryptPassword(cipherText string) (string, error) {
	initKeys()
	decodedCipherText, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}
	plainText, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, decodedCipherText)
	if err != nil {
		return "", err
	}
	return string(plainText), nil
}

func GetPublicKey(r *ghttp.Request) {
	publicKey := GetPublicKeyBase()
	r.Response.WriteJson(g.Map{
		"code":      200,
		"publicKey": publicKey,
	})
}
