package models

import (
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	Model
	Email            string         `gorm:"uniqueIndex;not null" json:"email"`
	Password         string         `json:"-"`
	Name             string         `json:"name"`
	Role             string         `gorm:"default:'user'" json:"role"`
	Enabled          bool           `gorm:"default:true" json:"enabled"`
	LastLoginAt      *time.Time     `json:"last_login_at,omitempty"`
	StorageNamespace string         `gorm:"uniqueIndex;not null" json:"storage_namespace"`
	Identities       []UserIdentity `json:"identities,omitempty"`
	GitlabConfigs    []GitlabConfig `json:"gitlab_configs,omitempty"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.StorageNamespace = uuid.New().String()

	// Create user data directory
	dataDir := "/data"
	userDir := filepath.Join(dataDir, u.StorageNamespace)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return err
	}

	return
}
