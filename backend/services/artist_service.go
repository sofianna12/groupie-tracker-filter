package services

import (
	"groupie_tracker/backend/models"
	"groupie_tracker/db"
)

type ArtistService struct {
	store db.Store
}

func NewArtistService(store db.Store) *ArtistService {
	return &ArtistService{store: store}
}

func (s *ArtistService) Create(a models.Artist) models.Artist {
	return s.store.Create(a)
}

func (s *ArtistService) GetAll() []models.Artist {
	return s.store.GetAll()
}

func (s *ArtistService) GetByID(id int) (models.Artist, bool) {
	return s.store.GetByID(id)
}

func (s *ArtistService) Update(id int, a models.Artist) (models.Artist, bool) {
	return s.store.Update(id, a)
}

func (s *ArtistService) Delete(id int) bool {
	return s.store.Delete(id)
}

func (s *ArtistService) Filter(req models.FilterRequest) []models.Artist {
	return s.store.Filter(req)
}
