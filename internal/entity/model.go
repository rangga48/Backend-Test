package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EntityType string

const (
	TypeVehicle   EntityType = "Vehicle"
	TypeIoTDevice EntityType = "IoT Device"
	TypeFacility  EntityType = "Facility"
	TypeOther     EntityType = "Other"
)

type EntityStatus string

const (
	StatusActive      EntityStatus = "Active"
	StatusInactive    EntityStatus = "Inactive"
	StatusMaintenance EntityStatus = "Maintenance"
)

// ValidEntityTypes returns the allowed entity types
func ValidEntityTypes() map[string]bool {
	return map[string]bool{
		string(TypeVehicle):   true,
		string(TypeIoTDevice): true,
		string(TypeFacility):  true,
		string(TypeOther):     true,
	}
}

// ValidEntityStatuses returns the allowed entity statuses
func ValidEntityStatuses() map[string]bool {
	return map[string]bool{
		string(StatusActive):      true,
		string(StatusInactive):    true,
		string(StatusMaintenance): true,
	}
}

// Entity represents a geospatial entity stored in the database
type Entity struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string       `gorm:"type:varchar(100);not null" json:"name"`
	Type        EntityType   `gorm:"type:varchar(50);not null" json:"type"`
	Status      EntityStatus `gorm:"type:varchar(50);not null" json:"status"`
	Description string       `gorm:"type:text" json:"description"`
	Latitude    float64      `gorm:"type:double precision;not null" json:"latitude"`
	Longitude   float64      `gorm:"type:double precision;not null" json:"longitude"`
	CreatedAt   time.Time    `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"not null" json:"updated_at"`
}

// BeforeCreate hook to generate UUID if not provided
func (e *Entity) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}
