package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"groupie_tracker/backend/events"
	"groupie_tracker/backend/models"
	"groupie_tracker/backend/services"
	"groupie_tracker/db"
)

func setupTestHandler() (*ArtistHandler, *db.ArtistStore) {
	store := db.NewArtistStore()
	service := services.NewArtistService(store)
	searchChan := make(chan events.SearchEvent, 10)
	go events.StartSearchWorker(store, searchChan)
	handler := NewArtistHandler(service, searchChan)
	return handler, store
}

func TestHandleArtists_GET(t *testing.T) {
	handler, store := setupTestHandler()
	store.Create(models.Artist{Name: "Test Band"})

	req := httptest.NewRequest(http.MethodGet, "/api/artists", nil)
	w := httptest.NewRecorder()

	handler.handleArtists(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var artists []models.Artist
	if err := json.NewDecoder(w.Body).Decode(&artists); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(artists) != 1 {
		t.Errorf("expected 1 artist, got %d", len(artists))
	}
}

func TestHandleArtists_POST(t *testing.T) {
	handler, _ := setupTestHandler()

	artist := models.Artist{
		Name:    "New Band",
		Members: []string{"Member1"},
	}
	body, _ := json.Marshal(artist)

	req := httptest.NewRequest(http.MethodPost, "/api/artists", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleArtists(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var created models.Artist
	if err := json.NewDecoder(w.Body).Decode(&created); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if created.Name != artist.Name {
		t.Errorf("expected name %s, got %s", artist.Name, created.Name)
	}
	if created.ID == 0 {
		t.Error("expected ID to be assigned")
	}
}

func TestHandleArtistByID_GET(t *testing.T) {
	handler, store := setupTestHandler()
	artist := store.Create(models.Artist{Name: "Test Band"})

	req := httptest.NewRequest(http.MethodGet, "/api/artists/1", nil)
	w := httptest.NewRecorder()

	handler.handleArtistByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var found models.Artist
	if err := json.NewDecoder(w.Body).Decode(&found); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if found.ID != artist.ID {
		t.Errorf("expected ID %d, got %d", artist.ID, found.ID)
	}
}

func TestHandleArtistByID_NotFound(t *testing.T) {
	handler, _ := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/artists/999", nil)
	w := httptest.NewRecorder()

	handler.handleArtistByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestHandleArtistByID_PUT(t *testing.T) {
	handler, store := setupTestHandler()
	artist := store.Create(models.Artist{Name: "Original"})

	updated := models.Artist{Name: "Updated"}
	body, _ := json.Marshal(updated)

	req := httptest.NewRequest(http.MethodPut, "/api/artists/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleArtistByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var result models.Artist
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result.Name != "Updated" {
		t.Errorf("expected name 'Updated', got %s", result.Name)
	}
	if result.ID != artist.ID {
		t.Errorf("expected ID to remain %d, got %d", artist.ID, result.ID)
	}
}

func TestHandleArtistByID_DELETE(t *testing.T) {
	handler, store := setupTestHandler()
	store.Create(models.Artist{Name: "Test"})

	req := httptest.NewRequest(http.MethodDelete, "/api/artists/1", nil)
	w := httptest.NewRecorder()

	handler.handleArtistByID(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}

	// Verify deleted
	_, ok := store.GetByID(1)
	if ok {
		t.Error("artist should be deleted")
	}
}

func TestHandleSearch(t *testing.T) {
	handler, store := setupTestHandler()
	store.Create(models.Artist{Name: "Queen"})
	store.Create(models.Artist{Name: "The Beatles"})

	searchBody := map[string]string{"query": "queen"}
	body, _ := json.Marshal(searchBody)

	req := httptest.NewRequest(http.MethodPost, "/api/search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleSearch(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var results []models.Artist
	if err := json.NewDecoder(w.Body).Decode(&results); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].Name != "Queen" {
		t.Errorf("expected 'Queen', got %s", results[0].Name)
	}
}

func TestHandleArtists_InvalidMethod(t *testing.T) {
	handler, _ := setupTestHandler()

	req := httptest.NewRequest(http.MethodPatch, "/api/artists", nil)
	w := httptest.NewRecorder()

	handler.handleArtists(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleArtistByID_InvalidID(t *testing.T) {
	handler, _ := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/artists/invalid", nil)
	w := httptest.NewRecorder()

	handler.handleArtistByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHandleArtists_POST_InvalidJSON(t *testing.T) {
	handler, _ := setupTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/artists", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleArtists(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHandleArtistByID_PUT_InvalidJSON(t *testing.T) {
	handler, store := setupTestHandler()
	store.Create(models.Artist{Name: "Test"})

	req := httptest.NewRequest(http.MethodPut, "/api/artists/1", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleArtistByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHandleArtistByID_PUT_NotFound(t *testing.T) {
	handler, _ := setupTestHandler()

	updated := models.Artist{Name: "Updated"}
	body, _ := json.Marshal(updated)

	req := httptest.NewRequest(http.MethodPut, "/api/artists/999", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleArtistByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestHandleArtistByID_DELETE_NotFound(t *testing.T) {
	handler, _ := setupTestHandler()

	req := httptest.NewRequest(http.MethodDelete, "/api/artists/999", nil)
	w := httptest.NewRecorder()

	handler.handleArtistByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestHandleSearch_InvalidMethod(t *testing.T) {
	handler, _ := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/search", nil)
	w := httptest.NewRecorder()

	handler.handleSearch(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleSearch_InvalidJSON(t *testing.T) {
	handler, _ := setupTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/search", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleSearch(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHandleArtistByID_InvalidMethod(t *testing.T) {
	handler, _ := setupTestHandler()

	req := httptest.NewRequest(http.MethodPatch, "/api/artists/1", nil)
	w := httptest.NewRecorder()

	handler.handleArtistByID(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestRegisterRoutes(t *testing.T) {
	handler, store := setupTestHandler()
	store.Create(models.Artist{Name: "Test"})
	mux := http.NewServeMux()

	handler.RegisterRoutes(mux)

	// Test that routes are registered and respond correctly
	tests := []struct {
		method       string
		path         string
		expectedCode int
	}{
		{http.MethodGet, "/api/artists", http.StatusOK},
		{http.MethodPost, "/api/search", http.StatusBadRequest}, // No body
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code == http.StatusNotFound {
			t.Errorf("%s %s: route not registered (got 404)", tt.method, tt.path)
		}
	}
}

func TestNewArtistHandler(t *testing.T) {
	store := db.NewArtistStore()
	service := services.NewArtistService(store)
	searchChan := make(chan events.SearchEvent, 10)

	handler := NewArtistHandler(service, searchChan)

	if handler == nil {
		t.Fatal("NewArtistHandler returned nil")
	}
	if handler.service == nil {
		t.Error("handler service is nil")
	}
	if handler.searchChan == nil {
		t.Error("handler searchChan is nil")
	}
}
