# RBAC Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add Role-Based Access Control with 3 roles (admin, user, provider), JWT + API key authentication, and permission-based endpoint protection.

**Architecture:** New models (User, APIKey) with repositories and services. Middleware layer for auth extraction and permission checking. Modified existing handlers to enforce ownership rules.

**Tech Stack:** Go, GORM, Gin, bcrypt, golang-jwt/jwt/v5, Redis caching

**Design Doc:** `docs/plans/2025-12-26-rbac-design.md`

---

## Phase 1: Foundation (Models & Database)

### Task 1.1: Add Dependencies

**Files:**
- Modify: `go.mod`

**Step 1: Add required packages**

```bash
cd D:/temp/aigateway/worktrees/feature-rbac
go get golang.org/x/crypto/bcrypt
go get github.com/golang-jwt/jwt/v5
go get github.com/google/uuid
```

**Step 2: Verify dependencies**

```bash
go mod tidy
cat go.mod | grep -E "(bcrypt|jwt|uuid)"
```

Expected: 3 new dependencies listed

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add bcrypt, jwt, uuid dependencies"
```

---

### Task 1.2: Create Role Enum

**Files:**
- Create: `models/role.enum.go`

**Step 1: Create role enum file**

```go
// models/role.enum.go
package models

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleUser     Role = "user"
	RoleProvider Role = "provider"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleUser, RoleProvider:
		return true
	}
	return false
}

func (r Role) CanAccessAI() bool {
	return r == RoleAdmin || r == RoleUser
}

func (r Role) CanManageUsers() bool {
	return r == RoleAdmin
}

func (r Role) CanManageAccounts() bool {
	return r == RoleAdmin || r == RoleProvider
}

func (r Role) CanManageAPIKeys() bool {
	return r == RoleAdmin || r == RoleUser
}
```

**Step 2: Verify compiles**

```bash
go build ./models/...
```

Expected: No errors

**Step 3: Commit**

```bash
git add models/role.enum.go
git commit -m "feat(models): add Role enum with permission helpers"
```

---

### Task 1.3: Create User Model

**Files:**
- Create: `models/user.model.go`

**Step 1: Create user model file**

```go
// models/user.model.go
package models

import "time"

