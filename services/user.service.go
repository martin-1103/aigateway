// services/user.service.go
package services

import (
	"errors"

	"aigateway/models"
	"aigateway/repositories"

	"github.com/google/uuid"
)

type UserService struct {
	repo     *repositories.UserRepository
	password *PasswordService
}

func NewUserService(repo *repositories.UserRepository, password *PasswordService) *UserService {
	return &UserService{repo: repo, password: password}
}

func (s *UserService) Create(username, password string, role models.Role) (*models.User, error) {
	if !role.IsValid() {
		return nil, errors.New("invalid role")
	}

	hash, err := s.password.Hash(password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           uuid.New().String(),
		Username:     username,
		PasswordHash: hash,
		Role:         role,
		IsActive:     true,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetByID(id string) (*models.User, error) {
	return s.repo.GetByID(id)
}

func (s *UserService) GetByUsername(username string) (*models.User, error) {
	return s.repo.GetByUsername(username)
}

func (s *UserService) List(limit, offset int) ([]*models.User, int64, error) {
	return s.repo.List(limit, offset)
}

func (s *UserService) Update(user *models.User) error {
	return s.repo.Update(user)
}

func (s *UserService) ChangePassword(userID, newPassword string) error {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		return err
	}

	hash, err := s.password.Hash(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hash
	return s.repo.Update(user)
}

func (s *UserService) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *UserService) VerifyPassword(user *models.User, password string) bool {
	return s.password.Verify(password, user.PasswordHash)
}
