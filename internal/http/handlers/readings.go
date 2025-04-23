package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// PostReadingsHandler queries and returns the list of readings from the database.
//
// A request body is expected with the following schema:
//
//	{
//	   "serial_number": string,
//	   "num_readings": uint32
//	}
func PostReadingsHandler(c *gin.Context) {
	var request struct {
		SerialNumber string `json:"serial_number" binding:"required"`
		NumReadings  uint32 `json:"num_readings" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		config.log.Error("Failed to bind JSON: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Validate the serial number and number of readings.
	if request.SerialNumber == "" {
		config.log.Error("Serial number is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Serial number is required"})
		return
	} else if request.NumReadings == 0 {
		config.log.Error("Number of readings must be greater than 0")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Number of readings must be greater than 0"})
		return
	}

	// Get the readings from the database.
	readings, err := config.db.GetLastReadings(request.SerialNumber, request.NumReadings)
	if err != nil {
		config.log.Error("Failed to get readings: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get readings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"readings": readings})
}
