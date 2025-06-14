package main

import (
	"1brc-challange/delivery"
	http_delivery "1brc-challange/delivery/http"
	"log"

	"runtime"

	"github.com/gin-gonic/gin"
)

func main() {
	// Set up CPU profiling if needed
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)

	// Initialize services
	clientHandler := http_delivery.NewClientHandler(numCPU)

	router := delivery.RouteConfig{
		Router:        gin.Default(),
		ClientHandler: clientHandler,
	}
	// Enable pprof for debugging
	router.SetupRoutes()

	// Set up routes
	err := router.Router.Run(":8080")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
