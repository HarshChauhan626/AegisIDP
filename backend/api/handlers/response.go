package handlers

import "github.com/gin-gonic/gin"

// successResponse wraps a successful API response in the standard envelope.
func successResponse(data interface{}) gin.H {
	return gin.H{
		"data":  data,
		"error": nil,
	}
}

// errorResponse wraps an error in the standard envelope.
func errorResponse(code, message string) gin.H {
	return gin.H{
		"data": nil,
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	}
}

// paginatedResponse wraps a paginated list response.
func paginatedResponse(data interface{}, total int64, limit, offset int) gin.H {
	return gin.H{
		"data": data,
		"meta": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
		"error": nil,
	}
}
