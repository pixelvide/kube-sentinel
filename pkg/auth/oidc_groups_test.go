package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pixelvide/kube-sentinel/pkg/common"
	"github.com/pixelvide/kube-sentinel/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestOIDCGroupsInJWT(t *testing.T) {
	common.JwtSecret = "testsecret"
	om := NewOAuthManager()

	groups := []string{"group1", "group2"}
	user := &model.User{
		Model:      model.Model{ID: 1},
		Username:   "testuser",
		Provider:   "testprovider",
		OIDCGroups: groups,
	}

	tokenString, err := om.GenerateJWT(user, "refresh_token")
	assert.NoError(t, err)

	// Parse the token without validating signature
	parser := jwt.NewParser()
	var claims jwt.MapClaims
	_, _, err = parser.ParseUnverified(tokenString, &claims)
	assert.NoError(t, err)

	// Check if oidc_groups are present
	oidcGroups, ok := claims["oidc_groups"]
	assert.True(t, ok, "oidc_groups claim missing")

	if ok {
		groupsInterface, ok := oidcGroups.([]interface{})
		assert.True(t, ok, "oidc_groups is not a list")
		assert.Equal(t, len(groups), len(groupsInterface))
		// Check content (conversion might be needed depending on json decoding)
		// jwt.MapClaims usually decodes arrays as []interface{}
		hasGroup1 := false
		for _, g := range groupsInterface {
			if g == "group1" {
				hasGroup1 = true
			}
		}
		assert.True(t, hasGroup1, "group1 missing")
	}
}

func TestRefreshJWTPreservesGroups(t *testing.T) {
	common.JwtSecret = "testsecret"
	om := NewOAuthManager()

	// Create a token manually with groups
	// We use the same signing method and secret as OAuthManager uses (HS256)
	claims := jwt.MapClaims{
		"user_id":     1.0, // JWT numbers are float64 by default
		"username":    "testuser",
		"provider":    "testprovider",
		"oidc_groups": []interface{}{"group1", "group2"},
		// No refresh token to trigger the "no refresh token" path in RefreshJWT
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(common.JwtSecret))

	refreshedToken, err := om.RefreshJWT(nil, tokenString)
	assert.NoError(t, err)

	// Verify the new token has groups
	parser := jwt.NewParser()
	var newClaims jwt.MapClaims
	_, _, err = parser.ParseUnverified(refreshedToken, &newClaims)
	assert.NoError(t, err)

	oidcGroups, ok := newClaims["oidc_groups"]
	assert.True(t, ok, "oidc_groups claim missing in refreshed token")

	if ok {
		groupsInterface, ok := oidcGroups.([]interface{})
		assert.True(t, ok, "oidc_groups is not a list in refreshed token")
		assert.Equal(t, 2, len(groupsInterface))
	}
}
