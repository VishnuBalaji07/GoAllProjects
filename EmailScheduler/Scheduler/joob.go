package scheduler

import (
	"EmailScheduler/service"
	"log"
	"os"

	"github.com/robfig/cron/v3"
)

func StartScheduler(mailer *service.Mailer) {
	c := cron.New()

	// Runs every 10 seconds for testing (replace with "@daily" for production)
	_, err := c.AddFunc("@every 10s", func() {
		mailer.Send(
			os.Getenv("EMAIL_TO"), // To
			"Kowsi loves pradeep", // Subject
			"true love",           // Body
		)
	})

	if err != nil {
		log.Fatal("Failed to schedule job:", err)
	}

	log.Println("Email Scheduler started")
	c.Start()

	select {} // keep the program running
}
