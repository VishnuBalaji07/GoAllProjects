package database

import (
	"ChatApiServer/models"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := "root:Vishnu@tj@tcp(127.0.0.1:3306)/ChatMessagedb?parseTime=true"
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database")
	}

	DB.AutoMigrate(&models.User{}, &models.Chat{}, &models.Message{}, &models.Reaction{}, &models.ChatMember{}, &models.MessageStatus{})
	fmt.Println("Database connected and migrated!")
}
