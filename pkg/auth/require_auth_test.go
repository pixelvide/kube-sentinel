package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pixelvide/kube-sentinel/pkg/common"
	"github.com/pixelvide/kube-sentinel/pkg/model"
	"github.com/pixelvide/kube-sentinel/pkg/rbac"
	"github.com/stretchr/testify/assert"
)

func TestRequireAuthRestoresOIDCGroups(t *testing.T) {
	// Initialize RBACConfig to avoid panic
	rbac.RBACConfig = &common.RolesConfig{
		Roles:       []common.Role{},
		RoleMapping: []common.RoleMapping{},
	}

	// Setup DB
	common.DBType = "sqlite"
	common.DBDSN = "file::memory:?cache=shared"
	common.JwtSecret = "testsecret"

	// Reset DB (if already initialized by other tests, though unlikely in this package)
	if model.DB != nil {
		// Clean tables
		model.DB.Exec("DELETE FROM users")
		model.DB.Exec("DELETE FROM user_configs")
	} else {
		model.InitDB()
	}

	// Create a user
	user := &model.User{
		Username: "testuser",
		Provider: "oidc",
		Enabled:  true,
		// Password is required for some checks but here we bypass password login
	}
	// We need to handle created_at/updated_at manually if not using GORM Create? No, Create handles it.
	err := model.DB.Create(user).Error
	assert.NoError(t, err)

	// User config is needed
	model.DB.Create(&model.UserConfig{UserID: user.ID})

	// Add OIDC groups to the user object (transient)
	groups := []string{"dev-team", "admins"}
	user.OIDCGroups = groups

	// Generate JWT
	om := NewOAuthManager()
	token, err := om.GenerateJWT(user, "")
	assert.NoError(t, err)

	// Setup Gin
	gin.SetMode(gin.TestMode)
	r := gin.New()
	authHandler := NewAuthHandler(nil) // ClusterManager is nil, should be fine for RequireAuth

	// Protected route
	r.GET("/protected", authHandler.RequireAuth(), func(c *gin.Context) {
		u, exists := c.Get("user")
		assert.True(t, exists)
		userObj := u.(model.User)

		// This is the crucial check
		assert.Equal(t, len(groups), len(userObj.OIDCGroups))
		// We can't rely on order, so check containment
		hasDev := false
		hasAdmin := false
		for _, g := range userObj.OIDCGroups {
			if g == "dev-team" {
				hasDev = true
			}
			if g == "admins" {
				hasAdmin = true
			}
		}
		assert.True(t, hasDev, "dev-team group missing")
		assert.True(t, hasAdmin, "admins group missing")

		c.Status(http.StatusOK)
	})

	// Make request
	req := httptest.NewRequest("GET", "/protected", nil)
	// Set cookie
	cookie := &http.Cookie{
		Name:  "auth_token",
		Value: token,
	}
	req.AddCookie(cookie)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
