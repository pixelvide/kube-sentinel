package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetPortForwards returns a placeholder list of port forwards
func GetPortForwards(c *gin.Context) {
	// For now, return an empty list as we haven't implemented port forwarding management yet
	c.JSON(http.StatusOK, gin.H{"portforwards": []interface{}{}})
}
