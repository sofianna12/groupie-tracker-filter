package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"groupie_tracker/backend/models"
	"groupie_tracker/db"
)

const baseURL = "https://groupietrackers.herokuapp.com/api"

type artistsResponse []struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	Image        string   `json:"image"`
	FirstAlbum   string   `json:"firstAlbum"`
	CreationDate int      `json:"creationDate"`
	Members      []string `json:"members"`
}

type locationsResponse struct {
	Index []struct {
		ID        int      `json:"id"`
		Locations []string `json:"locations"`
	} `json:"index"`
}

type datesResponse struct {
	Index []struct {
		ID    int      `json:"id"`
		Dates []string `json:"dates"`
	} `json:"index"`
}

type relationResponse struct {
	Index []struct {
		ID             int                 `json:"id"`
		DatesLocations map[string][]string `json:"datesLocations"`
	} `json:"index"`
}

func FetchAndLoad(store db.Store) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var (
		artists   artistsResponse
		locations locationsResponse
		dates     datesResponse
		relations relationResponse
		wg        sync.WaitGroup
		errChan   = make(chan error, 4)
	)

	wg.Add(4)

	go fetch(client, baseURL+"/artists", &artists, &wg, errChan)
	go fetch(client, baseURL+"/locations", &locations, &wg, errChan)
	go fetch(client, baseURL+"/dates", &dates, &wg, errChan)
	go fetch(client, baseURL+"/relation", &relations, &wg, errChan)

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	// Create lookup maps
	locationMap := make(map[int][]string)
	for _, l := range locations.Index {
		locationMap[l.ID] = l.Locations
	}

	dateMap := make(map[int][]string)
	for _, d := range dates.Index {
		dateMap[d.ID] = d.Dates
	}

	relationMap := make(map[int]map[string][]string)
	for _, r := range relations.Index {
		relationMap[r.ID] = r.DatesLocations
	}

	// Build unified Artist models
	for _, a := range artists {
		artist := models.Artist{
			ID:             a.ID,
			Name:           a.Name,
			Image:          a.Image,
			FirstAlbum:     a.FirstAlbum,
			CreationDate:   a.CreationDate,
			Members:        a.Members,
			Locations:      locationMap[a.ID],
			Dates:          dateMap[a.ID],
			DatesLocations: relationMap[a.ID],
		}

		store.Create(artist)
	}

	return nil
}

func fetch(client *http.Client, url string, target interface{}, wg *sync.WaitGroup, errChan chan error) {
	defer wg.Done()

	resp, err := client.Get(url)
	if err != nil {
		errChan <- fmt.Errorf("failed to fetch %s: %w", url, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errChan <- fmt.Errorf("non-200 response from %s", url)
		return
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		errChan <- fmt.Errorf("decode error from %s: %w", url, err)
		return
	}

	errChan <- nil
}
