package models

import (
	"encoding/json"
	"testing"
)

func TestArtist_JSONMarshaling(t *testing.T) {
	artist := Artist{
		ID:           1,
		Name:         "Test Band",
		Image:        "http://example.com/image.jpg",
		FirstAlbum:   "01-01-2020",
		CreationDate: 2020,
		Members:      []string{"Member1", "Member2"},
		Locations:    []string{"New York", "London"},
		Dates:        []string{"01-01-2020", "02-02-2020"},
	}

	// Marshal
	data, err := json.Marshal(artist)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Unmarshal
	var unmarshaled Artist
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify
	if unmarshaled.ID != artist.ID {
		t.Errorf("expected ID %d, got %d", artist.ID, unmarshaled.ID)
	}
	if unmarshaled.Name != artist.Name {
		t.Errorf("expected name %s, got %s", artist.Name, unmarshaled.Name)
	}
	if len(unmarshaled.Members) != len(artist.Members) {
		t.Errorf("expected %d members, got %d", len(artist.Members), len(unmarshaled.Members))
	}
}

func TestArtist_EmptyFields(t *testing.T) {
	artist := Artist{
		ID:   1,
		Name: "Minimal Band",
	}

	data, err := json.Marshal(artist)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled Artist
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.Name != "Minimal Band" {
		t.Errorf("expected name 'Minimal Band', got %s", unmarshaled.Name)
	}
	if unmarshaled.Members != nil {
		t.Error("expected nil members slice")
	}
}
