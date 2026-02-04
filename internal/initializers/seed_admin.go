package initializers

import (
	"log"
	"os"

	"github.com/vitub/CLabServer/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func SeedAdmin() {
	var count int64
	DB.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&count)

	if count == 0 {
		log.Println("Seeding Admin user...")

		adminEmail := os.Getenv("ADMIN_EMAIL")
		adminPassword := os.Getenv("ADMIN_PASSWORD")

		if adminEmail == "" {
			adminEmail = "admin@clab.ide"
		}
		if adminPassword == "" {
			adminPassword = "admin" // Default for dev, should be changed in prod
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), 10)
		if err != nil {
			log.Printf("Failed to hash admin password: %v", err)
			return
		}

		admin := models.User{
			Name:     "System Admin",
			Email:    adminEmail,
			Password: string(hash),
			Role:     models.RoleAdmin,
		}

		if err := DB.Create(&admin).Error; err != nil {
			log.Printf("Failed to create admin user: %v", err)
		} else {
			log.Println("Admin user created successfully.")
		}
	} else {
		log.Println("Admin user already exists.")
	}
}
