package model

import "github.com/pixelvide/kube-sentinel/pkg/common"

type AuditLog struct {
	Model
	AppID        uint   `json:"appId" gorm:"index"`
	Action       string `json:"action" gorm:"type:varchar(50);index"`
	ActorID      uint   `json:"actorId" gorm:"index"`
	IPAddress    string `json:"ipAddress" gorm:"type:varchar(50)"`
	UserAgent    string `json:"userAgent" gorm:"type:text"`
	Payload      string `json:"payload" gorm:"type:text"`
	Success      bool   `json:"success"`
	ErrorMessage string `json:"errorMessage" gorm:"type:text"`

	// Relationships
	App   *App  `json:"app" gorm:"foreignKey:AppID;constraint:OnDelete:CASCADE"`
	Actor *User `json:"actor" gorm:"foreignKey:ActorID;constraint:OnDelete:CASCADE"`
}

func (AuditLog) TableName() string {
	return common.GetAppTableName("audit_logs")
}
