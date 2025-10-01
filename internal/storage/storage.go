package storage

import "github.com/iiviie/go-scrapper/internal/models"

// Storage defines the interface for post storage
type Storage interface {
	// IsProcessed checks if a post has already been processed
	IsProcessed(postID string) (bool, error)

	// SavePost saves a post to storage
	SavePost(post *models.Post) error

	// GetAllPosts retrieves all processed posts
	GetAllPosts() ([]*models.Post, error)

	// Close closes the storage connection
	Close() error
}
