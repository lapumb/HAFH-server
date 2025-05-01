package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const ApiVersionMajor = "v1"
const ApiVersionMinor = "0"
const ApiVersionPatch = "0"
const ApiVersion = ApiVersionMajor + "." + ApiVersionMinor + "." + ApiVersionPatch

// GetApiVersion returns the API version.
func GetApiVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version": string(ApiVersion),
	})
}
