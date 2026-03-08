package httpapi

import (
	"net/http"

	"charon/backend/internal/config"

	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "charon-api",
			"status":  "scaffolded",
			"docs": []string{
				"COMPREHENSIVE_SPEC.md",
				"API_SPEC.md",
				"ADMIN_CASHIER_API_SPEC.md",
				"DRIVER_SERVICE_API_SPEC.md",
				"STUDENT_SELF_SERVICE_API_SPEC.md",
				"SYSTEM_OPS_API_SPEC.md",
			},
		})
	})

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	for _, pattern := range []string{
		"/auth/*path",
		"/wallet/*path",
		"/boardings/*path",
		"/driver/*path",
		"/admin/*path",
		"/public/*path",
		"/ws",
	} {
		router.Any(pattern, notImplementedHandler(pattern))
	}

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error_code": "NOT_FOUND",
			"message":    "No route matches the requested path.",
		})
	})

	return router
}

func notImplementedHandler(pattern string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error_code": "NOT_IMPLEMENTED",
			"message":    "This endpoint is scaffolded but not implemented yet.",
			"route":      pattern,
		})
	}
}
