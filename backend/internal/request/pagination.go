package request

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/response"
)

func Page(c *gin.Context) (int, int, bool) {
	pageNumber, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || pageNumber < 1 {
		response.Fail(c, 400, "invalid_query", "page must be a positive integer")
		return 0, 0, false
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		response.Fail(c, 400, "invalid_query", "pageSize must be an integer from 1 through 100")
		return 0, 0, false
	}
	if pageNumber > math.MaxInt32/pageSize {
		response.Fail(c, 400, "invalid_query", "page is too large")
		return 0, 0, false
	}
	return pageNumber, pageSize, true
}
