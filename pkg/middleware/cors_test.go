package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pixelvide/kube-sentinel/pkg/common"
	"github.com/stretchr/testify/assert"
)

func TestCORS_Secure(t *testing.T) {
	// Setup Gin
	gin.SetMode(gin.TestMode)

	// Save original AllowedOrigins
	originalOrigins := common.AllowedOrigins
	defer func() { common.AllowedOrigins = originalOrigins }()

	t.Run("Default_NoAllowedOrigins", func(t *testing.T) {
		common.AllowedOrigins = []string{}
		r := gin.New()
		r.Use(CORS())
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://evil.com")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("AllowedOrigin", func(t *testing.T) {
		common.AllowedOrigins = []string{"http://good.com"}
		r := gin.New()
		r.Use(CORS())
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		// Good origin
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://good.com")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, "http://good.com", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))

		// Bad origin
		reqBad, _ := http.NewRequest("GET", "/test", nil)
		reqBad.Header.Set("Origin", "http://evil.com")
		wBad := httptest.NewRecorder()
		r.ServeHTTP(wBad, reqBad)

		assert.Empty(t, wBad.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("WildcardOrigin", func(t *testing.T) {
		common.AllowedOrigins = []string{"*"}
		r := gin.New()
		r.Use(CORS())
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://any.com")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
	})
}
