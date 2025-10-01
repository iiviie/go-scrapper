package config

import (
	"log"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Scraper       ScraperConfig       `mapstructure:"scraper"`
	Storage       StorageConfig       `mapstructure:"storage"`
	Server        ServerConfig        `mapstructure:"server"`
	Notifications NotificationConfig  `mapstructure:"notifications"`
	AI            AIConfig            `mapstructure:"ai"`
}

// ScraperConfig holds scraper-related configuration
type ScraperConfig struct {
	Subreddits   []string `mapstructure:"subreddits"`
	PollInterval string   `mapstructure:"poll_interval"`
	BaseURL      string   `mapstructure:"base_url"`
}

// StorageConfig holds storage-related configuration
type StorageConfig struct {
	Type string `mapstructure:"type"`
	Path string `mapstructure:"path"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port string `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

// NotificationConfig holds notification settings
type NotificationConfig struct {
	Discord DiscordConfig `mapstructure:"discord"`
}

// DiscordConfig holds Discord webhook settings
type DiscordConfig struct {
	WebhookURL string `mapstructure:"webhook_url"`
}

// AIConfig holds AI provider settings
type AIConfig struct {
	Provider string `mapstructure:"provider"`
	Model    string `mapstructure:"model"`
	APIKey   string `mapstructure:"api_key"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set defaults
	viper.SetDefault("scraper.subreddits", []string{"internships"})
	viper.SetDefault("scraper.poll_interval", "5m")
	viper.SetDefault("scraper.base_url", "https://old.reddit.com")
	viper.SetDefault("storage.type", "sqlite")
	viper.SetDefault("storage.path", "./data/posts.db")
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "0.0.0.0")

	// Environment variable bindings
	viper.AutomaticEnv()
	viper.BindEnv("notifications.discord.webhook_url", "DISCORD_WEBHOOK_URL")
	viper.BindEnv("ai.api_key", "OPENAI_API_KEY")
	viper.BindEnv("ai.provider", "AI_PROVIDER")
	viper.BindEnv("ai.model", "AI_MODEL")

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
		log.Println("No config file found, using defaults and environment variables")
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
