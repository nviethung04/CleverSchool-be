package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse chuẩn hóa cấu trúc phản hồi API
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"` // `omitempty` giúp loại bỏ nếu Data = nil
}

// JSONResponse giúp gửi phản hồi theo chuẩn API
func JSONResponse(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, APIResponse{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

func SuccessResponse(c *gin.Context, data ...interface{}) {
	message := "Operation successful"

	if len(data) > 0 {
		if msg, ok := data[len(data)-1].(string); ok {
			message = msg
			data = data[:len(data)-1]
		}
	}

	JSONResponse(c, http.StatusOK, message, data)
}

func DeleteResponse(c *gin.Context) {
	message := "Delete successful"

	c.JSON(http.StatusOK, APIResponse{
		Code:    http.StatusOK,
		Message: message,
	})
}

func ErrorResponse(c *gin.Context, code int, message string, details ...interface{}) {
	var data interface{}
	if len(details) > 0 {
		data = details[0]
	}

	c.JSON(code, APIResponse{
		Code:    code,
		Message: message,
		Data:    data,
	})
}
