package db

import (
	"testing"

	"groupie_tracker/backend/models"
)

func TestNewArtistStore(t *testing.T) {
	store := NewArtistStore()
	if store == nil {
		t.Fatal("NewArtistStore returned nil")
	}
	if store.artists == nil {
		t.Error("artists map not initialized")
	}
	if store.nextID != 1 {
		t.Errorf("expected nextID to be 1, got %d", store.nextID)
	}
}

func TestCreate(t *testing.T) {
	store := NewArtistStore()
	artist := models.Artist{
		Name:         "Test Band",
		Image:        "http://example.com/image.jpg",
		Members:      []string{"Member1", "Member2"},
		CreationDate: 2020,
		FirstAlbum:   "01-01-2020",
	}

	created := store.Create(artist)

	if created.ID != 1 {
		t.Errorf("expected ID 1, got %d", created.ID)
	}
	if created.Name != artist.Name {
		t.Errorf("expected name %s, got %s", artist.Name, created.Name)
	}
	if store.nextID != 2 {
		t.Errorf("expected nextID to be 2, got %d", store.nextID)
	}
}

func TestCreate_PreservesOriginalID(t *testing.T) {
	store := NewArtistStore()
	artist := models.Artist{ID: 42, Name: "API Band"}

	created := store.Create(artist)

	if created.ID != 42 {
		t.Errorf("expected ID 42 to be preserved, got %d", created.ID)
	}
	if store.nextID != 43 {
		t.Errorf("expected nextID to be 43, got %d", store.nextID)
	}

	next := store.Create(models.Artist{Name: "Another Band"})
	if next.ID != 43 {
		t.Errorf("expected next ID 43, got %d", next.ID)
	}
}

func TestGetAll(t *testing.T) {
	store := NewArtistStore()

	// Empty store
	all := store.GetAll()
	if len(all) != 0 {
		t.Errorf("expected 0 artists, got %d", len(all))
	}

	// Add artists
	store.Create(models.Artist{Name: "Band1"})
	store.Create(models.Artist{Name: "Band2"})

	all = store.GetAll()
	if len(all) != 2 {
		t.Errorf("expected 2 artists, got %d", len(all))
	}
}

func TestGetByID(t *testing.T) {
	store := NewArtistStore()
	artist := store.Create(models.Artist{Name: "Test Band"})

	// Existing artist
	found, ok := store.GetByID(artist.ID)
	if !ok {
		t.Error("expected to find artist")
	}
	if found.Name != "Test Band" {
		t.Errorf("expected name 'Test Band', got %s", found.Name)
	}

	// Non-existing artist
	_, ok = store.GetByID(999)
	if ok {
		t.Error("expected not to find artist with ID 999")
	}
}

func TestUpdate(t *testing.T) {
	store := NewArtistStore()
	artist := store.Create(models.Artist{Name: "Original Name"})

	// Update existing
	updated, ok := store.Update(artist.ID, models.Artist{Name: "Updated Name"})
	if !ok {
		t.Error("expected update to succeed")
	}
	if updated.Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got %s", updated.Name)
	}

	// Update non-existing
	_, ok = store.Update(999, models.Artist{Name: "Test"})
	if ok {
		t.Error("expected update to fail for non-existing artist")
	}
}

func TestDelete(t *testing.T) {
	store := NewArtistStore()
	artist := store.Create(models.Artist{Name: "Test Band"})

	// Delete existing
	ok := store.Delete(artist.ID)
	if !ok {
		t.Error("expected delete to succeed")
	}

	// Verify deleted
	_, found := store.GetByID(artist.ID)
	if found {
		t.Error("artist should be deleted")
	}

	// Delete non-existing
	ok = store.Delete(999)
	if ok {
		t.Error("expected delete to fail for non-existing artist")
	}
}

func TestSearch(t *testing.T) {
	store := NewArtistStore()
	store.Create(models.Artist{Name: "Queen"})
	store.Create(models.Artist{Name: "The Beatles"})
	store.Create(models.Artist{Name: "Pink Floyd"})

	tests := []struct {
		query    string
		expected int
	}{
		{"queen", 1},
		{"QUEEN", 1},
		{"the", 1},
		{"pink", 1},
		{"xyz", 0},
		{"", 3},
	}

	for _, tt := range tests {
		results := store.Search(tt.query)
		if len(results) != tt.expected {
			t.Errorf("search(%q): expected %d results, got %d", tt.query, tt.expected, len(results))
		}
	}
}

func TestFilter_CreationDateRange(t *testing.T) {
	store := NewArtistStore()
	store.Create(models.Artist{Name: "Old Band", CreationDate: 1960})
	store.Create(models.Artist{Name: "Mid Band", CreationDate: 1990})
	store.Create(models.Artist{Name: "New Band", CreationDate: 2010})

	tests := []struct {
		from, to int
		expected int
	}{
		{1980, 2000, 1}, // only Mid Band
		{1950, 2020, 3}, // all three
		{2005, 0, 1},    // New Band (no upper bound)
		{0, 1970, 1},    // Old Band (no lower bound)
		{0, 0, 3},       // no filter
	}

	for _, tt := range tests {
		req := models.FilterRequest{CreationDateFrom: tt.from, CreationDateTo: tt.to}
		results := store.Filter(req)
		if len(results) != tt.expected {
			t.Errorf("creationDate[%d-%d]: expected %d, got %d", tt.from, tt.to, tt.expected, len(results))
		}
	}
}

