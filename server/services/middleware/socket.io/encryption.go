package socketioMiddleware

import (
	conf "github.com/cherrai/SAaSS/config"
	"github.com/cherrai/SAaSS/protos"
	"github.com/cherrai/SAaSS/services/encryption"
	socketiomid "github.com/cherrai/SAaSS/services/nyanyago-utils/socketio-mid"
	"github.com/cherrai/SAaSS/services/response"
	"encoding/base64"
	"encoding/json"
)

// 解密Request
func Decryption() socketiomid.HandlerFunc {
	return func(c *socketiomid.ConnContext) (err error) {
		var res response.ResponseProtobufType
		res.Code = 10008
		enData := c.GetParamsString("data")
		// requestId := c.GetParamsString("requestId")
		// fmt.Println("en requestId", requestId)
		// c.Set("requestId", requestId)
		var userAesKey string
		aes := encryption.AesEncrypt{
			Key:  "",
			Mode: "CFB",
		}
		data := new(protos.RequestEncryptDataType)
		// fmt.Println("DecryptionSoc中间件", enData)
		dataBase64, dataBase64Err := base64.StdEncoding.DecodeString(enData)
		if dataBase64Err != nil {
			res.Data = "[Encryption]" + dataBase64Err.Error()
			res.Code = 10008
			res.CallSocketIo(c)
			return
		}
		deErr := protos.Decode(dataBase64, data)
		if deErr != nil {
			res.Data = "[Encryption]" + deErr.Error()
			res.Code = 10008
			res.CallSocketIo(c)
			return
		}

		getAesKey, aesKeyErr := conf.EncryptionClient.GetUserAesKeyWithAesKey(data.Key)
		if aesKeyErr != nil {
			res.Data = "[Encryption]" + aesKeyErr.Error()
			res.Code = 10008
			res.CallSocketIo(c)
			return
		}
		userAesKey = getAesKey
		aes.Key = getAesKey
		c.Set("userAesKey", userAesKey)

		deStr, deStrErr := aes.DecryptWithString(data.Data)
		if deStrErr != nil {
			res.Data = "[Encryption]" + deStrErr.Error()
			res.Code = 10008
			res.CallSocketIo(c)
			return
		}
		var dataMap map[string]interface{}
		unErr := json.Unmarshal([]byte(deStr), &dataMap)
		if unErr != nil {
			res.Data = "[Encryption]" + unErr.Error()
			res.Code = 10008
			res.CallSocketIo(c)
			return
		}
		for key, item := range dataMap {
			// fmt.Println(key, item)
			c.Set(key, item)
		}

		c.Next()
		return
	}
}
