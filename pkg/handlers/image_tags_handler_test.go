package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetImageTags_SSRF_Vulnerability(t *testing.T) {
	// Setup Gin
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/image/tags", GetImageTags)

	// Test case 1: Attempt to access 127.0.0.1 (SSRF)
	// This should be blocked by IsSecureRegistry
	req, _ := http.NewRequest("GET", "/image/tags?image=127.0.0.1/myimage", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "access to loopback/unspecified addresses is restricted")

	// Test case 2: Attempt to access valid registry (e.g. quay.io)
	// This should NOT be blocked by IsSecureRegistry.
	// However, the connection will likely fail due to network isolation or just fail to connect,
	// causing GetTags to return nil, nil, which results in 200 OK with null body (as per current implementation)
	req2, _ := http.NewRequest("GET", "/image/tags?image=quay.io/coreos/etcd", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	// We don't check the body because it depends on network connectivity, but status code 200 means it passed the validation.
}
