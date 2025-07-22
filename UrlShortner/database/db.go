package database

import (
	"log"

	"UrlShortner/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("shortener.db"), &gorm.Config{}) // use sqlite or swap with MySQL/PostgreSQL
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&model.ShortUrl{})
	return db
}
