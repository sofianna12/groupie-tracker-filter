package events

import (
	"testing"
	"time"

	"groupie_tracker/backend/models"
	"groupie_tracker/db"
)

func TestSearchWorker(t *testing.T) {
	store := db.NewArtistStore()
	store.Create(models.Artist{Name: "Queen"})
	store.Create(models.Artist{Name: "The Beatles"})

	searchChan := make(chan SearchEvent, 10)
	go StartSearchWorker(store, searchChan)

	// Test search
	respChan := make(chan []models.Artist)
	searchChan <- SearchEvent{
		Query: "queen",
		Resp:  respChan,
	}

	select {
	case results := <-respChan:
		if len(results) != 1 {
			t.Errorf("expected 1 result, got %d", len(results))
		}
		if results[0].Name != "Queen" {
			t.Errorf("expected 'Queen', got %s", results[0].Name)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for search results")
	}
}

func TestSearchWorker_EmptyQuery(t *testing.T) {
	store := db.NewArtistStore()
	store.Create(models.Artist{Name: "Queen"})

	searchChan := make(chan SearchEvent, 10)
	go StartSearchWorker(store, searchChan)

	respChan := make(chan []models.Artist)
	searchChan <- SearchEvent{
		Query: "",
		Resp:  respChan,
	}

	select {
	case results := <-respChan:
		if len(results) != 1 {
			t.Errorf("expected 1 result for empty query, got %d", len(results))
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for search results")
	}
}

func TestSearchWorker_NoResults(t *testing.T) {
	store := db.NewArtistStore()
	store.Create(models.Artist{Name: "Queen"})

	searchChan := make(chan SearchEvent, 10)
	go StartSearchWorker(store, searchChan)

	respChan := make(chan []models.Artist)
	searchChan <- SearchEvent{
		Query: "xyz",
		Resp:  respChan,
	}

	select {
	case results := <-respChan:
		if len(results) != 0 {
			t.Errorf("expected 0 results, got %d", len(results))
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for search results")
	}
}

func TestSearchWorker_MultipleRequests(t *testing.T) {
	store := db.NewArtistStore()
	store.Create(models.Artist{Name: "Queen"})
	store.Create(models.Artist{Name: "The Beatles"})
	store.Create(models.Artist{Name: "Pink Floyd"})

	searchChan := make(chan SearchEvent, 10)
	go StartSearchWorker(store, searchChan)

	// Send multiple search requests
	queries := []string{"queen", "beatles", "pink"}
	for _, q := range queries {
		respChan := make(chan []models.Artist)
		searchChan <- SearchEvent{
			Query: q,
			Resp:  respChan,
		}

		select {
		case results := <-respChan:
			if len(results) != 1 {
				t.Errorf("query %q: expected 1 result, got %d", q, len(results))
			}
		case <-time.After(1 * time.Second):
			t.Fatalf("timeout waiting for search results for query %q", q)
		}
	}
}
