package middleware

import (
	"net/http"
	"strings"

	"github.com/cherrai/SAaSS/services/response"

	"github.com/gin-gonic/gin"
)

func CheckRouteMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isStaticServer := strings.Contains(c.Request.URL.Path, "/s/")
		if isStaticServer {
			c.Set("isStaticServer", true)
			c.Next()
			return
		}
		isWSServer := strings.Contains(c.Request.URL.Path, "/socket.io")
		if isWSServer {
			c.Set("WsServer", true)
			c.Next()
			return
		}
		isHttpServer := strings.Contains(c.Request.URL.Path, "/api")
		if isHttpServer {
			c.Next()
			return
		}

		res := response.ResponseType{}
		res.Code = 10013
		c.JSON(http.StatusOK, res.GetResponse())
		c.Abort()
	}
}
