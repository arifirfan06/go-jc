package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type JSONResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (app *application) writeJSON(c *gin.Context, status int, data interface{}, headers ...http.Header) error {
	if len(headers) > 0 {
		for key, value := range headers[0] {
			for _, val := range value {
				c.Header(key, val)
			}
		}
	}

	c.JSON(status, data)
	return nil
}

func (app *application) readJSON(c *gin.Context, data interface{}) error {
	return c.ShouldBindJSON(data)
}

func (app *application) errorJSON(c *gin.Context, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	var payload JSONResponse
	payload.Error = true
	payload.Message = err.Error()

	c.JSON(statusCode, payload)
	return nil
}