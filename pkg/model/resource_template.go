package model

import "github.com/pixelvide/kube-sentinel/pkg/common"

type ResourceTemplate struct {
	Model
	Name        string `json:"name" gorm:"type:varchar(255);uniqueIndex;not null"`
	Description string `json:"description"`
	YAML        string `json:"yaml" gorm:"type:text"`
}

func (ResourceTemplate) TableName() string {
	return common.GetAppTableName("k8s_resource_templates")
}
