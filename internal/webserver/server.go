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
        }

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            background: var(--bg-primary);
            color: var(--text-primary);
            min-height: 100vh;
        }

        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }

        header {
            background: var(--bg-secondary);
            padding: 20px;
            border-bottom: 2px solid var(--accent);
        }

        header h1 {
            display: flex;
            align-items: center;
            gap: 15px;
        }

        header img {
            border-radius: 50%;
            width: 48px;
            height: 48px;
        }

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

        .stat-card:hover {
            border-color: var(--accent);
        }

        .stat-card h3 {
            color: var(--text-secondary);
            font-size: 14px;
            text-transform: uppercase;
            margin-bottom: 10px;
        }

        .stat-card .value {
            font-size: 36px;
            font-weight: bold;
            color: var(--accent);
        }

        .guilds-section {
            margin-top: 30px;
        }

        .guilds-section h2 {
            margin-bottom: 20px;
            padding-bottom: 10px;
            border-bottom: 1px solid var(--bg-card);
        }

        .guild-list {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
            gap: 15px;
        }

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

        .guild-card:hover {
            transform: translateY(-2px);
            background: var(--bg-card);
        }

        .guild-card img {
            width: 48px;
            height: 48px;
            border-radius: 50%;
            background: var(--bg-card);
        }

        .guild-info h4 {
            margin-bottom: 5px;
        }

        .guild-info p {
            color: var(--text-secondary);
            font-size: 14px;
        }

        .loading {
            text-align: center;
            padding: 40px;
            color: var(--text-secondary);
        }

        .error {
            background: #ff6b6b20;
            border: 1px solid var(--accent);
            padding: 15px;
            border-radius: 8px;
            margin: 20px 0;
        }

        /* Modal styles */
        .modal {
            display: none;
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0,0,0,0.8);
            z-index: 1000;
            justify-content: center;
            align-items: center;
        }

        .modal.active {
            display: flex;
        }

        .modal-content {
            background: var(--bg-secondary);
            padding: 30px;
            border-radius: 15px;
            max-width: 600px;
            width: 90%;
            max-height: 80vh;
            overflow-y: auto;
        }

        .modal-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 20px;
            padding-bottom: 15px;
            border-bottom: 1px solid var(--bg-card);
        }

        .modal-close {
            background: none;
            border: none;
            color: var(--text-secondary);
            font-size: 24px;
            cursor: pointer;
        }

        .modal-close:hover {
            color: var(--accent);
        }

        .form-group {
            margin-bottom: 20px;
        }

        .form-group label {
            display: block;
            margin-bottom: 8px;
            color: var(--text-secondary);
            font-size: 14px;
        }

        .form-group input, .form-group select {
            width: 100%;
            padding: 10px 15px;
            background: var(--bg-card);
            border: 1px solid transparent;
            border-radius: 5px;
            color: var(--text-primary);
            font-size: 16px;
        }

        .form-group input:focus, .form-group select:focus {
            outline: none;
            border-color: var(--accent);
        }

        .btn {
            padding: 10px 20px;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            font-size: 14px;
            transition: background 0.2s;
        }

        .btn-primary {
            background: var(--accent);
            color: white;
        }

        .btn-primary:hover {
            background: var(--accent-light);
        }

        .btn-secondary {
            background: var(--bg-card);
            color: var(--text-primary);
        }

        footer {
            text-align: center;
            padding: 30px;
            color: var(--text-secondary);
            font-size: 14px;
        }

        footer a {
            color: var(--accent);
            text-decoration: none;
        }
    </style>
