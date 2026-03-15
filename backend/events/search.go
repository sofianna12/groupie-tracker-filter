package events

import (
	"groupie_tracker/backend/models"
	"groupie_tracker/db"
)

type SearchEvent struct {
	Query string
	Resp  chan []models.Artist
}

func StartSearchWorker(store db.Store, events chan SearchEvent) {
	for event := range events {
		results := store.Search(event.Query)
		event.Resp <- results
	}
}
