package middleware

import (
	conf "github.com/cherrai/SAaSS/config"
	"github.com/cherrai/SAaSS/services/response"
	"github.com/cherrai/SAaSS/services/typings"

	"github.com/gin-gonic/gin"
)

func CheckApp() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, isStaticServer := c.Get("isStaticServer"); isStaticServer {
			c.Next()
			return
		}
		res := response.ResponseType{}
		res.Code = 10004

		roles := new(RoleOptionsType)
		getRoles, isRoles := c.Get("roles")
		if isRoles {
			roles = getRoles.(*RoleOptionsType)
		}

		log.Info(roles.CheckApp, roles.CheckAppToken, roles.CheckApp || roles.CheckAppToken)
		if roles.CheckApp || roles.CheckAppToken {
			// 两者具备其一即可
			if roles.CheckApp {
				// 解析用户数据
				var appId string
				var appKey string

				switch c.Request.Method {
				case "GET":
					appId = c.Query("appId")
					appKey = c.Query("appKey")

				case "POST":
					appId = c.PostForm("appId")
					appKey = c.PostForm("appKey")
				default:
					break
				}
				if appKey != "" && appId != "" && conf.AppList[appId].AppKey == appKey {
					c.Set("appId", appId)
					c.Set("appKey", appKey)
					c.Next()
					return
				}
			}

			log.Info(roles.CheckAppToken)
			if roles.CheckAppToken {
				// 解析用户数据
				var appId string
				var appKey string
				var appToken string

				switch c.Request.Method {
				case "GET":
					appToken = c.Query("appToken")

				case "POST":
					appToken = c.PostForm("appToken")
				default:
					break
				}
				log.Info(conf.AppList[appId].AppKey, appKey)
				log.Info("appToken", appToken)
				if appToken != "" {
					rKey := conf.Redisdb.GetKey("AppToken")

					ati := new(typings.AppTokenInfo)
					if err := conf.Redisdb.GetStruct(rKey.GetKey(appToken), ati); err == nil {
						c.Set("appId", ati.AppId)
						c.Set("appKey", ati.AppKey)
						c.Set("appTokenInfo", ati)
						c.Next()
						return
					}
				}
			}

			res.Code = 10014
			res.Call(c)
			c.Abort()
			return
		}
		c.Next()
	}
}
