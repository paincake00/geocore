package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	delivery "github.com/paincake00/geocore/internal/delivery/http"
	"github.com/paincake00/geocore/internal/entity"
	"github.com/paincake00/geocore/internal/usecase"
)

// --- Моки ---

type MockIncidentRepo struct {
	Incidents map[int]*entity.Incident
	Stats     map[int]int
}

func NewMockIncidentRepo() *MockIncidentRepo {
	return &MockIncidentRepo{
		Incidents: make(map[int]*entity.Incident),
		Stats:     make(map[int]int),
	}
}

func (m *MockIncidentRepo) Create(ctx context.Context, i *entity.Incident) error {
	i.ID = len(m.Incidents) + 1
	i.CreatedAt = time.Now()
	m.Incidents[i.ID] = i
	return nil
}

func (m *MockIncidentRepo) GetByID(ctx context.Context, id int) (*entity.Incident, error) {
	if i, ok := m.Incidents[id]; ok {
		return i, nil
	}
	return nil, fmt.Errorf("not found")
}

func (m *MockIncidentRepo) GetAll(ctx context.Context, limit, offset int) ([]*entity.Incident, error) {
	var res []*entity.Incident
	for _, i := range m.Incidents {
		res = append(res, i)
	}
	return res, nil
}

func (m *MockIncidentRepo) GetAllActive(ctx context.Context) ([]*entity.Incident, error) {
	return m.GetAll(ctx, 0, 0)
}

func (m *MockIncidentRepo) Update(ctx context.Context, i *entity.Incident) error {
	if _, ok := m.Incidents[i.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.Incidents[i.ID] = i
	return nil
}

func (m *MockIncidentRepo) Delete(ctx context.Context, id int) error {
	if _, ok := m.Incidents[id]; !ok {
		return fmt.Errorf("not found")
	}
	delete(m.Incidents, id)
	return nil
}

func (m *MockIncidentRepo) GetStats(ctx context.Context, windowMinutes int) (map[int]int, error) {
	return m.Stats, nil
}

// Проверка реализации интерфейса
var _ usecase.IncidentRepository = (*MockIncidentRepo)(nil)

type MockLocationRepo struct{}

func (m *MockLocationRepo) CreateLocationCheck(ctx context.Context, check *entity.LocationCheck) error {
	check.ID = 1
	check.CheckedAt = time.Now()
	return nil
}
func (m *MockLocationRepo) RecordIncidentMatch(ctx context.Context, checkID, incidentID int) error {
	return nil
}

type MockQueueRepo struct{}

func (m *MockQueueRepo) Enqueue(ctx context.Context, task string, payload interface{}) error {
	return nil
}
func (m *MockQueueRepo) Dequeue(ctx context.Context, task string) (string, error) {
	return "", nil
}

type MockCache struct{}

func (m *MockCache) SetIncidents(ctx context.Context, incidents []*entity.Incident) error { return nil }
func (m *MockCache) GetIncidents(ctx context.Context) ([]*entity.Incident, error) {
	return nil, nil // Cache miss
}

type MockPinger struct{}

func (m *MockPinger) Ping(ctx context.Context) error { return nil }

// --- Вспомогательные функции ---

func setupHandler() (*gin.Engine, *MockIncidentRepo) {
	gin.SetMode(gin.TestMode)

	mockIncRepo := NewMockIncidentRepo()
	mockLocRepo := &MockLocationRepo{}
	mockQueue := &MockQueueRepo{}
	mockCache := &MockCache{}
	mockPinger := &MockPinger{}

	incidentService := usecase.NewIncidentService(mockIncRepo, mockCache)
	geoService := usecase.NewGeoService(mockIncRepo, mockLocRepo, mockQueue, mockCache)

	apiKey := "test-key"
	statsWindow := 30

	h := delivery.NewHandler(incidentService, geoService, mockPinger, mockPinger, apiKey, statsWindow)
	return h.InitRoutes(), mockIncRepo
}

// --- Тесты ---

func TestHealthCheck(t *testing.T) {
	router, _ := setupHandler()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/system/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestCreateIncident_Unauthorized(t *testing.T) {
	router, _ := setupHandler()

	body := []byte(`{"title":"Test","latitude":10,"longitude":10,"radius_meters":100}`)
	req, _ := http.NewRequest("POST", "/api/v1/incidents", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	// Нет API ключа

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestCreateIncident_Success(t *testing.T) {
	router, repo := setupHandler()

	body := []byte(`{"title":"Fire","description":"Big fire","latitude":55.0,"longitude":37.0,"radius_meters":500}`)
	req, _ := http.NewRequest("POST", "/api/v1/incidents", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-key")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	if len(repo.Incidents) != 1 {
		t.Errorf("Expected 1 incident in repo, got %d", len(repo.Incidents))
	}
}

func TestGetIncidents(t *testing.T) {
	router, repo := setupHandler()

	// Предзаполнение репозитория
	repo.Incidents[100] = &entity.Incident{ID: 100, Title: "Test Incident", Latitude: 1, Longitude: 1, RadiusMeters: 100}

	req, _ := http.NewRequest("GET", "/api/v1/incidents", nil)
	req.Header.Set("X-API-Key", "test-key")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var res []entity.Incident
	json.Unmarshal(w.Body.Bytes(), &res)
	if len(res) < 1 {
		t.Error("Expected at least 1 incident in response")
	}
}

func TestCheckLocation(t *testing.T) {
	router, repo := setupHandler()

	// Инцидент в координатах 10,10 с радиусом 1000м (~1км)
	repo.Incidents[1] = &entity.Incident{ID: 1, Title: "Danger Zone", Latitude: 10.0, Longitude: 10.0, RadiusMeters: 1000}

	body := []byte(`{"user_id":"u1","latitude":10.001,"longitude":10.001}`)
	req, _ := http.NewRequest("POST", "/api/v1/location/check", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var matches []entity.Incident
	json.Unmarshal(w.Body.Bytes(), &matches)
	if len(matches) != 1 {
		t.Errorf("Expected 1 match, got %d", len(matches))
	}
}
