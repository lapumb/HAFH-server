package handlers

import (
	"hafh-server/internal/database"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetPeripheralsHandler returns the list of peripherals from the database.
func GetPeripheralsHandler(c *gin.Context) {
	// Get the peripherals from the database.
	peripherals, err := config.db.GetAllPeripherals()
	if err != nil {
		config.log.Error("Failed to get peripherals: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get peripherals"})
		return
	}

	// Return the peripherals as JSON.
	c.JSON(http.StatusOK, gin.H{"peripherals": peripherals})
}

// PostConfigurePeripheralHandler sets the name of a peripheral in the database.
//
// A request body is expected with the following schema:
//
//		{
//		   "serial_number": string,
//		   "name": string,
//	       "type": PeripheralType
//		}
func PostConfigurePeripheralHandler(c *gin.Context) {
	var request struct {
		SerialNumber string `json:"serial_number" binding:"required"`
		Name         string `json:"name" binding:"required"`
		Type         uint8  `json:"type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		config.log.Error("Failed to bind JSON: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Validate the serial number and name.
	if request.SerialNumber == "" {
		config.log.Error("Serial number is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Serial number is required"})
		return
	} else if request.Name == "" {
		config.log.Error("Name is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	// Get the peripheral from the database.
	peripheral, err := config.db.GetPeripheralBySerial(request.SerialNumber)
	if err != nil {
		config.log.Error("Failed to get peripheral: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get peripheral"})
		return
	}

	// Check if the peripheral exists.
	if peripheral == nil {
		config.log.Error("Peripheral not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "Peripheral not found"})
		return
	}

	// Set the name of the peripheral in the database.
	peripheral.Name = request.Name
	peripheral.Type = database.PeripheralType(request.Type)
	err = config.db.UpdatePeripheral(peripheral)
	if err != nil {
		config.log.Error("Failed to set peripheral name: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set peripheral name"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Peripheral name set successfully"})
}
