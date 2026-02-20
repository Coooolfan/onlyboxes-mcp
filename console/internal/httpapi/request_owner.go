package httpapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const requestOwnerIDGinKey = "request_owner_id"

type requestOwnerIDContextKey struct{}

func setRequestOwnerID(c *gin.Context, ownerID string) {
	if c == nil {
		return
	}
	trimmedOwnerID := strings.TrimSpace(ownerID)
	if trimmedOwnerID == "" {
		return
	}
	c.Set(requestOwnerIDGinKey, trimmedOwnerID)
	if c.Request != nil {
		ctx := context.WithValue(c.Request.Context(), requestOwnerIDContextKey{}, trimmedOwnerID)
		c.Request = c.Request.WithContext(ctx)
	}
}

func requestOwnerIDFromGin(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if value, ok := c.Get(requestOwnerIDGinKey); ok {
		if ownerID, ok := value.(string); ok {
			trimmed := strings.TrimSpace(ownerID)
			if trimmed != "" {
				return trimmed
			}
		}
	}
	if c.Request != nil {
		return requestOwnerIDFromContext(c.Request.Context())
	}
	return ""
}

func requestOwnerIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value := ctx.Value(requestOwnerIDContextKey{})
	ownerID, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(ownerID)
}

func requireRequestOwnerID(c *gin.Context) (string, bool) {
	ownerID := requestOwnerIDFromGin(c)
	if ownerID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing token"})
		c.Abort()
		return "", false
	}
	return ownerID, true
}
