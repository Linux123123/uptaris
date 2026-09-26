package response

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func links(c *gin.Context, pageNumber, pageSize, total int) gin.H {
	pages := (total + pageSize - 1) / pageSize
	values := c.Request.URL.Query()
	values.Set("page", strconv.Itoa(pageNumber))
	values.Set("pageSize", strconv.Itoa(pageSize))
	self := c.Request.URL.Path + "?" + values.Encode()
	var next, prev any
	if pageNumber < pages {
		values.Set("page", strconv.Itoa(pageNumber+1))
		next = c.Request.URL.Path + "?" + values.Encode()
	}
	if pageNumber > 1 {
		values.Set("page", strconv.Itoa(pageNumber-1))
		prev = c.Request.URL.Path + "?" + values.Encode()
	}
	return gin.H{"self": self, "next": next, "previous": prev}
}

func List(c *gin.Context, data any, total int, pageNumber, pageSize int) {
	c.JSON(200, gin.H{
		"data":       data,
		"pagination": gin.H{"page": pageNumber, "pageSize": pageSize, "total": total, "totalPages": (total + pageSize - 1) / pageSize},
		"links":      links(c, pageNumber, pageSize, total),
	})
}
