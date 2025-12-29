// services/auth.service.go
package services

import (
	"errors"
	"log"

	"aigateway-backend/models"
)

type AuthService struct {
	userService   *UserService
	jwtService    *JWTService
	apiKeyService *APIKeyService
}

func NewAuthService(user *UserService, jwt *JWTService, apiKey *APIKeyService) *AuthService {
	return &AuthService{
		userService:   user,
		jwtService:    jwt,
		apiKeyService: apiKey,
	}
}

type LoginResponse struct {
	Token string        `json:"token"`
	User  *UserResponse `json:"user"`
}

type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (s *AuthService) Login(username, password string) (*LoginResponse, error) {
	user, err := s.userService.GetByUsername(username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !user.IsActive {
		return nil, errors.New("account disabled")
	}

	if !s.userService.VerifyPassword(user, password) {
		return nil, errors.New("invalid credentials")
	}

	// Auto-generate access key for existing users if not exists
	if user.AccessKey == nil || *user.AccessKey == "" {
		log.Printf("[Auth] User %s has no access key, generating...", user.Username)
		key, generated, err := s.userService.EnsureAccessKey(user.ID)
		if err != nil {
			log.Printf("[Auth] Failed to generate access key for user %s: %v", user.Username, err)
		} else {
			log.Printf("[Auth] Access key for user %s: generated=%v, key_prefix=%s", user.Username, generated, key[:10])
		}
	} else {
		log.Printf("[Auth] User %s already has access key", user.Username)
	}

	token, err := s.jwtService.Generate(user)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token: token,
		User: &UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Role:      string(user.Role),
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt.String(),
			UpdatedAt: user.UpdatedAt.String(),
		},
	}, nil
}

func (s *AuthService) ValidateJWT(token string) (*models.User, error) {
	claims, err := s.jwtService.Validate(token)
	if err != nil {
		return nil, err
	}

	user, err := s.userService.GetByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, errors.New("account disabled")
	}

	return user, nil
}

func (s *AuthService) ValidateAPIKey(rawKey string) (*models.User, error) {
	apiKey, err := s.apiKeyService.Validate(rawKey)
	if err != nil {
		return nil, err
	}

	if apiKey.User == nil {
		return nil, errors.New("user not found")
	}

	if !apiKey.User.IsActive {
		return nil, errors.New("account disabled")
	}

	return apiKey.User, nil
}
