package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/pixelvide/kube-sentinel/pkg/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTestDB() {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}
	model.DB = db
	model.DB.Exec("PRAGMA foreign_keys = ON")
	err = model.DB.AutoMigrate(&model.User{}, &model.GitlabHosts{}, &model.UserGitlabConfig{})
	if err != nil {
		panic("failed to migrate database: " + err.Error())
	}
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		// Mock auth middleware
		user := model.User{Username: "testuser"}
		err := model.DB.FirstOrCreate(&user, model.User{Username: "testuser"}).Error
		if err != nil {
			panic("failed to create test user: " + err.Error())
		}
		c.Set("user", user)
		c.Next()
	})

	api := r.Group("/api/v1")
	userGroup := api.Group("/settings/gitlab-configs")
	{
		userGroup.GET("/", ListUserGitlabConfigs)
		userGroup.POST("/", UpsertUserGitlabConfig)
		userGroup.POST("/:id/validate", ValidateUserGitlabConfig)
		userGroup.DELETE("/:id", DeleteUserGitlabConfig)
	}
	return r
}

func TestUserGitlabConfigHandlers(t *testing.T) {
	setupTestDB()
	r := setupRouter()

	// Seed Host
	host := model.GitlabHosts{Host: "gitlab.com"}
	err := model.DB.Create(&host).Error
	assert.NoError(t, err)

	// Create (Upsert - New)
	t.Run("Create (Upsert)", func(t *testing.T) {
		reqBody := UpsertUserGitlabConfigReq{
			GitlabHostID: host.ID,
			Token:        "test-token",
		}
		body, err := json.Marshal(reqBody)
		assert.NoError(t, err)
		req, err := http.NewRequest("POST", "/api/v1/settings/gitlab-configs/", bytes.NewBuffer(body))
		assert.NoError(t, err)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp model.UserGitlabConfig
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "test-token", string(resp.Token))
		assert.Equal(t, host.ID, resp.GitlabHostID)
	})

	// List
	t.Run("List", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/settings/gitlab-configs/", nil)
		assert.NoError(t, err)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp []model.UserGitlabConfig
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
	})

	// Get ID of created config
	var config model.UserGitlabConfig
	err = model.DB.First(&config).Error
	assert.NoError(t, err)

	// Update (Upsert - Existing)
	t.Run("Update (Upsert)", func(t *testing.T) {
		reqBody := UpsertUserGitlabConfigReq{
			GitlabHostID: host.ID,
			Token:        "updated-token",
		}
		body, err := json.Marshal(reqBody)
		assert.NoError(t, err)
		req, err := http.NewRequest("POST", "/api/v1/settings/gitlab-configs/", bytes.NewBuffer(body))
		assert.NoError(t, err)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp model.UserGitlabConfig
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "updated-token", string(resp.Token))
		assert.Equal(t, config.ID, resp.ID) // should be same ID
	})

	// Validate (Mocking GitLab) - Disabled as it requires glab CLI
	/*
		t.Run("Validate", func(t *testing.T) {
			// Start a mock GitLab server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v4/user", r.URL.Path)
				assert.Equal(t, "updated-token", r.Header.Get("PRIVATE-TOKEN"))
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			// Update Host to point to mock server
			// The Host field usually contains "gitlab.com". We need to change it to the mock server's host.
			// server.URL looks like "http://127.0.0.1:12345"
			// We need to parse it to set Host and IsHTTPS correctly if strictly needed,
			// but our handler handles "http://" prefix.

			// Let's create a specific host config for this test or update the existing one's related host
			var hostToUpdate model.GitlabHosts
			model.DB.First(&hostToUpdate, host.ID)
			hostToUpdate.Host = server.URL // set full URL
			model.DB.Save(&hostToUpdate)

			req, _ := http.NewRequest("POST", "/api/v1/settings/gitlab-configs/"+strconv.Itoa(int(config.ID))+"/validate", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			// Check IsValidated
			var updatedConfig model.UserGitlabConfig
			model.DB.First(&updatedConfig, config.ID)
			assert.True(t, updatedConfig.IsValidated)
		})
	*/

	// Delete
	t.Run("Delete", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", "/api/v1/settings/gitlab-configs/"+strconv.Itoa(int(config.ID)), nil)
		assert.NoError(t, err)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify deletion
		var count int64
		model.DB.Model(&model.UserGitlabConfig{}).Where("id = ?", config.ID).Count(&count)
		assert.Equal(t, int64(0), count)
	})
}
