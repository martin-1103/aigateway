package database

import (
	"log"

	"aigateway/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedDefaultAdmin(db *gorm.DB) error {
	var count int64
	db.Model(&models.User{}).Count(&count)

	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), 10)
	if err != nil {
		return err
	}

	admin := &models.User{
		ID:           uuid.New().String(),
		Username:     "admin",
		PasswordHash: string(hash),
		Role:         models.RoleAdmin,
		IsActive:     true,
	}

	if err := db.Create(admin).Error; err != nil {
		return err
	}

	log.Println("⚠️  Default admin created (admin/admin123) - CHANGE PASSWORD!")
	return nil
}
