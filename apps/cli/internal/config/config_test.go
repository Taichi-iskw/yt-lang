package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig_NoConfigFile(t *testing.T) {
	// Use temporary directory for test
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	_, err := NewConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration file not found")
	assert.Contains(t, err.Error(), "ytlang config init")
}

func TestNewConfig_ConfigFile(t *testing.T) {
	// Create temporary config directory
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".yt-lang")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	// Create test config file with custom URL
	configContent := `database_url: "postgres://myuser:mypass@myhost:5433/mydb?sslmode=require"`
	configPath := filepath.Join(configDir, "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	// Set temporary HOME
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	config, err := NewConfig()
	require.NoError(t, err)

	// Check config file URL was loaded
	assert.Equal(t, "postgres://myuser:mypass@myhost:5433/mydb?sslmode=require", config.DatabaseURL)
}

func TestNewConfig_EnvironmentOverride(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".yt-lang")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	// Create test config file
	configContent := `database_url: "postgres://fileuser:filepass@filehost:5433/filedb"`
	configPath := filepath.Join(configDir, "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	// Set environment variable to override config file
	os.Setenv("DATABASE_URL", "postgres://envuser:envpass@envhost:5434/envdb")
	defer os.Unsetenv("DATABASE_URL")

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	config, err := NewConfig()
	require.NoError(t, err)

	// Environment variable should override config file
	assert.Equal(t, "postgres://envuser:envpass@envhost:5434/envdb", config.DatabaseURL)
}

func TestInitConfig(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Test InitConfig with custom URL
	databaseURL := "postgres://testuser:testpass@testhost:5433/testdb"
	err := InitConfig(databaseURL)
	require.NoError(t, err)

	// Check config file was created with correct content
	configPath := filepath.Join(tempDir, ".yt-lang", "config.yaml")
	assert.FileExists(t, configPath)

	// Load and verify config content
	config, err := NewConfig()
	require.NoError(t, err)
	assert.Equal(t, databaseURL, config.DatabaseURL)
}

func TestInitConfig_AlreadyExists(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".yt-lang")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	// Create existing config file
	configPath := filepath.Join(configDir, "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte("database_url: existing"), 0644))

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// InitConfig should fail with existing file
	err := InitConfig("postgres://new:pass@host/db")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration file already exists")
}

func TestParseDatabaseURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected *DatabaseConfig
		wantErr  bool
	}{
		{
			name: "full URL",
			url:  "postgres://user:pass@host:5433/dbname?sslmode=require",
			expected: &DatabaseConfig{
				Host:     "host",
				Port:     5433,
				User:     "user",
				Password: "pass",
				DBName:   "dbname",
				SSLMode:  "require",
			},
			wantErr: false,
		},
		{
			name: "minimal URL",
			url:  "postgres://postgres@localhost/ytlang",
			expected: &DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "",
				DBName:   "ytlang",
				SSLMode:  "disable",
			},
			wantErr: false,
		},
		{
			name: "default values",
			url:  "postgres:///",
			expected: &DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "",
				DBName:   "ytlang",
				SSLMode:  "disable",
			},
			wantErr: false,
		},
		{
			name:     "invalid scheme",
			url:      "mysql://user@host/db",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := parseDatabaseURL(tt.url)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, config)
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)
				assert.Equal(t, tt.expected.Host, config.Host)
				assert.Equal(t, tt.expected.Port, config.Port)
				assert.Equal(t, tt.expected.User, config.User)
				assert.Equal(t, tt.expected.Password, config.Password)
				assert.Equal(t, tt.expected.DBName, config.DBName)
				assert.Equal(t, tt.expected.SSLMode, config.SSLMode)
			}
		})
	}
}

func TestConfig_ParseDatabaseConfig(t *testing.T) {
	config := &Config{
		DatabaseURL: "postgres://testuser:testpass@testhost:5433/testdb?sslmode=require",
	}

	dbConfig, err := config.ParseDatabaseConfig()
	require.NoError(t, err)

	assert.Equal(t, "testhost", dbConfig.Host)
	assert.Equal(t, 5433, dbConfig.Port)
	assert.Equal(t, "testuser", dbConfig.User)
	assert.Equal(t, "testpass", dbConfig.Password)
	assert.Equal(t, "testdb", dbConfig.DBName)
	assert.Equal(t, "require", dbConfig.SSLMode)
}
