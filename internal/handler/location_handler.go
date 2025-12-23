package handler

import (
	"geo-alert-core/internal/domain"
	"geo-alert-core/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// handler for checking coordinates
type LocationHandler struct {
	service *service.LocationService
}

func NewLocationHandler(locationService *service.LocationService) *LocationHandler {
	return &LocationHandler{
		service: locationService,
	}
}

func (h *LocationHandler) CheckLocation(c *gin.Context) {
	var req domain.LocationCheckRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	response, err := h.service.CheckLocation(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check location",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
