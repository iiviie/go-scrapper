package storage

import (
	"database/sql"
	"encoding/json"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/iiviie/go-scrapper/internal/models"
)

// SQLiteStorage implements Storage interface using SQLite
type SQLiteStorage struct {
	db *sql.DB
}

// NewSQLiteStorage creates a new SQLite storage instance
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	storage := &SQLiteStorage{db: db}
	if err := storage.initDB(); err != nil {
		return nil, err
	}

	return storage, nil
}

// initDB initializes the database schema
func (s *SQLiteStorage) initDB() error {
	query := `
	CREATE TABLE IF NOT EXISTS processed_posts (
		post_id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		body TEXT,
		author TEXT,
		url TEXT NOT NULL,
		timestamp DATETIME,
		comments TEXT,
		processed_at DATETIME NOT NULL,
		is_opportunity BOOLEAN DEFAULT FALSE
	);

	CREATE INDEX IF NOT EXISTS idx_timestamp ON processed_posts(timestamp);
	CREATE INDEX IF NOT EXISTS idx_processed_at ON processed_posts(processed_at);
	CREATE INDEX IF NOT EXISTS idx_is_opportunity ON processed_posts(is_opportunity);
	`

	_, err := s.db.Exec(query)
	return err
}

// IsProcessed checks if a post has already been processed
func (s *SQLiteStorage) IsProcessed(postID string) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM processed_posts WHERE post_id = ?", postID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// SavePost saves a post to storage
func (s *SQLiteStorage) SavePost(post *models.Post) error {
	commentsJSON, err := json.Marshal(post.Comments)
	if err != nil {
		return err
	}

	query := `
	INSERT INTO processed_posts (post_id, title, body, author, url, timestamp, comments, processed_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(post_id) DO UPDATE SET
		title = excluded.title,
		body = excluded.body,
		author = excluded.author,
		url = excluded.url,
		timestamp = excluded.timestamp,
		comments = excluded.comments,
		processed_at = excluded.processed_at
	`

	_, err = s.db.Exec(query,
		post.ID,
		post.Title,
		post.Body,
		post.Author,
		post.URL,
		post.Timestamp,
		string(commentsJSON),
		time.Now(),
	)

	return err
}

// GetAllPosts retrieves all processed posts
func (s *SQLiteStorage) GetAllPosts() ([]*models.Post, error) {
	query := `SELECT post_id, title, body, author, url, timestamp, comments, processed_at
	          FROM processed_posts ORDER BY timestamp DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		var post models.Post
		var commentsJSON string
		var timestamp, processedAt string

		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Body,
			&post.Author,
			&post.URL,
			&timestamp,
			&commentsJSON,
			&processedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse timestamps
		post.Timestamp, _ = time.Parse(time.RFC3339, timestamp)
		post.ProcessedAt, _ = time.Parse(time.RFC3339, processedAt)

		// Parse comments
		json.Unmarshal([]byte(commentsJSON), &post.Comments)

		posts = append(posts, &post)
	}

	return posts, nil
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}
