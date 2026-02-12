package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pixelvide/kube-sentinel/pkg/model"
	"gorm.io/datatypes"
)

// ListKnowledge returns knowledge entries for a specific cluster.
func ListKnowledge(c *gin.Context) {
	clusterIDStr := c.Param("id")
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cluster ID"})
		return
	}

	knowledgeBase, err := model.ListKnowledge(uint(clusterID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch knowledge base"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": knowledgeBase})
}

// AddKnowledge adds a new knowledge entry to a cluster.
func AddKnowledge(c *gin.Context) {
	clusterIDStr := c.Param("id")
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cluster ID"})
		return
	}

	var req struct {
		Content  string         `json:"content" binding:"required"`
		AddedBy  string         `json:"added_by"`
		Metadata map[string]any `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get real user from context if authenticated
	addedBy := req.AddedBy
	if addedBy == "" {
		addedBy = "User (Manual)"
	}

	metaJSON, err := json.Marshal(req.Metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process metadata"})
		return
	}

	kb := model.ClusterKnowledgeBase{
		ClusterID: uint(clusterID),
		Content:   req.Content,
		AddedBy:   addedBy,
		Metadata:  datatypes.JSON(metaJSON),
	}

	if err := model.AddKnowledge(&kb); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add knowledge"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": kb})
}

// DeleteKnowledge removes a knowledge entry.
func DeleteKnowledge(c *gin.Context) {
	// clusterID is in path but we just need knn_id to delete
	// We might check if knn_id belongs to clusterID for strictness, but ignoring for now for simplicity
	idStr := c.Param("knn_id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid knowledge ID"})
		return
	}

	if err := model.DeleteKnowledge(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete knowledge"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Knowledge deleted successfully"})
}
