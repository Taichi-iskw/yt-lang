package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the application
type Config struct {
	DatabaseURL string `yaml:"database_url"`
}

// DatabaseConfig holds parsed database connection configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

// NewConfig loads configuration with the following priority:
// Environment variables > Config file (required)
func NewConfig() (*Config, error) {
	// Load from config file (required)
	config := &Config{}
	if err := loadConfigFile(config); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("configuration file not found. Please run 'ytlang config init' to create it")
		}
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	// Apply environment variables (can override config file)
	if envURL := os.Getenv("DATABASE_URL"); envURL != "" {
		config.DatabaseURL = envURL
	}

	return config, nil
}

// ParseDatabaseConfig parses the DATABASE_URL into DatabaseConfig
func (c *Config) ParseDatabaseConfig() (*DatabaseConfig, error) {
	if c.DatabaseURL == "" {
		return nil, fmt.Errorf("database URL is empty")
	}

	return parseDatabaseURL(c.DatabaseURL)
}

// InitConfig creates a new configuration file with example DATABASE_URL
func InitConfig(databaseURL string) error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	// Check if config file already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("configuration file already exists: %s", configPath)
	}

	// Create config with provided DATABASE_URL
	if databaseURL == "" {
		databaseURL = "postgres://user:password@localhost:5432/ytlang?sslmode=disable"
	}

	// Prepare YAML content with comments
	yamlContent := fmt.Sprintf(`# yt-lang configuration file
# Database connection URL format:
# postgres://[user[:password]@]host[:port]/dbname[?param1=value1&...]

database_url: "%s"
`, databaseURL)

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfigPath returns the path to the configuration file
func GetConfigPath() (string, error) {
	return getConfigFilePath()
}

// getConfigDir returns the configuration directory path (~/.yt-lang)
func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".yt-lang"), nil
}

// getConfigFilePath returns the full path to the config file
func getConfigFilePath() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.yaml"), nil
}

// loadConfigFile loads configuration from ~/.yt-lang/config.yaml
func loadConfigFile(config *Config) error {
	configPath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

// parseDatabaseURL parses DATABASE_URL format (postgres://user:pass@host:port/dbname?params)
func parseDatabaseURL(databaseURL string) (*DatabaseConfig, error) {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid database URL: %w", err)
	}

	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return nil, fmt.Errorf("unsupported scheme: %s (expected postgres or postgresql)", u.Scheme)
	}

	// Extract components
	host := u.Hostname()
	if host == "" {
		host = "localhost"
	}

	port := 5432 // default
	if u.Port() != "" {
		if p, err := strconv.Atoi(u.Port()); err == nil {
			port = p
		}
	}

	user := "postgres" // default
	if u.User != nil {
		user = u.User.Username()
	}

	password := ""
	if u.User != nil {
		if pass, ok := u.User.Password(); ok {
			password = pass
		}
	}

	dbname := "ytlang" // default
	if u.Path != "" && u.Path != "/" {
		dbname = u.Path[1:] // remove leading slash
	}

	// Parse query parameters
	sslMode := "disable" // default for local development
	if ssl := u.Query().Get("sslmode"); ssl != "" {
		sslMode = ssl
	}

	return &DatabaseConfig{
		Host:            host,
		Port:            port,
		User:            user,
		Password:        password,
		DBName:          dbname,
		SSLMode:         sslMode,
		MaxConns:        10,
		MinConns:        1,
		MaxConnLifetime: 60 * time.Minute,
		MaxConnIdleTime: 10 * time.Minute,
	}, nil
}

// ConnectionString returns PostgreSQL connection string
func (db *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		db.Host, db.Port, db.User, db.Password, db.DBName, db.SSLMode,
	)
}
