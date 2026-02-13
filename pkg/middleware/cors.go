package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pixelvide/kube-sentinel/pkg/common"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowed := false

		// Check if origin is allowed
		if len(common.AllowedOrigins) > 0 {
			for _, o := range common.AllowedOrigins {
				if o == "*" {
					allowed = true
					c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
					// Credentials cannot be used with wildcard origin
					break
				}
				if o == origin {
					allowed = true
					c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
					c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
					break
				}
			}
		}

		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Cluster-Name")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
