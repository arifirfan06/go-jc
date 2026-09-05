package main

import (
	"backend/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// Home displays the status of the api, as JSON.
func (app *application) Home(c *gin.Context) {
	var payload = struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Version string `json:"version"`
	}{
		Status:  "active",
		Message: "Go Movies up and running",
		Version: "1.0.0",
	}

	_ = app.writeJSON(c, http.StatusOK, payload)
}

// AllMovies returns a slice of all movies as JSON.
func (app *application) AllMovies(c *gin.Context) {
	movies, err := app.DB.AllMovies()
	if err != nil {
		app.errorJSON(c, err)
		return
	}

	_ = app.writeJSON(c, http.StatusOK, movies)
}

// authenticate authenticates a user, and returns a JWT.
func (app *application) authenticate(c *gin.Context) {
	// read json payload
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(c, &requestPayload)
	if err != nil {
		app.errorJSON(c, err, http.StatusBadRequest)
		return
	}

	// validate user against database
	user, err := app.DB.GetUserByEmail(requestPayload.Email)
	if err != nil {
		app.errorJSON(c, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	// check password
	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		app.errorJSON(c, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	// create a jwt user
	u := jwtUser{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	// generate tokens
	tokens, err := app.auth.GenerateTokenPair(&u)
	if err != nil {
		app.errorJSON(c, err)
		return
	}

	refreshCookie := app.auth.GetRefreshCookie(tokens.RefreshToken)
	http.SetCookie(c.Writer, refreshCookie)

	app.writeJSON(c, http.StatusAccepted, tokens)
}

// refreshToken checks for a valid refresh cookie, and returns a JWT if it finds one.
func (app *application) refreshToken(c *gin.Context) {
	for _, cookie := range c.Request.Cookies() {
		if cookie.Name == app.auth.CookieName {
			claims := &Claims{}
			refreshToken := cookie.Value

			// parse the token to get the claims
			_, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(app.JWTSecret), nil
			})
			if err != nil {
				app.errorJSON(c, errors.New("unauthorized"), http.StatusUnauthorized)
				return
			}

			// get the user id from the token claims
			userID, err := strconv.Atoi(claims.Subject)
			if err != nil {
				app.errorJSON(c, errors.New("unknown user"), http.StatusUnauthorized)
				return
			}

			user, err := app.DB.GetUserByID(userID)
			if err != nil {
				app.errorJSON(c, errors.New("unknown user"), http.StatusUnauthorized)
				return
			}

			u := jwtUser{
				ID:        user.ID,
				FirstName: user.FirstName,
				LastName:  user.LastName,
			}

			tokenPairs, err := app.auth.GenerateTokenPair(&u)
			if err != nil {
				app.errorJSON(c, errors.New("error generating tokens"), http.StatusUnauthorized)
				return
			}

			http.SetCookie(c.Writer, app.auth.GetRefreshCookie(tokenPairs.RefreshToken))

			app.writeJSON(c, http.StatusOK, tokenPairs)
			return
		}
	}

	app.errorJSON(c, errors.New("unauthorized"), http.StatusUnauthorized)
}

// logout logs the user out by sending an expired cookie to delete the refresh cookie.
func (app *application) logout(c *gin.Context) {
	http.SetCookie(c.Writer, app.auth.GetExpiredRefreshCookie())
	c.Status(http.StatusAccepted)
}

// MovieCatalog returns a list of all movies as JSON
func (app *application) MovieCatalog(c *gin.Context) {
	movies, err := app.DB.AllMovies()
	if err != nil {
		app.errorJSON(c, err)
		return
	}

	_ = app.writeJSON(c, http.StatusOK, movies)
}

// GetMovie returns one movie, as JSON.
func (app *application) GetMovie(c *gin.Context) {
	id := c.Param("id")
	movieID, err := strconv.Atoi(id)
	if err != nil {
		app.errorJSON(c, err)
		return
	}

	movie, err := app.DB.OneMovie(movieID)
	if err != nil {
		app.errorJSON(c, err)
		return
	}

	_ = app.writeJSON(c, http.StatusOK, movie)
}

// MovieForEdit returns a JSON payload for a given movie and a list of all genres, for edit.
func (app *application) MovieForEdit(c *gin.Context) {
	id := c.Param("id")
	movieID, err := strconv.Atoi(id)
	if err != nil {
		app.errorJSON(c, err)
		return
	}

	movie, genres, err := app.DB.OneMovieForEdit(movieID)
	if err != nil {
		app.errorJSON(c, err)
		return
	}

	var payload = struct {
		Movie  *models.Movie   `json:"movie"`
		Genres []*models.Genre `json:"genres"`
	}{
		movie,
		genres,
	}

	_ = app.writeJSON(c, http.StatusOK, payload)
}

// AllGenres returns a slice of all genres as JSON.
func (app *application) AllGenres(c *gin.Context) {
	genres, err := app.DB.AllGenres()
	if err != nil {
		app.errorJSON(c, err)
		return
	}

	_ = app.writeJSON(c, http.StatusOK, genres)
}

// InsertMovie receives a JSON payload and tries to insert a movie into the database.
func (app *application) InsertMovie(c *gin.Context) {
	var movie models.Movie

	err := app.readJSON(c, &movie)
	if err != nil {
		app.errorJSON(c, err)
		return
	}

	// try to get an image
	movie = app.getPoster(movie)

	movie.CreatedAt = time.Now()
	movie.UpdatedAt = time.Now()

	newID, err := app.DB.InsertMovie(movie)
	if err != nil {
		app.errorJSON(c, err)
		return
	}

	// now handle genres
	err = app.DB.UpdateMovieGenres(newID, movie.GenresArray)
	if err != nil {
		app.errorJSON(c, err)
		return
	}

	resp := JSONResponse{
		Error:   false,
		Message: "movie updated",
	}

	app.writeJSON(c, http.StatusAccepted, resp)
}

// getPoster tries to get a poster image from themoviedb.org.
func (app *application) getPoster(movie models.Movie) models.Movie {
	type TheMovieDB struct {
		Page    int `json:"page"`
		Results []struct {
			PosterPath string `json:"poster_path"`
		} `json:"results"`
		TotalPages int `json:"total_pages"`
	}

	client := &http.Client{}
	theUrl := fmt.Sprintf("https://api.themoviedb.org/3/search/movie?api_key=%s", app.APIKey)

	req, err := http.NewRequest("GET", theUrl+"&query="+url.QueryEscape(movie.Title), nil)
	if err != nil {
		log.Println(err)
		return movie
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return movie
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return movie
	}

	var responseObject TheMovieDB

	json.Unmarshal(bodyBytes, &responseObject)

	if len(responseObject.Results) > 0 {
		movie.Image = responseObject.Results[0].PosterPath
	}

	return movie
}

// UpdateMovie updates a movie in the database, based on a JSON payload.
func (app *application) UpdateMovie(c *gin.Context) {
	var payload models.Movie

	err := app.readJSON(c, &payload)
	if err != nil {
		app.errorJSON(c, err)
		return
	}

	movie, err := app.DB.OneMovie(payload.ID)
	if err != nil {
		app.errorJSON(c, err)
		return
	}

	movie.Title = payload.Title
	movie.ReleaseDate = payload.ReleaseDate
	movie.Description = payload.Description
	movie.MPAARating = payload.MPAARating
	movie.RunTime = payload.RunTime
	movie.UpdatedAt = time.Now()

	err = app.DB.UpdateMovie(*movie)
	if err != nil {
		app.errorJSON(c, err)
		return
	}

	err = app.DB.UpdateMovieGenres(movie.ID, payload.GenresArray)
	if err != nil {
		app.errorJSON(c, err)
		return
	}

	resp := JSONResponse{
		Error:   false,
		Message: "movie updated",
	}

	app.writeJSON(c, http.StatusAccepted, resp)
}

// DeleteMovie deletes a movie from the database, by ID.
func (app *application) DeleteMovie(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		app.errorJSON(c, err)
		return
	}

	err = app.DB.DeleteMovie(id)
	if err != nil {
		app.errorJSON(c, err)
		return
	}

	resp := JSONResponse{
		Error:   false,
		Message: "movie deleted",
	}

	app.writeJSON(c, http.StatusAccepted, resp)
}

func (app *application) AllMoviesByGenre(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		app.errorJSON(c, err)
		return
	}

	movies, err := app.DB.AllMovies(id)
	if err != nil {
		app.errorJSON(c, err)
		return
	}

	app.writeJSON(c, http.StatusOK, movies)
}

