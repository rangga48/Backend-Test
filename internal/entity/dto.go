package entity

import (
	"encoding/json"
	"strconv"
	"strings"
)

// CreateOrUpdateEntityRequest represents the request body for creating or updating an entity
type CreateOrUpdateEntityRequest struct {
	Name        *string  `json:"name"`
	Type        *string  `json:"type"`
	Status      *string  `json:"status"`
	Description *string  `json:"description"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`

	// parseErrors holds any type conversion errors encountered during JSON unmarshaling
	parseErrors map[string]string
}

type rawEntityRequest struct {
	Name        *string     `json:"name"`
	Type        *string     `json:"type"`
	Status      *string     `json:"status"`
	Description *string     `json:"description"`
	Latitude    interface{} `json:"latitude"`
	Longitude   interface{} `json:"longitude"`
	Longtitude  interface{} `json:"longtitude"` // support common spelling typo
}

// UnmarshalJSON custom unmarshaler to accept numbers or numeric strings for coordinates
func (r *CreateOrUpdateEntityRequest) UnmarshalJSON(data []byte) error {
	var raw rawEntityRequest
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	r.Name = raw.Name
	r.Type = raw.Type
	r.Status = raw.Status
	r.Description = raw.Description
	r.parseErrors = make(map[string]string)

	// Process Latitude (support float or numeric string)
	if raw.Latitude != nil {
		switch v := raw.Latitude.(type) {
		case float64:
			r.Latitude = &v
		case string:
			trimmed := strings.TrimSpace(v)
			if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
				r.Latitude = &f
			} else {
				r.parseErrors["latitude"] = "must be a valid number between -90 and 90"
			}
		default:
			r.parseErrors["latitude"] = "must be a valid number between -90 and 90"
		}
	}

	// Process Longitude (support raw.Longitude or typo raw.Longtitude)
	lonRaw := raw.Longitude
	if lonRaw == nil && raw.Longtitude != nil {
		lonRaw = raw.Longtitude
	}

	if lonRaw != nil {
		switch v := lonRaw.(type) {
		case float64:
			r.Longitude = &v
		case string:
			trimmed := strings.TrimSpace(v)
			if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
				r.Longitude = &f
			} else {
				r.parseErrors["longitude"] = "must be a valid number between -180 and 180"
			}
		default:
			r.parseErrors["longitude"] = "must be a valid number between -180 and 180"
		}
	}

	return nil
}

// Validate validates the CreateOrUpdateEntityRequest according to PRD requirements
func (r *CreateOrUpdateEntityRequest) Validate() map[string]string {
	details := make(map[string]string)

	// Copy parsing errors if any
	for k, v := range r.parseErrors {
		details[k] = v
	}

	// Validate Name
	if r.Name == nil || strings.TrimSpace(*r.Name) == "" {
		details["name"] = "is required"
	} else if len([]rune(strings.TrimSpace(*r.Name))) > 100 {
		details["name"] = "must not exceed 100 characters"
	}

	// Validate Type
	if r.Type == nil || strings.TrimSpace(*r.Type) == "" {
		details["type"] = "is required"
	} else {
		allowedTypes := ValidEntityTypes()
		if !allowedTypes[strings.TrimSpace(*r.Type)] {
			details["type"] = "must be one of: Vehicle, IoT Device, Facility, Other"
		}
	}

	// Validate Status
	if r.Status == nil || strings.TrimSpace(*r.Status) == "" {
		details["status"] = "is required"
	} else {
		allowedStatuses := ValidEntityStatuses()
		if !allowedStatuses[strings.TrimSpace(*r.Status)] {
			details["status"] = "must be one of: Active, Inactive, Maintenance"
		}
	}

	// Validate Latitude
	if _, hasError := details["latitude"]; !hasError {
		if r.Latitude == nil {
			details["latitude"] = "is required"
		} else if *r.Latitude < -90 || *r.Latitude > 90 {
			details["latitude"] = "must be between -90 and 90"
		}
	}

	// Validate Longitude
	if _, hasError := details["longitude"]; !hasError {
		if r.Longitude == nil {
			details["longitude"] = "is required"
		} else if *r.Longitude < -180 || *r.Longitude > 180 {
			details["longitude"] = "must be between -180 and 180"
		}
	}

	return details
}

// Response represents the standardized API response envelope
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// NewSuccessResponse creates a standardized success response
func NewSuccessResponse(message string, data interface{}) Response {
	return Response{
		Status:  "success",
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse creates a standardized error response
func NewErrorResponse(message string, data interface{}) Response {
	return Response{
		Status:  "error",
		Message: message,
		Data:    data,
	}
}

// ErrorDetail kept for backward compatibility if needed
type ErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// ErrorResponse kept for backward compatibility if needed
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// MessageResponse kept for backward compatibility if needed
type MessageResponse struct {
	Message string `json:"message"`
}
