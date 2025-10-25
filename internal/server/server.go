package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jayk0001/my-go-next-todo/internal/config"
	"github.com/jayk0001/my-go-next-todo/internal/database"
)

// Server holds the HTTP server and dependencies
type Server struct {
	router *gin.Engine
	db     *database.DB
	config *config.Config
}

// New creates a new server instance
func New(cfg *config.Config, db *database.DB) *Server {
	// Set gin mode based on environment
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	server := &Server{
		router: router,
		db:     db,
		config: cfg,
	}

	server.setupRoutes()
	return server

}

// setupRoutes configures all the routes
func (s *Server) setupRoutes() {
	// Health check endpoints
	health := s.router.Group("/health")
	{
		health.GET("/", s.healthCheck)
		health.GET("ready", s.readinessCheck)
		health.GET("live", s.livenessCheck)
	}

	// API version 1
	v1 := s.router.Group("/api/v1")
	{
		// TODO: add more routes here as we develop features
		v1.GET("/ping", s.ping)
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%s", s.config.Server.Host, s.config.Server.Port)
	fmt.Printf("Starting server on %s\n", addr)
	return s.router.Run(addr)
}

// Router returns the gin router (useful for testing)
func (s *Server) Router() *gin.Engine {
	return s.router
}

// healthCheck handles GET /health
func (s *Server) healthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Check database connectivity
	if err := s.db.Health(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"error":   "database connection failed",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "0.1.0",
		"services": gin.H{
			"database": "healthy",
		},
	})
}

func (s *Server) readinessCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := s.db.Health(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"error":  "database not ready",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// livenessCheck handles get /ealth/live
func (s *Server) livenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}

// ping handles GET /api/v1/ping
func (s *Server) ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

// cordMiddleware Handles CORS headrs
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credential", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
