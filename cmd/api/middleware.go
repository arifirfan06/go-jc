package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *application) enableCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")

		if c.Request.Method == "OPTIONS" {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, X-CSRF-Token, Authorization")
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

func (app *application) authRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, _, err := app.auth.GetTokenFromHeaderAndVerify(c.Writer, c.Request)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	}
}

