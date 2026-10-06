package entity

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	// ErrNotFound is returned when an entity is not found in the database
	ErrNotFound = errors.New("entity not found")
)

// Repository defines the contract for entity data access operations
type Repository interface {
	FindAll(ctx context.Context) ([]Entity, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Entity, error)
	Create(ctx context.Context, entity *Entity) error
	Update(ctx context.Context, entity *Entity) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository creates a new GORM-backed entity repository
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

// FindAll retrieves all entities ordered by creation date descending
func (r *gormRepository) FindAll(ctx context.Context) ([]Entity, error) {
	var entities []Entity
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&entities).Error; err != nil {
		return nil, err
	}
	if entities == nil {
		entities = make([]Entity, 0)
	}
	return entities, nil
}

// FindByID retrieves a single entity by its UUID
func (r *gormRepository) FindByID(ctx context.Context, id uuid.UUID) (*Entity, error) {
	var item Entity
	if err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

// Create persists a new entity
func (r *gormRepository) Create(ctx context.Context, entity *Entity) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// Update persists modifications to an existing entity
func (r *gormRepository) Update(ctx context.Context, entity *Entity) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// Delete removes an entity by UUID
func (r *gormRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&Entity{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
