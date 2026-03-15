package models

// FilterRequest holds all supported filter criteria.
// Zero values are treated as "no filter applied":
//   - range filters: 0 means unbounded on that side
//   - slice filters: empty slice means no restriction
type FilterRequest struct {
	// Range filter: creation year (e.g. 1950–2023)
	CreationDateFrom int `json:"creationDateFrom"`
	CreationDateTo   int `json:"creationDateTo"`

	// Range filter: first album release year
	FirstAlbumFrom int `json:"firstAlbumFrom"`
	FirstAlbumTo   int `json:"firstAlbumTo"`

	// Checkbox filter: allowed number-of-members values (e.g. [1,2,4])
	MembersCount []int `json:"membersCount"`

	// Checkbox filter: concert location substrings (e.g. ["usa", "washington"])
	// Matching is case-insensitive and hierarchical:
	//   filter "washington, usa" matches artist location "seattle, washington, usa"
	Locations []string `json:"locations"`
}