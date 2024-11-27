# 接口文档

## 状态码定义

| 状态码 | 解释                 |
| ------ | -------------------- |
| 200    | 执行成功             |
| 10001  | 生成图形验证码失败   |
| 10002  | 解密密码失败         |
| 10003  | 二次校验密码不通过   |
| 10004  | LDAP密码重置失败     |
| 10005  | 无效的验证方式       |
| 10006  | 无效的验证码         |
| 10007  | LDAP初始化失败       |
| 10008  | 获取用户信息失败     |
| 10009  | 验证码发送间隔太短   |
| 10010  | 邮箱验证码发送失败   |
| 10011  | 手机号验证码发送失败 |
| 10012  | 未查找到用户         |

##  /api/get-user-info 

用途：查找用户的手机和邮箱，用于设定重置方式

请求方法：POST

请求参数：

| 参数名称 | 类型   | 说明                     |
| -------- | ------ | ------------------------ |
| username | String | 域用户名称或者手机或邮箱 |

返回示例：

~~~json
{
	"code": 200,
	"mail": "chu***********@oe*******.com",
	"mobile": "152****1"
}
~~~

说明：

| 字段   | 说明     |
| ------ | -------- |
| code   | 状态码   |
| mail   | 用户邮箱 |
| mobile | 用户手机 |

> 验证类型可以是mail(邮箱)或mobile(手机)

## /api/generate-captcha

用途：创建验证码

请求方法：GET

请求参数：无

返回示例：

~~~json
{
	"code": 200,
	"mail": "chu***********@oe*******.com",
	"mobile": "152****1"
}
~~~

说明：

| 字段 | 说明           |
| ---- | -------------- |
| code | 状态码         |
| id   | 此验证码ID     |
| url  | 验证码图片地址 |

## /api/send-code

用途：发送短信或邮箱验证码

请求方法：POST

请求参数：

| 字段       | 说明                     |
| ---------- | ------------------------ |
| username   | 域用户名称或者手机或邮箱 |
| type       | 验证类型                 |
| verifyID   | 验证码ID                 |
| verifyCode | 验证码答案               |

返回示例：

~~~json
{
	"code": 200,
	"message": "Success"
}
~~~

说明：

| 字段    | 说明   |
| ------- | ------ |
| code    | 状态码 |
| message | 消息   |

## /api/verification-code 

用途：验证短信或邮箱验证码

请求方法：POST

请求参数：

| 字段       | 说明                     |
| ---------- | ------------------------ |
| verifyCode | 接收到的短信或邮箱验证码 |
| username   | 域用户名称或者手机或邮箱 |
| type       | 验证类型                 |

返回示例：

~~~json
{
	"code": 200,
	"message": "Success"
}
~~~

说明：

| 字段    | 说明   |
| ------- | ------ |
| code    | 状态码 |
| message | 消息   |

## /api/public-key  

用途：获取公钥，用于加密参数

请求方法：GET

请求参数：无

返回示例：

~~~json
{
	"code": 200,
	"publicKey": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAx5U/bLOd5lTYBqH9DrmI\nAdk+x0xrVpRDIJdPqnyRt7ceCEQlCzCgmLSCkZrZ2p3jF/6QtDOFMcYmmVqdHDmy\nvWYDoLtOi7QsbrU+bzF2bbIZxzn3X/ntcQiHyHHEoJVyGmBVdWuE3TIzxLQjXm5R\naXzQ4ROc4ABmc4GslMgjjWTCG1RKKzO9M7wu3uybyZjJm+CGUuzGa+6isNw4GAGh\nmIUJps0HFah22FJPwCP+1NBk88YB4Ck3nN9jQbRBDX4Cwz3qxmcaWgE8h04lSAtQ\n+T5qktpPyiLjSLHb2zG+O6bUpCMT5yqKYjkOrcMhoujKZK4DM7EUN+jGXvqk9mAX\nbQIDAQAB\n-----END PUBLIC KEY-----\n"
}
~~~

## /api/reset-password 

用途：重置密码

请求方法：POST

请求参数：

| 字段        | 说明                     |
| ----------- | ------------------------ |
| username    | 域用户名称或者手机或邮箱 |
| newPassword | 新密码(需要公钥加密)     |
| type        | 验证类型                 |
| verifyCode  | 短信或者邮箱验证码       |

返回示例：

~~~json
{
	"code": 200,
	"message": "Success"
}
~~~

说明：

| 字段    | 说明   |
| ------- | ------ |
| code    | 状态码 |
| message | 消息   |

密码加密示例：

~~~js
async function encryptPassword(password: string): Promise<string> {
  // 获取后端公钥
  const response = await fetch("/api/public-key");
  const data = await response.json();
  const publicKey = data.publicKey;

  // 使用公钥加密密码
  const encryptor = new JSEncrypt();
  encryptor.setPublicKey(publicKey);
  const encrypted = encryptor.encrypt(password);
  if (!encrypted) {
    throw new Error("Failed to encrypt password");
  }
  return encrypted;
}
~~~

