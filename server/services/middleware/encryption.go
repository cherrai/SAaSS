package middleware

import (
	conf "github.com/cherrai/SAaSS/config"
	"github.com/cherrai/SAaSS/services/encryption"
	"github.com/cherrai/SAaSS/services/response"
	"bytes"
	"encoding/hex"
	"encoding/json"

	"github.com/gin-gonic/gin"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func Encryption() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, isWsServer := c.Get("WsServer"); isWsServer {
			c.Next()
			return
		}
		roles := new(RoleOptionsType)
		getRoles, isRoles := c.Get("roles")
		if isRoles {
			roles = getRoles.(*RoleOptionsType)
		}
		if isRoles && roles.isHttpServer {
			var userAesKey string
			var res response.ResponseProtobufType
			res.Code = 10008

			// Reponse
			// defer func() {
			// 	fmt.Println("Encryption middleware", roles.ResponseEncryption)
			// 	// 暂时全部开放
			// 	if roles.ResponseEncryption == true {
			// 		res.Encryption(c)
			// 		// fmt.Println("Response解析成功！！！！！！！！！！")
			// 	}
			// }()
			// fmt.Println("Encryption middleware.")
			// Request

			if roles.RequestEncryption == true {
				var data string
				var key string
				var tempKey string
				var initKey string
				switch c.Request.Method {
				case "GET":
					data = c.Query("data")
					key = c.Query("key")
					tempKey = c.Query("tempKey")
					initKey = c.Query("initKey")
					break

				case "POST":
					data = c.PostForm("data")
					key = c.PostForm("key")
					tempKey = c.PostForm("tempKey")
					initKey = c.PostForm("initKey")
					// fmt.Println("aeskey enc", aeskey)
					break
				default:
					break
				}
				if data == "" {
					res.Code = 10002
					res.Call(c)
					c.Abort()
					return
				}
				var dataMap map[string]interface{}
				aes := encryption.AesEncrypt{
					Key:  "",
					Mode: "CFB",
				}
				if initKey != "" && tempKey == "" && key == "" {
					aes.Key = initKey
				}
				// 当没有临时AES秘钥时，前端传秘钥并公钥加密
				if tempKey != "" && key == "" {
					keyHex, keyHexErr := hex.DecodeString(tempKey)
					if keyHexErr != nil {
						res.Data = "[Encryption hex.DecodeString]" + keyHexErr.Error()
						res.Code = 10008
						res.Call(c)
						c.Abort()
						return
					}
					// RSA私钥解密
					deKey := conf.EncryptionClient.RsaKey.Decrypt(keyHex, nil)

					aes.Key = string(deKey)
				}
				// 当利用Public AESKey生成UID的Key存在的时候
				if key != "" && tempKey == "" {
					getAesKey, aesKeyErr := conf.EncryptionClient.GetUserAesKeyWithAesKey(key)
					if aesKeyErr != nil {
						res.Data = "[Encryption]" + aesKeyErr.Error()
						res.Code = 10008
						res.Call(c)
						c.Abort()
						return
					}
					userAesKey = getAesKey
					c.Set("userAesKey", userAesKey)
					aes.Key = getAesKey
				}

				if aes.Key == "" {
					res.Data = "[Encryption] AesKey does not exist."
					res.Code = 10008
					res.Call(c)
					c.Abort()
					return
				}

				deStr, deStrErr := aes.DecryptWithString(data)
				if deStrErr != nil {
					res.Data = "[Encryption aes.DecryptWithString]" + deStrErr.Error()
					res.Code = 10008
					res.Call(c)
					c.Abort()
					return
				}
				err := json.Unmarshal([]byte(deStr), &dataMap)
				if err != nil {
					res.Data = "[Encryption json.Unmarshal]" + err.Error()
					res.Code = 10008
					res.Call(c)
					c.Abort()
					return
				}

				// fmt.Println("dataMap", dataMap)
				// 为Gin请求体赋值
				for key, item := range dataMap {
					c.Set(key, item)
				}
				// fmt.Println("Request解析成功！！！！！！！！！！")
			}
			c.Next()
		} else {
			c.Next()
		}
	}
}
