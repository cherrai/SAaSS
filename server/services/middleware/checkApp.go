package middleware

import (
	conf "github.com/cherrai/SAaSS/config"
	"github.com/cherrai/SAaSS/services/response"

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

		if roles.CheckApp {
			// 解析用户数据
			var appId string
			var appKey string

			switch c.Request.Method {
			case "GET":
				appId = c.Query("appId")
				appKey = c.Query("appKey")
				break

			case "POST":
				appId = c.PostForm("appId")
				appKey = c.PostForm("appKey")
				break
			default:
				break
			}
			// log.Info(conf.AppList[appId].AppKey, appKey)
			if conf.AppList[appId].AppKey == appKey {
				c.Next()
				return
			}
			res.Code = 10014
			res.Call(c)
			c.Abort()
			return
		}
		c.Next()
	}
}
