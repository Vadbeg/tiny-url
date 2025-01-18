package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteDatabase struct {
	db *sql.DB
}

func InitializeSQLiteDatabase(dbFilePath string) (*SQLiteDatabase, error) {
	db, err := sql.Open("sqlite3", dbFilePath)

	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS urls (
		short_hash TEXT PRIMARY KEY,
		full_url TEXT NOT NULL
	)`)

	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &SQLiteDatabase{db: db}, nil
}

func (s *SQLiteDatabase) InsertURL(shortHash, fullURL string) error {
	_, err := s.db.Exec(
		"INSERT INTO urls (short_hash, full_url) VALUES (?, ?)", shortHash, fullURL,
	)
	return err
}

func (s *SQLiteDatabase) GetFullURL(shortHash string) (string, error) {
	var fullURL string
	err := s.db.QueryRow("SELECT full_url FROM urls WHERE short_hash = ?", shortHash).Scan(&fullURL)
	return fullURL, err
}

func (s *SQLiteDatabase) GetShortURL(fullURL string) (string, error) {
	var shortHash string
	err := s.db.QueryRow("SELECT short_hash FROM urls WHERE full_url = ?", fullURL).Scan(&shortHash)
	return shortHash, err
}

func (s *SQLiteDatabase) GetAllBindings() (map[string]string, error) {
	allBindings := make(map[string]string)

	rows, err := s.db.Query("SELECT full_url, short_hash FROM urls")

	if err != nil {
		return allBindings, err
	}

	for rows.Next() {
		var fullURL string
		var shortHash string
		err = rows.Scan(&fullURL, &shortHash)
		if err == nil {
			allBindings[fullURL] = shortHash
		}
	}

	return allBindings, nil
}

func (s *SQLiteDatabase) Close() error {
	return s.db.Close()
}
