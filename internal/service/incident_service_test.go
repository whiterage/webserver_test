package service

import (
	"context"
	"geo-alert-core/internal/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockIncidentRepository - мок для тестирования
type MockIncidentRepository struct {
	mock.Mock
}

func (m *MockIncidentRepository) Create(ctx context.Context, incident *domain.Incident) error {
	args := m.Called(ctx, incident)
	return args.Error(0)
}

func (m *MockIncidentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Incident, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Incident), args.Error(1)
}

func (m *MockIncidentRepository) GetAll(ctx context.Context, limit, offset int) ([]*domain.Incident, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*domain.Incident), args.Error(1)
}

func (m *MockIncidentRepository) GetActiveIncidents(ctx context.Context) ([]*domain.Incident, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.Incident), args.Error(1)
}

func (m *MockIncidentRepository) Update(ctx context.Context, id uuid.UUID, incident *domain.Incident) error {
	args := m.Called(ctx, id, incident)
	return args.Error(0)
}

func (m *MockIncidentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockIncidentRepository) FindNearbyIncidents(ctx context.Context, latitude, longitude float64) ([]*domain.Incident, error) {
	args := m.Called(ctx, latitude, longitude)
	return args.Get(0).([]*domain.Incident), args.Error(1)
}

func (m *MockIncidentRepository) GetStats(ctx context.Context, minutes int) ([]*domain.IncidentStats, error) {
	args := m.Called(ctx, minutes)
	return args.Get(0).([]*domain.IncidentStats), args.Error(1)
}

func TestIncidentService_CreateIncident(t *testing.T) {
	tests := []struct {
		name    string
		req     *domain.CreateIncidentRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid incident",
			req: &domain.CreateIncidentRequest{
				Title:       "Test",
				Description: "Test description",
				Latitude:    55.7558,
				Longitude:   37.6173,
				Radius:      100.0,
			},
			wantErr: false,
		},
		{
			name: "invalid latitude",
			req: &domain.CreateIncidentRequest{
				Title:     "Test",
				Latitude:  100.0, // Invalid
				Longitude: 37.6173,
				Radius:    100.0,
			},
			wantErr: true,
			errMsg:  "latitude must be between -90 and 90",
		},
		{
			name: "invalid radius",
			req: &domain.CreateIncidentRequest{
				Title:     "Test",
				Latitude:  55.7558,
				Longitude: 37.6173,
				Radius:    -10.0, // Invalid
			},
			wantErr: true,
			errMsg:  "radius must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockIncidentRepository)
			service := NewIncidentService(mockRepo)

			if !tt.wantErr {
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
			}

			_, err := service.CreateIncident(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

func TestIncidentService_GetAllIncidents(t *testing.T) {
	mockRepo := new(MockIncidentRepository)
	service := NewIncidentService(mockRepo)

	tests := []struct {
		name     string
		page     int
		pageSize int
		expected int
	}{
		{"default pagination", 1, 20, 20},
		{"invalid page", 0, 20, 20},         // Должно стать 1
		{"invalid page size", 1, 0, 20},     // Должно стать 20
		{"too large page size", 1, 200, 20}, // Должно стать 20
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.On("GetAll", mock.Anything, tt.expected, mock.Anything).Return([]*domain.Incident{}, nil)

			_, err := service.GetAllIncidents(context.Background(), tt.page, tt.pageSize)
			assert.NoError(t, err)
		})
	}
}
