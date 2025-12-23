package service

import (
	"context"
	"fmt"
	"geo-alert-core/internal/domain"
	"geo-alert-core/internal/repository"

	"github.com/google/uuid"
)

type IncidentService struct {
	repo            repository.IncidentRepository
	locationService *LocationService // Для инвалидации кэша
}

func NewIncidentService(repo repository.IncidentRepository) *IncidentService {
	return &IncidentService{repo: repo}
}

func (s *IncidentService) SetLocationService(locationService *LocationService) {
	s.locationService = locationService
}

func (s *IncidentService) CreateIncident(ctx context.Context, req *domain.CreateIncidentRequest) (*domain.Incident, error) {
	if err := s.validateCoordinates(req.Latitude, req.Longitude); err != nil {
		return nil, err
	}

	if req.Radius <= 0 {
		return nil, domain.ErrInvalidRadius
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

	// Инвалидируем кэш активных инцидентов
	if s.locationService != nil {
		_ = s.locationService.InvalidateCache(ctx)
	}

	return incident, nil
}

func (s *IncidentService) GetIncident(ctx context.Context, id uuid.UUID) (*domain.Incident, error) {
	incident, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return incident, nil
}

func (s *IncidentService) GetAllIncidents(ctx context.Context, page, pageSize int) ([]*domain.Incident, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	return s.repo.GetAll(ctx, pageSize, offset)
}

func (s *IncidentService) UpdateIncident(ctx context.Context, id uuid.UUID, req *domain.UpdateIncidentRequest) (*domain.Incident, error) {
	incident, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

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
			return nil, domain.ErrInvalidRadius
		}
		incident.Radius = *req.Radius
	}
	if req.IsActive != nil {
		incident.IsActive = *req.IsActive
	}

	if req.Latitude != nil || req.Longitude != nil {
		if err := s.validateCoordinates(incident.Latitude, incident.Longitude); err != nil {
			return nil, err
		}
	}

	if err := s.repo.Update(ctx, id, incident); err != nil {
		return nil, fmt.Errorf("failed to update incident: %w", err)
	}

	// Инвалидируем кэш активных инцидентов
	if s.locationService != nil {
		_ = s.locationService.InvalidateCache(ctx)
	}

	return incident, nil
}

func (s *IncidentService) DeleteIncident(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Инвалидируем кэш активных инцидентов
	if s.locationService != nil {
		_ = s.locationService.InvalidateCache(ctx)
	}

	return nil
}

func (s *IncidentService) validateCoordinates(lat, lon float64) error {
	if lat < -90 || lat > 90 {
		return fmt.Errorf("%w: latitude must be between -90 and 90", domain.ErrInvalidCoordinates)
	}
	if lon < -180 || lon > 180 {
		return fmt.Errorf("%w: longitude must be between -180 and 180", domain.ErrInvalidCoordinates)
	}
	return nil
}
