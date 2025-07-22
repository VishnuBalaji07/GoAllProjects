package main

import (
	"EmailScheduler/service"
	"EmailScheduler/Scheduler"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mailer := &service.Mailer{
		From:     os.Getenv("EMAIL_FROM"),
		Password: os.Getenv("EMAIL_PASSWORD"),
		Host:     "smtp.gmail.com",
		Port:     "587",
	}

	scheduler.StartScheduler(mailer)
}
