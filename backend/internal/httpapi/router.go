package httpapi

import (
	"net/http"

	"charon/backend/internal/config"
	"charon/backend/internal/domain/auth"

	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config, deps Dependencies) (*gin.Engine, error) {
	if err := deps.Validate(); err != nil {
		return nil, err
	}

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(requestIDMiddleware(), gin.Logger(), gin.Recovery())

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

	registerAuthRoutes(router.Group("/auth"), deps)

	wallet := router.Group("/wallet")
	wallet.Use(authenticationMiddleware(deps.Auth), requireRoles(auth.RoleStudent))
	registerWalletRoutes(wallet, deps)
	wallet.POST("/emergency-voucher/issue", notImplementedHandler("POST /wallet/emergency-voucher/issue"))

	boardings := router.Group("/boardings")
	boardings.Use(authenticationMiddleware(deps.Auth), requireRoles(auth.RoleStudent))
	boardings.GET("/preview", notImplementedHandler("GET /boardings/preview"))
	boardings.POST("", notImplementedHandler("POST /boardings"))

	driver := router.Group("/driver")
	driver.Use(authenticationMiddleware(deps.Auth), requireRoles(auth.RoleDriver))
	driver.Any("/*path", notImplementedHandler("/driver"))

	admin := router.Group("/admin")
	admin.Use(authenticationMiddleware(deps.Auth), requireRoles(auth.RoleCashier, auth.RoleAdmin, auth.RoleTechnicalAdmin))
	admin.Any("/*path", notImplementedHandler("/admin"))

	router.GET("/ws", authenticationMiddleware(deps.Auth), requireRoles(auth.RoleStudent, auth.RoleDriver), notImplementedHandler("GET /ws"))

	public := router.Group("/public")
	public.Any("/*path", notImplementedHandler("/public"))

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error_code": "NOT_FOUND",
			"message":    "No route matches the requested path.",
			"trace_id":   traceIDFromContext(c),
		})
	})

	return router, nil
}

func notImplementedHandler(pattern string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error_code": "NOT_IMPLEMENTED",
			"message":    "This endpoint is scaffolded but not implemented yet.",
			"trace_id":   traceIDFromContext(c),
			"route":      pattern,
		})
	}
}
