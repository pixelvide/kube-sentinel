package models

type KubeConfig struct {
	Model
	UserID    uint   `gorm:"not null;index" json:"user_id"`
	Name      string `gorm:"not null" json:"name"`
	Content   string `gorm:"type:text;not null" json:"content"`
	IsDefault bool   `gorm:"default:false" json:"is_default"` // The main config
}
