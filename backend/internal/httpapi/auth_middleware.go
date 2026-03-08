package httpapi

import (
	"strings"

	"charon/backend/internal/domain/auth"

	"github.com/gin-gonic/gin"
)

const authIdentityContextKey = "auth_identity"

func authenticationMiddleware(authService AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := bearerToken(c.GetHeader("Authorization"))
		if err != nil {
			respondAuthError(c, err)
			c.Abort()
			return
		}

		identity, err := authService.AuthenticateAccessToken(c.Request.Context(), accessToken)
		if err != nil {
			respondAuthError(c, err)
			c.Abort()
			return
		}

		c.Set(authIdentityContextKey, identity)
		c.Next()
	}
}

func requireRoles(allowedRoles ...auth.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		identity, ok := currentIdentity(c)
		if !ok {
			respondAuthError(c, auth.ErrAuthorizationRequired)
			c.Abort()
			return
		}

		for _, role := range allowedRoles {
			if identity.User.Role == role {
				c.Next()
				return
			}
		}

		respondAuthError(c, auth.ErrInsufficientRole)
		c.Abort()
	}
}

func currentIdentity(c *gin.Context) (auth.AuthenticatedIdentity, bool) {
	identity, ok := c.Get(authIdentityContextKey)
	if !ok {
		return auth.AuthenticatedIdentity{}, false
	}

	value, ok := identity.(auth.AuthenticatedIdentity)
	return value, ok
}

func bearerToken(header string) (string, error) {
	header = strings.TrimSpace(header)
	if header == "" {
		return "", auth.ErrAuthorizationRequired
	}

	prefix, token, found := strings.Cut(header, " ")
	if !found || !strings.EqualFold(prefix, "Bearer") || strings.TrimSpace(token) == "" {
		return "", auth.ErrAuthorizationRequired
	}

	return strings.TrimSpace(token), nil
}