type User struct {
	ID           string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	Username     string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	Role         Role      `gorm:"type:varchar(20);not null" json:"role"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
```

**Step 2: Verify compiles**

```bash
go build ./models/...
```

**Step 3: Commit**

```bash
git add models/user.model.go
git commit -m "feat(models): add User model"
```

---

### Task 1.4: Create APIKey Model

**Files:**
- Create: `models/apikey.model.go`

**Step 1: Create apikey model file**

```go
// models/apikey.model.go
package models

import "time"

type APIKey struct {
	ID         string     `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID     string     `gorm:"type:varchar(36);index;not null" json:"user_id"`
	KeyHash    string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"-"`
	KeyPrefix  string     `gorm:"type:varchar(12);not null" json:"key_prefix"`
	Label      string     `gorm:"type:varchar(100)" json:"label"`
	IsActive   bool       `gorm:"default:true" json:"is_active"`
	LastUsedAt *time.Time `json:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (APIKey) TableName() string {
	return "api_keys"
}
```

**Step 2: Verify compiles**

```bash
go build ./models/...
```

**Step 3: Commit**

```bash
git add models/apikey.model.go
git commit -m "feat(models): add APIKey model"
```

---

### Task 1.5: Modify Account Model

**Files:**
- Modify: `models/account.model.go`

**Step 1: Add CreatedBy field to Account struct**

Find the Account struct and add:

```go
CreatedBy *string `gorm:"type:varchar(36);index" json:"created_by,omitempty"`
```

**Step 2: Verify compiles**

```bash
go build ./models/...
```

**Step 3: Commit**

```bash
git add models/account.model.go
git commit -m "feat(models): add CreatedBy field to Account"
```

---

### Task 1.6: Modify ModelMapping Model

**Files:**
- Modify: `models/model.mapping.go`

**Step 1: Add OwnerID field to ModelMapping struct**

Find the ModelMapping struct and add:

```go
OwnerID *string `gorm:"type:varchar(36);index" json:"owner_id,omitempty"`
```

**Step 2: Verify compiles**

```bash
go build ./models/...
```

**Step 3: Commit**

```bash
git add models/model.mapping.go
git commit -m "feat(models): add OwnerID field to ModelMapping"
```

---

### Task 1.7: Update Database AutoMigrate

**Files:**
- Modify: `database/mysql.go`

**Step 1: Add User and APIKey to AutoMigrate call**

Find the AutoMigrate call and add User, APIKey:

```go
db.AutoMigrate(
    // existing models...
    &models.User{},
    &models.APIKey{},
)
```

**Step 2: Verify compiles**

```bash
go build ./database/...
```

**Step 3: Commit**

```bash
git add database/mysql.go
git commit -m "feat(database): add User and APIKey to AutoMigrate"
```

---

### Task 1.8: Create Seed Logic

**Files:**
- Create: `database/seed.go`

**Step 1: Create seed file**

```go
// database/seed.go
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
```

**Step 2: Verify compiles**

```bash
go build ./database/...
```

**Step 3: Commit**

```bash
git add database/seed.go
git commit -m "feat(database): add SeedDefaultAdmin function"
```

---

## Phase 2: Repositories

### Task 2.1: Create User Repository

**Files:**
- Create: `repositories/user.repository.go`

**Step 1: Create user repository**

```go
// repositories/user.repository.go
package repositories

import (
	"aigateway/models"

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
```

**Step 2: Verify compiles**

```bash
go build ./repositories/...
```

**Step 3: Commit**

```bash
git add repositories/user.repository.go
git commit -m "feat(repositories): add UserRepository"
```

---

### Task 2.2: Create APIKey Repository

**Files:**
- Create: `repositories/apikey.repository.go`

**Step 1: Create apikey repository**

```go
// repositories/apikey.repository.go
package repositories

import (
	"time"

	"aigateway/models"

	"gorm.io/gorm"
)

type APIKeyRepository struct {
	db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) Create(key *models.APIKey) error {
	return r.db.Create(key).Error
}

func (r *APIKeyRepository) GetByID(id string) (*models.APIKey, error) {
	var key models.APIKey
	err := r.db.Preload("User").Where("id = ?", id).First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *APIKeyRepository) GetByHash(hash string) (*models.APIKey, error) {
	var key models.APIKey
	err := r.db.Preload("User").Where("key_hash = ? AND is_active = ?", hash, true).First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *APIKeyRepository) ListByUserID(userID string) ([]*models.APIKey, error) {
	var keys []*models.APIKey
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&keys).Error
	return keys, err
}

func (r *APIKeyRepository) ListAll(limit, offset int) ([]*models.APIKey, int64, error) {
	var keys []*models.APIKey
	var total int64

	r.db.Model(&models.APIKey{}).Count(&total)
	err := r.db.Preload("User").Limit(limit).Offset(offset).Order("created_at DESC").Find(&keys).Error

	return keys, total, err
}

func (r *APIKeyRepository) UpdateLastUsed(id string) error {
	now := time.Now()
	return r.db.Model(&models.APIKey{}).Where("id = ?", id).Update("last_used_at", &now).Error
}

func (r *APIKeyRepository) Revoke(id string) error {
	return r.db.Model(&models.APIKey{}).Where("id = ?", id).Update("is_active", false).Error
}

func (r *APIKeyRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.APIKey{}).Error
}
```

**Step 2: Verify compiles**

```bash
go build ./repositories/...
```

**Step 3: Commit**

```bash
git add repositories/apikey.repository.go
git commit -m "feat(repositories): add APIKeyRepository"
```

---

## Phase 3: Services

### Task 3.1: Create Password Service

**Files:**
- Create: `services/password.service.go`

**Step 1: Create password service**

```go
// services/password.service.go
package services

