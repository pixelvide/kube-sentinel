package middleware

import (
	"fmt"
	"strings"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
)

var unlogPath = []string{
	"/healthz",
	"/assets/",
	"/favicon.ico",
	"/metrics",
}

func Logger() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			// Skip logging for health check and static asset paths
			for _, path := range unlogPath {
				if param.Path == path || strings.HasPrefix(param.Path, path) {
					return ""
				}
			}

			userEmail := "-"
			if v, ok := param.Keys["user"]; ok {
				if user, ok := v.(*models.User); ok {
					userEmail = user.Email
				}
			}

			// Get Kubernetes context from header if present
			kubeContext := param.Request.Header.Get("x-kube-context")
			if kubeContext == "" {
				kubeContext = "-"
			}

			return fmt.Sprintf("%s - %s \"%s %s\" %d %s %s %s\n",
				param.ClientIP,
				param.TimeStamp.Format("2006-01-02 15:04:05"),
				param.Method,
				param.Path,
				param.StatusCode,
				param.Latency,
				kubeContext,
				userEmail,
			)
		},
	})
}
