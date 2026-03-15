package db

import "groupie_tracker/backend/models"

type Store interface {
	Create(artist models.Artist) models.Artist
	GetAll() []models.Artist
	GetByID(id int) (models.Artist, bool)
	Update(id int, artist models.Artist) (models.Artist, bool)
	Delete(id int) bool
	Search(query string) []models.Artist
	Filter(req models.FilterRequest) []models.Artist
}
