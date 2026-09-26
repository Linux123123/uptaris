package request

import (
	"math"
	"slices"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/response"
)

func ID(c *gin.Context, key string) (uint, bool) {
	resourceID, err := strconv.ParseUint(c.Param(key), 10, 0)
	if err != nil || resourceID == 0 || resourceID > math.MaxInt64 {
		response.Fail(c, 400, "invalid_id", "resource id must be a positive integer")
		return 0, false
	}
	return uint(resourceID), true
}

func Filter(c *gin.Context, parameter string, allowed ...string) (string, bool) {
	value := c.Query(parameter)
	if value != "" && !slices.Contains(allowed, value) {
		response.Fail(c, 400, "invalid_query", "invalid "+parameter)
		return "", false
	}
	return value, true
}
