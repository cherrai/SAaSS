package middleware

import (
	"fmt"
	"reflect"
	"runtime"

	"github.com/cherrai/SAaSS/services/response"

	"github.com/cherrai/nyanyago-utils/nlog"
	"github.com/gin-gonic/gin"
)

var (
	Log = nlog.New()
)

func Error() gin.HandlerFunc {
	return func(c *gin.Context) {
		// roles := c.MustGet("roles").(*RoleOptionsType)
		defer func() {
			// fmt.Println("Error middleware.2222222222222")
			// fmt.Println("roles1", roles)
			// fmt.Println("Error mid getRoles", roles.ResponseDataType)
			if err := recover(); err != nil {
				Log.Error("<"+c.Request.URL.Path+">", "Gin Error: ", err.(error))
				for i := 2; i < 10; i++ {
					_, fn, line, _ := runtime.Caller(i)
					fmt.Println("file:", fn, "line:", line)
				}
				var res response.ResponseProtobufType
				res.Code = 10001
				switch reflect.TypeOf(err).String() {
				case "string":
					res.Error = err.(string)
					break
				case "*errors.errorString":
					res.Error = err.(error).Error()
					break
				case "runtime.errorString":
					res.Error = err.(error).Error()
					break

				}
				res.Call(c)
				c.Abort()
				// fmt.Println("=========Error=========")
				// // fmt.Println("roles2", roles)
				// fmt.Println("Error middleware:", err)
				// switch roles.ResponseDataType {
				// case "protobuf":
				// 	fmt.Println("输出protobuf Err")
				// 	userAesKey, _ := c.Get("userAesKey")
				// 	fmt.Println(userAesKey)
				// 	var res response.ResponseProtobufType
				// 	res.Code = 10001
				// 	res.Call(c)
				// 	break

				// default:
				// 	res := response.ResponseType{
				// 		Code: 10001,
				// 	}
				// 	switch reflect.TypeOf(err).String() {
				// 	case "string":
				// 		res.Data = err.(string)
				// 		break
				// 	case "*errors.errorString":
				// 		res.Data = err.(error).Error()
				// 		break

				// 	}
				// 	res.Call(c)
				// 	break
				// }

				// fmt.Println("=========Error=========")
			}
		}()
		c.Next()
	}
}