import "golang.org/x/crypto/bcrypt"

type PasswordService struct {
	cost int
}

func NewPasswordService() *PasswordService {
	return &PasswordService{cost: 10}
}

func (s *PasswordService) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s *PasswordService) Verify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
```

**Step 2: Verify compiles**

```bash
go build ./services/...
```

**Step 3: Commit**

```bash
git add services/password.service.go
git commit -m "feat(services): add PasswordService"
```

---

### Task 3.2: Create JWT Service

**Files:**
- Create: `services/auth.jwt.service.go`

**Step 1: Create JWT service**

```go
// services/auth.jwt.service.go
package services

import (
	"errors"
	"time"

	"aigateway/models"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID   string      `json:"user_id"`
	Username string      `json:"username"`
	Role     models.Role `json:"role"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secret     []byte
	expiration time.Duration
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{
		secret:     []byte(secret),
		expiration: 24 * time.Hour,
	}
}

func (s *JWTService) Generate(user *models.User) (string, error) {
	claims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *JWTService) Validate(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (s *JWTService) ExpiresIn() int {
	return int(s.expiration.Seconds())
}
```

**Step 2: Verify compiles**

```bash
go build ./services/...
```

**Step 3: Commit**

```bash
git add services/auth.jwt.service.go
git commit -m "feat(services): add JWTService"
```

---

### Task 3.3: Create APIKey Service

**Files:**
- Create: `services/auth.apikey.service.go`

**Step 1: Create APIKey service**

```go
// services/auth.apikey.service.go
package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"aigateway/models"
	"aigateway/repositories"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type APIKeyService struct {
	repo  *repositories.APIKeyRepository
	redis *redis.Client
}

func NewAPIKeyService(repo *repositories.APIKeyRepository, redis *redis.Client) *APIKeyService {
	return &APIKeyService{repo: repo, redis: redis}
}

func (s *APIKeyService) Generate(userID, label string) (*models.APIKey, string, error) {
	rawKey := s.generateRawKey()
	hash := s.hashKey(rawKey)
	prefix := rawKey[:12]

	apiKey := &models.APIKey{
		ID:        uuid.New().String(),
		UserID:    userID,
		KeyHash:   hash,
		KeyPrefix: prefix,
		Label:     label,
		IsActive:  true,
	}

	if err := s.repo.Create(apiKey); err != nil {
		return nil, "", err
	}

	return apiKey, rawKey, nil
}

func (s *APIKeyService) Validate(rawKey string) (*models.APIKey, error) {
	hash := s.hashKey(rawKey)

	// Check cache first
	ctx := context.Background()
	cacheKey := fmt.Sprintf("apikey:%s", hash)

	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var key models.APIKey
		if json.Unmarshal([]byte(cached), &key) == nil {
			go s.repo.UpdateLastUsed(key.ID)
			return &key, nil
		}
	}

	// Lookup in DB
	key, err := s.repo.GetByHash(hash)
	if err != nil {
		return nil, err
	}

	// Cache for 5 minutes
	data, _ := json.Marshal(key)
	s.redis.Set(ctx, cacheKey, data, 5*time.Minute)

	go s.repo.UpdateLastUsed(key.ID)

	return key, nil
}

func (s *APIKeyService) ListByUser(userID string) ([]*models.APIKey, error) {
	return s.repo.ListByUserID(userID)
}

func (s *APIKeyService) ListAll(limit, offset int) ([]*models.APIKey, int64, error) {
	return s.repo.ListAll(limit, offset)
}

func (s *APIKeyService) Revoke(id string) error {
	return s.repo.Revoke(id)
}

func (s *APIKeyService) GetByID(id string) (*models.APIKey, error) {
	return s.repo.GetByID(id)
}

func (s *APIKeyService) generateRawKey() string {
	bytes := make([]byte, 24)
	rand.Read(bytes)
	return "ak_" + hex.EncodeToString(bytes)
}

func (s *APIKeyService) hashKey(rawKey string) string {
	hash := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(hash[:])
}
```

**Step 2: Verify compiles**

```bash
go build ./services/...
```

**Step 3: Commit**

```bash
git add services/auth.apikey.service.go
git commit -m "feat(services): add APIKeyService with Redis caching"
```

---

### Task 3.4: Create User Service

**Files:**
- Create: `services/user.service.go`

**Step 1: Create user service**

```go
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
```

**Step 2: Verify compiles**

```bash
go build ./services/...
```

**Step 3: Commit**

```bash
git add services/user.service.go
git commit -m "feat(services): add UserService"
```

---

### Task 3.5: Create Auth Service

**Files:**
- Create: `services/auth.service.go`

**Step 1: Create auth service**

```go
// services/auth.service.go
package services

import (
	"errors"

	"aigateway/models"
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
```

**Step 2: Verify compiles**

```bash
go build ./services/...
```

**Step 3: Commit**

```bash
git add services/auth.service.go
git commit -m "feat(services): add AuthService"
```

---

## Phase 4: Middleware

### Task 4.1: Create Auth Context Helper

**Files:**
- Create: `middleware/auth.context.go`

**Step 1: Create context helper**

```go
// middleware/auth.context.go
package middleware

import (
	"aigateway/models"

	"github.com/gin-gonic/gin"
)

const UserContextKey = "current_user"

func SetCurrentUser(c *gin.Context, user *models.User) {
	c.Set(UserContextKey, user)
}

func GetCurrentUser(c *gin.Context) *models.User {
	val, exists := c.Get(UserContextKey)
	if !exists {
		return nil
	}
	user, ok := val.(*models.User)
	if !ok {
		return nil
	}
	return user
}

func GetCurrentUserID(c *gin.Context) string {
	user := GetCurrentUser(c)
	if user == nil {
		return ""
	}
	return user.ID
}

func GetCurrentRole(c *gin.Context) models.Role {
	user := GetCurrentUser(c)
	if user == nil {
		return ""
	}
	return user.Role
}
```

**Step 2: Verify compiles**

```bash
go build ./middleware/...
```

**Step 3: Commit**

```bash
git add middleware/auth.context.go
git commit -m "feat(middleware): add auth context helpers"
```

---

### Task 4.2: Create Auth Middleware

**Files:**
- Create: `middleware/auth.middleware.go`

**Step 1: Create auth middleware**

```go
// middleware/auth.middleware.go
package middleware

import (
	"strings"

	"aigateway/services"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	authService *services.AuthService
}

func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

func (m *AuthMiddleware) ExtractAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try X-API-Key header first
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" {
			user, err := m.authService.ValidateAPIKey(apiKey)
			if err == nil {
				SetCurrentUser(c, user)
			}
			c.Next()
			return
		}

		// Try Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 {
			c.Next()
			return
		}

		scheme := strings.ToLower(parts[0])
		token := parts[1]

		switch scheme {
		case "bearer":
			// Could be JWT or API key
			if strings.HasPrefix(token, "ak_") {
				user, err := m.authService.ValidateAPIKey(token)
				if err == nil {
					SetCurrentUser(c, user)
				}
			} else {
				user, err := m.authService.ValidateJWT(token)
				if err == nil {
					SetCurrentUser(c, user)
				}
			}
		}

		c.Next()
	}
}
```

**Step 2: Verify compiles**

```bash
go build ./middleware/...
```

**Step 3: Commit**

```bash
git add middleware/auth.middleware.go
git commit -m "feat(middleware): add auth extraction middleware"
```

---

### Task 4.3: Create Require Middleware

**Files:**
- Create: `middleware/auth.require.middleware.go`

**Step 1: Create require middleware**

```go
// middleware/auth.require.middleware.go
package middleware

import (
	"net/http"

	"aigateway/models"

	"github.com/gin-gonic/gin"
)

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
			})
			return
		}
		c.Next()
	}
}

func RequireRole(roles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
			})
			return
		}

		for _, role := range roles {
			if user.Role == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "insufficient permissions",
		})
	}
}

func RequireAdmin() gin.HandlerFunc {
	return RequireRole(models.RoleAdmin)
}

func RequireAIAccess() gin.HandlerFunc {
	return RequireRole(models.RoleAdmin, models.RoleUser)
}

func RequireAccountAccess() gin.HandlerFunc {
	return RequireRole(models.RoleAdmin, models.RoleProvider)
}
```

**Step 2: Verify compiles**

```bash
go build ./middleware/...
```

**Step 3: Commit**

```bash
git add middleware/auth.require.middleware.go
git commit -m "feat(middleware): add role requirement middleware"
```

---

## Phase 5: Handlers

### Task 5.1: Create Auth Handler

**Files:**
- Create: `handlers/auth.handler.go`

**Step 1: Create auth handler**

```go
// handlers/auth.handler.go
package handlers

import (
	"net/http"

	"aigateway/middleware"
	"aigateway/services"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
	userService *services.UserService
}

func NewAuthHandler(auth *services.AuthService, user *services.UserService) *AuthHandler {
	return &AuthHandler{authService: auth, userService: user}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Me(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"role":     user.Role,
	})
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !h.userService.VerifyPassword(user, req.CurrentPassword) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "current password incorrect"})
		return
	}

	if err := h.userService.ChangePassword(user.ID, req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password changed"})
}
```

**Step 2: Verify compiles**

```bash
go build ./handlers/...
```

**Step 3: Commit**

```bash
git add handlers/auth.handler.go
git commit -m "feat(handlers): add AuthHandler for login and password"
```

---

### Task 5.2: Create User Handler

**Files:**
- Create: `handlers/user.handler.go`

**Step 1: Create user handler**

```go
// handlers/user.handler.go
package handlers

import (
	"net/http"
	"strconv"

	"aigateway/models"
	"aigateway/services"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	users, total, err := h.userService.List(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  users,
		"total": total,
	})
}

func (h *UserHandler) Get(c *gin.Context) {
	id := c.Param("id")

	user, err := h.userService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

type CreateUserRequest struct {
	Username string      `json:"username" binding:"required"`
	Password string      `json:"password" binding:"required,min=6"`
	Role     models.Role `json:"role" binding:"required"`
}

func (h *UserHandler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.Create(req.Username, req.Password, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

type UpdateUserRequest struct {
	Role     *models.Role `json:"role"`
	IsActive *bool        `json:"is_active"`
}

func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")

	user, err := h.userService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Role != nil && req.Role.IsValid() {
		user.Role = *req.Role
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := h.userService.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.userService.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}
```

**Step 2: Verify compiles**

```bash
go build ./handlers/...
```

**Step 3: Commit**

```bash
git add handlers/user.handler.go
git commit -m "feat(handlers): add UserHandler for CRUD"
```

---

### Task 5.3: Create APIKey Handler

**Files:**
- Create: `handlers/apikey.handler.go`

**Step 1: Create apikey handler**

```go
// handlers/apikey.handler.go
package handlers

import (
	"net/http"
	"strconv"

	"aigateway/middleware"
	"aigateway/models"
	"aigateway/services"

	"github.com/gin-gonic/gin"
)

type APIKeyHandler struct {
	apiKeyService *services.APIKeyService
}

func NewAPIKeyHandler(apiKeyService *services.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{apiKeyService: apiKeyService}
}

func (h *APIKeyHandler) List(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	// Admin can filter by user_id
	if user.Role == models.RoleAdmin {
		userID := c.Query("user_id")
		if userID != "" {
			keys, err := h.apiKeyService.ListByUser(userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": keys})
			return
		}

		// List all
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		keys, total, err := h.apiKeyService.ListAll(limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": keys, "total": total})
		return
	}

	// User can only list own keys
	keys, err := h.apiKeyService.ListByUser(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": keys})
}

type CreateAPIKeyRequest struct {
	Label string `json:"label"`
}

func (h *APIKeyHandler) Create(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var req CreateAPIKeyRequest
	c.ShouldBindJSON(&req)

	apiKey, rawKey, err := h.apiKeyService.Generate(user.ID, req.Label)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         apiKey.ID,
		"key":        rawKey,
		"key_prefix": apiKey.KeyPrefix,
		"label":      apiKey.Label,
		"message":    "Save this key - it will not be shown again",
	})
}

func (h *APIKeyHandler) Revoke(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	id := c.Param("id")

	// Check ownership unless admin
	apiKey, err := h.apiKeyService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "api key not found"})
		return
	}

	if user.Role != models.RoleAdmin && apiKey.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not your api key"})
		return
	}

	if err := h.apiKeyService.Revoke(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "api key revoked"})
}
```

**Step 2: Verify compiles**

```bash
go build ./handlers/...
```

**Step 3: Commit**

```bash
git add handlers/apikey.handler.go
git commit -m "feat(handlers): add APIKeyHandler"
```

---

## Phase 6: Route Integration

### Task 6.1: Update Routes

**Files:**
- Modify: `routes/routes.go`

**Step 1: Read current routes file**

Read the file to understand current structure.

**Step 2: Add imports and new routes**

Add to imports:
```go
"aigateway/middleware"
```

Add middleware initialization and new route groups after existing routes. The exact integration depends on current file structure - add:

```go
// Auth middleware (add after handler initialization)
authMiddleware := middleware.NewAuthMiddleware(authService)

// Apply global auth extraction
r.Use(authMiddleware.ExtractAuth())

// Auth routes (public)
auth := r.Group("/api/v1/auth")
{
    auth.POST("/login", authHandler.Login)
    auth.GET("/me", middleware.RequireAuth(), authHandler.Me)
    auth.PUT("/password", middleware.RequireAuth(), authHandler.ChangePassword)
}

// User routes (admin only)
users := r.Group("/api/v1/users")
users.Use(middleware.RequireAdmin())
{
    users.GET("", userHandler.List)
    users.POST("", userHandler.Create)
    users.GET("/:id", userHandler.Get)
    users.PUT("/:id", userHandler.Update)
    users.DELETE("/:id", userHandler.Delete)
}

// API Key routes (admin + user)
apiKeys := r.Group("/api/v1/api-keys")
apiKeys.Use(middleware.RequireAuth(), middleware.RequireRole(models.RoleAdmin, models.RoleUser))
{
    apiKeys.GET("", apiKeyHandler.List)
    apiKeys.POST("", apiKeyHandler.Create)
    apiKeys.DELETE("/:id", apiKeyHandler.Revoke)
}

// Protect AI proxy routes
r.POST("/v1/messages", middleware.RequireAIAccess(), proxyHandler.HandleProxy)
r.POST("/v1/chat/completions", middleware.RequireAIAccess(), proxyHandler.HandleProxy)

// Protect account routes
accounts := r.Group("/api/v1/accounts")
accounts.Use(middleware.RequireAccountAccess())
```

**Step 3: Verify compiles**

```bash
go build ./...
```

**Step 4: Commit**

```bash
git add routes/routes.go
git commit -m "feat(routes): add auth routes and middleware"
```

---

### Task 6.2: Update Main to Seed Admin

**Files:**
- Modify: `cmd/main.go`

**Step 1: Add seed call after AutoMigrate**

Find where database is initialized and add after AutoMigrate:

```go
// After db connection and AutoMigrate
if err := database.SeedDefaultAdmin(db); err != nil {
    log.Fatalf("Failed to seed admin: %v", err)
}
```

**Step 2: Verify compiles**

```bash
go build ./cmd/...
```

**Step 3: Commit**

```bash
git add cmd/main.go
git commit -m "feat(main): add auto-seed for default admin"
```

---

## Phase 7: Modify Existing Handlers

### Task 7.1: Add Ownership to Account Handler

**Files:**
- Modify: `handlers/account.handler.go`

**Step 1: Modify Create to set CreatedBy**

In the Create handler, add:
```go
user := middleware.GetCurrentUser(c)
if user != nil {
    account.CreatedBy = &user.ID
}
```

**Step 2: Modify List to filter by ownership for providers**

```go
user := middleware.GetCurrentUser(c)
if user != nil && user.Role == models.RoleProvider {
    // Only show accounts created by this provider
    // Add filter: WHERE created_by = user.ID
}
```

**Step 3: Verify compiles**

```bash
go build ./handlers/...
```

**Step 4: Commit**

```bash
git add handlers/account.handler.go
git commit -m "feat(handlers): add ownership filter to AccountHandler"
```

---

### Task 7.2: Add Ownership to ModelMapping Handler

**Files:**
- Modify: `handlers/model.mapping.handler.go`

**Step 1: Modify Create to set OwnerID**

For non-admin users, set OwnerID:
```go
user := middleware.GetCurrentUser(c)
if user != nil && user.Role != models.RoleAdmin {
    mapping.OwnerID = &user.ID
}
```

**Step 2: Modify List to return global + owned**

```go
user := middleware.GetCurrentUser(c)
// Return: WHERE owner_id IS NULL OR owner_id = user.ID
```

**Step 3: Modify Update/Delete to check ownership**

```go
// Only allow if admin OR owner_id = user.ID
```

**Step 4: Verify compiles**

```bash
go build ./handlers/...
```

**Step 5: Commit**

```bash
git add handlers/model.mapping.handler.go
git commit -m "feat(handlers): add ownership to ModelMappingHandler"
```

---

## Phase 8: Final Integration & Testing

### Task 8.1: Full Build Test

**Step 1: Build entire project**

```bash
cd D:/temp/aigateway/worktrees/feature-rbac
go build ./...
```

Expected: No errors

**Step 2: Run any existing tests**

```bash
go test ./...
```

**Step 3: Commit any fixes**

---

### Task 8.2: Manual Testing Checklist

**Step 1: Start server**

```bash
go run cmd/main.go
```

Expected: "Default admin created" message on first run

**Step 2: Test login**

```bash
curl -X POST http://localhost:8088/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

Expected: JWT token response

**Step 3: Test create user**

```bash
curl -X POST http://localhost:8088/api/v1/users \
  -H "Authorization: Bearer <JWT>" \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"test123","role":"user"}'
