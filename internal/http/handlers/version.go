package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Handles the /api/version endpoint.
func VersionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version": "1.0.0",
		"status":  "success",
	})
}
