package middleware

import (
	"encoding/json"

	conf "github.com/cherrai/SAaSS/config"
	"github.com/cherrai/SAaSS/services/methods"
	"github.com/cherrai/SAaSS/services/response"

	sso "github.com/cherrai/saki-sso-go"

	"github.com/gin-gonic/gin"
)

func Authorize() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log.Info("------Authorize------")
		if _, isStaticServer := c.Get("isStaticServer"); isStaticServer {
			c.Next()
			return
		}
		if _, isWsServer := c.Get("WsServer"); isWsServer {
			c.Next()
			return
		}
		res := response.ResponseType{}
		res.Code = 10015

		roles := new(RoleOptionsType)
		getRoles, isRoles := c.Get("roles")
		if isRoles {
			roles = getRoles.(*RoleOptionsType)
		}

		if roles.Authorize {
			// 解析用户数据
			var token string
			token = c.Query("token")
			configInfo, err := methods.ParseToken(token)
			if err != nil {
				res.Code = 10015
				res.Call(c)
				c.Abort()
				return
			}
			getToken, err := conf.Redisdb.Get("file_" + configInfo.FileInfo.Hash)
			if err != nil {
				res.Code = 10015
				res.Call(c)
				c.Abort()
			}
			if getToken.String() != token {
				res.Code = 10015
				res.Call(c)
				c.Abort()
			}
			c.Set("token", token)
			c.Set("fileConfigInfo", configInfo)
			return
		}

		c.Next()
	}
}

func ConvertResponseJson(jsonStr []byte) (sso.UserInfo, error) {
	var m sso.UserInfo
	err := json.Unmarshal([]byte(jsonStr), &m)
	if err != nil {
		Log.Info("Unmarshal with error: %+v\n", err)
		return m, err
	}
	return m, nil
}