```

Expected: User created

**Step 4: Test create API key**

```bash
curl -X POST http://localhost:8088/api/v1/api-keys \
  -H "Authorization: Bearer <JWT>" \
  -H "Content-Type: application/json" \
  -d '{"label":"test"}'
```

Expected: API key returned (save it!)

**Step 5: Test AI proxy with API key**

```bash
curl -X POST http://localhost:8088/v1/messages \
  -H "X-API-Key: ak_..." \
  -H "Content-Type: application/json" \
  -d '{"model":"claude-sonnet-4-20250514","messages":[{"role":"user","content":"Hi"}]}'
```

Expected: Should work with valid API key, 401 without

---

### Task 8.3: Final Commit & Merge Prep

**Step 1: Review all changes**

```bash
git log --oneline
git diff master
```

**Step 2: Ensure all tests pass**

```bash
go test ./...
go build ./...
```

**Step 3: Ready for merge**

Use skill: superpowers:finishing-a-development-branch

---

## Summary

| Phase | Tasks | Description |
|-------|-------|-------------|
| 1 | 1.1-1.8 | Models & Database foundation |
| 2 | 2.1-2.2 | Repositories |
| 3 | 3.1-3.5 | Services |
| 4 | 4.1-4.3 | Middleware |
| 5 | 5.1-5.3 | New Handlers |
| 6 | 6.1-6.2 | Route Integration |
| 7 | 7.1-7.2 | Modify Existing Handlers |
| 8 | 8.1-8.3 | Integration & Testing |

**Total: ~25 tasks, each 2-5 minutes**
