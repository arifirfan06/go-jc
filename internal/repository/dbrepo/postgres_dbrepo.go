package dbrepo

import (
	"backend/internal/models"

	"gorm.io/gorm"
)

type PostgresDBRepo struct {
	DB *gorm.DB
}

// Connection returns underlying connection pool.
func (m *PostgresDBRepo) Connection() *gorm.DB {
	return m.DB
}

// AllMovies returns a slice of movies, sorted by name. If the optional parameter genre
// is supplied, then only all movies for a particular genre is returned.
func (m *PostgresDBRepo) AllMovies(genre ...int) ([]*models.Movie, error) {
	var movies []*models.Movie
	db := m.DB.Order("title ASC")

	if len(genre) > 0 {
		db = db.Joins("JOIN movies_genres ON movies_genres.movie_id = movies.id").
			Where("movies_genres.genre_id = ?", genre[0])
	}

	err := db.Find(&movies).Error
	if err != nil {
		return nil, err
	}

	return movies, nil
}

// OneMovie returns a single movie and associated genres, if any.
func (m *PostgresDBRepo) OneMovie(id int) (*models.Movie, error) {
	var movie models.Movie

	err := m.DB.Preload("Genres", func(db *gorm.DB) *gorm.DB {
		return db.Order("genres.genre ASC")
	}).First(&movie, id).Error

	if err != nil {
		return nil, err
	}

	return &movie, nil
}

// OneMovieForEdit returns a single movie and associated genres, if any, for edit.
func (m *PostgresDBRepo) OneMovieForEdit(id int) (*models.Movie, []*models.Genre, error) {
	movie, err := m.OneMovie(id)
	if err != nil {
		return nil, nil, err
	}

	var genresArray []int
	for _, g := range movie.Genres {
		genresArray = append(genresArray, g.ID)
	}
	movie.GenresArray = genresArray

	allGenres, err := m.AllGenres()
	if err != nil {
		return nil, nil, err
	}

	return movie, allGenres, nil
}

// GetUserByEmail returns one user, by email.
func (m *PostgresDBRepo) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := m.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID returns one user, by ID.
func (m *PostgresDBRepo) GetUserByID(id int) (*models.User, error) {
	var user models.User
	err := m.DB.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// AllGenres returns a slice of genres, sorted by name.
func (m *PostgresDBRepo) AllGenres() ([]*models.Genre, error) {
	var genres []*models.Genre
	err := m.DB.Order("genre ASC").Find(&genres).Error
	if err != nil {
		return nil, err
	}
	return genres, nil
}

// InsertMovie inserts one movie into the database.
func (m *PostgresDBRepo) InsertMovie(movie models.Movie) (int, error) {
	err := m.DB.Create(&movie).Error
	if err != nil {
		return 0, err
	}
	return movie.ID, nil
}

// UpdateMovie updates one movie in the database.
func (m *PostgresDBRepo) UpdateMovie(movie models.Movie) error {
	return m.DB.Model(&models.Movie{ID: movie.ID}).Updates(map[string]interface{}{
		"title":        movie.Title,
		"description":  movie.Description,
		"release_date": movie.ReleaseDate,
		"runtime":      movie.RunTime,
		"mpaa_rating":  movie.MPAARating,
		"updated_at":   movie.UpdatedAt,
		"image":        movie.Image,
	}).Error
}

// UpdateMovieGenres first deletes all genres associated with a movie, and
// then inserts the ones stored in genreIDs.
func (m *PostgresDBRepo) UpdateMovieGenres(id int, genreIDs []int) error {
	err := m.DB.Where("movie_id = ?", id).Delete(&models.MovieGenre{}).Error
	if err != nil {
		return err
	}

	for _, gID := range genreIDs {
		mg := models.MovieGenre{
			MovieID: id,
			GenreID: gID,
		}
		if err := m.DB.Create(&mg).Error; err != nil {
			return err
		}
	}

	return nil
}

// DeleteMovie deletes one movie, by id.
func (m *PostgresDBRepo) DeleteMovie(id int) error {
	return m.DB.Delete(&models.Movie{}, id).Error
}

