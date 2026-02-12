package model

import (
	"encoding/json"

	"github.com/pixelvide/kube-sentinel/pkg/common"
	"gorm.io/datatypes"
)

// ClusterKnowledgeBase represents a knowledge entry associated with a cluster.
// It stores patterns, rules, or observations that help the AI understand the cluster's resources.
type ClusterKnowledgeBase struct {
	Model
	ClusterID uint           `json:"cluster_id" gorm:"index;not null"`
	Content   string         `json:"content" gorm:"type:text;not null"`
	AddedBy   string         `json:"added_by" gorm:"type:varchar(100)"`
	Metadata  datatypes.JSON `json:"metadata" gorm:"type:json"`
}

func (ClusterKnowledgeBase) TableName() string {
	return common.GetAppTableName("k8s_cluster_knowledge_bases")
}

// AddKnowledge adds a new knowledge entry.
func AddKnowledge(kb *ClusterKnowledgeBase) error {
	return DB.Create(kb).Error
}

// ListKnowledge retrieves knowledge entries for a specific cluster.
func ListKnowledge(clusterID uint) ([]ClusterKnowledgeBase, error) {
	var kbList []ClusterKnowledgeBase
	err := DB.Where("cluster_id = ?", clusterID).Find(&kbList).Error
	return kbList, err
}

// DeleteKnowledge removes a knowledge entry by ID.
func DeleteKnowledge(id uint) error {
	return DB.Delete(&ClusterKnowledgeBase{}, id).Error
}

// GetKnowledgeByID retrieves a specific knowledge entry.
func GetKnowledgeByID(id uint) (*ClusterKnowledgeBase, error) {
	var kb ClusterKnowledgeBase
	if err := DB.First(&kb, id).Error; err != nil {
		return nil, err
	}
	return &kb, nil
}

// UnmarshalMetadata helper to parse the JSON metadata
func (kb *ClusterKnowledgeBase) UnmarshalMetadata() (map[string]interface{}, error) {
	var meta map[string]interface{}
	if len(kb.Metadata) == 0 {
		return meta, nil
	}
	err := json.Unmarshal(kb.Metadata, &meta)
	return meta, err
}
