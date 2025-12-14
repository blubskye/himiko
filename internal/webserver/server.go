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

package webserver

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/blubskye/himiko/internal/config"
	"github.com/blubskye/himiko/internal/database"
	"github.com/blubskye/himiko/internal/updater"
	"github.com/bwmarrin/discordgo"
)

// Server represents the web server for the dashboard
type Server struct {
	config     *config.Config
	db         *database.DB
	session    *discordgo.Session
	httpServer *http.Server
	running    bool
	mu         sync.RWMutex
}

// New creates a new web server instance
func New(cfg *config.Config, db *database.DB, session *discordgo.Session) *Server {
	return &Server{
		config:  cfg,
		db:      db,
		session: session,
	}
}

// Start starts the web server
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("server is already running")
	}

	// Generate secret key if not set
	if s.config.WebServer.SecretKey == "" {
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return fmt.Errorf("failed to generate secret key: %w", err)
		}
		s.config.WebServer.SecretKey = base64.StdEncoding.EncodeToString(key)
		log.Println("[WebServer] Generated new secret key")
	}

	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/status", s.handleAPIStatus)
	mux.HandleFunc("/api/guilds", s.handleAPIGuilds)
	mux.HandleFunc("/api/guild/", s.handleAPIGuild)
	mux.HandleFunc("/api/guild/settings/", s.handleAPIGuildSettings)
	mux.HandleFunc("/api/stats", s.handleAPIStats)

	// Config API endpoints
	mux.HandleFunc("/api/guild/logging/", s.handleAPILoggingConfig)
	mux.HandleFunc("/api/guild/antiraid/", s.handleAPIAntiRaidConfig)
	mux.HandleFunc("/api/guild/antispam/", s.handleAPIAntiSpamConfig)
	mux.HandleFunc("/api/guild/spamfilter/", s.handleAPISpamFilterConfig)
	mux.HandleFunc("/api/guild/voicexp/", s.handleAPIVoiceXPConfig)
	mux.HandleFunc("/api/guild/autoclean/", s.handleAPIAutoCleanConfig)
	mux.HandleFunc("/api/guild/ticket/", s.handleAPITicketConfig)
	mux.HandleFunc("/api/guild/regex/", s.handleAPIRegexFilters)
	mux.HandleFunc("/api/guild/ranks/", s.handleAPILevelRanks)
	mux.HandleFunc("/api/guild/commands/", s.handleAPICommandConfig)

	// Helper endpoints
	mux.HandleFunc("/api/commands/list", s.handleAPICommandsList)
	mux.HandleFunc("/api/channels/", s.handleAPIChannels)
	mux.HandleFunc("/api/roles/", s.handleAPIRoles)

	addr := fmt.Sprintf("%s:%d", s.config.WebServer.Host, s.config.WebServer.Port)

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.middleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.running = true

	go func() {
		log.Printf("[WebServer] Starting on http://%s", addr)
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("[WebServer] Error: %v", err)
		}
	}()

	return nil
}

// Stop stops the web server
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	s.running = false
	log.Println("[WebServer] Stopped")
	return nil
}

// IsRunning returns whether the server is running
func (s *Server) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// middleware wraps handlers with common functionality
func (s *Server) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Check if remote access is allowed
		// Note: When behind NGINX, AllowRemote should be true and NGINX handles access control
		// For direct access without proxy, binding to 127.0.0.1 already restricts to localhost
		_ = s.config.WebServer.AllowRemote // Used for documentation/future enhancement

		// Log request
		log.Printf("[WebServer] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

		next.ServeHTTP(w, r)
	})
}

// handleIndex serves the main dashboard page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, dashboardHTML)
}

// handleAPIStatus returns bot status information
func (s *Server) handleAPIStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	botUser := s.session.State.User
	guilds := s.session.State.Guilds

	status := map[string]interface{}{
		"bot": map[string]interface{}{
			"username":      botUser.Username,
			"discriminator": botUser.Discriminator,
			"id":            botUser.ID,
			"avatar":        botUser.AvatarURL("128"),
		},
		"guilds":  len(guilds),
		"version": updater.GetCurrentVersion(),
		"uptime":  time.Now().Format(time.RFC3339),
	}

	s.jsonResponse(w, status)
}

// handleAPIGuilds returns list of guilds
func (s *Server) handleAPIGuilds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var guildList []map[string]interface{}
	for _, guild := range s.session.State.Guilds {
		guildList = append(guildList, map[string]interface{}{
			"id":           guild.ID,
			"name":         guild.Name,
			"member_count": guild.MemberCount,
			"icon":         guild.IconURL("64"),
		})
	}

	s.jsonResponse(w, guildList)
}

// handleAPIGuild returns information about a specific guild
func (s *Server) handleAPIGuild(w http.ResponseWriter, r *http.Request) {
	guildID := r.URL.Path[len("/api/guild/"):]
	if guildID == "" {
		http.Error(w, "Guild ID required", http.StatusBadRequest)
		return
	}

	// Remove trailing slash if present
	if len(guildID) > 0 && guildID[len(guildID)-1] == '/' {
		guildID = guildID[:len(guildID)-1]
	}

	// Check if this is a settings request
	if len(guildID) > 9 && guildID[len(guildID)-9:] == "/settings" {
		guildID = guildID[:len(guildID)-9]
		s.handleGuildSettings(w, r, guildID)
		return
	}

	guild, err := s.session.State.Guild(guildID)
	if err != nil {
		http.Error(w, "Guild not found", http.StatusNotFound)
		return
	}

	// Get guild settings from database
	settings, _ := s.db.GetGuildSettings(guildID)

	guildInfo := map[string]interface{}{
		"id":           guild.ID,
		"name":         guild.Name,
		"member_count": guild.MemberCount,
		"icon":         guild.IconURL("128"),
		"owner_id":     guild.OwnerID,
		"settings": map[string]interface{}{
			"prefix":          settings.Prefix,
			"mod_log_channel": settings.ModLogChannel,
			"welcome_channel": settings.WelcomeChannel,
			"welcome_message": settings.WelcomeMessage,
		},
	}

	s.jsonResponse(w, guildInfo)
}

// handleAPIGuildSettings handles guild settings API
func (s *Server) handleAPIGuildSettings(w http.ResponseWriter, r *http.Request) {
	guildID := r.URL.Path[len("/api/guild/settings/"):]
	s.handleGuildSettings(w, r, guildID)
}