func TestFilter_FirstAlbumRange(t *testing.T) {
	store := NewArtistStore()
	store.Create(models.Artist{Name: "A", FirstAlbum: "01-06-1965"})
	store.Create(models.Artist{Name: "B", FirstAlbum: "15-03-1991"})
	store.Create(models.Artist{Name: "C", FirstAlbum: "20-11-2005"})

	tests := []struct {
		from, to int
		expected int
	}{
		{1990, 2000, 1}, // only B (1991)
		{1960, 2010, 3}, // all
		{2000, 0, 1},    // C (no upper bound)
		{0, 1970, 1},    // A (no lower bound)
	}

	for _, tt := range tests {
		req := models.FilterRequest{FirstAlbumFrom: tt.from, FirstAlbumTo: tt.to}
		results := store.Filter(req)
		if len(results) != tt.expected {
			t.Errorf("firstAlbum[%d-%d]: expected %d, got %d", tt.from, tt.to, tt.expected, len(results))
		}
	}
}

func TestFilter_MembersCount(t *testing.T) {
	store := NewArtistStore()
	store.Create(models.Artist{Name: "Solo", Members: []string{"Alice"}})
	store.Create(models.Artist{Name: "Duo", Members: []string{"A", "B"}})
	store.Create(models.Artist{Name: "Quartet", Members: []string{"A", "B", "C", "D"}})

	tests := []struct {
		counts   []int
		expected int
	}{
		{[]int{1}, 1},    // solo only
		{[]int{2}, 1},    // duo only
		{[]int{1, 2}, 2}, // solo + duo
		{[]int{4}, 1},    // quartet only
		{[]int{5}, 0},    // none
		{nil, 3},         // no filter
	}

	for _, tt := range tests {
		req := models.FilterRequest{MembersCount: tt.counts}
		results := store.Filter(req)
		if len(results) != tt.expected {
			t.Errorf("membersCount=%v: expected %d, got %d", tt.counts, tt.expected, len(results))
		}
	}
}

func TestFilter_Locations(t *testing.T) {
	store := NewArtistStore()
	store.Create(models.Artist{
		Name:      "US Band",
		Locations: []string{"seattle-washington-usa", "new_york-usa"},
	})
	store.Create(models.Artist{
		Name:      "UK Band",
		Locations: []string{"london-uk", "manchester-uk"},
	})
	store.Create(models.Artist{
		Name:      "World Band",
		Locations: []string{"paris-france", "tokyo-japan"},
	})

	tests := []struct {
		locs     []string
		expected int
	}{
		{[]string{"usa"}, 1},                              // US Band only
		{[]string{"uk"}, 1},                               // UK Band only
		{[]string{"usa", "uk"}, 2},                        // US + UK bands
		{[]string{"washington"}, 1},                       // hierarchical match
		{[]string{"london-uk"}, 1},                        // exact match
		{[]string{"germany"}, 0},                          // no match
		{nil, 3},                                          // no filter
	}

	for _, tt := range tests {
		req := models.FilterRequest{Locations: tt.locs}
		results := store.Filter(req)
		if len(results) != tt.expected {
			t.Errorf("locations=%v: expected %d, got %d", tt.locs, tt.expected, len(results))
		}
	}
}

func TestFilter_Combined(t *testing.T) {
	store := NewArtistStore()
	store.Create(models.Artist{
		Name:         "Match",
		CreationDate: 1990,
		FirstAlbum:   "01-01-1992",
		Members:      []string{"A", "B", "C"},
		Locations:    []string{"new_york-usa"},
	})
	store.Create(models.Artist{
		Name:         "No Match",
		CreationDate: 2005,
		FirstAlbum:   "01-01-2006",
		Members:      []string{"X"},
		Locations:    []string{"berlin-germany"},
	})

	req := models.FilterRequest{
		CreationDateFrom: 1985,
		CreationDateTo:   2000,
		FirstAlbumFrom:   1990,
		FirstAlbumTo:     2000,
		MembersCount:     []int{3},
		Locations:        []string{"usa"},
	}
	results := store.Filter(req)
	if len(results) != 1 {
		t.Fatalf("combined filter: expected 1, got %d", len(results))
	}
	if results[0].Name != "Match" {
		t.Errorf("expected 'Match', got %s", results[0].Name)
	}
}

func TestParseFirstAlbumYear(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"02-01-1966", 1966},
		{"15-11-2003", 2003},
		{"1990", 1990},
		{"", 0},
		{"invalid", 0},
	}

	for _, tt := range tests {
		got := parseFirstAlbumYear(tt.input)
		if got != tt.expected {
			t.Errorf("parseFirstAlbumYear(%q): expected %d, got %d", tt.input, tt.expected, got)
		}
	}
}

func TestConcurrentAccess(t *testing.T) {
	store := NewArtistStore()
	done := make(chan bool)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(n int) {
			store.Create(models.Artist{Name: "Band"})
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	all := store.GetAll()
	if len(all) != 10 {
		t.Errorf("expected 10 artists after concurrent writes, got %d", len(all))
	}
}
