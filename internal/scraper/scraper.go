package scraper

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/iiviie/go-scrapper/internal/models"
	"github.com/iiviie/go-scrapper/internal/storage"
)

// RedditScraper handles scraping Reddit for posts
type RedditScraper struct {
	collector  *colly.Collector
	storage    storage.Storage
	baseURL    string
	subreddits []string
}

// NewRedditScraper creates a new Reddit scraper instance
func NewRedditScraper(baseURL string, subreddits []string, storage storage.Storage) *RedditScraper {
	c := colly.NewCollector(
		colly.AllowedDomains("old.reddit.com"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"),
	)

	// Set rate limiting to be respectful
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*reddit.com*",
		Delay:       2 * time.Second,
		RandomDelay: 1 * time.Second,
	})

	return &RedditScraper{
		collector:  c,
		storage:    storage,
		baseURL:    baseURL,
		subreddits: subreddits,
	}
}

// ScrapeNew scrapes new posts from configured subreddits
func (rs *RedditScraper) ScrapeNew() ([]*models.Post, error) {
	var newPosts []*models.Post

	for _, subreddit := range rs.subreddits {
		posts, err := rs.scrapeSubreddit(subreddit)
		if err != nil {
			log.Printf("Error scraping r/%s: %v", subreddit, err)
			continue
		}
		newPosts = append(newPosts, posts...)
	}

	return newPosts, nil
}

// scrapeSubreddit scrapes a specific subreddit
func (rs *RedditScraper) scrapeSubreddit(subreddit string) ([]*models.Post, error) {
	var posts []*models.Post
	url := fmt.Sprintf("%s/r/%s/new", rs.baseURL, subreddit)

	// Clone the collector for this specific request
	c := rs.collector.Clone()

	// Extract post information
	c.OnHTML("div.thing", func(e *colly.HTMLElement) {
		postID := e.Attr("data-fullname")

		// Check if already processed
		processed, err := rs.storage.IsProcessed(postID)
		if err != nil {
			log.Printf("Error checking if post is processed: %v", err)
			return
		}
		if processed {
			return
		}

		post := &models.Post{
			ID:        postID,
			Title:     e.ChildText("a.title"),
			Author:    e.ChildAttr("a.author", "href"),
			URL:       e.Request.AbsoluteURL(e.ChildAttr("a.title", "href")),
			Timestamp: time.Now(),
		}

		// Extract author from href (format: /user/username)
		if strings.HasPrefix(post.Author, "/user/") {
			post.Author = strings.TrimPrefix(post.Author, "/user/")
		}

		// Get the post's permalink to visit for full details
		permalink := e.ChildAttr("a.comments", "href")
		if permalink != "" {
			fullURL := rs.baseURL + permalink
			rs.scrapePostDetails(post, fullURL)
		}

		posts = append(posts, post)
		log.Printf("Found new post: %s (ID: %s)", post.Title, post.ID)
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error scraping %s: %v", r.Request.URL, err)
	})

	c.OnRequest(func(r *colly.Request) {
		log.Printf("Visiting: %s", r.URL.String())
	})

	err := c.Visit(url)
	return posts, err
}

// scrapePostDetails fetches the full post body and comments
func (rs *RedditScraper) scrapePostDetails(post *models.Post, postURL string) {
	c := rs.collector.Clone()

	// Extract post body
	c.OnHTML("div.usertext-body", func(e *colly.HTMLElement) {
		if post.Body == "" { // Only set the first occurrence (the post body, not comments)
			post.Body = e.Text
		}
	})

	// Extract top-level comments
	c.OnHTML("div.entry.unvoted div.usertext-body", func(e *colly.HTMLElement) {
		// Get the comment author
		author := e.DOM.ParentsUntil("div.entry").Parent().Find("a.author").First().Text()
		commentBody := e.Text

		if author != "" && commentBody != "" {
			comment := models.Comment{
				Author: author,
				Body:   commentBody,
			}
			post.Comments = append(post.Comments, comment)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error fetching post details for %s: %v", postURL, err)
	})

	c.Visit(postURL)
}

// SavePosts saves posts to storage
func (rs *RedditScraper) SavePosts(posts []*models.Post) error {
	for _, post := range posts {
		if err := rs.storage.SavePost(post); err != nil {
			log.Printf("Error saving post %s: %v", post.ID, err)
			continue
		}
		log.Printf("Saved post: %s", post.ID)
	}
	return nil
}
