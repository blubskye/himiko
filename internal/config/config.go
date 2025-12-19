// Himiko Discord Bot
// Copyright (C) 2025 Himiko Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const MaxConfigBackups = 3

type Config struct {
	Token        string   `json:"token"`
	Prefix       string   `json:"prefix"`
	DatabasePath string   `json:"database_path"`
	OwnerID      string   `json:"owner_id"`       // Single owner (backwards compatible)
	OwnerIDs     []string `json:"owner_ids"`      // Multiple owners

	// API Keys for various services
	APIs struct {
		WeatherAPIKey      string `json:"weather_api_key"`
		GoogleAPIKey       string `json:"google_api_key"`
		SpotifyID          string `json:"spotify_client_id"`
		SpotifySecret      string `json:"spotify_client_secret"`
		OpenAIKey          string `json:"openai_api_key"`
		OpenAIBaseURL      string `json:"openai_base_url"`
		OpenAIModel        string `json:"openai_model"`
		YouTubeAPIKey      string `json:"youtube_api_key"`
		SoundCloudAuthToken string `json:"soundcloud_auth_token"`
	} `json:"apis"`

	// Feature toggles
	Features struct {
		DMLogging           bool   `json:"dm_logging"`
		CommandHistory      bool   `json:"command_history"`
		DeleteTimer         int    `json:"delete_timer"` // seconds, 0 = disabled
		WebhookNotify       bool   `json:"webhook_notify"`
		WebhookURL          string `json:"webhook_url"`
		AutoUpdate          bool   `json:"auto_update"`           // Check for updates on startup
		AutoUpdateApply     bool   `json:"auto_update_apply"`     // Automatically apply updates (requires restart)
		UpdateCheckHours    int    `json:"update_check_hours"`    // Hours between periodic update checks (0 = disabled)
		UpdateNotifyChannel string `json:"update_notify_channel"` // Channel ID to post update notifications
		DebugMode           bool   `json:"debug_mode"`            // Enable verbose logging and stack traces
	} `json:"features"`

	// Web server configuration
	WebServer struct {
		Enabled     bool   `json:"enabled"`      // Enable/disable the web server
		Port        int    `json:"port"`         // Port to listen on (default: 8080)
		Host        string `json:"host"`         // Host to bind to (default: "127.0.0.1" for local only)
		SecretKey   string `json:"secret_key"`   // Secret key for session encryption
		AllowRemote bool   `json:"allow_remote"` // Allow connections from non-localhost (for NGINX proxy)
	} `json:"webserver"`

	// Field-level encryption for sensitive database data
	Encryption struct {
		Enabled bool   `json:"enabled"` // Enable/disable field encryption
		Key     string `json:"key"`     // Encryption passphrase (required if enabled)
	} `json:"encryption"`
}

func Load(path string) (*Config, error) {
	// Check if config file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create default config
		cfg := &Config{
			Token:        "YOUR_BOT_TOKEN_HERE",
			Prefix:       "/",
			DatabasePath: "himiko.db",
			OwnerID:      "",
		}
		cfg.APIs.OpenAIBaseURL = "https://api.openai.com/v1"
		cfg.APIs.OpenAIModel = "gpt-3.5-turbo"
		cfg.Features.CommandHistory = true

		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, data, 0600); err != nil {
			return nil, err
		}
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Set defaults if not specified
	if cfg.Prefix == "" {
		cfg.Prefix = "/"
	}
	if cfg.DatabasePath == "" {
		cfg.DatabasePath = "himiko.db"
	}
	if cfg.APIs.OpenAIBaseURL == "" {
		cfg.APIs.OpenAIBaseURL = "https://api.openai.com/v1"
	}
	if cfg.APIs.OpenAIModel == "" {
		cfg.APIs.OpenAIModel = "gpt-3.5-turbo"
	}
	// Set webserver defaults
	if cfg.WebServer.Port == 0 {
		cfg.WebServer.Port = 8080
	}
	if cfg.WebServer.Host == "" {
		cfg.WebServer.Host = "127.0.0.1"
	}

	// Check if migration is needed (new fields added)
	migrated := migrateConfig(&cfg, data, path)
	if migrated {
		// Save the migrated config
		if err := cfg.Save(path); err != nil {
			fmt.Printf("Warning: Failed to save migrated config: %v\n", err)
		}
	}

	return &cfg, nil
}

// migrateConfig checks if the config needs migration and backs up old config if needed
func migrateConfig(cfg *Config, originalData []byte, path string) bool {
	// Check if config has new fields by comparing JSON
	var rawMap map[string]interface{}
	if err := json.Unmarshal(originalData, &rawMap); err != nil {
		return false
	}

	needsMigration := false

	// Check for owner_ids field (new in 1.5.3)
	if _, exists := rawMap["owner_ids"]; !exists {
		needsMigration = true
	}

	// Check for debug_mode in features (new in 1.5.5)
	if features, ok := rawMap["features"].(map[string]interface{}); ok {
		if _, exists := features["debug_mode"]; !exists {
			needsMigration = true
		}
	}

	if needsMigration {
		// Backup old config
		backupPath := fmt.Sprintf("%s.backup.%s", path, time.Now().Format("20060102-150405"))
		if err := os.WriteFile(backupPath, originalData, 0600); err != nil {
			fmt.Printf("Warning: Failed to backup config to %s: %v\n", backupPath, err)
		} else {
			fmt.Printf("Config backed up to %s\n", backupPath)
			// Clean up old backups, keeping only the most recent MaxConfigBackups
			cleanupOldBackups(path)
		}

		// Migrate owner_id to owner_ids if owner_id is set and owner_ids is empty
		if cfg.OwnerID != "" && len(cfg.OwnerIDs) == 0 {
			cfg.OwnerIDs = []string{cfg.OwnerID}
			fmt.Println("Migrated owner_id to owner_ids array")
		}

		fmt.Println("Config migrated to new format")
		return true
	}

	return false
}

// cleanupOldBackups removes old config backups, keeping only the most recent MaxConfigBackups
func cleanupOldBackups(configPath string) {
	dir := filepath.Dir(configPath)
	base := filepath.Base(configPath)
	pattern := base + ".backup."

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	// Find all backup files
	var backups []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), pattern) {
			backups = append(backups, filepath.Join(dir, entry.Name()))
		}
	}

	// If we have more than MaxConfigBackups, delete the oldest ones
	if len(backups) <= MaxConfigBackups {
		return
	}

	// Sort by name (which includes timestamp, so older files come first)
	sort.Strings(backups)

	// Delete oldest backups, keeping only the most recent MaxConfigBackups
	toDelete := len(backups) - MaxConfigBackups
	for i := 0; i < toDelete; i++ {
		if err := os.Remove(backups[i]); err != nil {
			fmt.Printf("Warning: Failed to delete old backup %s: %v\n", backups[i], err)
		} else {
			fmt.Printf("Deleted old config backup: %s\n", filepath.Base(backups[i]))
		}
	}
}

func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// IsOwner checks if the given user ID is a bot owner
func (c *Config) IsOwner(userID string) bool {
	// Check single owner (backwards compatibility)
	if c.OwnerID != "" && c.OwnerID == userID {
		return true
	}

	// Check multiple owners
	for _, id := range c.OwnerIDs {
		if id == userID {
			return true
		}
	}

	return false
}
