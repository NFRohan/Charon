package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const traceIDContextKey = "trace_id"

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := sanitizeTraceID(c.GetHeader("X-Request-ID"))
		if traceID == "" {
			traceID = newTraceID()
		}

		c.Set(traceIDContextKey, traceID)
		c.Writer.Header().Set("X-Request-ID", traceID)
		c.Next()
	}
}

func sanitizeTraceID(value string) string {
	value = strings.TrimSpace(value)
	if len(value) == 0 || len(value) > 128 {
		return ""
	}

	for _, r := range value {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= 'A' && r <= 'Z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		switch r {
		case '-', '_', '.', ':':
			continue
		default:
			return ""
		}
	}

	return value
}

func traceIDFromContext(c *gin.Context) string {
	if traceID, ok := c.Get(traceIDContextKey); ok {
		if value, ok := traceID.(string); ok && value != "" {
			return value
		}
	}

	return newTraceID()
}

func newTraceID() string {
	buffer := make([]byte, 12)
	if _, err := rand.Read(buffer); err == nil {
		return "req_" + hex.EncodeToString(buffer)
	}

	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
