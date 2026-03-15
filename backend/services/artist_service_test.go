package services

import (
	"testing"

	"groupie_tracker/backend/models"
	"groupie_tracker/db"
)

func TestNewArtistService(t *testing.T) {
	store := db.NewArtistStore()
	service := NewArtistService(store)

	if service == nil {
		t.Fatal("NewArtistService returned nil")
	}
}

func TestArtistService_Create(t *testing.T) {
	store := db.NewArtistStore()
	service := NewArtistService(store)

	artist := models.Artist{
		Name:    "Test Band",
		Members: []string{"Member1"},
	}

	created := service.Create(artist)
	if created.ID == 0 {
		t.Error("expected artist to have ID assigned")
	}
	if created.Name != artist.Name {
		t.Errorf("expected name %s, got %s", artist.Name, created.Name)
	}
}

func TestArtistService_GetAll(t *testing.T) {
	store := db.NewArtistStore()
	service := NewArtistService(store)

	service.Create(models.Artist{Name: "Band1"})
	service.Create(models.Artist{Name: "Band2"})

	all := service.GetAll()
	if len(all) != 2 {
		t.Errorf("expected 2 artists, got %d", len(all))
	}
}

func TestArtistService_GetByID(t *testing.T) {
	store := db.NewArtistStore()
	service := NewArtistService(store)

	created := service.Create(models.Artist{Name: "Test Band"})

	found, ok := service.GetByID(created.ID)
	if !ok {
		t.Error("expected to find artist")
	}
	if found.Name != "Test Band" {
		t.Errorf("expected name 'Test Band', got %s", found.Name)
	}

	_, ok = service.GetByID(999)
	if ok {
		t.Error("expected not to find non-existing artist")
	}
}

func TestArtistService_Update(t *testing.T) {
	store := db.NewArtistStore()
	service := NewArtistService(store)

	created := service.Create(models.Artist{Name: "Original"})

	updated, ok := service.Update(created.ID, models.Artist{Name: "Updated"})
	if !ok {
		t.Error("expected update to succeed")
	}
	if updated.Name != "Updated" {
		t.Errorf("expected name 'Updated', got %s", updated.Name)
	}
}

func TestArtistService_Delete(t *testing.T) {
	store := db.NewArtistStore()
	service := NewArtistService(store)

	created := service.Create(models.Artist{Name: "Test"})

	ok := service.Delete(created.ID)
	if !ok {
		t.Error("expected delete to succeed")
	}

	_, found := service.GetByID(created.ID)
	if found {
		t.Error("artist should be deleted")
	}
}

func TestArtistService_Filter_CreationDateRange(t *testing.T) {
	store := db.NewArtistStore()
	service := NewArtistService(store)

	service.Create(models.Artist{Name: "Old", CreationDate: 1965})
	service.Create(models.Artist{Name: "New", CreationDate: 2010})

	results := service.Filter(models.FilterRequest{CreationDateFrom: 2000, CreationDateTo: 2020})
	if len(results) != 1 || results[0].Name != "New" {
		t.Errorf("expected 1 artist 'New', got %d results", len(results))
	}
}

func TestArtistService_Filter_MembersCheckbox(t *testing.T) {
	store := db.NewArtistStore()
	service := NewArtistService(store)

	service.Create(models.Artist{Name: "Solo", Members: []string{"A"}})
	service.Create(models.Artist{Name: "Duo", Members: []string{"A", "B"}})

	results := service.Filter(models.FilterRequest{MembersCount: []int{1}})
	if len(results) != 1 || results[0].Name != "Solo" {
		t.Errorf("expected 1 result 'Solo', got %d", len(results))
	}

	results = service.Filter(models.FilterRequest{MembersCount: []int{1, 2}})
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestArtistService_Filter_LocationsCheckbox(t *testing.T) {
	store := db.NewArtistStore()
	service := NewArtistService(store)

	service.Create(models.Artist{Name: "USBand", Locations: []string{"new_york-usa"}})
	service.Create(models.Artist{Name: "UKBand", Locations: []string{"london-uk"}})

	results := service.Filter(models.FilterRequest{Locations: []string{"usa"}})
	if len(results) != 1 || results[0].Name != "USBand" {
		t.Errorf("expected 1 result 'USBand', got %d", len(results))
	}
}

func TestArtistService_Filter_EmptyRequest(t *testing.T) {
	store := db.NewArtistStore()
	service := NewArtistService(store)

	service.Create(models.Artist{Name: "A"})
	service.Create(models.Artist{Name: "B"})

	results := service.Filter(models.FilterRequest{})
	if len(results) != 2 {
		t.Errorf("empty filter should return all, got %d", len(results))
	}
}
