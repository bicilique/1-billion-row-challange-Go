package delivery

import (
	http_delivery "1brc-challange/delivery/http"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type RouteConfig struct {
	Router        *gin.Engine
	ClientHandler *http_delivery.ClientHandler
}

func (c *RouteConfig) SetupRoutes() {
	// Set up Prometheus metrics
	prometheus.MustRegister(httpRequests)
	prometheus.MustRegister(httpDuration)
	// prometheus.MustRegister(collectors.NewGoCollector())                                       // goroutines, GC, mem
	// prometheus.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})) // CPU, memory, FD

	c.Router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "Welcome to 1BRC Challenge API",
		})
	})

	c.Router.Use(PrometheusMiddleware())
	c.Router.POST("/one-billion-row-challenge", c.ClientHandler.OneBillionRowChallange)
	c.Router.POST("/anomaly-detection", c.ClientHandler.AnomalyDetection)

	c.Router.GET("/health", c.ClientHandler.HealthCheck)
	c.Router.GET("/numcpu", c.ClientHandler.GetNumCPU)
	c.Router.GET("/debug/pprof/", gin.WrapH(http.DefaultServeMux))
	c.Router.GET("/metrics", gin.WrapH(promhttp.Handler()))
}

var (
	httpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path", "method", "status"},
	)

	httpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path // fallback
		}

		status := fmt.Sprint(c.Writer.Status())
		method := c.Request.Method

		httpRequests.WithLabelValues(path, method, status).Inc()
		httpDuration.WithLabelValues(path, method).Observe(duration)
	}
}
