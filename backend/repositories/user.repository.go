// repositories/user.repository.go
package repositories

import (
	"log"

	"aigateway-backend/models"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) GetByID(id string) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) List(limit, offset int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	r.db.Model(&models.User{}).Count(&total)
	err := r.db.Limit(limit).Offset(offset).Order("created_at DESC").Find(&users).Error

	return users, total, err
}

func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.User{}).Error
}

func (r *UserRepository) GetByAccessKey(key string) (*models.User, error) {
	var user models.User
	err := r.db.Where("access_key = ? AND is_active = ?", key, true).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateAccessKey(userID, key string) error {
	log.Printf("[UserRepo] UpdateAccessKey: userID=%s, key=%s...", userID, key[:10])
	result := r.db.Model(&models.User{}).Where("id = ?", userID).Update("access_key", key)
	if result.Error != nil {
		log.Printf("[UserRepo] UpdateAccessKey error: %v", result.Error)
		return result.Error
	}
	log.Printf("[UserRepo] UpdateAccessKey rows affected: %d", result.RowsAffected)
	return nil
}
