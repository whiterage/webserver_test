// internal/service/incident_service.go
package service

import (
	"context"
	"fmt"
	"geo-alert-core/internal/domain"
	"geo-alert-core/internal/repository"

	"github.com/google/uuid"
)

// business logic for incidents
type IncidentService struct {
	repo repository.IncidentRepository
}

func NewIncidentService(repo repository.IncidentRepository) *IncidentService {
	return &IncidentService{repo: repo}
}

func (s *IncidentService) CreateIncident(ctx context.Context, req *domain.CreateIncidentRequest) (*domain.Incident, error) {
	// validation of coordinates
	if err := s.validateCoordinates(req.Latitude, req.Longitude); err != nil {
		return nil, err
	}

	// validation of radius
	if req.Radius <= 0 {
		return nil, fmt.Errorf("radius must be positive")
	}

	incident := &domain.Incident{
		Title:       req.Title,
		Description: req.Description,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Radius:      req.Radius,
		IsActive:    true,
	}

	if err := s.repo.Create(ctx, incident); err != nil {
		return nil, fmt.Errorf("failed to create incident: %w", err)
	}

	return incident, nil
}

// get incident by id
func (s *IncidentService) GetIncident(ctx context.Context, id uuid.UUID) (*domain.Incident, error) {
	return s.repo.GetByID(ctx, id)
}

// get all incidents with pagination
func (s *IncidentService) GetAllIncidents(ctx context.Context, page, pageSize int) ([]*domain.Incident, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20 // дефолтное значение
	}

	offset := (page - 1) * pageSize
	return s.repo.GetAll(ctx, pageSize, offset)
}

func (s *IncidentService) UpdateIncident(ctx context.Context, id uuid.UUID, req *domain.UpdateIncidentRequest) (*domain.Incident, error) {
	incident, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// update only passed fields
	if req.Title != nil {
		incident.Title = *req.Title
	}
	if req.Description != nil {
		incident.Description = *req.Description
	}
	if req.Latitude != nil {
		incident.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		incident.Longitude = *req.Longitude
	}
	if req.Radius != nil {
		if *req.Radius <= 0 {
			return nil, fmt.Errorf("radius must be positive")
		}
		incident.Radius = *req.Radius
	}
	if req.IsActive != nil {
		incident.IsActive = *req.IsActive
	}

	// validation of coordinates if they changed
	if req.Latitude != nil || req.Longitude != nil {
		if err := s.validateCoordinates(incident.Latitude, incident.Longitude); err != nil {
			return nil, err
		}
	}

	if err := s.repo.Update(ctx, id, incident); err != nil {
		return nil, fmt.Errorf("failed to update incident: %w", err)
	}

	return incident, nil
}

func (s *IncidentService) DeleteIncident(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// proverka validnosti koordinat
func (s *IncidentService) validateCoordinates(lat, lon float64) error {
	if lat < -90 || lat > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}
	if lon < -180 || lon > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}
	return nil
}
