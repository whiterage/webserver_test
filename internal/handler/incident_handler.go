package handler

import (
	"errors"
	"geo-alert-core/internal/domain"
	"geo-alert-core/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// handler for incidents
type IncidentHandler struct {
	service *service.IncidentService
}

func NewIncidentHandler(incidentService *service.IncidentService) *IncidentHandler {
	return &IncidentHandler{
		service: incidentService,
	}
}

// create new incident
// POST /api/v1/incidents
func (h *IncidentHandler) Create(c *gin.Context) {
	var req domain.CreateIncidentRequest

	// validation of input data (Gin automatically checks binding tags)
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	incident, err := h.service.CreateIncident(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCoordinates) || errors.Is(err, domain.ErrInvalidRadius) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation failed",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create incident",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, incident)
}

// get incident by id
// GET /api/v1/incidents/:id
func (h *IncidentHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid incident ID",
		})
		return
	}

	incident, err := h.service.GetIncident(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrIncidentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Incident not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get incident",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, incident)
}

func (h *IncidentHandler) GetAll(c *gin.Context) {
	// parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	incidents, err := h.service.GetAllIncidents(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get incidents",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      incidents,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *IncidentHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid incident ID",
		})
		return
	}

	var req domain.UpdateIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	incident, err := h.service.UpdateIncident(c.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, domain.ErrIncidentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Incident not found",
			})
			return
		}

		if errors.Is(err, domain.ErrInvalidCoordinates) || errors.Is(err, domain.ErrInvalidRadius) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation failed",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update incident",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, incident)
}

func (h *IncidentHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid incident ID",
		})
		return
	}

	if err := h.service.DeleteIncident(c.Request.Context(), id); err != nil {
		if errors.Is(err, domain.ErrIncidentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Incident not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete incident",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Incident deleted successfully",
	})
}
