# Reddit Internship Monitor

A Go-based web scraper that monitors r/internships for new internship opportunities and sends Discord notifications.

## High-Level Design (HLD)

### Overview
This system scrapes Reddit's r/internships subreddit using Colly, uses AI to classify posts as internship opportunities, and sends real-time notifications via Discord webhooks.

### Architecture Diagram
```
                         Reddit Monitor
                              |
                              v
                    1. Web Scraper (Colly)
  - Scrape old.reddit.com/r/internships/new
  - Poll every 5-10 minutes
  - Extract post metadata (title, body, author, URL)
  - Extract top-level comments
                              |
                              v
                    2. Data Storage Layer
  - SQLite/JSON file to track processed posts
  - Store post IDs to avoid duplicate processing
  - Track timestamp of last poll
                              |
                              v
                    3. AI Classification Engine
  - Send post content to LLM (OpenAI/Claude API)
  - Prompt: "Is this an internship opportunity?"
  - Returns: YES/NO + confidence score
  - Filters out spam, questions, or irrelevant posts
                              |
                              v
                    4. Notification System
  - Discord Webhook Integration
  - Send formatted message with:
    * Post title
    * Direct Reddit link
    * Key details extracted by AI
    * Timestamp
                              |
                              v
                        Discord Channel
```

### Component Details

#### 1. Web Scraper (Colly)
- **Technology**: Colly web scraping framework
- **Target**: old.reddit.com (HTML-based, easier to parse)
- **Responsibilities**:
  - Scrape old.reddit.com/r/internships/new
  - Extract post data from HTML elements
  - Handle pagination if needed
  - Respect rate limiting (avoid detection)
  - Parse post title, body, author, timestamp, and URL
- **Config**: Subreddit name (dynamically configurable)

#### 2. Data Storage
- **Technology**: SQLite or JSON file
- **Schema**:
  ```
  processed_posts {
    post_id: string (primary key)
    title: string
    processed_at: timestamp
    is_opportunity: boolean
  }
  ```
- **Purpose**: Prevent duplicate processing and maintain history

#### 3. AI Classification
- **Technology**: OpenAI API / Anthropic Claude API
- **Input**: Post title + body + top comments
- **Prompt Template**:
  ```
  Analyze this Reddit post and determine if it's a legitimate internship opportunity.

  Title: {title}
  Body: {body}
  Comments: {comments}

  Return YES if it's an internship posting/opportunity, NO otherwise.
  ```
- **Output**: Boolean classification + reasoning

#### 4. Discord Notification
- **Technology**: Discord Webhooks (HTTP POST)
- **Message Format**:
  ```
  🎯 New Internship Opportunity Found!

  Title: [Post Title]
  Link: [Reddit URL]
  Posted: [Timestamp]

  AI Analysis: [Brief summary]
  ```

### Data Flow

1. **Polling Loop** (every 5-10 minutes):
   ```
   START -> Scrape old.reddit.com -> Parse HTML -> For each post:
                                              |
                                              v
                                         Check if seen?
                                              |
                                         +----+----+
                                         |         |
                                        NO        YES
                                         |         |
                                         v         v
                           Get post + comments   Skip
                                    |
                                    v
                           Send to AI classifier
                                    |
                                    v
                           Is opportunity?
                                    |
                               +----+----+
                               |         |
                              YES        NO
                               |         |
                               v         v
                    Send Discord alert  Log & skip
                               |         |
                               +---------+
                                    |
                                    v
                           Mark as processed
                                    |
                                    v
                                  LOOP
   ```

### Configuration (Future)

```yaml
scraper:
  subreddits:
    - internships
    - cscareerquestions
    - forhire
  poll_interval: 5m
  base_url: https://old.reddit.com

ai:
  provider: openai  # or anthropic
  model: gpt-4

notifications:
  discord:
    webhook_url: ${DISCORD_WEBHOOK}

storage:
  type: sqlite  # or json
  path: ./data/posts.db
```

### Scalability Considerations

- **Multiple Subreddits**: Config-driven subreddit list
- **Multiple Notification Channels**: Plugin architecture for Slack, Email, Telegram
- **Distributed Processing**: Add message queue (Redis) for high volume
- **Caching**: Redis cache for scraped pages
- **AI Optimization**: Batch processing to reduce API costs
- **Rate Limiting**: Implement delays and user-agent rotation to avoid detection

### Tech Stack

- **Language**: Go 1.21+
- **Web Scraper**: Colly (`github.com/gocolly/colly`)
- **AI**: OpenAI/Anthropic SDK
- **Storage**: SQLite (`mattn/go-sqlite3`) or JSON
- **Scheduler**: `robfig/cron` or simple `time.Ticker`
- **Config**: `viper` for YAML/ENV management

### Environment Variables

```bash
OPENAI_API_KEY=your_openai_key
# OR
ANTHROPIC_API_KEY=your_claude_key

DISCORD_WEBHOOK_URL=your_discord_webhook
```

### Project Phases

#### Phase 1: Reddit Scraper
- Set up Colly scraper
- Scrape old.reddit.com/r/internships/new
- Extract post data from HTML
- Print new posts to console
- Track seen posts

#### Phase 2: AI Classification
- Integrate OpenAI/Claude API
- Build classification prompt
- Filter posts based on AI response

#### Phase 3: Discord Notifications
- Set up Discord webhook
- Format and send alerts
- Add error handling

#### Phase 4: Configuration & Scaling
- YAML config for multiple subreddits
- Support multiple notification channels
- Add logging and monitoring
- Dockerize the application

---

## Getting Started

(To be implemented)

## License

MIT
