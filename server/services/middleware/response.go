package middleware

import (
	"net/http"

	"github.com/cherrai/SAaSS/protos"
	"github.com/cherrai/SAaSS/services/response"

	"github.com/gin-gonic/gin"
)

func Response() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, isStaticServer := c.Get("isStaticServer"); isStaticServer {
			c.Next()
			return
		}
		if _, isWsServer := c.Get("WsServer"); isWsServer {
			c.Next()
			return
		}
		roles := new(RoleOptionsType)
		getRoles, isRoles := c.Get("roles")

		if isRoles {
			roles = getRoles.(*RoleOptionsType)
		} else {
			// res := response.ResponseType{}
			// res.Code = 10013
			// c.JSON(http.StatusOK, res.GetResponse())
			// return
		}
		if isRoles && roles.isHttpServer {
			defer func() {
				roles := c.MustGet("roles").(*RoleOptionsType)
				userAesKey := c.GetString("userAesKey")
				// Log.Info("Response middleware", roles.ResponseEncryption)
				if roles.isHttpServer {
					switch roles.ResponseDataType {
					case "protobuf":
						var res response.ResponseProtobufType
						getProtobufDataResponse, _ := c.Get("protobuf")

						if getProtobufDataResponse == nil {
							getBodyDataResponse, _ := c.Get("body")
							if roles.ResponseEncryption == true {
								if getBodyDataResponse == nil {
									res.Code = 10001
									c.JSON(http.StatusOK, res.Encryption(userAesKey, res))
								} else {
									// 当需要加密的时候
									c.JSON(http.StatusOK, res.Encryption(userAesKey, getProtobufDataResponse))
								}
							} else {
								if getBodyDataResponse == nil {
									res.Code = 10001
									c.JSON(http.StatusOK, res)
								} else {
									c.JSON(http.StatusOK, getBodyDataResponse)
								}
							}
						} else {
							// fmt.Println("输出protobuf Res")
							if roles.ResponseEncryption == true {
								c.Writer.Header().Set("Content-Type", "application/x-protobuf")
								c.String(http.StatusOK, res.Encryption(userAesKey, getProtobufDataResponse))
								// fmt.Println("Response解析成功！！！！！！！！！！")
							} else {
								c.Writer.Header().Set("Content-Type", "application/x-protobuf")
								c.ProtoBuf(http.StatusOK, getProtobufDataResponse.(*protos.ResponseEncryptDataType))
							}
						}
						break

					default:
						if roles.ResponseEncryption == true {
							// 当需要加密的时候
						} else {
							getResponse, _ := c.Get("body")
							c.JSON(http.StatusOK, getResponse)
						}
						break
					}
				}
			}()
			c.Next()
		} else {
			c.Next()
		}
	}
}

// fmt.Println(test.GetName())

// data, err := proto.Marshal(test)
// var msgData interface{}
// msgData = test
// msg, _ := proto.Marshal(msgData.(proto.Message))
// fmt.Println("msg", msg, "::::", string(msg))
// fmt.Println("data", data)
// // fmt.Println("data", string(data))
// if err != nil {
// 	log.Fatal("marshaling error: ", err)
// }
// newTest := &Student{}
// err = proto.Unmarshal(data, newTest)
// fmt.Println("newTest", newTest)
// if err != nil {
// 	log.Fatal("unmarshaling error: ", err)
// }
// // Now test and newTest contain the same data.
// if test.GetName() != newTest.GetName() {
// 	log.Fatalf("data mismatch %q != %q", test.GetName(), newTest.GetName())
// }
