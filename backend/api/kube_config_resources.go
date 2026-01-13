package api

import (
	"net/http"
	"time"

	"cloud-sentinel-k8s/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetConfigMaps lists config maps for a given namespace and context
func GetConfigMaps(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := c.Query("context")

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	namespaces := ParseNamespaces(ns)

	type ConfigMapInfo struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Age       string `json:"age"`
		DataCount int    `json:"data_count"`
	}

	var configMaps []ConfigMapInfo

	for _, singleNs := range namespaces {
		list, err := clientset.CoreV1().ConfigMaps(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, item := range list.Items {
			configMaps = append(configMaps, ConfigMapInfo{
				Name:      item.Name,
				Namespace: item.Namespace,
				Age:       item.CreationTimestamp.Time.Format(time.RFC3339),
				DataCount: len(item.Data),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"configmaps": configMaps})
}

// GetSecrets lists secrets for a given namespace and context
func GetSecrets(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := c.Query("context")

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	namespaces := ParseNamespaces(ns)

	type SecretInfo struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Age       string `json:"age"`
		Type      string `json:"type"`
		DataCount int    `json:"data_count"`
	}

	var secrets []SecretInfo

	for _, singleNs := range namespaces {
		list, err := clientset.CoreV1().Secrets(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, item := range list.Items {
			secrets = append(secrets, SecretInfo{
				Name:      item.Name,
				Namespace: item.Namespace,
				Age:       item.CreationTimestamp.Time.Format(time.RFC3339),
				Type:      string(item.Type),
				DataCount: len(item.Data),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"secrets": secrets})
}

// GetResourceQuotas lists resource quotas for a given namespace and context
func GetResourceQuotas(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := c.Query("context")

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	namespaces := ParseNamespaces(ns)

	type ResourceQuotaInfo struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Age       string `json:"age"`
	}

	var resourceQuotas []ResourceQuotaInfo

	for _, singleNs := range namespaces {
		list, err := clientset.CoreV1().ResourceQuotas(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, item := range list.Items {
			resourceQuotas = append(resourceQuotas, ResourceQuotaInfo{
				Name:      item.Name,
				Namespace: item.Namespace,
				Age:       item.CreationTimestamp.Time.Format(time.RFC3339),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"resourcequotas": resourceQuotas})
}

// GetLimitRanges lists limit ranges for a given namespace and context
func GetLimitRanges(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := c.Query("context")

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	namespaces := ParseNamespaces(ns)

	type LimitRangeInfo struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Age       string `json:"age"`
	}

	var limitRanges []LimitRangeInfo

	for _, singleNs := range namespaces {
		list, err := clientset.CoreV1().LimitRanges(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, item := range list.Items {
			limitRanges = append(limitRanges, LimitRangeInfo{
				Name:      item.Name,
				Namespace: item.Namespace,
				Age:       item.CreationTimestamp.Time.Format(time.RFC3339),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"limitranges": limitRanges})
}

// GetHPAs lists horizontal pod autoscalers for a given namespace and context
func GetHPAs(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := c.Query("context")

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	namespaces := ParseNamespaces(ns)

	type HPAInfo struct {
		Name         string `json:"name"`
		Namespace    string `json:"namespace"`
		Age          string `json:"age"`
		Reference    string `json:"reference"`
		MinReplicas  int32  `json:"min_replicas"`
		MaxReplicas  int32  `json:"max_replicas"`
		CurrReplicas int32  `json:"curr_replicas"`
	}

	var hpas []HPAInfo

	for _, singleNs := range namespaces {
		list, err := clientset.AutoscalingV2().HorizontalPodAutoscalers(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, item := range list.Items {
			minReplicas := int32(0)
			if item.Spec.MinReplicas != nil {
				minReplicas = *item.Spec.MinReplicas
			}
			hpas = append(hpas, HPAInfo{
				Name:         item.Name,
				Namespace:    item.Namespace,
				Age:          item.CreationTimestamp.Time.Format(time.RFC3339),
				Reference:    item.Spec.ScaleTargetRef.Kind + "/" + item.Spec.ScaleTargetRef.Name,
				MinReplicas:  minReplicas,
				MaxReplicas:  item.Spec.MaxReplicas,
				CurrReplicas: item.Status.CurrentReplicas,
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"hpas": hpas})
}
