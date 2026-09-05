package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func openDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (app *application) connectToDB() (*gorm.DB, error) {
	connection, err := openDB(app.DSN)
	if err != nil {
		return nil, err
	}

	log.Println("Connected to Postgres via GORM!")
	return connection, nil
}