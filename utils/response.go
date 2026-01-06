package utils

import (
	"github.com/gin-gonic/gin"
)

func ErrorResponse(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"error":   true,
		"message": message,
	})
}

func SuccessResponse(c *gin.Context, status int, data interface{}) {
	c.JSON(status, gin.H{
		"error": false,
		"data":  data,
	})
}
