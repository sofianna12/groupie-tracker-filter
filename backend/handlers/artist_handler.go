package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"groupie_tracker/backend/events"
	"groupie_tracker/backend/models"
	"groupie_tracker/backend/services"
)

type ArtistHandler struct {
	service    *services.ArtistService
	searchChan chan events.SearchEvent
}

func NewArtistHandler(s *services.ArtistService, sc chan events.SearchEvent) *ArtistHandler {
	return &ArtistHandler{s, sc}
}

func (h *ArtistHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/artists", h.handleArtists)
	mux.HandleFunc("/api/artists/filter", h.handleFilter)
	mux.HandleFunc("/api/artists/", h.handleArtistByID)
	mux.HandleFunc("/api/search", h.handleSearch)
}

func (h *ArtistHandler) handleArtists(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		json.NewEncoder(w).Encode(h.service.GetAll())
	case http.MethodPost:
		var a models.Artist
		if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		created := h.service.Create(a)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(created)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ArtistHandler) handleArtistByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := strings.TrimPrefix(r.URL.Path, "/api/artists/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		a, ok := h.service.GetByID(id)
		if !ok {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(a)

	case http.MethodPut:
		var a models.Artist
		if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		updated, ok := h.service.Update(id, a)
		if !ok {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(updated)

	case http.MethodDelete:
		if !h.service.Delete(id) {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ArtistHandler) handleFilter(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.FilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	results := h.service.Filter(req)
	if results == nil {
		results = []models.Artist{}
	}
	json.NewEncoder(w).Encode(results)
}

func (h *ArtistHandler) handleSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Query string `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	respChan := make(chan []models.Artist)
	h.searchChan <- events.SearchEvent{
		Query: body.Query,
		Resp:  respChan,
	}

	results := <-respChan
	json.NewEncoder(w).Encode(results)
}
