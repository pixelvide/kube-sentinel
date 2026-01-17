package models

type GitlabK8sAgentConfig struct {
	Model
	UserID         uint   `gorm:"not null" json:"user_id"`
	GitlabConfigID uint   `gorm:"not null" json:"gitlab_config_id"`
	AgentID        string `gorm:"not null" json:"agent_id"`
	AgentRepo      string `gorm:"not null" json:"agent_repo"`
	IsConfigured   bool   `gorm:"default:false" json:"is_configured"`

	// Relationships
	GitlabConfig GitlabConfig `gorm:"foreignKey:GitlabConfigID" json:"gitlab_config"`
}
