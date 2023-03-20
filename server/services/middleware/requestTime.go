package middleware

import (
	"github.com/gin-gonic/gin"
)

func RequestTime() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, isStaticServer := c.Get("isStaticServer"); isStaticServer {
			c.Next()
			return
		}
		lt := log.Time()
		c.Next()
		lt.TimeEnd(c.Request.URL.Path + ", Request Time =>")
	}
}
