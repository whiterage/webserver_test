package repository

import (
	"context"
	"database/sql"
	"fmt"
	"geo-alert-core/internal/domain"
	"time"

	"github.com/google/uuid"
)

// interface dlya raboty s incidetami
type IncidentRepository interface {
	Create(ctx context.Context, incident *domain.Incident) error

	// get by id
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Incident, error)

	// get all
	GetAll(ctx context.Context) ([]*domain.Incident, error)

	// only active
	GetActiveIncidents(ctx context.Context) ([]*domain.Incident, error)

	// update
	Update(ctx context.Context, incident *domain.Incident) error

	// delete
	Delete(ctx context.Context, id uuid.UUID) error

	// find locations
	FindNearbyIncidents(ctx context.Context, latitude, longitude float64, radius float64) ([]*domain.Incident, error)

	// statistics
	GetStats(ctx context.Context, zoneID uuid.UUID) (*domain.IncidentStats, error)
}

type postgresIncidentRepository struct {
	db *sql.DB
}

// new realization of incedent repository
func NewPostgresIncidentRepository(db *sql.DB) IncidentRepository {
	return &postgresIncidentRepository{db: db}
}

func (r *postgresIncidentRepository) Create(ctx context.Context, incident *domain.Incident) error {
	query := `
		INSERT INTO incidents (id, title, description, latitude, longitude, radius, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	now := time.Now()
	incident.ID = uuid.New()
	incident.CreatedAt = now
	incident.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		incident.ID,
		incident.Title,
		incident.Description,
		incident.Latitude,
		incident.Longitude,
		incident.Radius,
		incident.IsActive,
		incident.CreatedAt,
		incident.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create incident: %w", err)
	}

	return nil
}

func (r *postgresIncidentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Incident, error) {
	query := `
		SELECT id, title, description, latitude, longitude, radius, is_active, created_at, updated_at
		FROM incidents
		WHERE id = $1
	`

	var incident domain.Incident
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&incident.ID,
		&incident.Title,
		&incident.Description,
		&incident.Latitude,
		&incident.Longitude,
		&incident.Radius,
		&incident.IsActive,
		&incident.CreatedAt,
		&incident.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("incident not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get incident: %w", err)
	}

	return &incident, nil
}

func (r *postgresIncidentRepository) GetAll(ctx context.Context, limit, offset int) ([]*domain.Incident, error) {
	query := `
		SELECT id, title, description, latitude, longitude, radius, is_active, created_at, updated_at
		FROM incidents
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get incidents: %w", err)
	}
	defer rows.Close()

	var incidents []*domain.Incident
	for rows.Next() {
		var incident domain.Incident
		err := rows.Scan(
			&incident.ID,
			&incident.Title,
			&incident.Description,
			&incident.Latitude,
			&incident.Longitude,
			&incident.Radius,
			&incident.IsActive,
			&incident.CreatedAt,
			&incident.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan incident: %w", err)
		}
		incidents = append(incidents, &incident)
	}

	return incidents, nil
}

func (r *postgresIncidentRepository) GetActiveIncidents(ctx context.Context) ([]*domain.Incident, error) {
	query := `
		SELECT id, title, description, latitude, longitude, radius, is_active, created_at, updated_at
		FROM incidents
		WHERE is_active = true
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active incidents: %w", err)
	}
	defer rows.Close()

	var incidents []*domain.Incident
	for rows.Next() {
		var incident domain.Incident
		err := rows.Scan(
			&incident.ID,
			&incident.Title,
			&incident.Description,
			&incident.Latitude,
			&incident.Longitude,
			&incident.Radius,
			&incident.IsActive,
			&incident.CreatedAt,
			&incident.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan incident: %w", err)
		}
		incidents = append(incidents, &incident)
	}

	return incidents, nil
}

func (r *postgresIncidentRepository) Update(ctx context.Context, id uuid.UUID, incident *domain.Incident) error {
	query := `
		UPDATE incidents
		SET title = $1, description = $2, latitude = $3, longitude = $4, 
		    radius = $5, is_active = $6, updated_at = $7
		WHERE id = $8
	`

	incident.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		incident.Title,
		incident.Description,
		incident.Latitude,
		incident.Longitude,
		incident.Radius,
		incident.IsActive,
		incident.UpdatedAt,
		id,
	)

	if err != nil {
		return fmt.Errorf("failed to update incident: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("incident not found")
	}

	return nil
}

func (r *postgresIncidentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Soft delete
	query := `UPDATE incidents SET is_active = false, updated_at = $1 WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete incident: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("incident not found")
	}

	return nil
}

func (r *postgresIncidentRepository) FindNearbyIncidents(ctx context.Context, latitude, longitude float64) ([]*domain.Incident, error) {
	query := `
		SELECT id, title, description, latitude, longitude, radius, is_active, created_at, updated_at
		FROM incidents
		WHERE is_active = true
		AND ST_DWithin(
			ST_MakePoint(longitude, latitude)::geography,
			ST_MakePoint($1, $2)::geography,
			radius
		)
	`

	rows, err := r.db.QueryContext(ctx, query, longitude, latitude)
	if err != nil {
		return nil, fmt.Errorf("failed to find nearby incidents: %w", err)
	}
	defer rows.Close()

	var incidents []*domain.Incident
	for rows.Next() {
		var incident domain.Incident
		err := rows.Scan(
			&incident.ID,
			&incident.Title,
			&incident.Description,
			&incident.Latitude,
			&incident.Longitude,
			&incident.Radius,
			&incident.IsActive,
			&incident.CreatedAt,
			&incident.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan incident: %w", err)
		}
		incidents = append(incidents, &incident)
	}

	return incidents, nil
}

func (r *postgresIncidentRepository) GetStats(ctx context.Context, minutes int) ([]*domain.IncidentStats, error) {
	query := `
		SELECT 
			i.id as zone_id,
			COUNT(DISTINCT lc.user_id) as user_count
		FROM incidents i
		LEFT JOIN location_check_incidents lci ON i.id = lci.incident_id
		LEFT JOIN location_checks lc ON lci.location_check_id = lc.id
			AND lc.checked_at >= NOW() - INTERVAL '%d minutes'
		WHERE i.is_active = true
		GROUP BY i.id
		ORDER BY user_count DESC
	`

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(query, minutes))
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	defer rows.Close()

	var stats []*domain.IncidentStats
	for rows.Next() {
		var stat domain.IncidentStats
		err := rows.Scan(&stat.ZoneID, &stat.UserCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stat: %w", err)
		}
		stats = append(stats, &stat)
	}

	return stats, nil
}
