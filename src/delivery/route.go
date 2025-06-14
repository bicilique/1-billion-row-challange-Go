package delivery

import (
	http_delivery "1brc-challange/delivery/http"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RouteConfig struct {
	Router        *gin.Engine
	ClientHandler *http_delivery.ClientHandler
}

func (c *RouteConfig) SetupRoutes() {
	c.Router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "Welcome to 1BRC Challenge API",
		})
	})
	c.Router.POST("/one-billion-row-challenge", c.ClientHandler.OneBillionRowChallange)
	c.Router.POST("/anomaly-detection", c.ClientHandler.AnomalyDetection)
	c.Router.GET("/health", c.ClientHandler.HealthCheck)
	c.Router.GET("/numcpu", c.ClientHandler.GetNumCPU)
	c.Router.GET("/debug/pprof/", gin.WrapH(http.DefaultServeMux))
}
