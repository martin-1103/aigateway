package handlers

import (
	"aigateway/middleware"
	"aigateway/models"
	"aigateway/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AccountHandler struct {
	service *services.AccountService
}

func NewAccountHandler(service *services.AccountService) *AccountHandler {
	return &AccountHandler{service: service}
}

func (h *AccountHandler) List(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var accounts []*models.Account
	var total int64
	var err error

	// Provider can only see accounts they created
	if user != nil && user.Role == models.RoleProvider {
		accounts, total, err = h.service.ListByCreator(user.ID, limit, offset)
	} else {
		accounts, total, err = h.service.List(limit, offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   accounts,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *AccountHandler) Get(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	id := c.Param("id")

	account, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}

	// Provider can only see accounts they created
	if user != nil && user.Role == models.RoleProvider {
		if account.CreatedBy == nil || *account.CreatedBy != user.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	}

	c.JSON(http.StatusOK, account)
}

func (h *AccountHandler) Create(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	var account models.Account
	if err := c.ShouldBindJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account.ID = uuid.New().String()

	// Set creator
	if user != nil {
		account.CreatedBy = &user.ID
	}

	if err := h.service.Create(&account); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, account)
}

func (h *AccountHandler) Update(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	id := c.Param("id")

	// Get existing account to check ownership
	existing, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}

	// Provider can only update accounts they created
	if user != nil && user.Role == models.RoleProvider {
		if existing.CreatedBy == nil || *existing.CreatedBy != user.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	}

	var account models.Account
	if err := c.ShouldBindJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account.ID = id
	account.CreatedBy = existing.CreatedBy // Preserve creator

	if err := h.service.Update(&account); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, account)
}

func (h *AccountHandler) Delete(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	id := c.Param("id")

	// Only admin can delete (provider cannot delete per design)
	if user != nil && user.Role != models.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admin can delete accounts"})
		return
	}

	if err := h.service.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account deleted"})
}
