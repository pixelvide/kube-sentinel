package model

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/pixelvide/kube-sentinel/pkg/common"
	"gorm.io/gorm"
)

func TestAISettingsEncryption(t *testing.T) {
	// Setup encryption key
	originalKey := common.KubeSentinelEncryptKey
	common.KubeSentinelEncryptKey = "test_encryption_key_12345"
	defer func() { common.KubeSentinelEncryptKey = originalKey }()

	// Setup in-memory DB
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	// AutoMigrate
	err = db.AutoMigrate(&AISettings{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	apiKey := "sk-1234567890-test-key"
	settings := AISettings{
		UserID:    1,
		ProfileID: 1,
		APIKey:    SecretString(apiKey),
		IsActive:  true,
	}

	if err := db.Create(&settings).Error; err != nil {
		t.Fatalf("failed to create settings: %v", err)
	}

	// Verify it is stored encrypted
	// We query using a struct with string field to bypass SecretString Scan/Value logic?
	// No, GORM maps struct fields to columns. If we query into a different struct that maps to the same table, we can inspect raw value.

	type AISettingsRaw struct {
		ID     uint
		APIKey string
	}

	var raw AISettingsRaw
	if err := db.Table(settings.TableName()).First(&raw, settings.ID).Error; err != nil {
		t.Fatalf("failed to query raw: %v", err)
	}

	if raw.APIKey == apiKey {
		t.Errorf("APIKey is stored in plaintext! Got: %s", raw.APIKey)
	}

	// Verify it is decrypted when read back as AISettings
	var readSettings AISettings
	if err := db.First(&readSettings, settings.ID).Error; err != nil {
		t.Fatalf("failed to read settings: %v", err)
	}

	if string(readSettings.APIKey) != apiKey {
		t.Errorf("expected %s, got %s", apiKey, readSettings.APIKey)
	}
}
