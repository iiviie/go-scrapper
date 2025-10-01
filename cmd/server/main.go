package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iiviie/go-scrapper/internal/config"
	"github.com/iiviie/go-scrapper/internal/scraper"
	"github.com/iiviie/go-scrapper/internal/storage"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize storage
	var store storage.Storage
	switch cfg.Storage.Type {
	case "sqlite":
		store, err = storage.NewSQLiteStorage(cfg.Storage.Path)
		if err != nil {
			log.Fatalf("Failed to initialize storage: %v", err)
		}
		defer store.Close()
	default:
		log.Fatalf("Unsupported storage type: %s", cfg.Storage.Type)
	}

	// Initialize scraper
	sc := scraper.NewRedditScraper(
		cfg.Scraper.BaseURL,
		cfg.Scraper.Subreddits,
		store,
	)

	// Parse poll interval
	pollInterval, err := time.ParseDuration(cfg.Scraper.PollInterval)
	if err != nil {
		log.Fatalf("Invalid poll interval: %v", err)
	}

	// Start scraping in a goroutine
	go startScraping(sc, pollInterval)

	// Initialize Gin server
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.IndentedJSON(200, gin.H{
			"status": "ok",
			"timestamp": time.Now(),
		})
	})

	// Get all processed posts
	router.GET("/posts", func(c *gin.Context) {
		posts, err := store.GetAllPosts()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(200, gin.H{
			"count": len(posts),
			"posts": posts,
		})
	})

	// Trigger manual scrape
	router.POST("/scrape", func(c *gin.Context) {
		log.Println("Manual scrape triggered")
		posts, err := sc.ScrapeNew()
		if err != nil {
			c.IndentedJSON(500, gin.H{"error": err.Error()})
			return
		}

		if len(posts) > 0 {
			if err := sc.SavePosts(posts); err != nil {
				c.IndentedJSON(500, gin.H{"error": err.Error()})
				return
			}
		}

		c.IndentedJSON(200, gin.H{
			"message": "Scrape completed",
			"new_posts": len(posts),
		})
	})

	// Start server
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Starting server on %s", addr)
	log.Printf("Scraping subreddits: %v", cfg.Scraper.Subreddits)
	log.Printf("Poll interval: %s", pollInterval)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// startScraping runs the scraper at regular intervals
func startScraping(sc *scraper.RedditScraper, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run once immediately
	log.Println("Starting initial scrape...")
	runScrape(sc)

	// Then run on interval
	for range ticker.C {
		log.Println("Running scheduled scrape...")
		runScrape(sc)
	}
}

// runScrape executes a single scrape operation
func runScrape(sc *scraper.RedditScraper) {
	posts, err := sc.ScrapeNew()
	if err != nil {
		log.Printf("Scrape error: %v", err)
		return
	}

	log.Printf("Found %d new posts", len(posts))

	if len(posts) > 0 {
		if err := sc.SavePosts(posts); err != nil {
			log.Printf("Error saving posts: %v", err)
		}
	}
}
