package model

import "github.com/pixelvide/kube-sentinel/pkg/common"

type GitlabHosts struct {
	Model
	Host    string `gorm:"not null;uniqueIndex:idx_user_host" json:"gitlab_host"`
	IsHTTPS *bool  `gorm:"default:true" json:"is_https"`
}

func (GitlabHosts) TableName() string {
	return common.GetAppTableName("gitlab_hosts")
}
