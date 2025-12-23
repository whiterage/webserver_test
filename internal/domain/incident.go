// internal/domain/incident.go
package domain

import (
	"time"

	"github.com/google/uuid"
)

type Incident struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	Latitude    float64   `json:"latitude" db:"latitude"`
	Longitude   float64   `json:"longitude" db:"longitude"`
	Radius      float64   `json:"radius" db:"radius"` // радиус в метрах
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CreateIncidentRequest - запрос на создание инцидента
type CreateIncidentRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description string  `json:"description"`
	Latitude    float64 `json:"latitude" binding:"required"`
	Longitude   float64 `json:"longitude" binding:"required"`
	Radius      float64 `json:"radius" binding:"required"`
}

// UpdateIncidentRequest - запрос на обновление инцидента
type UpdateIncidentRequest struct {
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
	Radius      *float64 `json:"radius"`
	IsActive    *bool    `json:"is_active"`
}
