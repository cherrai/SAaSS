package socketioMiddleware

import (
	socketiomid "github.com/cherrai/SAaSS/services/nyanyago-utils/socketio-mid"
	"github.com/cherrai/SAaSS/services/response"
)

func Response() socketiomid.HandlerFunc {
	return func(c *socketiomid.ConnContext) error {
		defer func() {
			var res response.ResponseProtobufType
			// fmt.Println("Response")
			getProtobufDataResponse, _ := c.Get("protobuf")
			// fmt.Println("getProtobufDataResponse", getProtobufDataResponse)
			userAesKey := c.GetString("userAesKey")
			// fmt.Println("userAesKey", userAesKey)
			requestId := c.GetParamsString("requestId")
			// fmt.Println("requestId", requestId)
			// fmt.Println(res.Encryption(userAesKey, getProtobufDataResponse))
			c.Body(map[string]interface{}{
				"data":      res.Encryption(userAesKey, getProtobufDataResponse),
				"requestId": requestId,
			})

		}()

		c.Next()
		return nil
	}
}
