package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	_ "github.com/lib/pq"

	"groupie_tracker/backend/models"
)

// PostgresStore implements Store backed by PostgreSQL.
type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(connStr string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	s := &PostgresStore{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *PostgresStore) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS artists (
			id              SERIAL PRIMARY KEY,
			name            TEXT    NOT NULL DEFAULT '',
			image           TEXT    NOT NULL DEFAULT '',
			first_album     TEXT    NOT NULL DEFAULT '',
			creation_date   INT     NOT NULL DEFAULT 0,
			members         JSONB   NOT NULL DEFAULT '[]',
			locations       JSONB   NOT NULL DEFAULT '[]',
			dates           JSONB   NOT NULL DEFAULT '[]',
			dates_locations JSONB   NOT NULL DEFAULT '{}'
		)
	`)
	return err
}

// --- helpers ---

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func scanArtist(row interface {
	Scan(dest ...any) error
}) (models.Artist, error) {
	var a models.Artist
	var membersJSON, locationsJSON, datesJSON, datesLocJSON []byte

	if err := row.Scan(
		&a.ID, &a.Name, &a.Image, &a.FirstAlbum, &a.CreationDate,
		&membersJSON, &locationsJSON, &datesJSON, &datesLocJSON,
	); err != nil {
		return a, err
	}

	_ = json.Unmarshal(membersJSON, &a.Members)
	_ = json.Unmarshal(locationsJSON, &a.Locations)
	_ = json.Unmarshal(datesJSON, &a.Dates)
	_ = json.Unmarshal(datesLocJSON, &a.DatesLocations)
	return a, nil
}

// --- Store interface ---

func (s *PostgresStore) Create(artist models.Artist) models.Artist {
	row := s.db.QueryRow(`
		INSERT INTO artists (name, image, first_album, creation_date, members, locations, dates, dates_locations)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, name, image, first_album, creation_date, members, locations, dates, dates_locations`,
		artist.Name, artist.Image, artist.FirstAlbum, artist.CreationDate,
		mustJSON(artist.Members), mustJSON(artist.Locations),
		mustJSON(artist.Dates), mustJSON(artist.DatesLocations),
	)
	created, _ := scanArtist(row)
	return created
}

func (s *PostgresStore) GetAll() []models.Artist {
	rows, err := s.db.Query(`
		SELECT id, name, image, first_album, creation_date, members, locations, dates, dates_locations
		FROM artists ORDER BY id`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var list []models.Artist
	for rows.Next() {
		a, err := scanArtist(rows)
		if err == nil {
			list = append(list, a)
		}
	}
	return list
}

func (s *PostgresStore) GetByID(id int) (models.Artist, bool) {
	row := s.db.QueryRow(`
		SELECT id, name, image, first_album, creation_date, members, locations, dates, dates_locations
		FROM artists WHERE id = $1`, id)
	a, err := scanArtist(row)
	if err != nil {
		return models.Artist{}, false
	}
	return a, true
}

func (s *PostgresStore) Update(id int, artist models.Artist) (models.Artist, bool) {
	row := s.db.QueryRow(`
		UPDATE artists SET
			name=$1, image=$2, first_album=$3, creation_date=$4,
			members=$5, locations=$6, dates=$7, dates_locations=$8
		WHERE id=$9
		RETURNING id, name, image, first_album, creation_date, members, locations, dates, dates_locations`,
		artist.Name, artist.Image, artist.FirstAlbum, artist.CreationDate,
		mustJSON(artist.Members), mustJSON(artist.Locations),
		mustJSON(artist.Dates), mustJSON(artist.DatesLocations),
		id,
	)
	updated, err := scanArtist(row)
	if err != nil {
		return models.Artist{}, false
	}
	return updated, true
}

func (s *PostgresStore) Delete(id int) bool {
	res, err := s.db.Exec(`DELETE FROM artists WHERE id=$1`, id)
	if err != nil {
		return false
	}
	n, _ := res.RowsAffected()
	return n > 0
}

func (s *PostgresStore) Search(query string) []models.Artist {
	rows, err := s.db.Query(`
		SELECT id, name, image, first_album, creation_date, members, locations, dates, dates_locations
		FROM artists WHERE LOWER(name) LIKE $1 ORDER BY id`,
		"%"+strings.ToLower(query)+"%",
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var list []models.Artist
	for rows.Next() {
		a, err := scanArtist(rows)
		if err == nil {
			list = append(list, a)
		}
	}
	return list
}

// Filter fetches all artists and applies criteria in Go (dataset is small).
func (s *PostgresStore) Filter(req models.FilterRequest) []models.Artist {
	all := s.GetAll()
	var results []models.Artist
	for _, a := range all {
		if matchesFilter(a, req) {
			results = append(results, a)
		}
	}
	return results
}