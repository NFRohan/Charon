package httpapi

import (
	"net/http"
	"strings"

	"charon/backend/internal/domain/auth"

	"github.com/gin-gonic/gin"
)

type authHandler struct {
	authService AuthService
}

type loginRequest struct {
	LoginID  string `json:"login_id"`
	Password string `json:"password"`
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func registerAuthRoutes(router *gin.RouterGroup, deps Dependencies) {
	handler := authHandler{authService: deps.Auth}

	router.POST("/login", handler.login)
	router.POST("/refresh", handler.refresh)
	router.POST("/logout", handler.logout)
}

func (h authHandler) login(c *gin.Context) {
	var request loginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondValidationError(c, "Request body must be valid JSON.")
		return
	}

	request.LoginID = strings.TrimSpace(request.LoginID)
	if request.LoginID == "" {
		respondValidationError(c, "One or more fields are invalid.", fieldError{
			Field:   "login_id",
			Message: "Login ID is required.",
		})
		return
	}

	if request.Password == "" {
		respondValidationError(c, "One or more fields are invalid.", fieldError{
			Field:   "password",
			Message: "Password is required.",
		})
		return
	}

	result, err := h.authService.Login(c.Request.Context(), auth.LoginInput{
		LoginID:  request.LoginID,
		Password: request.Password,
	})
	if err != nil {
		respondAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":             result.AccessToken,
		"access_token_expires_at":  result.AccessTokenExpiresAt.UTC(),
		"refresh_token":            result.RefreshToken,
		"refresh_token_expires_at": result.RefreshTokenExpiresAt.UTC(),
		"role":                     result.User.Role,
		"user_id":                  result.User.ID,
		"profile_summary": gin.H{
			"name":        result.User.Name,
			"status":      result.User.Status,
			"fare_exempt": result.User.FareExempt,
		},
	})
}

func (h authHandler) refresh(c *gin.Context) {
	refreshToken, ok := readRefreshToken(c)
	if !ok {
		return
	}

	result, err := h.authService.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		respondAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":             result.AccessToken,
		"access_token_expires_at":  result.AccessTokenExpiresAt.UTC(),
		"refresh_token":            result.RefreshToken,
		"refresh_token_expires_at": result.RefreshTokenExpiresAt.UTC(),
	})
}

func (h authHandler) logout(c *gin.Context) {
	refreshToken, ok := readRefreshToken(c)
	if !ok {
		return
	}

	if err := h.authService.Logout(c.Request.Context(), refreshToken); err != nil {
		respondAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logged_out": true,
	})
}

func readRefreshToken(c *gin.Context) (string, bool) {
	var request refreshTokenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondValidationError(c, "Request body must be valid JSON.")
		return "", false
	}

	request.RefreshToken = strings.TrimSpace(request.RefreshToken)
	if request.RefreshToken == "" {
		respondValidationError(c, "One or more fields are invalid.", fieldError{
			Field:   "refresh_token",
			Message: "Refresh token is required.",
		})
		return "", false
	}

	return request.RefreshToken, true
}
