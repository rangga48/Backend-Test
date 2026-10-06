package entity

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// mockRepository implements Repository for unit testing
type mockRepository struct {
	items map[uuid.UUID]Entity
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		items: make(map[uuid.UUID]Entity),
	}
}

func (m *mockRepository) FindAll(ctx context.Context) ([]Entity, error) {
	result := make([]Entity, 0, len(m.items))
	for _, item := range m.items {
		result = append(result, item)
	}
	return result, nil
}

func (m *mockRepository) FindByID(ctx context.Context, id uuid.UUID) (*Entity, error) {
	item, ok := m.items[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &item, nil
}

func (m *mockRepository) Create(ctx context.Context, entity *Entity) error {
	m.items[entity.ID] = *entity
	return nil
}

func (m *mockRepository) Update(ctx context.Context, entity *Entity) error {
	if _, ok := m.items[entity.ID]; !ok {
		return ErrNotFound
	}
	m.items[entity.ID] = *entity
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if _, ok := m.items[id]; !ok {
		return ErrNotFound
	}
	delete(m.items, id)
	return nil
}

func setupTestRouter(repo Repository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	service := NewService(repo)
	handler := NewHandler(service)
	api := r.Group("/api")
	handler.RegisterRoutes(api)
	return r
}

type genericResponse struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func TestHandler_GetAll(t *testing.T) {
	repo := newMockRepository()
	id := uuid.New()
	repo.items[id] = Entity{
		ID:        id,
		Name:      "Vehicle #001",
		Type:      TypeVehicle,
		Status:    StatusActive,
		Latitude:  -6.9175,
		Longitude: 107.6191,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	router := setupTestRouter(repo)

	req, _ := http.NewRequest(http.MethodGet, "/api/entities", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var res genericResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response envelope: %v", err)
	}

	if res.Status != "success" {
		t.Errorf("expected status 'success', got '%s'", res.Status)
	}

	var entities []Entity
	if err := json.Unmarshal(res.Data, &entities); err != nil {
		t.Fatalf("failed to decode entities data: %v", err)
	}
	if len(entities) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(entities))
	}
}

func TestHandler_GetByID_SuccessAndNotFound(t *testing.T) {
	repo := newMockRepository()
	id := uuid.New()
	repo.items[id] = Entity{
		ID:        id,
		Name:      "Test Facility",
		Type:      TypeFacility,
		Status:    StatusActive,
		Latitude:  -6.2,
		Longitude: 106.8,
	}

	router := setupTestRouter(repo)

	// Test found
	req, _ := http.NewRequest(http.MethodGet, "/api/entities/"+id.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var res genericResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("expected status 'success', got '%s'", res.Status)
	}

	// Test not found
	randomID := uuid.New().String()
	reqNF, _ := http.NewRequest(http.MethodGet, "/api/entities/"+randomID, nil)
	wNF := httptest.NewRecorder()
	router.ServeHTTP(wNF, reqNF)
	if wNF.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", wNF.Code)
	}

	var resNF genericResponse
	if err := json.Unmarshal(wNF.Body.Bytes(), &resNF); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resNF.Status != "error" {
		t.Errorf("expected status 'error', got '%s'", resNF.Status)
	}

	// Test invalid UUID
	reqInv, _ := http.NewRequest(http.MethodGet, "/api/entities/invalid-uuid", nil)
	wInv := httptest.NewRecorder()
	router.ServeHTTP(wInv, reqInv)
	if wInv.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid UUID, got %d", wInv.Code)
	}
}

func TestHandler_Create_Success(t *testing.T) {
	repo := newMockRepository()
	router := setupTestRouter(repo)

	body := `{
		"name": "Vehicle #001",
		"type": "Vehicle",
		"status": "Active",
		"description": "Delivery vehicle",
		"latitude": -6.9175,
		"longitude": 107.6191
	}`

	req, _ := http.NewRequest(http.MethodPost, "/api/entities", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var res genericResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response envelope: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("expected status 'success', got '%s'", res.Status)
	}

	var created Entity
	if err := json.Unmarshal(res.Data, &created); err != nil {
		t.Fatalf("failed to decode created entity: %v", err)
	}

	if created.Name != "Vehicle #001" {
		t.Errorf("expected name 'Vehicle #001', got '%s'", created.Name)
	}
}

func TestHandler_Create_ValidationError(t *testing.T) {
	repo := newMockRepository()
	router := setupTestRouter(repo)

	// Invalid latitude
	body := `{
		"name": "Vehicle #001",
		"type": "Vehicle",
		"status": "Active",
		"latitude": 120.0,
		"longitude": 107.6191
	}`

	req, _ := http.NewRequest(http.MethodPost, "/api/entities", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var res genericResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if res.Status != "error" {
		t.Errorf("expected status 'error', got '%s'", res.Status)
	}

	var details map[string]string
	if err := json.Unmarshal(res.Data, &details); err != nil {
		t.Fatalf("failed to decode details: %v", err)
	}
	if details["latitude"] != "must be between -90 and 90" {
		t.Errorf("expected latitude error, got: %s", details["latitude"])
	}
}

func TestHandler_Update_Success(t *testing.T) {
	repo := newMockRepository()
	id := uuid.New()
	repo.items[id] = Entity{
		ID:        id,
		Name:      "Old Name",
		Type:      TypeOther,
		Status:    StatusInactive,
		Latitude:  0.0,
		Longitude: 0.0,
	}

	router := setupTestRouter(repo)

	body := `{
		"name": "Updated Name",
		"type": "Vehicle",
		"status": "Active",
		"latitude": -7.25,
		"longitude": 112.75
	}`

	req, _ := http.NewRequest(http.MethodPut, "/api/entities/"+id.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var res genericResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("expected status 'success', got '%s'", res.Status)
	}
}

func TestHandler_Update_QueryParam_Success(t *testing.T) {
	repo := newMockRepository()
	id := uuid.New()
	repo.items[id] = Entity{
		ID:        id,
		Name:      "Old Name",
		Type:      TypeOther,
		Status:    StatusInactive,
		Latitude:  0.0,
		Longitude: 0.0,
	}

	router := setupTestRouter(repo)

	body := `{
		"name": "Updated via Query",
		"type": "Vehicle",
		"status": "Active",
		"latitude": -7.25,
		"longitude": 112.75
	}`

	req, _ := http.NewRequest(http.MethodPut, "/api/entities?id="+id.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var res genericResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("expected status 'success', got '%s'", res.Status)
	}
}

func TestHandler_Delete_Success(t *testing.T) {
	repo := newMockRepository()
	id := uuid.New()
	repo.items[id] = Entity{
		ID:   id,
		Name: "To Delete",
	}

	router := setupTestRouter(repo)

	req, _ := http.NewRequest(http.MethodDelete, "/api/entities/"+id.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var res genericResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("expected status 'success', got '%s'", res.Status)
	}

	if len(repo.items) != 0 {
		t.Fatalf("expected item to be deleted from repo")
	}
}

func TestHandler_Delete_QueryParam_Success(t *testing.T) {
	repo := newMockRepository()
	id := uuid.New()
	repo.items[id] = Entity{
		ID:   id,
		Name: "To Delete Via Query",
	}

	router := setupTestRouter(repo)

	req, _ := http.NewRequest(http.MethodDelete, "/api/entities?id="+id.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var res genericResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("expected status 'success', got '%s'", res.Status)
	}

	if len(repo.items) != 0 {
		t.Fatalf("expected item to be deleted from repo")
	}
}
