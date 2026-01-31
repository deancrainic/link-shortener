package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"link-shortener/internal/model"
	"link-shortener/internal/storage"
)

type Store struct {
	db *sql.DB
}

func New(path string) (*Store, error) {
	dbPath := strings.TrimSpace(path)
	if dbPath == "" {
		dbPath = "data.db"
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return nil, err
	}
	if err := ensureSchema(db); err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func ensureSchema(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS links (
			code TEXT PRIMARY KEY,
			original_url TEXT NOT NULL,
			created_at TEXT NOT NULL,
			expires_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS clicks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			code TEXT NOT NULL,
			timestamp TEXT NOT NULL,
			ip TEXT,
			country TEXT,
			user_agent TEXT,
			FOREIGN KEY(code) REFERENCES links(code) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS unique_ips (
			code TEXT NOT NULL,
			ip TEXT NOT NULL,
			PRIMARY KEY (code, ip),
			FOREIGN KEY(code) REFERENCES links(code) ON DELETE CASCADE
		);`,
	}
	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) Save(link *model.Link) error {
	_, err := s.db.Exec(
		`INSERT INTO links (code, original_url, created_at, expires_at)
		 VALUES (?, ?, ?, ?)`,
		link.Code,
		link.OriginalURL,
		formatTime(link.CreatedAt),
		formatTime(link.ExpiresAt),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return storage.ErrCodeExists
		}
		return err
	}
	return nil
}

func (s *Store) Upsert(link *model.Link) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM clicks WHERE code = ?`, link.Code); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM unique_ips WHERE code = ?`, link.Code); err != nil {
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO links (code, original_url, created_at, expires_at)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(code) DO UPDATE SET
		   original_url = excluded.original_url,
		   created_at = excluded.created_at,
		   expires_at = excluded.expires_at`,
		link.Code,
		link.OriginalURL,
		formatTime(link.CreatedAt),
		formatTime(link.ExpiresAt),
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) Get(code string) (*model.Link, bool) {
	row := s.db.QueryRow(
		`SELECT code, original_url, created_at, expires_at
		 FROM links WHERE code = ?`,
		code,
	)

	var link model.Link
	var created, expires string
	if err := row.Scan(&link.Code, &link.OriginalURL, &created, &expires); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false
		}
		return nil, false
	}

	createdAt, err := parseTime(created)
	if err != nil {
		return nil, false
	}
	expiresAt, err := parseTime(expires)
	if err != nil {
		return nil, false
	}
	link.CreatedAt = createdAt
	link.ExpiresAt = expiresAt

	clicks, err := s.loadClicks(code)
	if err != nil {
		return nil, false
	}
	link.Clicks = clicks

	uniqueIPs, err := s.loadUniqueIPs(code)
	if err != nil {
		return nil, false
	}
	link.UniqueIPs = uniqueIPs

	return &link, true
}

func (s *Store) List() []*model.Link {
	rows, err := s.db.Query(
		`SELECT code, original_url, created_at, expires_at
		 FROM links ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var links []*model.Link
	for rows.Next() {
		var link model.Link
		var created, expires string
		if err := rows.Scan(&link.Code, &link.OriginalURL, &created, &expires); err != nil {
			continue
		}
		createdAt, err := parseTime(created)
		if err != nil {
			continue
		}
		expiresAt, err := parseTime(expires)
		if err != nil {
			continue
		}
		link.CreatedAt = createdAt
		link.ExpiresAt = expiresAt

		totalClicks, err := s.loadClickCount(link.Code)
		if err == nil {
			link.Clicks = make([]model.Click, totalClicks)
		}
		uniqueIPs, err := s.loadUniqueIPs(link.Code)
		if err == nil {
			link.UniqueIPs = uniqueIPs
		}
		links = append(links, &link)
	}
	return links
}

func (s *Store) RecordClick(code string, click model.Click) (*model.Link, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var exists int
	if err := tx.QueryRow(`SELECT 1 FROM links WHERE code = ?`, code).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}

	_, err = tx.Exec(
		`INSERT INTO clicks (code, timestamp, ip, country, user_agent)
		 VALUES (?, ?, ?, ?, ?)`,
		code,
		formatTime(click.Timestamp),
		click.IP,
		click.Country,
		click.UserAgent,
	)
	if err != nil {
		return nil, err
	}

	if click.IP != "" {
		if _, err := tx.Exec(
			`INSERT OR IGNORE INTO unique_ips (code, ip) VALUES (?, ?)`,
			code,
			click.IP,
		); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	link, ok := s.Get(code)
	if !ok {
		return nil, storage.ErrNotFound
	}
	return link, nil
}

func (s *Store) loadClicks(code string) ([]model.Click, error) {
	rows, err := s.db.Query(
		`SELECT timestamp, ip, country, user_agent
		 FROM clicks WHERE code = ? ORDER BY timestamp`,
		code,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clicks []model.Click
	for rows.Next() {
		var click model.Click
		var timestamp string
		if err := rows.Scan(&timestamp, &click.IP, &click.Country, &click.UserAgent); err != nil {
			return nil, err
		}
		parsed, err := parseTime(timestamp)
		if err != nil {
			return nil, err
		}
		click.Timestamp = parsed
		clicks = append(clicks, click)
	}
	return clicks, nil
}

func (s *Store) loadUniqueIPs(code string) (map[string]struct{}, error) {
	rows, err := s.db.Query(`SELECT ip FROM unique_ips WHERE code = ?`, code)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	unique := make(map[string]struct{})
	for rows.Next() {
		var ip string
		if err := rows.Scan(&ip); err != nil {
			return nil, err
		}
		if ip != "" {
			unique[ip] = struct{}{}
		}
	}
	return unique, nil
}

func (s *Store) loadClickCount(code string) (int, error) {
	row := s.db.QueryRow(`SELECT COUNT(*) FROM clicks WHERE code = ?`, code)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func parseTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, fmt.Errorf("empty time")
	}
	t, err := time.Parse(time.RFC3339Nano, value)
	if err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, value)
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") || strings.Contains(msg, "constraint")
}
