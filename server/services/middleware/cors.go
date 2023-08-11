package middleware

import (
	"net/http"

	"github.com/cherrai/nyanyago-utils/narrays"
	"github.com/gin-gonic/gin"
)

func Cors(allowOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// log.Info(allowOrigins, narrays.Includes(allowOrigins, "*"))
		if c.Request.Referer() != "" {

			if narrays.Includes(allowOrigins, "*") {
				c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				origin := ""
				// 这里预留，还没写代码
				origin = c.Request.Referer()
				origin = origin[0 : len(origin)-1]
				// log.Info("origin", origin)
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			}
			// Log.Info(c.Request.URL)
			// Log.Info("Cors 当前Referer: ", c.Request.Referer())
			// c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}
		c.Writer.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")
		c.Writer.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Content-Length, X-CSRF-Token, Token, session, Origin, Host, Connection, Accept-Encoding, Accept-Language, X-Requested-With")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Request.Header.Del("Origin")

		c.Next()
	}
}
