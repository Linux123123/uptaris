package middleware

import (
	"mime"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/config"
	"github.com/uptaris/uptaris/backend/internal/response"
)

func RequireJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength == 0 {
			return
		}

		mediaType, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
		if err != nil || mediaType != "application/json" {
			response.Fail(c, http.StatusUnsupportedMediaType, "unsupported_media_type", "Content-Type must be application/json")
			return
		}

		c.Next()
	}
}

func Security(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Cache-Control", "no-store")
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead && c.Request.Method != http.MethodOptions {
			origin := c.GetHeader("Origin")
			if origin != "" && !slices.Contains(cfg.CORSOrigins, origin) {
				response.Fail(c, 403, "origin_not_allowed", "request origin not allowed")
				return
			}
		}

		if c.Request.ContentLength > 64<<10 {
			response.Fail(c, 413, "body_too_large", "request body exceeds 64 KiB")
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 64<<10)
		c.Next()
	}
}
