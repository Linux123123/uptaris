package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/auth"
	"github.com/uptaris/uptaris/backend/internal/response"
	"golang.org/x/time/rate"
)

func AuthLimit() gin.HandlerFunc {
	limiter := rate.NewLimiter(5, 20)
	slots := make(chan struct{}, 4)
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.Header("Retry-After", "1")
			response.Fail(c, 429, "rate_limited", "too many authentication attempts")
			return
		}
		select {
		case slots <- struct{}{}:
			defer func() { <-slots }()
			c.Next()
		default:
			c.Header("Retry-After", "1")
			response.Fail(c, 429, "rate_limited", "authentication busy; retry shortly")
		}
	}
}

// Require authenticates a bearer token and writes any authorization failure.
func Require(c *gin.Context, service *auth.Service, roles ...string) (auth.Identity, bool) {
	identity, err := service.Authenticate(c.Request.Context(), c.GetHeader("Authorization"), roles...)
	if err != nil {
		response.Error(c, err)
		return auth.Identity{}, false
	}
	return identity, true
}
