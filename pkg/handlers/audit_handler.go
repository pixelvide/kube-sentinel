package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pixelvide/kube-sentinel/pkg/model"
)

func ListAuditLogs(c *gin.Context) {
	page := 1
	size := 20

	if p := strings.TrimSpace(c.Query("page")); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page parameter"})
			return
		}
	}
	if s := strings.TrimSpace(c.Query("size")); s != "" {
		if parsed, err := strconv.Atoi(s); err == nil && parsed > 0 {
			size = parsed
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid size parameter"})
			return
		}
	}

	actorID := uint64(0)
	if op := strings.TrimSpace(c.Query("operatorId")); op != "" {
		parsed, err := strconv.ParseUint(op, 10, 64)
		if err == nil {
			actorID = parsed
		}
	}

	search := strings.TrimSpace(c.Query("search"))
	operation := strings.TrimSpace(c.Query("operation"))
	clusterName := strings.TrimSpace(c.Query("cluster"))
	resourceType := strings.TrimSpace(c.Query("resourceType"))
	resourceName := strings.TrimSpace(c.Query("resourceName"))
	namespace := strings.TrimSpace(c.Query("namespace"))

	query := model.DB.Model(&model.AuditLog{})
	if actorID > 0 {
		query = query.Where("actor_id = ?", actorID)
	}
	if clusterName != "" {
		query = query.Where("payload LIKE ?", "%"+clusterName+"%")
	}
	if resourceType != "" {
		query = query.Where("payload LIKE ?", "%"+resourceType+"%")
	}
	if resourceName != "" {
		query = query.Where("payload LIKE ?", "%"+resourceName+"%")
	}
	if namespace != "" {
		query = query.Where("payload LIKE ?", "%"+namespace+"%")
	}
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("payload LIKE ?", like)
	}
	if operation != "" {
		query = query.Where("action = ?", operation)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logs := []model.AuditLog{}
	if err := query.Preload("Actor").Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Map to readable format
	history := make([]map[string]interface{}, 0, len(logs))
	for _, l := range logs {
		var p map[string]interface{}
		_ = json.Unmarshal([]byte(l.Payload), &p)
		if p == nil {
			p = make(map[string]interface{})
		}
		actorName := ""
		if l.Actor != nil {
			actorName = l.Actor.Username
		}
		history = append(history, map[string]interface{}{
			"id":            l.ID,
			"createdAt":     l.CreatedAt,
			"updatedAt":     l.UpdatedAt,
			"clusterName":   p["clusterName"],
			"resourceType":  p["resourceType"],
			"resourceName":  p["resourceName"],
			"namespace":     p["namespace"],
			"operationType": l.Action,
			"actor":         actorName,
			"operator": map[string]interface{}{
				"username": actorName,
			},
			"success":      l.Success,
			"errorMessage": l.ErrorMessage,
			"ipAddress":    l.IPAddress,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  history,
		"total": total,
		"page":  page,
		"size":  size,
	})
}