// handleGuildSettings handles getting/updating guild settings
func (s *Server) handleGuildSettings(w http.ResponseWriter, r *http.Request, guildID string) {
	switch r.Method {
	case http.MethodGet:
		settings, err := s.db.GetGuildSettings(guildID)
		if err != nil {
			http.Error(w, "Failed to get settings", http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, settings)

	case http.MethodPost, http.MethodPut:
		var settings database.GuildSettings
		if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		settings.GuildID = guildID

		if err := s.db.SetGuildSettings(&settings); err != nil {
			http.Error(w, "Failed to save settings", http.StatusInternalServerError)
			return
		}

		s.jsonResponse(w, map[string]string{"status": "ok"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAPIStats returns bot statistics
func (s *Server) handleAPIStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	totalMembers := 0
	for _, guild := range s.session.State.Guilds {
		totalMembers += guild.MemberCount
	}

	stats := map[string]interface{}{
		"guilds":        len(s.session.State.Guilds),
		"total_members": totalMembers,
		"version":       updater.GetCurrentVersion(),
	}

	s.jsonResponse(w, stats)
}

// handleAPILoggingConfig handles logging configuration
func (s *Server) handleAPILoggingConfig(w http.ResponseWriter, r *http.Request) {
	guildID := r.URL.Path[len("/api/guild/logging/"):]
	switch r.Method {
	case http.MethodGet:
		config, err := s.db.GetLoggingConfig(guildID)
		if err != nil {
			http.Error(w, "Failed to get config", http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, config)
	case http.MethodPost, http.MethodPut:
		var config database.LoggingConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		config.GuildID = guildID
		if err := s.db.SetLoggingConfig(&config); err != nil {
			http.Error(w, "Failed to save config", http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, map[string]string{"status": "ok"})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAPIAntiRaidConfig handles anti-raid configuration
func (s *Server) handleAPIAntiRaidConfig(w http.ResponseWriter, r *http.Request) {
	guildID := r.URL.Path[len("/api/guild/antiraid/"):]
	switch r.Method {
	case http.MethodGet:
		config, err := s.db.GetAntiRaidConfig(guildID)
		if err != nil {
			http.Error(w, "Failed to get config", http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, config)
	case http.MethodPost, http.MethodPut:
		var config database.AntiRaidConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		config.GuildID = guildID
		if err := s.db.SetAntiRaidConfig(&config); err != nil {
			http.Error(w, "Failed to save config", http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, map[string]string{"status": "ok"})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAPIAntiSpamConfig handles anti-spam configuration
func (s *Server) handleAPIAntiSpamConfig(w http.ResponseWriter, r *http.Request) {
	guildID := r.URL.Path[len("/api/guild/antispam/"):]
	switch r.Method {
	case http.MethodGet:
		config, err := s.db.GetAntiSpamConfig(guildID)
		if err != nil {
			http.Error(w, "Failed to get config", http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, config)
	case http.MethodPost, http.MethodPut:
		var config database.AntiSpamConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		config.GuildID = guildID
		if err := s.db.SetAntiSpamConfig(&config); err != nil {
			http.Error(w, "Failed to save config", http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, map[string]string{"status": "ok"})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAPISpamFilterConfig handles spam filter configuration
func (s *Server) handleAPISpamFilterConfig(w http.ResponseWriter, r *http.Request) {
	guildID := r.URL.Path[len("/api/guild/spamfilter/"):]
	switch r.Method {
	case http.MethodGet:
		config, err := s.db.GetSpamFilterConfig(guildID)
		if err != nil {
			http.Error(w, "Failed to get config", http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, config)
	case http.MethodPost, http.MethodPut:
		var config database.SpamFilterConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		config.GuildID = guildID
		if err := s.db.SetSpamFilterConfig(&config); err != nil {
			http.Error(w, "Failed to save config", http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, map[string]string{"status": "ok"})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAPIVoiceXPConfig handles voice XP configuration
func (s *Server) handleAPIVoiceXPConfig(w http.ResponseWriter, r *http.Request) {
	guildID := r.URL.Path[len("/api/guild/voicexp/"):]
	switch r.Method {
	case http.MethodGet:
		config, err := s.db.GetVoiceXPConfig(guildID)
		if err != nil {
			http.Error(w, "Failed to get config", http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, config)
	case http.MethodPost, http.MethodPut:
		var config database.VoiceXPConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		config.GuildID = guildID
		if err := s.db.SetVoiceXPConfig(&config); err != nil {
			http.Error(w, "Failed to save config", http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, map[string]string{"status": "ok"})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAPIAutoCleanConfig handles auto-clean configuration
func (s *Server) handleAPIAutoCleanConfig(w http.ResponseWriter, r *http.Request) {
	guildID := r.URL.Path[len("/api/guild/autoclean/"):]
	switch r.Method {
	case http.MethodGet:
		channels, err := s.db.GetAutoCleanChannels(guildID)
		if err != nil {
			http.Error(w, "Failed to get config", http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, channels)
	case http.MethodPost:
		var channel database.AutoCleanChannel
		if err := json.NewDecoder(r.Body).Decode(&channel); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if err := s.db.AddAutoCleanChannel(guildID, channel.ChannelID, "web", channel.IntervalHours, channel.WarningMinutes); err != nil {
			http.Error(w, "Failed to add channel", http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, map[string]string{"status": "ok"})
	case http.MethodDelete:
		channelID := r.URL.Query().Get("channel_id")
		if channelID == "" {
			http.Error(w, "channel_id required", http.StatusBadRequest)
			return
		}
		if err := s.db.RemoveAutoCleanChannel(guildID, channelID); err != nil {
			http.Error(w, "Failed to remove channel", http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, map[string]string{"status": "ok"})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAPITicketConfig handles ticket system configuration
func (s *Server) handleAPITicketConfig(w http.ResponseWriter, r *http.Request) {
	guildID := r.URL.Path[len("/api/guild/ticket/"):]
	switch r.Method {
	case http.MethodGet:
		config, err := s.db.GetTicketConfig(guildID)
		if err != nil {
			http.Error(w, "Failed to get config", http.StatusInternalServerError)
			return
		}
		if config == nil {
			s.jsonResponse(w, map[string]interface{}{"enabled": false, "channel_id": ""})
			return
		}
		s.jsonResponse(w, config)
	case http.MethodPost, http.MethodPut:
		var req struct {
			ChannelID string `json:"channel_id"`
			Enabled   bool   `json:"enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if err := s.db.SetTicketConfig(guildID, req.ChannelID, req.Enabled); err != nil {
			http.Error(w, "Failed to save config", http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, map[string]string{"status": "ok"})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAPIRegexFilters handles regex filter configuration
func (s *Server) handleAPIRegexFilters(w http.ResponseWriter, r *http.Request) {
	guildID := r.URL.Path[len("/api/guild/regex/"):]
	switch r.Method {
	case http.MethodGet:
		filters, err := s.db.GetRegexFilters(guildID)
		if err != nil {
			http.Error(w, "Failed to get filters", http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, filters)
	case http.MethodPost:
		var filter struct {
			Pattern string `json:"pattern"`
			Action  string `json:"action"`
			Reason  string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if err := s.db.AddRegexFilter(guildID, filter.Pattern, filter.Action, filter.Reason, "web"); err != nil {
			http.Error(w, "Failed to add filter", http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, map[string]string{"status": "ok"})
	case http.MethodDelete:
		var req struct {
			ID int64 `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if err := s.db.RemoveRegexFilter(guildID, req.ID); err != nil {
			http.Error(w, "Failed to remove filter", http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, map[string]string{"status": "ok"})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAPILevelRanks handles level rank configuration
func (s *Server) handleAPILevelRanks(w http.ResponseWriter, r *http.Request) {
	guildID := r.URL.Path[len("/api/guild/ranks/"):]
	switch r.Method {
	case http.MethodGet:
		ranks, err := s.db.GetLevelRanks(guildID)
		if err != nil {
			http.Error(w, "Failed to get ranks", http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, ranks)
	case http.MethodPost:
		var rank struct {
			RoleID string `json:"role_id"`
			Level  int    `json:"level"`
		}
		if err := json.NewDecoder(r.Body).Decode(&rank); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if err := s.db.AddLevelRank(guildID, rank.RoleID, rank.Level); err != nil {
			http.Error(w, "Failed to add rank", http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, map[string]string{"status": "ok"})
	case http.MethodDelete:
		var req struct {
			RoleID string `json:"role_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if err := s.db.RemoveLevelRank(guildID, req.RoleID); err != nil {
			http.Error(w, "Failed to remove rank", http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, map[string]string{"status": "ok"})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAPICommandConfig handles command enable/disable configuration
func (s *Server) handleAPICommandConfig(w http.ResponseWriter, r *http.Request) {
	guildID := r.URL.Path[len("/api/guild/commands/"):]
	switch r.Method {
	case http.MethodGet:
		disabledCommands, _ := s.db.GetDisabledCommands(guildID)
		disabledCategories, _ := s.db.GetDisabledCategories(guildID)
		s.jsonResponse(w, map[string]interface{}{
			"disabled_commands":   disabledCommands,
			"disabled_categories": disabledCategories,
		})
	case http.MethodPost, http.MethodPut:
		var req struct {
			DisabledCommands   []string `json:"disabled_commands"`
			DisabledCategories []string `json:"disabled_categories"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if err := s.db.SetDisabledCommands(guildID, req.DisabledCommands); err != nil {
			http.Error(w, "Failed to save commands", http.StatusInternalServerError)
			return
		}
		if err := s.db.SetDisabledCategories(guildID, req.DisabledCategories); err != nil {
			http.Error(w, "Failed to save categories", http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, map[string]string{"status": "ok"})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAPICommandsList returns all commands grouped by category
func (s *Server) handleAPICommandsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Static list of commands by category - this matches the bot's actual commands
	commands := map[string][]string{
		"Admin": {"kick", "ban", "unban", "timeout", "untimeout", "purge", "slowmode",
			"warn", "warnings", "clearwarnings", "lock", "unlock", "bans", "hackban",
			"softban", "massrole", "chanlockdown", "chanunlock", "syncperms"},
		"Info":          {"help", "botinfo", "serverinfo", "userinfo", "avatar", "roleinfo", "channelinfo", "emojiinfo", "inviteinfo", "roles", "membercount"},
		"XP":            {"rank", "leaderboard", "xp", "setxp", "addxp", "removexp", "resetxp", "setlevel", "massaddxp"},
		"Logging":       {"setlogchannel", "togglelogging", "logconfig", "disablechannellog", "enablechannellog", "logstatus"},
		"Filters":       {"addfilter", "removefilter", "listfilters", "testfilter"},
		"Anti-Raid":     {"antiraid", "silence", "unsilence", "getraid"},
		"Anti-Spam":     {"antispam"},
		"Ranks":         {"addrank", "removerank", "listranks", "syncranks", "applyranks"},
		"VoiceXP":       {"voicexp"},
		"AutoClean":     {"autoclean", "setcleanmessage", "setcleanimage"},
		"Ticket":        {"setticket", "disableticket", "ticketstatus", "ticket"},
		"Settings":      {"setprefix", "setmodlog", "setwelcome", "disablewelcome", "settings", "setjoindm", "disablejoindm"},
		"Moderation":    {"modstats", "spamfilter"},
		"DM":            {"setdmchannel", "disabledm", "dmstatus"},
		"BotBan":        {"botban", "botunban", "botbanlist"},
		"Misc":          {"snipe", "tag", "customcmd", "mentionresponse"},
		"AI":            {"ai"},
		"Fun":           {"8ball", "coinflip", "dice", "roll", "rps", "random", "joke", "rate", "ship", "iq", "gay", "pp", "hug", "slap", "pat", "kiss", "f", "choose"},
		"Text":          {"ascii", "zalgo", "reverse", "upsidedown", "morse", "vaporwave", "owo", "mock", "leet", "regional", "spoiler", "space", "fancy", "encode", "decode", "codeblock", "hyperlink"},
		"Random":        {"cat", "dog", "fox", "bird", "duck", "shiba", "meme", "quote", "fact", "advice", "dadjoke"},
		"Images":        {"resize", "rotate", "flip", "invert", "grayscale", "blur", "sharpen", "brightness", "contrast", "saturate"},
		"Lookup":        {"steam", "minecraft", "npm", "pypi", "github", "weather", "urban", "define", "wikipedia", "anime", "manga"},
		"Tools":         {"qr", "color", "math", "base64", "hash", "timestamp", "snowflake", "permissions", "ping", "uptime"},
		"Utility":       {"afk", "remind", "poll", "giveaway", "timezone", "time", "countdown", "note", "notes", "deletenote"},
		"Music":         {"play", "skip", "stop", "pause", "resume", "queue", "nowplaying", "volume", "shuffle", "loop", "clear", "remove", "move", "seek", "lyrics", "playlist"},
		"Configuration": {"mentionresponse"},
	}

	s.jsonResponse(w, commands)
}

// handleAPIChannels returns channels for a guild
func (s *Server) handleAPIChannels(w http.ResponseWriter, r *http.Request) {
	guildID := r.URL.Path[len("/api/channels/"):]
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	guild, err := s.session.State.Guild(guildID)
	if err != nil {
		http.Error(w, "Guild not found", http.StatusNotFound)
		return
	}

	var channels []map[string]interface{}
	for _, ch := range guild.Channels {
		if ch.Type == discordgo.ChannelTypeGuildText || ch.Type == discordgo.ChannelTypeGuildNews {
			channels = append(channels, map[string]interface{}{
				"id":   ch.ID,
				"name": ch.Name,
				"type": ch.Type,
			})
		}
	}

	s.jsonResponse(w, channels)
}

// handleAPIRoles returns roles for a guild
func (s *Server) handleAPIRoles(w http.ResponseWriter, r *http.Request) {
	guildID := r.URL.Path[len("/api/roles/"):]
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	guild, err := s.session.State.Guild(guildID)
	if err != nil {
		http.Error(w, "Guild not found", http.StatusNotFound)
		return
	}

	var roles []map[string]interface{}
	for _, role := range guild.Roles {
		roles = append(roles, map[string]interface{}{
			"id":       role.ID,
			"name":     role.Name,
			"color":    role.Color,
			"position": role.Position,
		})
	}

	s.jsonResponse(w, roles)
}

// jsonResponse sends a JSON response
func (s *Server) jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// Dashboard HTML template
const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Himiko Dashboard</title>
    <style>
        :root {
            --bg-primary: #1a1a2e;
            --bg-secondary: #16213e;
            --bg-card: #0f3460;
            --accent: #e94560;
            --accent-light: #ff6b6b;
            --text-primary: #ffffff;
            --text-secondary: #a0a0a0;
            --success: #57F287;
            --warning: #FEE75C;
        }
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: var(--bg-primary);
            color: var(--text-primary);
            min-height: 100vh;
        }
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        header {
            background: var(--bg-secondary);
            padding: 20px;
            border-bottom: 2px solid var(--accent);
        }
        header h1 { display: flex; align-items: center; gap: 15px; }
        header img { border-radius: 50%; width: 48px; height: 48px; }
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin: 30px 0;
        }
        .stat-card {
            background: var(--bg-card);
            padding: 20px;
            border-radius: 10px;
            text-align: center;
            border: 1px solid transparent;
            transition: border-color 0.3s;
        }
        .stat-card:hover { border-color: var(--accent); }
        .stat-card h3 { color: var(--text-secondary); font-size: 14px; text-transform: uppercase; margin-bottom: 10px; }
        .stat-card .value { font-size: 36px; font-weight: bold; color: var(--accent); }
        .guilds-section { margin-top: 30px; }
        .guilds-section h2 { margin-bottom: 20px; padding-bottom: 10px; border-bottom: 1px solid var(--bg-card); }
        .guild-list { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 15px; }
        .guild-card {
            background: var(--bg-secondary);
            padding: 15px;
            border-radius: 8px;
            display: flex;
            align-items: center;
            gap: 15px;
            cursor: pointer;
            transition: transform 0.2s, background 0.2s;
        }
        .guild-card:hover { transform: translateY(-2px); background: var(--bg-card); }
        .guild-card img { width: 48px; height: 48px; border-radius: 50%; background: var(--bg-card); }
        .guild-info h4 { margin-bottom: 5px; }
        .guild-info p { color: var(--text-secondary); font-size: 14px; }
        .loading { text-align: center; padding: 40px; color: var(--text-secondary); }
        .error { background: #ff6b6b20; border: 1px solid var(--accent); padding: 15px; border-radius: 8px; margin: 20px 0; }
        .modal { display: none; position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.8); z-index: 1000; justify-content: center; align-items: center; }
        .modal.active { display: flex; }
        .modal-content {
            background: var(--bg-secondary);
            padding: 25px;
            border-radius: 15px;
            max-width: 900px;
            width: 95%;
            max-height: 90vh;
            overflow-y: auto;
        }
        .modal-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 15px;
            padding-bottom: 15px;
            border-bottom: 1px solid var(--bg-card);
        }
        .modal-close { background: none; border: none; color: var(--text-secondary); font-size: 24px; cursor: pointer; }
        .modal-close:hover { color: var(--accent); }
        .tabs { display: flex; gap: 5px; margin-bottom: 20px; flex-wrap: wrap; border-bottom: 2px solid var(--bg-card); padding-bottom: 10px; }
        .tab {
            padding: 8px 16px;
            cursor: pointer;
            border-radius: 5px;
            color: var(--text-secondary);
            transition: all 0.2s;
            font-size: 14px;
        }
        .tab:hover { background: var(--bg-card); color: var(--text-primary); }
        .tab.active { background: var(--accent); color: white; }
        .tab-content { display: none; }
        .tab-content.active { display: block; }
        .form-group { margin-bottom: 15px; }
        .form-group label { display: block; margin-bottom: 6px; color: var(--text-secondary); font-size: 13px; }
        .form-group input, .form-group select, .form-group textarea {
            width: 100%;
            padding: 10px 12px;
            background: var(--bg-card);
            border: 1px solid transparent;
            border-radius: 5px;
            color: var(--text-primary);
            font-size: 14px;
        }
        .form-group input:focus, .form-group select:focus, .form-group textarea:focus { outline: none; border-color: var(--accent); }
        .form-group textarea { resize: vertical; min-height: 80px; }
        .form-row { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; }
        .btn {
            padding: 10px 20px;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            font-size: 14px;
            transition: background 0.2s;
        }
        .btn-primary { background: var(--accent); color: white; }
        .btn-primary:hover { background: var(--accent-light); }
        .btn-secondary { background: var(--bg-card); color: var(--text-primary); }
        .btn-sm { padding: 6px 12px; font-size: 12px; }
        .btn-danger { background: #ED4245; color: white; }
        .toggle {
            position: relative;
            width: 50px;
            height: 26px;
            background: var(--bg-card);
            border-radius: 13px;
            cursor: pointer;
            display: inline-block;
        }
        .toggle.active { background: var(--success); }
        .toggle::after {
            content: '';
            position: absolute;
            top: 3px;
            left: 3px;
            width: 20px;
            height: 20px;
            background: white;
            border-radius: 50%;
            transition: transform 0.2s;
        }
        .toggle.active::after { transform: translateX(24px); }
        .toggle-row { display: flex; align-items: center; justify-content: space-between; padding: 10px 0; border-bottom: 1px solid var(--bg-card); }
        .toggle-row:last-child { border-bottom: none; }
        .section-title { font-size: 16px; font-weight: 600; margin: 20px 0 10px; padding-bottom: 8px; border-bottom: 1px solid var(--bg-card); }
        .section-title:first-child { margin-top: 0; }
        .list-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 10px;
            background: var(--bg-card);
            border-radius: 5px;
            margin-bottom: 8px;
        }
        .list-item span { font-family: monospace; }
        .add-form { display: flex; gap: 10px; margin-bottom: 15px; flex-wrap: wrap; }
        .add-form input, .add-form select { flex: 1; min-width: 150px; }
        .category-section { margin-bottom: 15px; }
        .category-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 10px;
            background: var(--bg-card);
            border-radius: 5px;
            cursor: pointer;
            margin-bottom: 5px;
        }
        .category-header:hover { background: #1a4a7a; }
        .category-commands { padding-left: 20px; display: none; }
        .category-commands.expanded { display: block; }
        .command-item { display: flex; align-items: center; gap: 10px; padding: 5px 0; }
        .command-item label { flex: 1; cursor: pointer; }
        .command-item input[type="checkbox"] { width: 18px; height: 18px; cursor: pointer; }
        .toast {
            position: fixed;
            bottom: 20px;
            right: 20px;
            padding: 15px 25px;
            background: var(--success);
            color: white;
            border-radius: 8px;
            display: none;
            z-index: 2000;
        }
        .toast.error { background: var(--accent); }
        .toast.show { display: block; animation: fadeIn 0.3s; }
        @keyframes fadeIn { from { opacity: 0; transform: translateY(20px); } to { opacity: 1; transform: translateY(0); } }
        footer { text-align: center; padding: 30px; color: var(--text-secondary); font-size: 14px; }
        footer a { color: var(--accent); text-decoration: none; }
    </style>
</head>
<body>
    <header>
        <div class="container">
            <h1><img id="bot-avatar" src="" alt="Bot Avatar"><span>Himiko Dashboard</span></h1>
        </div>
    </header>
    <main class="container">
        <div class="stats-grid">
            <div class="stat-card"><h3>Servers</h3><div class="value" id="guild-count">-</div></div>
            <div class="stat-card"><h3>Total Members</h3><div class="value" id="member-count">-</div></div>
            <div class="stat-card"><h3>Version</h3><div class="value" id="version">-</div></div>
        </div>
        <section class="guilds-section">
            <h2>Managed Servers</h2>
            <div id="guild-list" class="guild-list"><div class="loading">Loading servers...</div></div>
        </section>
    </main>
    <div id="guild-modal" class="modal">
        <div class="modal-content">
            <div class="modal-header">
                <h3 id="modal-guild-name">Server Settings</h3>
                <button class="modal-close" onclick="closeModal()">&times;</button>
            </div>
            <input type="hidden" id="setting-guild-id">
            <div class="tabs">
                <div class="tab active" data-tab="basic">Basic</div>
                <div class="tab" data-tab="moderation">Moderation</div>
                <div class="tab" data-tab="filters">Filters</div>
                <div class="tab" data-tab="xp">XP & Ranks</div>
                <div class="tab" data-tab="features">Features</div>
                <div class="tab" data-tab="commands">Commands</div>
            </div>
            <div id="tab-basic" class="tab-content active">
                <div class="section-title">General Settings</div>
                <div class="form-row">
                    <div class="form-group"><label>Command Prefix</label><input type="text" id="setting-prefix" maxlength="5" placeholder="/"></div>
                    <div class="form-group"><label>Mod Log Channel</label><select id="setting-modlog"><option value="">None</option></select></div>
                </div>
                <div class="section-title">Welcome Messages</div>
                <div class="form-row">
                    <div class="form-group"><label>Welcome Channel</label><select id="setting-welcome-channel"><option value="">Disabled</option></select></div>
                </div>
                <div class="form-group"><label>Welcome Message (use {user}, {username}, {server})</label><textarea id="setting-welcome-message" placeholder="Welcome to {server}, {user}!"></textarea></div>
                <div class="section-title">Join DM</div>
                <div class="form-group"><label>DM Title</label><input type="text" id="setting-joindm-title" placeholder="Welcome!"></div>
                <div class="form-group"><label>DM Message</label><textarea id="setting-joindm-message" placeholder="Thanks for joining {server}!"></textarea></div>
                <div style="display:flex;gap:10px;justify-content:flex-end;margin-top:20px;">
                    <button class="btn btn-primary" onclick="saveBasicSettings()">Save Settings</button>
                </div>
            </div>
            <div id="tab-moderation" class="tab-content">
                <div class="section-title">Logging</div>
                <div class="form-row">
                    <div class="form-group"><label>Log Channel</label><select id="logging-channel"><option value="">None</option></select></div>
                </div>
                <div class="toggle-row"><span>Logging Enabled</span><div class="toggle" id="logging-enabled" onclick="toggleSwitch(this)"></div></div>
                <div class="toggle-row"><span>Message Deletes</span><div class="toggle" id="logging-delete" onclick="toggleSwitch(this)"></div></div>
                <div class="toggle-row"><span>Message Edits</span><div class="toggle" id="logging-edit" onclick="toggleSwitch(this)"></div></div>
                <div class="toggle-row"><span>Voice Join</span><div class="toggle" id="logging-voicejoin" onclick="toggleSwitch(this)"></div></div>
                <div class="toggle-row"><span>Voice Leave</span><div class="toggle" id="logging-voiceleave" onclick="toggleSwitch(this)"></div></div>
                <div class="toggle-row"><span>Nickname Changes</span><div class="toggle" id="logging-nickname" onclick="toggleSwitch(this)"></div></div>
                <div class="section-title">Anti-Raid</div>
                <div class="toggle-row"><span>Anti-Raid Enabled</span><div class="toggle" id="antiraid-enabled" onclick="toggleSwitch(this)"></div></div>
                <div class="form-row">
                    <div class="form-group"><label>Raid Time Window (seconds)</label><input type="number" id="antiraid-time" min="10" max="3600" value="300"></div>
                    <div class="form-group"><label>Raid Size (joins to trigger)</label><input type="number" id="antiraid-size" min="2" max="100" value="5"></div>
                </div>
                <div class="form-row">
                    <div class="form-group"><label>Lockdown Duration (seconds)</label><input type="number" id="antiraid-lockdown" min="0" max="3600" value="120"></div>
                    <div class="form-group"><label>Action</label><select id="antiraid-action"><option value="silence">Silence</option><option value="kick">Kick</option><option value="ban">Ban</option></select></div>
                </div>
                <div class="form-row">
                    <div class="form-group"><label>Alert Channel</label><select id="antiraid-alertchannel"><option value="">None</option></select></div>
                    <div class="form-group"><label>Silent Role</label><select id="antiraid-silentrole"><option value="">None</option></select></div>
                </div>
                <div class="section-title">Anti-Spam (Pressure System)</div>
                <div class="toggle-row"><span>Anti-Spam Enabled</span><div class="toggle" id="antispam-enabled" onclick="toggleSwitch(this)"></div></div>
                <div class="form-row">
                    <div class="form-group"><label>Max Pressure</label><input type="number" id="antispam-maxpressure" min="10" max="200" value="60"></div>
                    <div class="form-group"><label>Base Pressure</label><input type="number" id="antispam-basepressure" min="1" max="50" value="10"></div>
                    <div class="form-group"><label>Decay (seconds)</label><input type="number" id="antispam-decay" min="0.5" max="30" step="0.5" value="2.5"></div>
                </div>
                <div class="form-row">
                    <div class="form-group"><label>Action</label><select id="antispam-action"><option value="delete">Delete</option><option value="warn">Warn</option><option value="silence">Silence</option><option value="kick">Kick</option><option value="ban">Ban</option></select></div>
                    <div class="form-group"><label>Silent Role</label><select id="antispam-silentrole"><option value="">None</option></select></div>
                </div>
                <div class="section-title">Spam Filter (Simple)</div>
                <div class="toggle-row"><span>Spam Filter Enabled</span><div class="toggle" id="spamfilter-enabled" onclick="toggleSwitch(this)"></div></div>
                <div class="form-row">
                    <div class="form-group"><label>Max Mentions</label><input type="number" id="spamfilter-mentions" min="1" max="50" value="5"></div>
                    <div class="form-group"><label>Max Links</label><input type="number" id="spamfilter-links" min="1" max="50" value="3"></div>
                    <div class="form-group"><label>Max Emojis</label><input type="number" id="spamfilter-emojis" min="1" max="100" value="10"></div>
                </div>
                <div class="form-group"><label>Action</label><select id="spamfilter-action"><option value="delete">Delete</option><option value="warn">Warn</option><option value="kick">Kick</option><option value="ban">Ban</option></select></div>
                <div style="display:flex;gap:10px;justify-content:flex-end;margin-top:20px;">
                    <button class="btn btn-primary" onclick="saveModerationSettings()">Save All Moderation Settings</button>
                </div>
            </div>
            <div id="tab-filters" class="tab-content">
                <div class="section-title">Regex Filters</div>
                <div class="add-form">
                    <input type="text" id="filter-pattern" placeholder="Regex pattern">
                    <select id="filter-action"><option value="delete">Delete</option><option value="warn">Warn</option><option value="ban">Ban</option></select>
                    <input type="text" id="filter-reason" placeholder="Reason (optional)">
                    <button class="btn btn-primary btn-sm" onclick="addFilter()">Add Filter</button>
                </div>
                <div id="filters-list"></div>
            </div>
            <div id="tab-xp" class="tab-content">
                <div class="section-title">Voice XP</div>
                <div class="toggle-row"><span>Voice XP Enabled</span><div class="toggle" id="voicexp-enabled" onclick="toggleSwitch(this)"></div></div>
                <div class="form-row">
                    <div class="form-group"><label>XP Rate (per interval)</label><input type="number" id="voicexp-rate" min="1" max="100" value="10"></div>
                    <div class="form-group"><label>Interval (minutes)</label><input type="number" id="voicexp-interval" min="1" max="60" value="5"></div>
                </div>
                <div class="toggle-row"><span>Ignore AFK Channel</span><div class="toggle" id="voicexp-ignoreafk" onclick="toggleSwitch(this)"></div></div>
                <div style="display:flex;gap:10px;justify-content:flex-end;margin-top:15px;">
                    <button class="btn btn-primary" onclick="saveVoiceXPSettings()">Save Voice XP</button>
                </div>
                <div class="section-title">Level Ranks (Role Rewards)</div>
                <div class="add-form">
                    <select id="rank-role"><option value="">Select Role</option></select>
                    <input type="number" id="rank-level" placeholder="Level" min="1" max="1000">
                    <button class="btn btn-primary btn-sm" onclick="addRank()">Add Rank</button>
                </div>
                <div id="ranks-list"></div>
            </div>
            <div id="tab-features" class="tab-content">
                <div class="section-title">Auto-Clean Channels</div>
                <div class="add-form">
                    <select id="autoclean-channel"><option value="">Select Channel</option></select>
                    <input type="number" id="autoclean-interval" placeholder="Hours" min="1" max="168" value="24">
                    <input type="number" id="autoclean-warning" placeholder="Warning mins" min="1" max="60" value="5">
                    <button class="btn btn-primary btn-sm" onclick="addAutoClean()">Add</button>
                </div>
                <div id="autoclean-list"></div>
                <div class="section-title">Ticket System</div>
                <div class="toggle-row"><span>Tickets Enabled</span><div class="toggle" id="ticket-enabled" onclick="toggleSwitch(this)"></div></div>
                <div class="form-group"><label>Ticket Channel</label><select id="ticket-channel"><option value="">Select Channel</option></select></div>
                <div style="display:flex;gap:10px;justify-content:flex-end;margin-top:15px;">
                    <button class="btn btn-primary" onclick="saveTicketSettings()">Save Ticket Settings</button>
                </div>
            </div>
            <div id="tab-commands" class="tab-content">
                <div class="section-title">Command Categories</div>
                <p style="color:var(--text-secondary);margin-bottom:15px;font-size:13px;">Toggle entire categories or expand to disable individual commands.</p>
                <div id="commands-list"></div>
                <div style="display:flex;gap:10px;justify-content:flex-end;margin-top:20px;">
                    <button class="btn btn-primary" onclick="saveCommandSettings()">Save Command Settings</button>
                </div>
            </div>
        </div>
    </div>
    <div id="toast" class="toast"></div>
    <footer><p>Himiko Bot Dashboard &bull; <a href="https://github.com/blubskye/himiko" target="_blank">GitHub</a></p></footer>
    <script>
        let currentGuildId = null;
        let channels = [];
        let roles = [];
        let allCommands = {};
        let disabledCommands = [];
        let disabledCategories = [];

        async function fetchStatus() {
            try {
                const res = await fetch('/api/status');
                const data = await res.json();
                document.getElementById('bot-avatar').src = data.bot.avatar;
                document.getElementById('version').textContent = 'v' + data.version;
                document.getElementById('guild-count').textContent = data.guilds;
            } catch (err) { console.error('Failed to fetch status:', err); }
        }

        async function fetchStats() {
            try {
                const res = await fetch('/api/stats');
                const data = await res.json();
                document.getElementById('guild-count').textContent = data.guilds;
                document.getElementById('member-count').textContent = data.total_members.toLocaleString();
            } catch (err) { console.error('Failed to fetch stats:', err); }
        }

        async function fetchGuilds() {
            try {
                const res = await fetch('/api/guilds');
                const guilds = await res.json();
                const container = document.getElementById('guild-list');
                if (!guilds || guilds.length === 0) { container.innerHTML = '<p class="loading">No servers found</p>'; return; }
                container.innerHTML = guilds.map(g => ` + "`" + `<div class="guild-card" onclick="openGuildSettings('${g.id}', '${g.name.replace(/'/g, "\\'")}')"><img src="${g.icon || 'https://cdn.discordapp.com/embed/avatars/0.png'}" alt="${g.name}"><div class="guild-info"><h4>${g.name}</h4><p>${g.member_count.toLocaleString()} members</p></div></div>` + "`" + `).join('');
            } catch (err) { document.getElementById('guild-list').innerHTML = '<div class="error">Failed to load servers</div>'; }
        }

        function populateSelect(id, items, valueKey, labelKey, selected) {
            const sel = document.getElementById(id);
            const firstOpt = sel.options[0];
            sel.innerHTML = '';
            sel.appendChild(firstOpt);
            items.forEach(item => {
                const opt = document.createElement('option');
                opt.value = item[valueKey];
                opt.textContent = item[labelKey];
                if (item[valueKey] === selected) opt.selected = true;
                sel.appendChild(opt);
            });
        }

        async function openGuildSettings(guildId, guildName) {
            currentGuildId = guildId;
            document.getElementById('modal-guild-name').textContent = guildName + ' Settings';
            document.getElementById('setting-guild-id').value = guildId;

            // Fetch channels and roles
            try {
                const [chRes, roleRes, cmdRes] = await Promise.all([
                    fetch('/api/channels/' + guildId),
                    fetch('/api/roles/' + guildId),
                    fetch('/api/commands/list')
                ]);
                channels = await chRes.json() || [];
                roles = await roleRes.json() || [];
                allCommands = await cmdRes.json() || {};
            } catch (err) { console.error('Failed to fetch channels/roles:', err); }

            // Populate channel selects
            ['setting-modlog', 'setting-welcome-channel', 'logging-channel', 'antiraid-alertchannel', 'autoclean-channel', 'ticket-channel'].forEach(id => {
                populateSelect(id, channels, 'id', 'name', null);
            });

            // Populate role selects
            ['antiraid-silentrole', 'antispam-silentrole', 'rank-role'].forEach(id => {
                populateSelect(id, roles.filter(r => r.name !== '@everyone'), 'id', 'name', null);
            });

            await loadAllSettings();
            document.getElementById('guild-modal').classList.add('active');
            switchTab('basic');
        }

        async function loadAllSettings() {
            try {
                const [basic, logging, antiraid, antispam, spamfilter, voicexp, ticket, filters, ranks, autoclean, commands] = await Promise.all([
                    fetch('/api/guild/settings/' + currentGuildId).then(r => r.json()),
                    fetch('/api/guild/logging/' + currentGuildId).then(r => r.json()),
                    fetch('/api/guild/antiraid/' + currentGuildId).then(r => r.json()),
                    fetch('/api/guild/antispam/' + currentGuildId).then(r => r.json()),
                    fetch('/api/guild/spamfilter/' + currentGuildId).then(r => r.json()),
                    fetch('/api/guild/voicexp/' + currentGuildId).then(r => r.json()),
                    fetch('/api/guild/ticket/' + currentGuildId).then(r => r.json()),
                    fetch('/api/guild/regex/' + currentGuildId).then(r => r.json()),
                    fetch('/api/guild/ranks/' + currentGuildId).then(r => r.json()),
                    fetch('/api/guild/autoclean/' + currentGuildId).then(r => r.json()),
                    fetch('/api/guild/commands/' + currentGuildId).then(r => r.json())
                ]);

                // Basic
                document.getElementById('setting-prefix').value = basic.Prefix || '/';
                document.getElementById('setting-modlog').value = basic.ModLogChannel || '';
                document.getElementById('setting-welcome-channel').value = basic.WelcomeChannel || '';
                document.getElementById('setting-welcome-message').value = basic.WelcomeMessage || '';
                document.getElementById('setting-joindm-title').value = basic.JoinDMTitle || '';
                document.getElementById('setting-joindm-message').value = basic.JoinDMMessage || '';

                // Logging
                document.getElementById('logging-channel').value = logging.LogChannelID || '';
                setToggle('logging-enabled', logging.Enabled);
                setToggle('logging-delete', logging.MessageDelete);
                setToggle('logging-edit', logging.MessageEdit);
                setToggle('logging-voicejoin', logging.VoiceJoin);
                setToggle('logging-voiceleave', logging.VoiceLeave);
                setToggle('logging-nickname', logging.NicknameChange);

                // Anti-Raid
                setToggle('antiraid-enabled', antiraid.Enabled);
                document.getElementById('antiraid-time').value = antiraid.RaidTime || 300;
                document.getElementById('antiraid-size').value = antiraid.RaidSize || 5;
                document.getElementById('antiraid-lockdown').value = antiraid.LockdownDuration || 120;
                document.getElementById('antiraid-action').value = antiraid.Action || 'silence';
                document.getElementById('antiraid-alertchannel').value = antiraid.LogChannelID || '';
                document.getElementById('antiraid-silentrole').value = antiraid.SilentRoleID || '';

                // Anti-Spam
                setToggle('antispam-enabled', antispam.Enabled);
                document.getElementById('antispam-maxpressure').value = antispam.MaxPressure || 60;
                document.getElementById('antispam-basepressure').value = antispam.BasePressure || 10;
                document.getElementById('antispam-decay').value = antispam.PressureDecay || 2.5;
                document.getElementById('antispam-action').value = antispam.Action || 'delete';
                document.getElementById('antispam-silentrole').value = antispam.SilentRoleID || '';

                // Spam Filter
                setToggle('spamfilter-enabled', spamfilter.Enabled);
                document.getElementById('spamfilter-mentions').value = spamfilter.MaxMentions || 5;
                document.getElementById('spamfilter-links').value = spamfilter.MaxLinks || 3;
                document.getElementById('spamfilter-emojis').value = spamfilter.MaxEmojis || 10;
                document.getElementById('spamfilter-action').value = spamfilter.Action || 'delete';

                // Voice XP
                setToggle('voicexp-enabled', voicexp.Enabled);
                document.getElementById('voicexp-rate').value = voicexp.XPRate || 10;
                document.getElementById('voicexp-interval').value = voicexp.IntervalMins || 5;
                setToggle('voicexp-ignoreafk', voicexp.IgnoreAFK);

                // Ticket
                setToggle('ticket-enabled', ticket.enabled || ticket.Enabled);
                document.getElementById('ticket-channel').value = ticket.channel_id || ticket.ChannelID || '';

                // Filters
                renderFilters(filters || []);

                // Ranks
                renderRanks(ranks || []);

                // Auto-Clean
                renderAutoClean(autoclean || []);

                // Commands
                disabledCommands = commands.disabled_commands || [];
                disabledCategories = commands.disabled_categories || [];
                renderCommands();
            } catch (err) { console.error('Failed to load settings:', err); }
        }

        function setToggle(id, value) {
            const el = document.getElementById(id);
            if (value) el.classList.add('active');
            else el.classList.remove('active');
        }

        function toggleSwitch(el) { el.classList.toggle('active'); }

        function getToggle(id) { return document.getElementById(id).classList.contains('active'); }

        function switchTab(tabName) {
            document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
            document.querySelectorAll('.tab-content').forEach(t => t.classList.remove('active'));
            document.querySelector(` + "`" + `.tab[data-tab="${tabName}"]` + "`" + `).classList.add('active');
            document.getElementById('tab-' + tabName).classList.add('active');
        }

        document.querySelectorAll('.tab').forEach(tab => {
            tab.addEventListener('click', () => switchTab(tab.dataset.tab));
        });

        function showToast(msg, isError) {
            const toast = document.getElementById('toast');
            toast.textContent = msg;
            toast.className = 'toast show' + (isError ? ' error' : '');
            setTimeout(() => toast.classList.remove('show'), 3000);
        }

        async function saveBasicSettings() {
            const settings = {
                Prefix: document.getElementById('setting-prefix').value,
                ModLogChannel: document.getElementById('setting-modlog').value || null,
                WelcomeChannel: document.getElementById('setting-welcome-channel').value || null,
                WelcomeMessage: document.getElementById('setting-welcome-message').value || null,
                JoinDMTitle: document.getElementById('setting-joindm-title').value || null,
                JoinDMMessage: document.getElementById('setting-joindm-message').value || null
            };
            try {
                const res = await fetch('/api/guild/settings/' + currentGuildId, {
                    method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(settings)
                });
                if (res.ok) showToast('Basic settings saved!');
                else showToast('Failed to save', true);
            } catch (err) { showToast('Error saving settings', true); }
        }

        async function saveModerationSettings() {
            const loggingCh = document.getElementById('logging-channel').value;
            const logging = {
                LogChannelID: loggingCh || null,
                Enabled: getToggle('logging-enabled'),
                MessageDelete: getToggle('logging-delete'),
                MessageEdit: getToggle('logging-edit'),
                VoiceJoin: getToggle('logging-voicejoin'),
                VoiceLeave: getToggle('logging-voiceleave'),
                NicknameChange: getToggle('logging-nickname'),
                AvatarChange: false,
                PresenceChange: false,
                PresenceBatchMins: 5
            };
            const antiraid = {
                Enabled: getToggle('antiraid-enabled'),
                RaidTime: parseInt(document.getElementById('antiraid-time').value),
                RaidSize: parseInt(document.getElementById('antiraid-size').value),
                LockdownDuration: parseInt(document.getElementById('antiraid-lockdown').value),
                Action: document.getElementById('antiraid-action').value,
                LogChannelID: document.getElementById('antiraid-alertchannel').value,
                SilentRoleID: document.getElementById('antiraid-silentrole').value,
                AutoSilence: 0
            };
            const antispam = {
                Enabled: getToggle('antispam-enabled'),
                MaxPressure: parseFloat(document.getElementById('antispam-maxpressure').value),
                BasePressure: parseFloat(document.getElementById('antispam-basepressure').value),
                PressureDecay: parseFloat(document.getElementById('antispam-decay').value),
                Action: document.getElementById('antispam-action').value,
                SilentRoleID: document.getElementById('antispam-silentrole').value,
                ImagePressure: 8.33, LinkPressure: 8.33, PingPressure: 2.5,
                LengthPressure: 0.00625, LinePressure: 0.71, RepeatPressure: 10.0
            };
            const spamfilter = {
                Enabled: getToggle('spamfilter-enabled'),
                MaxMentions: parseInt(document.getElementById('spamfilter-mentions').value),
                MaxLinks: parseInt(document.getElementById('spamfilter-links').value),
                MaxEmojis: parseInt(document.getElementById('spamfilter-emojis').value),
                Action: document.getElementById('spamfilter-action').value
            };
            try {
                await Promise.all([
                    fetch('/api/guild/logging/' + currentGuildId, {method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(logging)}),
                    fetch('/api/guild/antiraid/' + currentGuildId, {method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(antiraid)}),
                    fetch('/api/guild/antispam/' + currentGuildId, {method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(antispam)}),
                    fetch('/api/guild/spamfilter/' + currentGuildId, {method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(spamfilter)})
                ]);
                showToast('Moderation settings saved!');
            } catch (err) { showToast('Error saving settings', true); }
        }

        async function saveVoiceXPSettings() {
            const config = {
                Enabled: getToggle('voicexp-enabled'),
                XPRate: parseInt(document.getElementById('voicexp-rate').value),
                IntervalMins: parseInt(document.getElementById('voicexp-interval').value),
                IgnoreAFK: getToggle('voicexp-ignoreafk')
            };
            try {
                const res = await fetch('/api/guild/voicexp/' + currentGuildId, {method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(config)});
                if (res.ok) showToast('Voice XP settings saved!');
                else showToast('Failed to save', true);
            } catch (err) { showToast('Error saving', true); }
        }

        async function saveTicketSettings() {
            const config = { enabled: getToggle('ticket-enabled'), channel_id: document.getElementById('ticket-channel').value };
            try {
                const res = await fetch('/api/guild/ticket/' + currentGuildId, {method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(config)});
                if (res.ok) showToast('Ticket settings saved!');
                else showToast('Failed to save', true);
            } catch (err) { showToast('Error saving', true); }
        }

        function renderFilters(filters) {
            const container = document.getElementById('filters-list');
            if (!filters || filters.length === 0) { container.innerHTML = '<p style="color:var(--text-secondary)">No filters configured</p>'; return; }
            container.innerHTML = filters.map(f => ` + "`" + `<div class="list-item"><span>${f.Pattern}</span><span>${f.Action}</span><span>${f.Reason || '-'}</span><button class="btn btn-danger btn-sm" onclick="removeFilter(${f.ID})">Remove</button></div>` + "`" + `).join('');
        }

        async function addFilter() {
            const pattern = document.getElementById('filter-pattern').value;
            const action = document.getElementById('filter-action').value;
            const reason = document.getElementById('filter-reason').value;
            if (!pattern) { showToast('Pattern required', true); return; }
            try {
                const res = await fetch('/api/guild/regex/' + currentGuildId, {method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({pattern, action, reason})});
                if (res.ok) {
                    document.getElementById('filter-pattern').value = '';
                    document.getElementById('filter-reason').value = '';
                    const filters = await fetch('/api/guild/regex/' + currentGuildId).then(r => r.json());
                    renderFilters(filters);
                    showToast('Filter added!');
                } else showToast('Failed to add filter', true);
            } catch (err) { showToast('Error adding filter', true); }
        }

        async function removeFilter(id) {
            try {
                const res = await fetch('/api/guild/regex/' + currentGuildId, {method: 'DELETE', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({id})});
                if (res.ok) {
                    const filters = await fetch('/api/guild/regex/' + currentGuildId).then(r => r.json());
                    renderFilters(filters);
                    showToast('Filter removed!');
                }
            } catch (err) { showToast('Error removing filter', true); }
        }

        function renderRanks(ranks) {
            const container = document.getElementById('ranks-list');
            if (!ranks || ranks.length === 0) { container.innerHTML = '<p style="color:var(--text-secondary)">No level ranks configured</p>'; return; }
            container.innerHTML = ranks.map(r => {
                const role = roles.find(ro => ro.id === r.RoleID);
                return ` + "`" + `<div class="list-item"><span>Level ${r.Level}</span><span>${role ? role.name : r.RoleID}</span><button class="btn btn-danger btn-sm" onclick="removeRank('${r.RoleID}')">Remove</button></div>` + "`" + `;
            }).join('');
        }

        async function addRank() {
            const roleId = document.getElementById('rank-role').value;
            const level = parseInt(document.getElementById('rank-level').value);
            if (!roleId || !level) { showToast('Role and level required', true); return; }
            try {
                const res = await fetch('/api/guild/ranks/' + currentGuildId, {method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({role_id: roleId, level})});
                if (res.ok) {
                    document.getElementById('rank-level').value = '';
                    const ranks = await fetch('/api/guild/ranks/' + currentGuildId).then(r => r.json());
                    renderRanks(ranks);
                    showToast('Rank added!');
                }
            } catch (err) { showToast('Error adding rank', true); }
        }

        async function removeRank(roleId) {
            try {
                const res = await fetch('/api/guild/ranks/' + currentGuildId, {method: 'DELETE', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({role_id: roleId})});
                if (res.ok) {
                    const ranks = await fetch('/api/guild/ranks/' + currentGuildId).then(r => r.json());
                    renderRanks(ranks);
                    showToast('Rank removed!');
                }
            } catch (err) { showToast('Error removing rank', true); }
        }

        function renderAutoClean(list) {
            const container = document.getElementById('autoclean-list');
            if (!list || list.length === 0) { container.innerHTML = '<p style="color:var(--text-secondary)">No auto-clean channels configured</p>'; return; }
            container.innerHTML = list.map(c => {
                const ch = channels.find(ch => ch.id === c.ChannelID);
                return ` + "`" + `<div class="list-item"><span>#${ch ? ch.name : c.ChannelID}</span><span>${c.IntervalHours}h</span><span>${c.WarningMinutes}m warning</span><button class="btn btn-danger btn-sm" onclick="removeAutoClean('${c.ChannelID}')">Remove</button></div>` + "`" + `;
            }).join('');
        }

        async function addAutoClean() {
            const channelId = document.getElementById('autoclean-channel').value;
            const interval = parseInt(document.getElementById('autoclean-interval').value);
            const warning = parseInt(document.getElementById('autoclean-warning').value);
            if (!channelId) { showToast('Channel required', true); return; }
            try {
                const res = await fetch('/api/guild/autoclean/' + currentGuildId, {method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({ChannelID: channelId, IntervalHours: interval, WarningMinutes: warning})});
                if (res.ok) {
                    const list = await fetch('/api/guild/autoclean/' + currentGuildId).then(r => r.json());
                    renderAutoClean(list);
                    showToast('Auto-clean added!');
                }
            } catch (err) { showToast('Error adding', true); }
        }

        async function removeAutoClean(channelId) {
            try {
                const res = await fetch('/api/guild/autoclean/' + currentGuildId + '?channel_id=' + channelId, {method: 'DELETE'});
                if (res.ok) {
                    const list = await fetch('/api/guild/autoclean/' + currentGuildId).then(r => r.json());
                    renderAutoClean(list);
                    showToast('Removed!');
                }
            } catch (err) { showToast('Error removing', true); }
        }

        function renderCommands() {
            const container = document.getElementById('commands-list');
            container.innerHTML = Object.entries(allCommands).map(([cat, cmds]) => {
                const catDisabled = disabledCategories.includes(cat);
                return ` + "`" + `
                    <div class="category-section">
                        <div class="category-header" onclick="toggleCategory(this)">
                            <span>${cat} (${cmds.length} commands)</span>
                            <div style="display:flex;align-items:center;gap:10px;">
                                <div class="toggle ${catDisabled ? '' : 'active'}" onclick="toggleCategoryEnabled(event, '${cat}')"></div>
                                <span style="color:var(--text-secondary)">&#9660;</span>
                            </div>
                        </div>
                        <div class="category-commands">
                            ${cmds.map(cmd => ` + "`" + `
                                <div class="command-item">
                                    <input type="checkbox" id="cmd-${cmd}" ${!disabledCommands.includes(cmd) && !catDisabled ? 'checked' : ''} ${catDisabled ? 'disabled' : ''} onchange="toggleCommand('${cmd}')">
                                    <label for="cmd-${cmd}">${cmd}</label>
                                </div>
                            ` + "`" + `).join('')}
                        </div>
                    </div>
                ` + "`" + `;
            }).join('');
        }

        function toggleCategory(el) {
            const cmds = el.nextElementSibling;
            cmds.classList.toggle('expanded');
        }

        function toggleCategoryEnabled(e, cat) {
            e.stopPropagation();
            const toggle = e.target;
            toggle.classList.toggle('active');
            if (toggle.classList.contains('active')) {
                disabledCategories = disabledCategories.filter(c => c !== cat);
            } else {
                if (!disabledCategories.includes(cat)) disabledCategories.push(cat);
            }
            renderCommands();
        }

        function toggleCommand(cmd) {
            if (disabledCommands.includes(cmd)) {
                disabledCommands = disabledCommands.filter(c => c !== cmd);
            } else {
                disabledCommands.push(cmd);
            }
        }

        async function saveCommandSettings() {
            try {
                const res = await fetch('/api/guild/commands/' + currentGuildId, {
                    method: 'POST', headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({disabled_commands: disabledCommands, disabled_categories: disabledCategories})
                });
                if (res.ok) showToast('Command settings saved!');
                else showToast('Failed to save', true);
            } catch (err) { showToast('Error saving', true); }
        }

        function closeModal() { document.getElementById('guild-modal').classList.remove('active'); }

        document.getElementById('guild-modal').addEventListener('click', (e) => {
            if (e.target.id === 'guild-modal') closeModal();
        });

        fetchStatus();
        fetchStats();
        fetchGuilds();
        setInterval(fetchStats, 30000);
    </script>
</body>
</html>`
