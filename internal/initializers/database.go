package initializers

import (
	"log"
	"os"

	"github.com/vitub/CLabServer/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectToDB() error {
	var err error
	dsn := os.Getenv("DATABASE_URL")
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	if err := DB.AutoMigrate(&models.User{}, &models.Classroom{}, &models.History{}, &models.Exercise{}, &models.ExerciseTopic{}, &models.ExamFolder{}); err != nil {
		return err
	}

	SeedAdmin()

	return nil
}
