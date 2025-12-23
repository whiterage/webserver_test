package repository

import (
	"context"
	"database/sql"
	"fmt"
	"geo-alert-core/internal/domain"
	"time"

	"github.com/google/uuid"
)

// proverka koordinat polzovatelya
type LocationCheckRepository interface {
	Create(ctx context.Context, check *domain.LocationCheck) error
	LinkToIncidents(ctx context.Context, checkID uuid.UUID, incidentIDs []uuid.UUID) error
}

// realization for postgres
type postgresLocationCheckRepository struct {
	db *sql.DB
}

// new realization for postgres
func NewPostgresLocationCheckRepository(db *sql.DB) LocationCheckRepository {
	return &postgresLocationCheckRepository{db: db}
}

func (r *postgresLocationCheckRepository) Create(ctx context.Context, check *domain.LocationCheck) error {
	query := `
		INSERT INTO location_checks (id, user_id, latitude, longitude, checked_at, webhook_sent)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	check.ID = uuid.New()
	check.CheckedAt = time.Now()
	check.WebhookSent = false

	_, err := r.db.ExecContext(ctx, query,
		check.ID,
		check.UserID,
		check.Latitude,
		check.Longitude,
		check.CheckedAt,
		check.WebhookSent,
	)

	if err != nil {
		return fmt.Errorf("failed to create location check: %w", err)
	}

	return nil
}

func (r *postgresLocationCheckRepository) LinkToIncidents(ctx context.Context, checkID uuid.UUID, incidentIDs []uuid.UUID) error {
	if len(incidentIDs) == 0 {
		return nil // no incidents to link
	}

	// use transaction for atomicity
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO location_check_incidents (location_check_id, incident_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, incidentID := range incidentIDs {
		_, err := stmt.ExecContext(ctx, checkID, incidentID)
		if err != nil {
			return fmt.Errorf("failed to link incident: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
