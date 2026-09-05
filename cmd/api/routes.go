package main

import (
	"github.com/gin-gonic/gin"
)

func (app *application) routes() *gin.Engine {
	r := gin.Default()

	r.Use(app.enableCORS())

	r.GET("/", app.Home)

	r.POST("/authenticate", app.authenticate)
	r.GET("/refresh", app.refreshToken)
	r.GET("/logout", app.logout)

	r.GET("/movies", app.AllMovies)
	r.GET("/movies/:id", app.GetMovie)

	r.GET("/genres", app.AllGenres)
	r.GET("/movies/genres/:id", app.AllMoviesByGenre)

	admin := r.Group("/admin")
	admin.Use(app.authRequired())
	{
		admin.GET("/movies", app.MovieCatalog)
		admin.GET("/movies/:id", app.MovieForEdit)
		admin.PUT("/movies/0", app.InsertMovie)
		admin.PATCH("/movies/:id", app.UpdateMovie)
		admin.DELETE("/movies/:id", app.DeleteMovie)
	}

	return r
}