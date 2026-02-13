package model_test

import (
	"testing"

	"github.com/pixelvide/kube-sentinel/pkg/common"
	"github.com/pixelvide/kube-sentinel/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestOAuthProvider_AppIsolation(t *testing.T) {
	// Setup DB
	common.DBType = "sqlite"
	common.DBDSN = "file::memory:?cache=shared"
	model.InitDB()

	// Helper to create an app and set it as CurrentApp
	createApp := func(name string) *model.App {
		app := model.App{Name: name, Enabled: true}
		model.DB.Create(&app)
		return &app
	}

	app1 := createApp("App1")
	app2 := createApp("App2")

	// Set CurrentApp to App1
	model.CurrentApp = app1

	// Create Provider in App1
	provider1 := model.OAuthProvider{
		Name:         "google",
		ClientID:     "id1",
		ClientSecret: "secret1",
	}
	err := model.CreateOAuthProvider(&provider1)
	assert.NoError(t, err)
	assert.Equal(t, app1.ID, provider1.AppID)

	// Verify we can find it
	p1, err := model.GetOAuthProviderByName("google")
	assert.NoError(t, err)
	assert.Equal(t, provider1.ID, p1.ID)

	// Set CurrentApp to App2
	model.CurrentApp = app2

	// Should NOT find provider1
	_, err = model.GetOAuthProviderByName("google")
	assert.Error(t, err) // Should be record not found

	// Create Provider in App2 with SAME name
	provider2 := model.OAuthProvider{
		Name:         "google", // Same name as App1's provider
		ClientID:     "id2",
		ClientSecret: "secret2",
	}
	err = model.CreateOAuthProvider(&provider2)
	assert.NoError(t, err)
	assert.Equal(t, app2.ID, provider2.AppID)

	// Verify we find provider2
	p2, err := model.GetOAuthProviderByName("google")
	assert.NoError(t, err)
	assert.Equal(t, provider2.ID, p2.ID)
	assert.NotEqual(t, provider1.ID, p2.ID)

	// Switch back to App1
	model.CurrentApp = app1
	p1_again, err := model.GetOAuthProviderByName("google")
	assert.NoError(t, err)
	assert.Equal(t, provider1.ID, p1_again.ID)

	// List providers
	providers1, err := model.GetAllOAuthProviders()
	assert.NoError(t, err)

	// We might have other providers from other tests or defaults, so we check if our provider is in the list
	found := false
	for _, p := range providers1 {
		if p.ID == provider1.ID {
			found = true
			break
		}
	}
	assert.True(t, found)

	// Switch to App2
	model.CurrentApp = app2
	providers2, err := model.GetAllOAuthProviders()
	assert.NoError(t, err)

	found = false
	for _, p := range providers2 {
		if p.ID == provider2.ID {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestOAuthProvider_UpdateDelete(t *testing.T) {
	// Setup DB
	common.DBType = "sqlite"
	common.DBDSN = "file::memory:?cache=shared"
	model.InitDB()

	// Create separate apps for this test to avoid conflicts
	app1 := model.App{Name: "UpdateDeleteApp1", Enabled: true}
	model.DB.Create(&app1)

	app2 := model.App{Name: "UpdateDeleteApp2", Enabled: true}
	model.DB.Create(&app2)

	model.CurrentApp = &app1

	// Create Provider
	provider := model.OAuthProvider{
		Name:         "github",
		ClientID:     "id",
		ClientSecret: "secret",
	}
	err := model.CreateOAuthProvider(&provider)
	assert.NoError(t, err)

	// Switch to App2
	model.CurrentApp = &app2

	// Try to update provider from App1 while in App2 context
	updates := map[string]interface{}{
		"client_id": "new_id",
	}
	// Even if we pass the struct which has ID, UpdateOAuthProvider filters by CurrentApp.ID
	err = model.UpdateOAuthProvider(&provider, updates)
	assert.NoError(t, err)

	// Switch back to App1 to verify no change
	model.CurrentApp = &app1
	reloaded, err := model.GetOAuthProviderByName("github")
	assert.NoError(t, err)
	assert.Equal(t, "id", reloaded.ClientID) // Should still be "id"

	// Now update correctly
	err = model.UpdateOAuthProvider(&reloaded, updates)
	assert.NoError(t, err)

	reloaded, err = model.GetOAuthProviderByName("github")
	assert.NoError(t, err)
	assert.Equal(t, "new_id", reloaded.ClientID)

	// Try delete from App2
	model.CurrentApp = &app2
	err = model.DeleteOAuthProvider(reloaded.ID)
	assert.NoError(t, err) // 0 rows affected

	// Verify still exists in App1
	model.CurrentApp = &app1
	_, err = model.GetOAuthProviderByName("github")
	assert.NoError(t, err)

	// Delete from App1
	err = model.DeleteOAuthProvider(reloaded.ID)
	assert.NoError(t, err)

	_, err = model.GetOAuthProviderByName("github")
	assert.Error(t, err)
}
