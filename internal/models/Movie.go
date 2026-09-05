package models

import "time"

type Movie struct {
	ID          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Title       string    `json:"title" gorm:"column:title"`
	ReleaseDate time.Time `json:"release_date" gorm:"column:release_date"`
	RunTime     int       `json:"runtime" gorm:"column:runtime"`
	MPAARating  string    `json:"mpaa_rating" gorm:"column:mpaa_rating"`
	Description string    `json:"description" gorm:"column:description"`
	Image       string    `json:"image" gorm:"column:image"`
	CreatedAt   time.Time `json:"-" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"-" gorm:"column:updated_at"`
	Genres      []*Genre  `json:"genres,omitempty" gorm:"many2many:movies_genres;foreignKey:ID;joinForeignKey:MovieID;references:ID;joinReferences:GenreID"`
	GenresArray []int     `json:"genres_array,omitempty" gorm:"-"`
}

type Genre struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Genre     string    `json:"genre" gorm:"column:genre"`
	Checked   bool      `json:"checked" gorm:"-"`
	CreatedAt time.Time `json:"-" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"-" gorm:"column:updated_at"`
}

type MovieGenre struct {
	ID      int `gorm:"primaryKey;autoIncrement"`
	MovieID int `gorm:"column:movie_id"`
	GenreID int `gorm:"column:genre_id"`
}

func (MovieGenre) TableName() string {
	return "movies_genres"
}


