package dbrepo

import (
	"backend/internal/models"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestInsertMovie(t *testing.T) {
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=movies sslmode=disable timezone=UTC connect_timeout=5"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping DB test, unable to connect to Postgres: %v", err)
	}

	repo := &PostgresDBRepo{DB: db}

	m := models.Movie{
		Title:       "Test Movie " + time.Now().Format("15:04:05"),
		ReleaseDate: time.Now(),
		RunTime:     120,
		MPAARating:  "PG13",
		Description: "Test movie description",
		Image:       "",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		GenresArray: []int{1, 5},
	}

	newID, err := repo.InsertMovie(m)
	if err != nil {
		t.Fatalf("InsertMovie failed: %v", err)
	}
	if newID <= 0 {
		t.Fatalf("Expected newID > 0, got %d", newID)
	}

	err = repo.UpdateMovieGenres(newID, m.GenresArray)
	if err != nil {
		t.Fatalf("UpdateMovieGenres failed: %v", err)
	}

	// cleanup test movie
	_ = repo.DeleteMovie(newID)
}
