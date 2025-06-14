package http

import (
	"1brc-challange/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ClientHandler struct {
	NumCPU         int
	ProcessService services.ProcessService
}

func NewClientHandler(numCPU int) *ClientHandler {
	return &ClientHandler{
		NumCPU:         numCPU,
		ProcessService: services.NewProcessService(numCPU),
	}
}

func (ch *ClientHandler) OneBillionRowChallange(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file upload"})
		return
	}
	defer file.Close()

	result, err := ch.ProcessService.OneBillionRowChallange(file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process file"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"result":  result,
		"num_cpu": ch.NumCPU,
		"message": "File processed successfully",
	})
}

func (ch *ClientHandler) AnomalyDetection(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file upload"})
		return
	}
	defer file.Close()

	result, err := ch.ProcessService.AnomalyDetection(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process file"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"result":  result,
		"num_cpu": ch.NumCPU,
		"message": "Anomaly detection completed successfully",
	})
}

func (ch *ClientHandler) GetNumCPU(c *gin.Context) {
	c.JSON(200, gin.H{
		"num_cpu": ch.NumCPU,
	})
}

func (ch *ClientHandler) HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}
