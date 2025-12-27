// services/auth.service.go
package services

import (
	"errors"

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
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
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

	token, err := s.jwtService.Generate(user)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token:     token,
		ExpiresIn: s.jwtService.ExpiresIn(),
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
