package handler

import (
	"geo-alert-core/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// handler for stats
type StatsHandler struct {
	service *service.StatsService
}

func NewStatsHandler(statsService *service.StatsService) *StatsHandler {
	return &StatsHandler{
		service: statsService,
	}
}

func (h *StatsHandler) GetStats(c *gin.Context) {
	minutesStr := c.Query("minutes")
	minutes := 0
	if minutesStr != "" {
		var err error
		minutes, err = strconv.Atoi(minutesStr)
		if err != nil || minutes <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid minutes parameter",
			})
			return
		}
	}

	stats, err := h.service.GetStats(c.Request.Context(), minutes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get stats",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}
