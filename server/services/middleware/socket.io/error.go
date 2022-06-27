package socketioMiddleware

import (
	"reflect"
	"runtime"

	socketiomid "github.com/cherrai/SAaSS/services/nyanyago-utils/socketio-mid"
	"github.com/cherrai/SAaSS/services/response"

	"github.com/cherrai/nyanyago-utils/nlog"
)

var (
	Log = nlog.New()
)

func Error() socketiomid.HandlerFunc {
	return func(c *socketiomid.ConnContext) error {
		// roles := c.MustGet("roles").(*RoleOptionsType)
		defer func() {
			// fmt.Println("Error middleware.2222222222222")
			if err := recover(); err != nil {
				_, fn, line, _ := runtime.Caller(2)
				Log.Error("<"+c.EventName()+">", "Socket Error: ", err.(error), "file:", fn, "line:", line)
				var res response.ResponseProtobufType
				res.Code = 10001
				switch reflect.TypeOf(err).String() {
				case "string":
					res.Data = err.(string)
					break
				case "*errors.errorString":
					res.Data = err.(error).Error()
					break
				}
				res.CallSocketIo(c)
			}
		}()
		c.Next()
		return nil
	}
}
