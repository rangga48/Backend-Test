package entity

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

// Service defines business logic operations for entities
type Service interface {
	GetAll(ctx context.Context) ([]Entity, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Entity, error)
	Create(ctx context.Context, req *CreateOrUpdateEntityRequest) (*Entity, error)
	Update(ctx context.Context, id uuid.UUID, req *CreateOrUpdateEntityRequest) (*Entity, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type entityService struct {
	repo Repository
}

// NewService creates a new entity service instance
func NewService(repo Repository) Service {
	return &entityService{repo: repo}
}

// GetAll retrieves all entities
func (s *entityService) GetAll(ctx context.Context) ([]Entity, error) {
	return s.repo.FindAll(ctx)
}

// GetByID retrieves a single entity by its UUID
func (s *entityService) GetByID(ctx context.Context, id uuid.UUID) (*Entity, error) {
	return s.repo.FindByID(ctx, id)
}

// Create builds and stores a new entity from validated request data
func (s *entityService) Create(ctx context.Context, req *CreateOrUpdateEntityRequest) (*Entity, error) {
	desc := ""
	if req.Description != nil {
		desc = strings.TrimSpace(*req.Description)
	}

	entity := &Entity{
		ID:          uuid.New(),
		Name:        strings.TrimSpace(*req.Name),
		Type:        EntityType(strings.TrimSpace(*req.Type)),
		Status:      EntityStatus(strings.TrimSpace(*req.Status)),
		Description: desc,
		Latitude:    *req.Latitude,
		Longitude:   *req.Longitude,
	}

	if err := s.repo.Create(ctx, entity); err != nil {
		return nil, err
	}

	return entity, nil
}

// Update modifies an existing entity
func (s *entityService) Update(ctx context.Context, id uuid.UUID, req *CreateOrUpdateEntityRequest) (*Entity, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	desc := ""
	if req.Description != nil {
		desc = strings.TrimSpace(*req.Description)
	}

	existing.Name = strings.TrimSpace(*req.Name)
	existing.Type = EntityType(strings.TrimSpace(*req.Type))
	existing.Status = EntityStatus(strings.TrimSpace(*req.Status))
	existing.Description = desc
	existing.Latitude = *req.Latitude
	existing.Longitude = *req.Longitude

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// Delete removes an entity by ID
func (s *entityService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
