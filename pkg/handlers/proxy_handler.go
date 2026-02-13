package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pixelvide/kube-sentinel/pkg/cluster"
	"github.com/pixelvide/kube-sentinel/pkg/kube"
	"github.com/pixelvide/kube-sentinel/pkg/model"
	"github.com/pixelvide/kube-sentinel/pkg/rbac"
)

type ProxyHandler struct{}

func NewProxyHandler() *ProxyHandler {
	return &ProxyHandler{}
}

func (h *ProxyHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/namespaces/:namespace/:kind/:name/proxy/*path", h.HandleProxy)
}

func (h *ProxyHandler) HandleProxy(c *gin.Context) {
	cs := c.MustGet("cluster").(*cluster.ClientSet)
	user := c.MustGet("user").(model.User)
	kind := c.Param("kind")
	if kind != "pods" && kind != "services" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid kind, must be 'pods' or 'services'"})
		return
	}
	name := c.Param("name")
	namespace := c.Param("namespace")
	path := c.Param("path")

	if err := ValidateProxyRequest(namespace, name, path); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !rbac.CanAccess(user, kind, "get", cs.Name, namespace) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	kube.HandleProxy(c, cs.K8sClient, kind, namespace, name, path)
}

func ValidateProxyRequest(namespace, name, path string) error {
	if strings.Contains(namespace, "/") || strings.Contains(namespace, "..") {
		return errors.New("invalid namespace")
	}
	if strings.Contains(name, "/") || strings.Contains(name, "..") {
		return errors.New("invalid name")
	}
	// Path should not allow traversing up
	if strings.Contains(path, "..") {
		return errors.New("invalid path: contains '..'")
	}
	return nil
}
