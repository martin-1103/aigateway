package handlers

import (
	"aigateway/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ModelsHandler struct {
	service *services.ModelsService
}

func NewModelsHandler(service *services.ModelsService) *ModelsHandler {
	return &ModelsHandler{service: service}
}

func (h *ModelsHandler) GetModels(c *gin.Context) {
	response, err := h.service.GetAvailableModels(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
