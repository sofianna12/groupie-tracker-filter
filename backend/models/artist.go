package models

type Artist struct {
	ID             int                 `json:"id"`
	Name           string              `json:"name"`
	Image          string              `json:"image"`
	FirstAlbum     string              `json:"firstAlbum"`
	CreationDate   int                 `json:"creationDate"`
	Members        []string            `json:"members"`
	Locations      []string            `json:"locations"`
	Dates          []string            `json:"dates"`
	DatesLocations map[string][]string `json:"datesLocations"`
}
