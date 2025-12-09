package utils

import "github.com/gin-gonic/gin"

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

func RespondJSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, data)
}

func RespondError(c *gin.Context, status int, message string) {
	c.JSON(status, ErrorResponse{Error: message})
}

func RespondSuccess(c *gin.Context, data interface{}) {
	c.JSON(200, SuccessResponse{
		Data: data,
	})
}
