package db

import (
	"strconv"
	"strings"
	"sync"

	"groupie_tracker/backend/models"
)

type ArtistStore struct {
	mu      sync.RWMutex
	artists map[int]models.Artist
	nextID  int
}

func NewArtistStore() *ArtistStore {
	return &ArtistStore{
		artists: make(map[int]models.Artist),
		nextID:  1,
	}
}

func (s *ArtistStore) Create(artist models.Artist) models.Artist {
	s.mu.Lock()
	defer s.mu.Unlock()

	if artist.ID == 0 {
		artist.ID = s.nextID
	}
	if artist.ID >= s.nextID {
		s.nextID = artist.ID + 1
	}
	s.artists[artist.ID] = artist

	return artist
}

func (s *ArtistStore) GetAll() []models.Artist {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var list []models.Artist
	for _, a := range s.artists {
		list = append(list, a)
	}
	return list
}

func (s *ArtistStore) GetByID(id int) (models.Artist, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	artist, ok := s.artists[id]
	return artist, ok
}

func (s *ArtistStore) Update(id int, artist models.Artist) (models.Artist, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.artists[id]
	if !ok {
		return models.Artist{}, false
	}

	artist.ID = id
	s.artists[id] = artist
	return artist, true
}

func (s *ArtistStore) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.artists[id]
	if !ok {
		return false
	}

	delete(s.artists, id)
	return true
}

func (s *ArtistStore) Search(query string) []models.Artist {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []models.Artist
	query = strings.ToLower(query)

	for _, a := range s.artists {
		if strings.Contains(strings.ToLower(a.Name), query) {
			results = append(results, a)
		}
	}

	return results
}

// Filter returns all artists that satisfy every active criterion in req.
// A criterion is inactive when its zero value is used (0 for ints, empty slice for slices).
func (s *ArtistStore) Filter(req models.FilterRequest) []models.Artist {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []models.Artist
	for _, a := range s.artists {
		if !matchesFilter(a, req) {
			continue
		}
		results = append(results, a)
	}
	return results
}

// matchesFilter returns true when the artist satisfies all active criteria.
func matchesFilter(a models.Artist, req models.FilterRequest) bool {
	// --- Range filter: creation date ---
	if req.CreationDateFrom > 0 && a.CreationDate < req.CreationDateFrom {
		return false
	}
	if req.CreationDateTo > 0 && a.CreationDate > req.CreationDateTo {
		return false
	}

	// --- Range filter: first album year ---
	if req.FirstAlbumFrom > 0 || req.FirstAlbumTo > 0 {
		year := parseFirstAlbumYear(a.FirstAlbum)
		if req.FirstAlbumFrom > 0 && year < req.FirstAlbumFrom {
			return false
		}
		if req.FirstAlbumTo > 0 && year > req.FirstAlbumTo {
			return false
		}
	}

	// --- Checkbox filter: number of members ---
	if len(req.MembersCount) > 0 {
		n := len(a.Members)
		found := false
		for _, c := range req.MembersCount {
			if n == c {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// --- Checkbox filter: concert locations (hierarchical contains match) ---
	// e.g. filter "washington, usa" matches artist location "seattle, washington, usa"
	if len(req.Locations) > 0 {
		matched := false
		for _, filterLoc := range req.Locations {
			fl := strings.ToLower(strings.TrimSpace(filterLoc))
			for _, artistLoc := range a.Locations {
				al := strings.ToLower(strings.TrimSpace(artistLoc))
				if strings.Contains(al, fl) || strings.Contains(fl, al) {
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

// parseFirstAlbumYear extracts the 4-digit year from strings like "02-01-1966" or "1966".
func parseFirstAlbumYear(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	// Format DD-MM-YYYY: year is the last segment
	parts := strings.Split(s, "-")
	if len(parts) == 3 {
		y, err := strconv.Atoi(parts[2])
		if err == nil {
			return y
		}
	}
	// Fallback: look for a 4-digit number anywhere in the string
	for _, p := range parts {
		if len(p) == 4 {
			y, err := strconv.Atoi(p)
			if err == nil {
				return y
			}
		}
	}
	return 0
}
