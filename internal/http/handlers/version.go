package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

const ApiVersionMajor = "v1"
const ApiVersionMinor = "0"
const ApiVersionPatch = "0"
const ApiVersion = ApiVersionMajor + "." + ApiVersionMinor + "." + ApiVersionPatch

// GetVersionHandler returns the API version.
func GetVersionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version": string(ApiVersion),
	})
}