</head>
<body>
    <header>
        <div class="container">
            <h1>
                <img id="bot-avatar" src="" alt="Bot Avatar">
                <span>Himiko Dashboard</span>
            </h1>
        </div>
    </header>

    <main class="container">
        <div class="stats-grid">
            <div class="stat-card">
                <h3>Servers</h3>
                <div class="value" id="guild-count">-</div>
            </div>
            <div class="stat-card">
                <h3>Total Members</h3>
                <div class="value" id="member-count">-</div>
            </div>
            <div class="stat-card">
                <h3>Version</h3>
                <div class="value" id="version">-</div>
            </div>
        </div>

        <section class="guilds-section">
            <h2>Managed Servers</h2>
            <div id="guild-list" class="guild-list">
                <div class="loading">Loading servers...</div>
            </div>
        </section>
    </main>

    <!-- Guild Settings Modal -->
    <div id="guild-modal" class="modal">
        <div class="modal-content">
            <div class="modal-header">
                <h3 id="modal-guild-name">Server Settings</h3>
                <button class="modal-close" onclick="closeModal()">&times;</button>
            </div>
            <form id="settings-form">
                <input type="hidden" id="setting-guild-id">
                <div class="form-group">
                    <label for="setting-prefix">Command Prefix</label>
                    <input type="text" id="setting-prefix" placeholder="/" maxlength="5">
                </div>
                <div class="form-group">
                    <label for="setting-welcome-message">Welcome Message</label>
                    <input type="text" id="setting-welcome-message" placeholder="Welcome to the server, {user}!">
                </div>
                <div style="display: flex; gap: 10px; justify-content: flex-end;">
                    <button type="button" class="btn btn-secondary" onclick="closeModal()">Cancel</button>
                    <button type="submit" class="btn btn-primary">Save Settings</button>
                </div>
            </form>
        </div>
    </div>

    <footer>
        <p>Himiko Bot Dashboard &bull; <a href="https://github.com/blubskye/himiko" target="_blank">GitHub</a></p>
    </footer>

    <script>
        // Fetch bot status
        async function fetchStatus() {
            try {
                const res = await fetch('/api/status');
                const data = await res.json();

                document.getElementById('bot-avatar').src = data.bot.avatar;
                document.getElementById('version').textContent = 'v' + data.version;
                document.getElementById('guild-count').textContent = data.guilds;
            } catch (err) {
                console.error('Failed to fetch status:', err);
            }
        }

        // Fetch stats
        async function fetchStats() {
            try {
                const res = await fetch('/api/stats');
                const data = await res.json();

                document.getElementById('guild-count').textContent = data.guilds;
                document.getElementById('member-count').textContent = data.total_members.toLocaleString();
            } catch (err) {
                console.error('Failed to fetch stats:', err);
            }
        }

        // Fetch guilds
        async function fetchGuilds() {
            try {
                const res = await fetch('/api/guilds');
                const guilds = await res.json();

                const container = document.getElementById('guild-list');

                if (!guilds || guilds.length === 0) {
                    container.innerHTML = '<p class="loading">No servers found</p>';
                    return;
                }

                container.innerHTML = guilds.map(guild => ` + "`" + `
                    <div class="guild-card" onclick="openGuildSettings('${guild.id}', '${guild.name.replace(/'/g, "\\'")}')">
                        <img src="${guild.icon || 'https://cdn.discordapp.com/embed/avatars/0.png'}" alt="${guild.name}">
                        <div class="guild-info">
                            <h4>${guild.name}</h4>
                            <p>${guild.member_count.toLocaleString()} members</p>
                        </div>
                    </div>
                ` + "`" + `).join('');
            } catch (err) {
                console.error('Failed to fetch guilds:', err);
                document.getElementById('guild-list').innerHTML = '<div class="error">Failed to load servers</div>';
            }
        }

        // Open guild settings modal
        async function openGuildSettings(guildId, guildName) {
            document.getElementById('modal-guild-name').textContent = guildName + ' Settings';
            document.getElementById('setting-guild-id').value = guildId;

            try {
                const res = await fetch('/api/guild/' + guildId);
                const data = await res.json();

                document.getElementById('setting-prefix').value = data.settings.prefix || '/';
                document.getElementById('setting-welcome-message').value = data.settings.welcome_message || '';
            } catch (err) {
                console.error('Failed to fetch guild settings:', err);
            }

            document.getElementById('guild-modal').classList.add('active');
        }

        // Close modal
        function closeModal() {
            document.getElementById('guild-modal').classList.remove('active');
        }

        // Save settings
        document.getElementById('settings-form').addEventListener('submit', async (e) => {
            e.preventDefault();

            const guildId = document.getElementById('setting-guild-id').value;
            const settings = {
                prefix: document.getElementById('setting-prefix').value,
                welcome_message: document.getElementById('setting-welcome-message').value
            };

            try {
                const res = await fetch('/api/guild/settings/' + guildId, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(settings)
                });

                if (res.ok) {
                    closeModal();
                    alert('Settings saved successfully!');
                } else {
                    alert('Failed to save settings');
                }
            } catch (err) {
                console.error('Failed to save settings:', err);
                alert('Failed to save settings');
            }
        });

        // Close modal on outside click
        document.getElementById('guild-modal').addEventListener('click', (e) => {
            if (e.target.id === 'guild-modal') {
                closeModal();
            }
        });

        // Initialize
        fetchStatus();
        fetchStats();
        fetchGuilds();

        // Auto-refresh stats every 30 seconds
        setInterval(fetchStats, 30000);
    </script>
</body>
</html>`
