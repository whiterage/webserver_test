package domain

import "errors"

var (
	ErrIncidentNotFound   = errors.New("incident not found")
	ErrInvalidCoordinates = errors.New("invalid coordinates")
	ErrInvalidRadius      = errors.New("radius must be positive")
)
