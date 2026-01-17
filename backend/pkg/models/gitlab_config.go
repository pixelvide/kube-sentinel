package models

type GitlabConfig struct {
	Model
	UserID      uint   `gorm:"not null;uniqueIndex:idx_user_host" json:"user_id"`
	User        User   `gorm:"foreignKey:UserID" json:"-"`
	Host        string `gorm:"not null;uniqueIndex:idx_user_host" json:"gitlab_host"`
	IsHTTPS     bool   `gorm:"default:true" json:"is_https"`
	Token       string `gorm:"not null" json:"token"` // TODO: Encrypt this field in production
	IsValidated bool   `gorm:"default:false" json:"is_validated"`
}
