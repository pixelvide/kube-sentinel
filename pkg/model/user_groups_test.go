package model

import (
	"fmt"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/pixelvide/kube-sentinel/pkg/common"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupUserGroupsTestDB(t *testing.T) {
	var err error
	DB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	err = DB.AutoMigrate(&User{}, &UserIdentity{}, &UserConfig{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	common.DBType = "sqlite"
}

func TestFindWithSubOrUpsertUser_UpdatesGroups(t *testing.T) {
	setupUserGroupsTestDB(t)

	// 1. Create initial user and identity
	initialGroups := SliceString{"group1", "group2"}
	user := &User{
		Username:   "testuser",
		Name:       "Test User",
		Email:      "test@example.com",
		Provider:   "oidc",
		Sub:        "12345",
		OIDCGroups: initialGroups,
	}

	// This simulates first login
	err := FindWithSubOrUpsertUser(user)
	assert.NoError(t, err)

	// Verify DB state
	var identity UserIdentity
	err = DB.Where("user_id = ?", user.ID).First(&identity).Error
	assert.NoError(t, err)
	assert.Equal(t, initialGroups, identity.OIDCGroups)

	// 2. Simulate second login with NEW groups
	newGroups := SliceString{"group1", "group3", "group4"}
	userLogin2 := &User{
		Username:   "testuser", // Username should match
		Provider:   "oidc",
		Sub:        "12345", // Sub matches
		OIDCGroups: newGroups,
	}

	err = FindWithSubOrUpsertUser(userLogin2)
	assert.NoError(t, err)

	// Verify user object is updated
	assert.Equal(t, newGroups, userLogin2.OIDCGroups)

	// Verify DB state is updated
	var identity2 UserIdentity
	err = DB.Where("user_id = ?", user.ID).First(&identity2).Error
	assert.NoError(t, err)

	// Convert to string for comparison to avoid slice pointer issues if any
	assert.Equal(t, fmt.Sprintf("%v", newGroups), fmt.Sprintf("%v", identity2.OIDCGroups))

	// 3. Simulate third login with EMPTY groups
	emptyGroups := SliceString{}
	userLogin3 := &User{
		Username:   "testuser",
		Provider:   "oidc",
		Sub:        "12345",
		OIDCGroups: emptyGroups,
	}

	err = FindWithSubOrUpsertUser(userLogin3)
	assert.NoError(t, err)

	var identity3 UserIdentity
	err = DB.Where("user_id = ?", user.ID).First(&identity3).Error
	assert.NoError(t, err)
	assert.Equal(t, emptyGroups, identity3.OIDCGroups)
}
