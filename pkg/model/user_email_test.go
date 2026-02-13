package model

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestUserEmail(t *testing.T) {
	// Setup localized DB for this test
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	DB = db
	// Need to migrate User, UserIdentity, and UserConfig
	err = DB.AutoMigrate(&User{}, &UserIdentity{}, &UserConfig{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	t.Run("Create User with Email", func(t *testing.T) {
		user := &User{
			Username: "user_with_email",
			Email:    "test@example.com",
			Enabled:  true,
		}
		err := DB.Create(user).Error
		assert.NoError(t, err)

		var fetched User
		err = DB.First(&fetched, user.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "test@example.com", fetched.Email)
	})

	t.Run("FindWithSubOrUpsertUser updates email for new user", func(t *testing.T) {
		user := &User{
			Username: "new_oidc_user",
			Provider: "oidc",
			Sub:      "sub1",
			Email:    "new@example.com",
		}

		err := FindWithSubOrUpsertUser(user)
		assert.NoError(t, err)

		var fetched User
		err = DB.Where("username = ?", "new_oidc_user").First(&fetched).Error
		assert.NoError(t, err)
		assert.Equal(t, "new@example.com", fetched.Email)
	})

	t.Run("FindWithSubOrUpsertUser does NOT update existing email", func(t *testing.T) {
		// Create initial user and identity
		user := &User{
			Username: "existing_oidc_user",
			Provider: "oidc",
			Sub:      "sub2",
			Email:    "old@example.com",
		}
		err := FindWithSubOrUpsertUser(user)
		assert.NoError(t, err)

		// New login with different email
		updatedUser := &User{
			Provider: "oidc",
			Sub:      "sub2",
			Email:    "updated@example.com",
		}
		err = FindWithSubOrUpsertUser(updatedUser)
		assert.NoError(t, err)
		// Should still be old email in returned object
		assert.Equal(t, "old@example.com", updatedUser.Email)

		var fetched User
		err = DB.Where("username = ?", "existing_oidc_user").First(&fetched).Error
		assert.NoError(t, err)
		// Should still be old email in DB
		assert.Equal(t, "old@example.com", fetched.Email)
	})

	t.Run("FindWithSubOrUpsertUser updates email when linking by username if empty", func(t *testing.T) {
		// Create user manually (no identity yet, no email)
		manualUser := &User{
			Username: "manual_user",
		}
		err := DB.Create(manualUser).Error
		assert.NoError(t, err)

		// Login via OIDC (matches username)
		oidcUser := &User{
			Username: "manual_user",
			Provider: "oidc",
			Sub:      "sub3",
			Email:    "oidc@example.com", // Email should update because it was empty
		}
		err = FindWithSubOrUpsertUser(oidcUser)
		assert.NoError(t, err)
		assert.Equal(t, "oidc@example.com", oidcUser.Email)

		var fetched User
		err = DB.Where("username = ?", "manual_user").First(&fetched).Error
		assert.NoError(t, err)
		assert.Equal(t, "oidc@example.com", fetched.Email)

		// Verify identity was created
		var identity UserIdentity
		err = DB.Where("provider = ? AND provider_id = ?", "oidc", "sub3").First(&identity).Error
		assert.NoError(t, err)
		assert.Equal(t, fetched.ID, identity.UserID)
	})

	t.Run("FindWithSubOrUpsertUser does NOT update email when linking by username if exists", func(t *testing.T) {
		// Create user manually (no identity yet, with email)
		manualUser := &User{
			Username: "manual_user_with_email",
			Email:    "manual@example.com",
		}
		err := DB.Create(manualUser).Error
		assert.NoError(t, err)

		// Login via OIDC (matches username)
		oidcUser := &User{
			Username: "manual_user_with_email",
			Provider: "oidc",
			Sub:      "sub4",
			Email:    "oidc@example.com", // Should be ignored
		}
		err = FindWithSubOrUpsertUser(oidcUser)
		assert.NoError(t, err)
		assert.Equal(t, "manual@example.com", oidcUser.Email)

		var fetched User
		err = DB.Where("username = ?", "manual_user_with_email").First(&fetched).Error
		assert.NoError(t, err)
		assert.Equal(t, "manual@example.com", fetched.Email)
	})

	t.Run("FindWithSubOrUpsertUser preserves existing email if new is empty", func(t *testing.T) {
		// Create user
		user := &User{
			Username: "preserve_email",
			Provider: "oidc",
			Sub:      "sub5",
			Email:    "keep@example.com",
		}
		err := FindWithSubOrUpsertUser(user)
		assert.NoError(t, err)

		// Login with empty email
		loginUser := &User{
			Provider: "oidc",
			Sub:      "sub5",
			Email:    "",
		}
		err = FindWithSubOrUpsertUser(loginUser)
		assert.NoError(t, err)

		var fetched User
		err = DB.Where("username = ?", "preserve_email").First(&fetched).Error
		assert.NoError(t, err)
		assert.Equal(t, "keep@example.com", fetched.Email)
	})
}
