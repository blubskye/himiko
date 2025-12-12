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
	"os"
)

type Config struct {
	Token        string `json:"token"`
	Prefix       string `json:"prefix"`
	DatabasePath string `json:"database_path"`
	OwnerID      string `json:"owner_id"`

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
		DMLogging            bool   `json:"dm_logging"`
		CommandHistory       bool   `json:"command_history"`
		DeleteTimer          int    `json:"delete_timer"` // seconds, 0 = disabled
		WebhookNotify        bool   `json:"webhook_notify"`
		WebhookURL           string `json:"webhook_url"`
		AutoUpdate           bool   `json:"auto_update"`            // Check for updates on startup
		AutoUpdateApply      bool   `json:"auto_update_apply"`      // Automatically apply updates (requires restart)
		UpdateCheckHours     int    `json:"update_check_hours"`     // Hours between periodic update checks (0 = disabled)
		UpdateNotifyChannel  string `json:"update_notify_channel"`  // Channel ID to post update notifications
	} `json:"features"`
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

	return &cfg, nil
}

func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
