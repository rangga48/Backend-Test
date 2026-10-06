package entity

import (
	"encoding/json"
	"testing"
)

func ptrString(s string) *string {
	return &s
}

func ptrFloat64(f float64) *float64 {
	return &f
}

func TestValidate_ValidRequest(t *testing.T) {
	req := CreateOrUpdateEntityRequest{
		Name:        ptrString("Vehicle #001"),
		Type:        ptrString("Vehicle"),
		Status:      ptrString("Active"),
		Description: ptrString("Delivery vehicle"),
		Latitude:    ptrFloat64(-6.9175),
		Longitude:   ptrFloat64(107.6191),
	}

	errs := req.Validate()
	if len(errs) != 0 {
		t.Fatalf("expected no validation errors, got: %v", errs)
	}
}

func TestValidate_AllFieldsMissing(t *testing.T) {
	req := CreateOrUpdateEntityRequest{}

	errs := req.Validate()
	if len(errs) != 5 {
		t.Fatalf("expected 5 validation errors, got %d: %v", len(errs), errs)
	}

	if errs["name"] != "is required" {
		t.Errorf("expected 'is required' for name, got: %s", errs["name"])
	}
	if errs["type"] != "is required" {
		t.Errorf("expected 'is required' for type, got: %s", errs["type"])
	}
	if errs["status"] != "is required" {
		t.Errorf("expected 'is required' for status, got: %s", errs["status"])
	}
	if errs["latitude"] != "is required" {
		t.Errorf("expected 'is required' for latitude, got: %s", errs["latitude"])
	}
	if errs["longitude"] != "is required" {
		t.Errorf("expected 'is required' for longitude, got: %s", errs["longitude"])
	}
}

func TestValidate_InvalidRangesAndTypes(t *testing.T) {
	longName := ""
	for i := 0; i < 105; i++ {
		longName += "a"
	}

	req := CreateOrUpdateEntityRequest{
		Name:      ptrString(longName),
		Type:      ptrString("Alien"),
		Status:    ptrString("Unknown"),
		Latitude:  ptrFloat64(95.0),
		Longitude: ptrFloat64(-200.0),
	}

	errs := req.Validate()
	if len(errs) != 5 {
		t.Fatalf("expected 5 validation errors, got %d: %v", len(errs), errs)
	}

	if errs["name"] != "must not exceed 100 characters" {
		t.Errorf("unexpected name error: %s", errs["name"])
	}
	if errs["type"] != "must be one of: Vehicle, IoT Device, Facility, Other" {
		t.Errorf("unexpected type error: %s", errs["type"])
	}
	if errs["status"] != "must be one of: Active, Inactive, Maintenance" {
		t.Errorf("unexpected status error: %s", errs["status"])
	}
	if errs["latitude"] != "must be between -90 and 90" {
		t.Errorf("unexpected latitude error: %s", errs["latitude"])
	}
	if errs["longitude"] != "must be between -180 and 180" {
		t.Errorf("unexpected longitude error: %s", errs["longitude"])
	}
}

func TestValidate_BoundaryCoordinates(t *testing.T) {
	req := CreateOrUpdateEntityRequest{
		Name:      ptrString("Sensor 01"),
		Type:      ptrString("IoT Device"),
		Status:    ptrString("Maintenance"),
		Latitude:  ptrFloat64(-90.0),
		Longitude: ptrFloat64(180.0),
	}

	errs := req.Validate()
	if len(errs) != 0 {
		t.Fatalf("expected boundary values -90 and 180 to be valid, got: %v", errs)
	}

	req2 := CreateOrUpdateEntityRequest{
		Name:      ptrString("Sensor 02"),
		Type:      ptrString("Facility"),
		Status:    ptrString("Inactive"),
		Latitude:  ptrFloat64(90.0),
		Longitude: ptrFloat64(-180.0),
	}

	errs2 := req2.Validate()
	if len(errs2) != 0 {
		t.Fatalf("expected boundary values 90 and -180 to be valid, got: %v", errs2)
	}
}

func TestUnmarshal_StringCoordinatesAndTypo(t *testing.T) {
	rawJSON := `{
		"name": "Vehicle #001",
		"type": "Vehicle",
		"status": "Active",
		"description": "Kendaraan operasional",
		"latitude": "-6.9175",
		"longtitude": "107.6191"
	}`

	var req CreateOrUpdateEntityRequest
	if err := json.Unmarshal([]byte(rawJSON), &req); err != nil {
		t.Fatalf("expected unmarshal to succeed, got: %v", err)
	}

	errs := req.Validate()
	if len(errs) != 0 {
		t.Fatalf("expected no validation errors, got: %v", errs)
	}

	if req.Latitude == nil || *req.Latitude != -6.9175 {
		t.Fatalf("expected parsed latitude -6.9175, got: %v", req.Latitude)
	}
	if req.Longitude == nil || *req.Longitude != 107.6191 {
		t.Fatalf("expected parsed longitude 107.6191, got: %v", req.Longitude)
	}
}

func TestUnmarshal_InvalidStringCoordinates(t *testing.T) {
	rawJSON := `{
		"name": "Vehicle #001",
		"type": "Vehicle",
		"status": "Active",
		"latitude": "bukan-angka",
		"longitude": 107.6191
	}`

	var req CreateOrUpdateEntityRequest
	if err := json.Unmarshal([]byte(rawJSON), &req); err != nil {
		t.Fatalf("expected unmarshal to succeed, got: %v", err)
	}

	errs := req.Validate()
	if errs["latitude"] == "" {
		t.Fatalf("expected latitude error, got: %v", errs)
	}
}
