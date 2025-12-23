package service

import (
	"context"
	"encoding/json"
	"fmt"
	"geo-alert-core/internal/domain"
	"geo-alert-core/internal/infrastructure/webhook"
	"geo-alert-core/internal/repository"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type LocationService struct {
	incidentRepo  repository.IncidentRepository
	checkRepo     repository.LocationCheckRepository
	redisClient   *redis.Client
	webhookSender *webhook.Sender
	cacheTTL      time.Duration
}

func NewLocationService(
	incidentRepo repository.IncidentRepository,
	checkRepo repository.LocationCheckRepository,
	redisClient *redis.Client,
	webhookSender *webhook.Sender,
) *LocationService {
	return &LocationService{
		incidentRepo:  incidentRepo,
		checkRepo:     checkRepo,
		redisClient:   redisClient,
		webhookSender: webhookSender,
		cacheTTL:      5 * time.Minute, // Cache active incidents for 5 minutes
	}
}

// proverka koordinat i vozvrat blizhayshix incidetov
func (s *LocationService) CheckLocation(ctx context.Context, req *domain.LocationCheckRequest) (*domain.LocationCheckResponse, error) {
	if req.Latitude < -90 || req.Latitude > 90 {
		return nil, fmt.Errorf("invalid latitude")
	}
	if req.Longitude < -180 || req.Longitude > 180 {
		return nil, fmt.Errorf("invalid longitude")
	}

	// get active incidents (with cache)
	incidents, err := s.getActiveIncidentsCached(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get incidents: %w", err)
	}

	// find nearby incidents
	nearbyIncidents, err := s.incidentRepo.FindNearbyIncidents(ctx, req.Latitude, req.Longitude)
	if err != nil {
		return nil, fmt.Errorf("failed to find nearby incidents: %w", err)
	}

	// save check to db
	check := &domain.LocationCheck{
		UserID:    req.UserID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	}

	if err := s.checkRepo.Create(ctx, check); err != nil {
		return nil, fmt.Errorf("failed to save location check: %w", err)
	}

	// link check to found incidents
	if len(nearbyIncidents) > 0 {
		incidentIDs := make([]uuid.UUID, len(nearbyIncidents))
		for i, inc := range nearbyIncidents {
			incidentIDs[i] = inc.ID
		}

		if err := s.checkRepo.LinkToIncidents(ctx, check.ID, incidentIDs); err != nil {
			fmt.Printf("Failed to link incidents: %v\n", err)
		}
	}

	if len(nearbyIncidents) > 0 {
		go s.sendWebhookAsync(check, nearbyIncidents)
	}

	return &domain.LocationCheckResponse{
		HasDanger: len(nearbyIncidents) > 0,
		Incidents: s.convertToDomainIncidents(nearbyIncidents),
	}, nil
}

// polychenie aktivnyh incidetov s keshirovaniem v Redis
func (s *LocationService) getActiveIncidentsCached(ctx context.Context) ([]*domain.Incident, error) {
	cacheKey := "active_incidents"

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var incidents []*domain.Incident
		if err := json.Unmarshal([]byte(cached), &incidents); err == nil {
			return incidents, nil
		}
	}

	// cache not found or error, load from db
	incidents, err := s.incidentRepo.GetActiveIncidents(ctx)
	if err != nil {
		return nil, err
	}

	// save to cache
	jsonData, err := json.Marshal(incidents)
	if err == nil {
		s.redisClient.Set(ctx, cacheKey, jsonData, s.cacheTTL)
	}

	return incidents, nil
}

// asynch otpravlyaem webhook
func (s *LocationService) sendWebhookAsync(check *domain.LocationCheck, incidents []*domain.Incident) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// convert incidents to webhook
	incidentInfos := make([]webhook.IncidentInfo, len(incidents))
	for i, inc := range incidents {
		incidentInfos[i] = webhook.IncidentInfo{
			ID:          inc.ID.String(),
			Title:       inc.Title,
			Description: inc.Description,
			Latitude:    inc.Latitude,
			Longitude:   inc.Longitude,
			Radius:      inc.Radius,
		}
	}

	payload := &webhook.WebhookPayload{
		UserID:    check.UserID,
		Latitude:  check.Latitude,
		Longitude: check.Longitude,
		Incidents: incidentInfos,
		CheckedAt: check.CheckedAt,
	}

	if err := s.webhookSender.Send(ctx, payload); err != nil {
		fmt.Printf("Failed to send webhook: %v\n", err)
	}
}

func (s *LocationService) convertToDomainIncidents(incidents []*domain.Incident) []domain.Incident {
	result := make([]domain.Incident, len(incidents))
	for i, inc := range incidents {
		result[i] = *inc
	}
	return result
}

// called when creating/updating/deleting incident
func (s *LocationService) InvalidateCache(ctx context.Context) error {
	return s.redisClient.Del(ctx, "active_incidents").Err()
}
