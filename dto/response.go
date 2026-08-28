package dto

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ResponseFormat merepresentasikan struktur JSON standar untuk seluruh endpoint API.
type ResponseFormat struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// RespondSuccess mengirimkan response JSON sukses dengan HTTP status code tertentu.
func RespondSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, ResponseFormat{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// RespondError mengirimkan response JSON error dengan HTTP status code tertentu.
func RespondError(c *gin.Context, statusCode int, message string, errs interface{}) {
	c.JSON(statusCode, ResponseFormat{
		Success: false,
		Message: message,
		Errors:  errs,
	})
}

// RespondValidationError mengembalikan format error validasi input yang konsisten.
func RespondValidationError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, ResponseFormat{
		Success: false,
		Message: "Validasi data input gagal",
		Errors:  err.Error(),
	})
}
