package domain

import (
	"time"

	"github.com/google/uuid"
)

// proverka koordinat polzovatelya
type LocationCheck struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      string    `json:"user_id" db:"user_id"`
	Latitude    float64   `json:"latitude" db:"latitude"`
	Longitude   float64   `json:"longitude" db:"longitude"`
	CheckedAt   time.Time `json:"checked_at" db:"checked_at"`
	WebhookSent bool      `json:"webhook_sent" db:"webhook_sent"`
}

// zapros
type LocationCheckRequest struct {
	UserID    string  `json:"user_id" binding:"required"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

// otvet
type LocationCheckResponse struct {
	HasDanger bool       `json:"has_danger"`
	Incidents []Incident `json:"incidents"`
}

type IncidentStats struct {
	ZoneID    uuid.UUID `json:"zone_id"`
	UserCount int       `json:"user_count"`
}
