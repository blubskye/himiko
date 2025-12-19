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

package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/blubskye/himiko/internal/crypto"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
	path      string
	encryptor *crypto.FieldEncryptor
}

// New creates a new database connection without encryption.
// Use NewWithEncryption to enable field-level encryption.
func New(path string) (*DB, error) {
	return NewWithEncryption(path, "")
}

// NewWithEncryption creates a new database connection with optional field-level encryption.
// If encryptionKey is empty, encryption is disabled.
func NewWithEncryption(path string, encryptionKey string) (*DB, error) {
	db, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	encryptor, err := crypto.NewFieldEncryptor(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	d := &DB{DB: db, path: path, encryptor: encryptor}
	if err := d.migrate(); err != nil {
		return nil, err
	}

	return d, nil
}

// GetPath returns the database file path
func (d *DB) GetPath() string {
	return d.path
}

// IsEncryptionEnabled returns whether field-level encryption is enabled
func (d *DB) IsEncryptionEnabled() bool {
	return d.encryptor.IsEnabled()
}

// Encrypt encrypts a string if encryption is enabled
func (d *DB) Encrypt(plaintext string) string {
	result, _ := d.encryptor.Encrypt(plaintext)
	return result
}

// Decrypt decrypts a string if encryption is enabled
func (d *DB) Decrypt(ciphertext string) string {
	result, _ := d.encryptor.Decrypt(ciphertext)
	return result
}

// EncryptNullable encrypts a nullable string
func (d *DB) EncryptNullable(plaintext *string) *string {
	result, _ := d.encryptor.EncryptNullable(plaintext)
	return result
}

// DecryptNullable decrypts a nullable string
func (d *DB) DecryptNullable(ciphertext *string) *string {
	result, _ := d.encryptor.DecryptNullable(ciphertext)
	return result
}

// IsDataEncrypted checks if a field value appears to be encrypted
func (d *DB) IsDataEncrypted(data string) bool {
	return d.encryptor.IsEncrypted(data)
}

func (d *DB) migrate() error {
	schema := `
	-- Guild settings
	CREATE TABLE IF NOT EXISTS guild_settings (
		guild_id TEXT PRIMARY KEY,
		prefix TEXT DEFAULT '/',
		mod_log_channel TEXT,
		welcome_channel TEXT,
		welcome_message TEXT,
		join_dm_title TEXT,
		join_dm_message TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Custom commands
	CREATE TABLE IF NOT EXISTS custom_commands (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guild_id TEXT NOT NULL,
		name TEXT NOT NULL,
		response TEXT NOT NULL,
		created_by TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		use_count INTEGER DEFAULT 0,
		UNIQUE(guild_id, name)
	);

	-- Command history
	CREATE TABLE IF NOT EXISTS command_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guild_id TEXT,
		channel_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		command TEXT NOT NULL,
		args TEXT,
		executed_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Warnings
	CREATE TABLE IF NOT EXISTS warnings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guild_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		moderator_id TEXT NOT NULL,
		reason TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Deleted messages log (for snipe command)
	CREATE TABLE IF NOT EXISTS deleted_messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guild_id TEXT,
		channel_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		content TEXT NOT NULL,
		deleted_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- User profiles/notes
	CREATE TABLE IF NOT EXISTS user_notes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guild_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		note TEXT NOT NULL,
		created_by TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(guild_id, user_id)
	);

	-- Scheduled messages
	CREATE TABLE IF NOT EXISTS scheduled_messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guild_id TEXT,
		channel_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		message TEXT NOT NULL,
		scheduled_for DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		executed INTEGER DEFAULT 0
	);

	-- AFK status
	CREATE TABLE IF NOT EXISTS afk_status (
		user_id TEXT PRIMARY KEY,
		message TEXT,
		set_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Reminders
	CREATE TABLE IF NOT EXISTS reminders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		channel_id TEXT NOT NULL,
		message TEXT NOT NULL,
		remind_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		completed INTEGER DEFAULT 0
	);

	-- Tags/snippets
	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guild_id TEXT NOT NULL,
		name TEXT NOT NULL,
		content TEXT NOT NULL,
		created_by TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		use_count INTEGER DEFAULT 0,
		UNIQUE(guild_id, name)
	);

	-- Keyword notifications
	CREATE TABLE IF NOT EXISTS keyword_notifications (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		guild_id TEXT,
		keyword TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, keyword)
	);

	CREATE INDEX IF NOT EXISTS idx_custom_commands_guild ON custom_commands(guild_id);
	CREATE INDEX IF NOT EXISTS idx_warnings_guild_user ON warnings(guild_id, user_id);
	CREATE INDEX IF NOT EXISTS idx_deleted_messages_channel ON deleted_messages(channel_id);
	CREATE INDEX IF NOT EXISTS idx_scheduled_messages_time ON scheduled_messages(scheduled_for);
	CREATE INDEX IF NOT EXISTS idx_reminders_time ON reminders(remind_at);

	-- XP/Leveling system
	CREATE TABLE IF NOT EXISTS user_xp (
		guild_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		xp INTEGER DEFAULT 0,
		level INTEGER DEFAULT 0,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (guild_id, user_id)
	);

	-- Regex filters for auto-moderation
	CREATE TABLE IF NOT EXISTS regex_filters (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guild_id TEXT NOT NULL,
		pattern TEXT NOT NULL,
		action TEXT NOT NULL,
		reason TEXT,
		created_by TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Auto-clean channels
	CREATE TABLE IF NOT EXISTS autoclean_channels (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guild_id TEXT NOT NULL,
		channel_id TEXT NOT NULL,
		interval_hours INTEGER DEFAULT 24,
		warning_minutes INTEGER DEFAULT 5,
		next_run DATETIME,
		clean_message INTEGER DEFAULT 1,
		clean_image INTEGER DEFAULT 1,
		created_by TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(guild_id, channel_id)
	);

	-- Logging configuration
	CREATE TABLE IF NOT EXISTS logging_config (
		guild_id TEXT PRIMARY KEY,
		log_channel_id TEXT,
		enabled INTEGER DEFAULT 0,
		message_delete INTEGER DEFAULT 1,
		message_edit INTEGER DEFAULT 1,
		voice_join INTEGER DEFAULT 1,
		voice_leave INTEGER DEFAULT 1,
		nickname_change INTEGER DEFAULT 1,
		avatar_change INTEGER DEFAULT 0,
		presence_change INTEGER DEFAULT 0,
		presence_batch_mins INTEGER DEFAULT 5
	);

	-- Disabled log channels (channels to ignore for logging)
	CREATE TABLE IF NOT EXISTS disabled_log_channels (
		guild_id TEXT NOT NULL,
		channel_id TEXT NOT NULL,
		PRIMARY KEY (guild_id, channel_id)
	);

	-- Voice XP configuration
	CREATE TABLE IF NOT EXISTS voice_xp_config (
		guild_id TEXT PRIMARY KEY,
		enabled INTEGER DEFAULT 0,
		xp_rate INTEGER DEFAULT 10,
		interval_mins INTEGER DEFAULT 5,
		ignore_afk INTEGER DEFAULT 1
	);

	-- Level ranks (role rewards)
	CREATE TABLE IF NOT EXISTS level_ranks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guild_id TEXT NOT NULL,
		role_id TEXT NOT NULL,
		level INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(guild_id, role_id)
	);

	-- DM forwarding configuration
	CREATE TABLE IF NOT EXISTS dm_config (
		guild_id TEXT PRIMARY KEY,
		channel_id TEXT NOT NULL,
		enabled INTEGER DEFAULT 1
	);

	-- Bot-level bans
	CREATE TABLE IF NOT EXISTS bot_bans (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		target_id TEXT NOT NULL UNIQUE,
		ban_type TEXT NOT NULL,
		reason TEXT,
		banned_by TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Moderation actions tracking
	CREATE TABLE IF NOT EXISTS mod_actions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guild_id TEXT NOT NULL,
		moderator_id TEXT NOT NULL,
		target_id TEXT NOT NULL,
		action TEXT NOT NULL,
		reason TEXT,
		timestamp INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Mention responses (custom triggers)
	CREATE TABLE IF NOT EXISTS mention_responses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guild_id TEXT NOT NULL,
		trigger_text TEXT NOT NULL,
		response TEXT NOT NULL,
		image_url TEXT,
		created_by TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(guild_id, trigger_text)
	);

	-- Spam filter configuration
	CREATE TABLE IF NOT EXISTS spam_filter_config (
		guild_id TEXT PRIMARY KEY,
		enabled INTEGER DEFAULT 0,
		max_mentions INTEGER DEFAULT 5,
		max_links INTEGER DEFAULT 3,
		max_emojis INTEGER DEFAULT 10,
		action TEXT DEFAULT 'delete'
	);

	-- Ticket system configuration
	CREATE TABLE IF NOT EXISTS ticket_config (
		guild_id TEXT PRIMARY KEY,
		channel_id TEXT NOT NULL,
		enabled INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Anti-raid configuration
	CREATE TABLE IF NOT EXISTS antiraid_config (
		guild_id TEXT PRIMARY KEY,
		enabled INTEGER DEFAULT 0,
		raid_time INTEGER DEFAULT 300,
		raid_size INTEGER DEFAULT 5,
		auto_silence INTEGER DEFAULT 0,
		lockdown_duration INTEGER DEFAULT 120,
		silent_role_id TEXT,
		alert_role_id TEXT,
		log_channel_id TEXT,
		action TEXT DEFAULT 'silence'
	);

	-- Member join tracking for raid detection
	CREATE TABLE IF NOT EXISTS member_joins (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guild_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		joined_at INTEGER NOT NULL,
		account_created_at INTEGER NOT NULL
	);

	-- Spam pressure tracking config
	CREATE TABLE IF NOT EXISTS antispam_config (
		guild_id TEXT PRIMARY KEY,
		enabled INTEGER DEFAULT 0,
		base_pressure REAL DEFAULT 10.0,
		image_pressure REAL DEFAULT 8.33,
		link_pressure REAL DEFAULT 8.33,
		ping_pressure REAL DEFAULT 2.5,
		length_pressure REAL DEFAULT 0.00625,
		line_pressure REAL DEFAULT 0.71,
		repeat_pressure REAL DEFAULT 10.0,
		max_pressure REAL DEFAULT 60.0,
		pressure_decay REAL DEFAULT 2.5,
		action TEXT DEFAULT 'delete',
		silent_role_id TEXT
	);

	-- Scheduled events (for timed unsilence, etc)
	CREATE TABLE IF NOT EXISTS scheduled_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guild_id TEXT NOT NULL,
		event_type TEXT NOT NULL,
		target_id TEXT NOT NULL,
		execute_at INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- User aliases (username/nickname history)
	CREATE TABLE IF NOT EXISTS user_aliases (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		alias TEXT NOT NULL,
		alias_type TEXT NOT NULL,
		first_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
		use_count INTEGER DEFAULT 1,
		UNIQUE(user_id, alias, alias_type)
	);

	-- User activity tracking (per guild)
	CREATE TABLE IF NOT EXISTS user_activity (
		guild_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		first_seen DATETIME,
		first_message DATETIME,
		last_seen DATETIME,
		message_count INTEGER DEFAULT 0,
		PRIMARY KEY (guild_id, user_id)
	);

	-- User timezone settings
	CREATE TABLE IF NOT EXISTS user_timezones (
		user_id TEXT PRIMARY KEY,
		timezone TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Music: Guild music settings
	CREATE TABLE IF NOT EXISTS music_settings (
		guild_id TEXT PRIMARY KEY,
		dj_role_id TEXT,
		mod_role_id TEXT,
		volume INTEGER DEFAULT 50,
		music_folder TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Music: Queue
	CREATE TABLE IF NOT EXISTS music_queue (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guild_id TEXT NOT NULL,
		channel_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		title TEXT NOT NULL,
		url TEXT NOT NULL,
		duration INTEGER DEFAULT 0,
		thumbnail TEXT,
		is_local INTEGER DEFAULT 0,
		position INTEGER NOT NULL,
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Music: Playback history
	CREATE TABLE IF NOT EXISTS music_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guild_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		title TEXT NOT NULL,
		url TEXT NOT NULL,
		played_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Disabled commands/categories per guild
	CREATE TABLE IF NOT EXISTS guild_disabled_commands (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guild_id TEXT NOT NULL,
		command_name TEXT,
		category TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(guild_id, command_name),
		UNIQUE(guild_id, category)
	);

	CREATE INDEX IF NOT EXISTS idx_user_xp_guild ON user_xp(guild_id);
	CREATE INDEX IF NOT EXISTS idx_member_joins_guild ON member_joins(guild_id, joined_at);
	CREATE INDEX IF NOT EXISTS idx_scheduled_events_time ON scheduled_events(execute_at);
	CREATE INDEX IF NOT EXISTS idx_regex_filters_guild ON regex_filters(guild_id);
	CREATE INDEX IF NOT EXISTS idx_level_ranks_guild ON level_ranks(guild_id);
	CREATE INDEX IF NOT EXISTS idx_mod_actions_guild ON mod_actions(guild_id);
	CREATE INDEX IF NOT EXISTS idx_mod_actions_moderator ON mod_actions(guild_id, moderator_id);
	CREATE INDEX IF NOT EXISTS idx_mod_actions_target ON mod_actions(guild_id, target_id);
	CREATE INDEX IF NOT EXISTS idx_user_aliases_user ON user_aliases(user_id);
	CREATE INDEX IF NOT EXISTS idx_user_activity_guild ON user_activity(guild_id);
	CREATE INDEX IF NOT EXISTS idx_music_queue_guild ON music_queue(guild_id, position);
	CREATE INDEX IF NOT EXISTS idx_music_history_guild ON music_history(guild_id);
	CREATE INDEX IF NOT EXISTS idx_disabled_commands_guild ON guild_disabled_commands(guild_id);

	-- Encryption metadata (tracks if data has been migrated to encrypted)
	CREATE TABLE IF NOT EXISTS encryption_metadata (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := d.Exec(schema)
	if err != nil {
		return err
	}

	// Run migrations for new columns
	migrations := []string{
		`ALTER TABLE guild_settings ADD COLUMN join_dm_title TEXT`,
		`ALTER TABLE guild_settings ADD COLUMN join_dm_message TEXT`,
	}

	for _, migration := range migrations {
		d.Exec(migration) // Ignore errors - column may already exist
	}

	return nil
}

// IsDataMigrated checks if data has been migrated to encrypted format
func (d *DB) IsDataMigrated() bool {
	var value string
	err := d.QueryRow(`SELECT value FROM encryption_metadata WHERE key = 'encrypted'`).Scan(&value)
	return err == nil && value == "true"
}

// SetDataMigrated marks that data has been migrated to encrypted format
func (d *DB) SetDataMigrated(migrated bool) error {
	val := "false"
	if migrated {
		val = "true"
	}
	_, err := d.Exec(`INSERT OR REPLACE INTO encryption_metadata (key, value, updated_at) VALUES ('encrypted', ?, CURRENT_TIMESTAMP)`, val)
	return err
}

// MigrateToEncrypted encrypts all sensitive fields in the database.
// This is a one-time migration when encryption is first enabled.
// It's safe to call multiple times - already encrypted data is skipped.
func (d *DB) MigrateToEncrypted() error {
	if !d.IsEncryptionEnabled() {
		return fmt.Errorf("encryption is not enabled")
	}

	if d.IsDataMigrated() {
		return nil // Already migrated
	}

	fmt.Println("[Database] Starting encryption migration...")

	// Migrate guild_settings (welcome_message, join_dm_title, join_dm_message)
	if err := d.migrateEncryptGuildSettings(); err != nil {
		return fmt.Errorf("failed to migrate guild_settings: %w", err)
	}

	// Migrate warnings (reason)
	if err := d.migrateEncryptWarnings(); err != nil {
		return fmt.Errorf("failed to migrate warnings: %w", err)
	}

	// Migrate deleted_messages (content)
	if err := d.migrateEncryptDeletedMessages(); err != nil {
		return fmt.Errorf("failed to migrate deleted_messages: %w", err)
	}

	// Migrate user_notes (note)
	if err := d.migrateEncryptUserNotes(); err != nil {
		return fmt.Errorf("failed to migrate user_notes: %w", err)
	}

	// Migrate scheduled_messages (message)
	if err := d.migrateEncryptScheduledMessages(); err != nil {
		return fmt.Errorf("failed to migrate scheduled_messages: %w", err)
	}

	// Migrate afk_status (message)
	if err := d.migrateEncryptAFKStatus(); err != nil {
		return fmt.Errorf("failed to migrate afk_status: %w", err)
	}

	// Migrate reminders (message)
	if err := d.migrateEncryptReminders(); err != nil {
		return fmt.Errorf("failed to migrate reminders: %w", err)
	}

	// Migrate tags (content)
	if err := d.migrateEncryptTags(); err != nil {
		return fmt.Errorf("failed to migrate tags: %w", err)
	}

	// Migrate custom_commands (response)
	if err := d.migrateEncryptCustomCommands(); err != nil {
		return fmt.Errorf("failed to migrate custom_commands: %w", err)
	}

	// Migrate bot_bans (reason)
	if err := d.migrateEncryptBotBans(); err != nil {
		return fmt.Errorf("failed to migrate bot_bans: %w", err)
	}

	// Migrate mod_actions (reason)
	if err := d.migrateEncryptModActions(); err != nil {
		return fmt.Errorf("failed to migrate mod_actions: %w", err)
	}

	// Migrate mention_responses (trigger, response, image_url)
	if err := d.migrateEncryptMentionResponses(); err != nil {
		return fmt.Errorf("failed to migrate mention_responses: %w", err)
	}

	// Migrate regex_filters (reason)
	if err := d.migrateEncryptRegexFilters(); err != nil {
		return fmt.Errorf("failed to migrate regex_filters: %w", err)
	}

	// Mark as migrated
	if err := d.SetDataMigrated(true); err != nil {
		return fmt.Errorf("failed to mark migration complete: %w", err)
	}

	fmt.Println("[Database] Encryption migration complete!")
	return nil
}

func (d *DB) migrateEncryptGuildSettings() error {
	rows, err := d.Query(`SELECT guild_id, welcome_message, join_dm_title, join_dm_message FROM guild_settings`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var guildID string
		var welcomeMsg, joinTitle, joinMsg *string
		if err := rows.Scan(&guildID, &welcomeMsg, &joinTitle, &joinMsg); err != nil {
			return err
		}

		// Only encrypt if not already encrypted
		needsUpdate := false
		if welcomeMsg != nil && *welcomeMsg != "" && !d.IsDataEncrypted(*welcomeMsg) {
			*welcomeMsg = d.Encrypt(*welcomeMsg)
			needsUpdate = true
		}
		if joinTitle != nil && *joinTitle != "" && !d.IsDataEncrypted(*joinTitle) {
			*joinTitle = d.Encrypt(*joinTitle)
			needsUpdate = true
		}
		if joinMsg != nil && *joinMsg != "" && !d.IsDataEncrypted(*joinMsg) {
			*joinMsg = d.Encrypt(*joinMsg)
			needsUpdate = true
		}

		if needsUpdate {
			_, err = d.Exec(`UPDATE guild_settings SET welcome_message = ?, join_dm_title = ?, join_dm_message = ? WHERE guild_id = ?`,
				welcomeMsg, joinTitle, joinMsg, guildID)
			if err != nil {
				return err
			}
		}
	}
	return rows.Err()
}

func (d *DB) migrateEncryptWarnings() error {
	rows, err := d.Query(`SELECT id, reason FROM warnings WHERE reason IS NOT NULL AND reason != ''`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var reason string
		if err := rows.Scan(&id, &reason); err != nil {
			return err
		}
		if !d.IsDataEncrypted(reason) {
			_, err = d.Exec(`UPDATE warnings SET reason = ? WHERE id = ?`, d.Encrypt(reason), id)
			if err != nil {
				return err
			}
		}
	}
	return rows.Err()
}

func (d *DB) migrateEncryptDeletedMessages() error {
	rows, err := d.Query(`SELECT id, content FROM deleted_messages WHERE content != ''`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var content string
		if err := rows.Scan(&id, &content); err != nil {
			return err
		}
		if !d.IsDataEncrypted(content) {
			_, err = d.Exec(`UPDATE deleted_messages SET content = ? WHERE id = ?`, d.Encrypt(content), id)
			if err != nil {
				return err
			}
		}
	}
	return rows.Err()
}

func (d *DB) migrateEncryptUserNotes() error {
	rows, err := d.Query(`SELECT id, note FROM user_notes WHERE note != ''`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var note string
		if err := rows.Scan(&id, &note); err != nil {
			return err
		}
		if !d.IsDataEncrypted(note) {
			_, err = d.Exec(`UPDATE user_notes SET note = ? WHERE id = ?`, d.Encrypt(note), id)
			if err != nil {
				return err
			}
		}
	}
	return rows.Err()
}

func (d *DB) migrateEncryptScheduledMessages() error {
	rows, err := d.Query(`SELECT id, message FROM scheduled_messages WHERE message != ''`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var message string
		if err := rows.Scan(&id, &message); err != nil {
			return err
		}
		if !d.IsDataEncrypted(message) {
			_, err = d.Exec(`UPDATE scheduled_messages SET message = ? WHERE id = ?`, d.Encrypt(message), id)
			if err != nil {
				return err
			}
		}
	}
	return rows.Err()
}

func (d *DB) migrateEncryptAFKStatus() error {
	rows, err := d.Query(`SELECT user_id, message FROM afk_status WHERE message IS NOT NULL AND message != ''`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var userID, message string
		if err := rows.Scan(&userID, &message); err != nil {
			return err
		}
		if !d.IsDataEncrypted(message) {
			_, err = d.Exec(`UPDATE afk_status SET message = ? WHERE user_id = ?`, d.Encrypt(message), userID)
			if err != nil {
				return err
			}
		}
	}
	return rows.Err()
}

func (d *DB) migrateEncryptReminders() error {
	rows, err := d.Query(`SELECT id, message FROM reminders WHERE message != ''`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var message string
		if err := rows.Scan(&id, &message); err != nil {
			return err
		}
		if !d.IsDataEncrypted(message) {
			_, err = d.Exec(`UPDATE reminders SET message = ? WHERE id = ?`, d.Encrypt(message), id)
			if err != nil {
				return err
			}
		}
	}
	return rows.Err()
}

func (d *DB) migrateEncryptTags() error {
	rows, err := d.Query(`SELECT id, content FROM tags WHERE content != ''`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var content string
		if err := rows.Scan(&id, &content); err != nil {
			return err
		}
		if !d.IsDataEncrypted(content) {
			_, err = d.Exec(`UPDATE tags SET content = ? WHERE id = ?`, d.Encrypt(content), id)
			if err != nil {
				return err
			}
		}
	}
	return rows.Err()
}

func (d *DB) migrateEncryptCustomCommands() error {
	rows, err := d.Query(`SELECT id, response FROM custom_commands WHERE response != ''`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var response string
		if err := rows.Scan(&id, &response); err != nil {
			return err
		}
		if !d.IsDataEncrypted(response) {
			_, err = d.Exec(`UPDATE custom_commands SET response = ? WHERE id = ?`, d.Encrypt(response), id)
			if err != nil {
				return err
			}
		}
	}
	return rows.Err()
}

func (d *DB) migrateEncryptBotBans() error {
	rows, err := d.Query(`SELECT target_id, reason FROM bot_bans WHERE reason IS NOT NULL AND reason != ''`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var targetID, reason string
		if err := rows.Scan(&targetID, &reason); err != nil {
			return err
		}
		if !d.IsDataEncrypted(reason) {
			_, err = d.Exec(`UPDATE bot_bans SET reason = ? WHERE target_id = ?`, d.Encrypt(reason), targetID)
			if err != nil {
				return err
			}
		}
	}
	return rows.Err()
}

func (d *DB) migrateEncryptModActions() error {
	rows, err := d.Query(`SELECT id, reason FROM mod_actions WHERE reason IS NOT NULL AND reason != ''`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var reason string
		if err := rows.Scan(&id, &reason); err != nil {
			return err
		}
		if !d.IsDataEncrypted(reason) {
			_, err = d.Exec(`UPDATE mod_actions SET reason = ? WHERE id = ?`, d.Encrypt(reason), id)
			if err != nil {
				return err
			}
		}
	}
	return rows.Err()
}

func (d *DB) migrateEncryptMentionResponses() error {
	rows, err := d.Query(`SELECT id, trigger_text, response, image_url FROM mention_responses`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var trigger, response string
		var imageURL *string
		if err := rows.Scan(&id, &trigger, &response, &imageURL); err != nil {
			return err
		}

		needsUpdate := false
		if !d.IsDataEncrypted(trigger) {
			trigger = d.Encrypt(trigger)
			needsUpdate = true
		}
		if !d.IsDataEncrypted(response) {
			response = d.Encrypt(response)
			needsUpdate = true
		}
		if imageURL != nil && *imageURL != "" && !d.IsDataEncrypted(*imageURL) {
			*imageURL = d.Encrypt(*imageURL)
			needsUpdate = true
		}

		if needsUpdate {
			_, err = d.Exec(`UPDATE mention_responses SET trigger_text = ?, response = ?, image_url = ? WHERE id = ?`,
				trigger, response, imageURL, id)
			if err != nil {
				return err
			}
		}
	}
	return rows.Err()
}

func (d *DB) migrateEncryptRegexFilters() error {
	rows, err := d.Query(`SELECT id, reason FROM regex_filters WHERE reason IS NOT NULL AND reason != ''`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var reason string
		if err := rows.Scan(&id, &reason); err != nil {
			return err
		}
		if !d.IsDataEncrypted(reason) {
			_, err = d.Exec(`UPDATE regex_filters SET reason = ? WHERE id = ?`, d.Encrypt(reason), id)
			if err != nil {
				return err
			}
		}
	}
	return rows.Err()
}

// Guild Settings
func (d *DB) GetGuildSettings(guildID string) (*GuildSettings, error) {
	var gs GuildSettings
	err := d.QueryRow(`SELECT guild_id, prefix, mod_log_channel, welcome_channel, welcome_message, join_dm_title, join_dm_message
		FROM guild_settings WHERE guild_id = ?`, guildID).Scan(
		&gs.GuildID, &gs.Prefix, &gs.ModLogChannel, &gs.WelcomeChannel, &gs.WelcomeMessage, &gs.JoinDMTitle, &gs.JoinDMMessage)
	if err == sql.ErrNoRows {
		return &GuildSettings{GuildID: guildID, Prefix: "/"}, nil
	}
	if err == nil {
		// Decrypt sensitive fields
		gs.WelcomeMessage = d.DecryptNullable(gs.WelcomeMessage)
		gs.JoinDMTitle = d.DecryptNullable(gs.JoinDMTitle)
		gs.JoinDMMessage = d.DecryptNullable(gs.JoinDMMessage)
	}
	return &gs, err
}

func (d *DB) SetGuildSettings(gs *GuildSettings) error {
	// Encrypt sensitive fields
	welcomeMsg := d.EncryptNullable(gs.WelcomeMessage)
	joinTitle := d.EncryptNullable(gs.JoinDMTitle)
	joinMsg := d.EncryptNullable(gs.JoinDMMessage)

	_, err := d.Exec(`INSERT INTO guild_settings (guild_id, prefix, mod_log_channel, welcome_channel, welcome_message, join_dm_title, join_dm_message, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(guild_id) DO UPDATE SET
		prefix = excluded.prefix,
		mod_log_channel = excluded.mod_log_channel,
		welcome_channel = excluded.welcome_channel,
		welcome_message = excluded.welcome_message,
		join_dm_title = excluded.join_dm_title,
		join_dm_message = excluded.join_dm_message,
		updated_at = CURRENT_TIMESTAMP`,
		gs.GuildID, gs.Prefix, gs.ModLogChannel, gs.WelcomeChannel, welcomeMsg, joinTitle, joinMsg)
	return err
}

// Custom Commands
func (d *DB) GetCustomCommand(guildID, name string) (*CustomCommand, error) {
	var cc CustomCommand
	err := d.QueryRow(`SELECT id, guild_id, name, response, created_by, use_count
		FROM custom_commands WHERE guild_id = ? AND name = ?`, guildID, name).Scan(
		&cc.ID, &cc.GuildID, &cc.Name, &cc.Response, &cc.CreatedBy, &cc.UseCount)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err == nil {
		cc.Response = d.Decrypt(cc.Response)
	}
	return &cc, err
}

func (d *DB) CreateCustomCommand(guildID, name, response, createdBy string) error {
	_, err := d.Exec(`INSERT INTO custom_commands (guild_id, name, response, created_by) VALUES (?, ?, ?, ?)`,
		guildID, name, d.Encrypt(response), createdBy)
	return err
}

func (d *DB) DeleteCustomCommand(guildID, name string) error {
	_, err := d.Exec(`DELETE FROM custom_commands WHERE guild_id = ? AND name = ?`, guildID, name)
	return err
}

func (d *DB) ListCustomCommands(guildID string) ([]CustomCommand, error) {
	rows, err := d.Query(`SELECT id, guild_id, name, response, created_by, use_count
		FROM custom_commands WHERE guild_id = ? ORDER BY name`, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commands []CustomCommand
	for rows.Next() {
		var cc CustomCommand
		if err := rows.Scan(&cc.ID, &cc.GuildID, &cc.Name, &cc.Response, &cc.CreatedBy, &cc.UseCount); err != nil {
			return nil, err
		}
		cc.Response = d.Decrypt(cc.Response)
		commands = append(commands, cc)
	}
	return commands, rows.Err()
}

func (d *DB) IncrementCommandUse(guildID, name string) error {
	_, err := d.Exec(`UPDATE custom_commands SET use_count = use_count + 1 WHERE guild_id = ? AND name = ?`,
		guildID, name)
	return err
}

// Command History
func (d *DB) LogCommand(guildID, channelID, userID, command, args string) error {
	_, err := d.Exec(`INSERT INTO command_history (guild_id, channel_id, user_id, command, args) VALUES (?, ?, ?, ?, ?)`,
		guildID, channelID, userID, command, args)
	return err
}

func (d *DB) GetCommandHistory(guildID string, limit int) ([]CommandHistory, error) {
	rows, err := d.Query(`SELECT id, guild_id, channel_id, user_id, command, args, executed_at
		FROM command_history WHERE guild_id = ? ORDER BY executed_at DESC LIMIT ?`, guildID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	history := make([]CommandHistory, 0, limit)
	for rows.Next() {
		var ch CommandHistory
		if err := rows.Scan(&ch.ID, &ch.GuildID, &ch.ChannelID, &ch.UserID, &ch.Command, &ch.Args, &ch.ExecutedAt); err != nil {
			return nil, err
		}
		history = append(history, ch)
	}
	return history, rows.Err()
}

// Warnings
func (d *DB) AddWarning(guildID, userID, moderatorID, reason string) error {
	_, err := d.Exec(`INSERT INTO warnings (guild_id, user_id, moderator_id, reason) VALUES (?, ?, ?, ?)`,
		guildID, userID, moderatorID, d.Encrypt(reason))
	return err
}

func (d *DB) GetWarnings(guildID, userID string) ([]Warning, error) {
	rows, err := d.Query(`SELECT id, guild_id, user_id, moderator_id, reason, created_at
		FROM warnings WHERE guild_id = ? AND user_id = ? ORDER BY created_at DESC`, guildID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var warnings []Warning
	for rows.Next() {
		var w Warning
		if err := rows.Scan(&w.ID, &w.GuildID, &w.UserID, &w.ModeratorID, &w.Reason, &w.CreatedAt); err != nil {
			return nil, err
		}
		w.Reason = d.DecryptNullable(w.Reason)
		warnings = append(warnings, w)
	}
	return warnings, rows.Err()
}

func (d *DB) ClearWarnings(guildID, userID string) error {
	_, err := d.Exec(`DELETE FROM warnings WHERE guild_id = ? AND user_id = ?`, guildID, userID)
	return err
}

func (d *DB) DeleteWarning(id int64) error {
	_, err := d.Exec(`DELETE FROM warnings WHERE id = ?`, id)
	return err
}

// Deleted Messages (for snipe)
func (d *DB) LogDeletedMessage(guildID, channelID, userID, content string) error {
	_, err := d.Exec(`INSERT INTO deleted_messages (guild_id, channel_id, user_id, content) VALUES (?, ?, ?, ?)`,
		guildID, channelID, userID, d.Encrypt(content))
	return err
}

func (d *DB) GetDeletedMessages(channelID string, limit int) ([]DeletedMessage, error) {
	rows, err := d.Query(`SELECT id, guild_id, channel_id, user_id, content, deleted_at
		FROM deleted_messages WHERE channel_id = ? ORDER BY deleted_at DESC LIMIT ?`, channelID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]DeletedMessage, 0, limit)
	for rows.Next() {
		var dm DeletedMessage
		if err := rows.Scan(&dm.ID, &dm.GuildID, &dm.ChannelID, &dm.UserID, &dm.Content, &dm.DeletedAt); err != nil {
			return nil, err
		}
		dm.Content = d.Decrypt(dm.Content)
		messages = append(messages, dm)
	}
	return messages, rows.Err()
}

func (d *DB) CleanOldDeletedMessages(olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	_, err := d.Exec(`DELETE FROM deleted_messages WHERE deleted_at < ?`, cutoff)
	return err
}

// Scheduled Messages
func (d *DB) ScheduleMessage(guildID, channelID, userID, message string, scheduledFor time.Time) error {
	_, err := d.Exec(`INSERT INTO scheduled_messages (guild_id, channel_id, user_id, message, scheduled_for) VALUES (?, ?, ?, ?, ?)`,
		guildID, channelID, userID, d.Encrypt(message), scheduledFor)
	return err
}

func (d *DB) GetPendingScheduledMessages() ([]ScheduledMessage, error) {
	rows, err := d.Query(`SELECT id, guild_id, channel_id, user_id, message, scheduled_for
		FROM scheduled_messages WHERE executed = 0 AND scheduled_for <= ? ORDER BY scheduled_for`, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []ScheduledMessage
	for rows.Next() {
		var sm ScheduledMessage
		if err := rows.Scan(&sm.ID, &sm.GuildID, &sm.ChannelID, &sm.UserID, &sm.Message, &sm.ScheduledFor); err != nil {
			return nil, err
		}
		sm.Message = d.Decrypt(sm.Message)
		messages = append(messages, sm)
	}
	return messages, rows.Err()
}

func (d *DB) MarkScheduledMessageExecuted(id int64) error {
	_, err := d.Exec(`UPDATE scheduled_messages SET executed = 1 WHERE id = ?`, id)
	return err
}

// AFK Status
func (d *DB) SetAFK(userID, message string) error {
	_, err := d.Exec(`INSERT INTO afk_status (user_id, message) VALUES (?, ?)
		ON CONFLICT(user_id) DO UPDATE SET message = excluded.message, set_at = CURRENT_TIMESTAMP`,
		userID, d.Encrypt(message))
	return err
}

func (d *DB) GetAFK(userID string) (*AFKStatus, error) {
	var afk AFKStatus
	err := d.QueryRow(`SELECT user_id, message, set_at FROM afk_status WHERE user_id = ?`, userID).Scan(
		&afk.UserID, &afk.Message, &afk.SetAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err == nil {
		afk.Message = d.DecryptNullable(afk.Message)
	}
	return &afk, err
}

func (d *DB) RemoveAFK(userID string) error {
	_, err := d.Exec(`DELETE FROM afk_status WHERE user_id = ?`, userID)
	return err
}

// Reminders
func (d *DB) AddReminder(userID, channelID, message string, remindAt time.Time) error {
	_, err := d.Exec(`INSERT INTO reminders (user_id, channel_id, message, remind_at) VALUES (?, ?, ?, ?)`,
		userID, channelID, d.Encrypt(message), remindAt)
	return err
}

func (d *DB) GetPendingReminders() ([]Reminder, error) {
	rows, err := d.Query(`SELECT id, user_id, channel_id, message, remind_at
		FROM reminders WHERE completed = 0 AND remind_at <= ? ORDER BY remind_at`, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reminders []Reminder
	for rows.Next() {
		var r Reminder
		if err := rows.Scan(&r.ID, &r.UserID, &r.ChannelID, &r.Message, &r.RemindAt); err != nil {
			return nil, err
		}
		r.Message = d.Decrypt(r.Message)
		reminders = append(reminders, r)
	}
	return reminders, rows.Err()
}

func (d *DB) MarkReminderCompleted(id int64) error {
	_, err := d.Exec(`UPDATE reminders SET completed = 1 WHERE id = ?`, id)
	return err
}

// Tags
func (d *DB) GetTag(guildID, name string) (*Tag, error) {
	var t Tag
	err := d.QueryRow(`SELECT id, guild_id, name, content, created_by, use_count
		FROM tags WHERE guild_id = ? AND name = ?`, guildID, name).Scan(
		&t.ID, &t.GuildID, &t.Name, &t.Content, &t.CreatedBy, &t.UseCount)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err == nil {
		t.Content = d.Decrypt(t.Content)
	}
	return &t, err
}

func (d *DB) CreateTag(guildID, name, content, createdBy string) error {
	_, err := d.Exec(`INSERT INTO tags (guild_id, name, content, created_by) VALUES (?, ?, ?, ?)`,
		guildID, name, d.Encrypt(content), createdBy)
	return err
}

func (d *DB) DeleteTag(guildID, name string) error {
	_, err := d.Exec(`DELETE FROM tags WHERE guild_id = ? AND name = ?`, guildID, name)
	return err
}

func (d *DB) ListTags(guildID string) ([]Tag, error) {
	rows, err := d.Query(`SELECT id, guild_id, name, content, created_by, use_count
		FROM tags WHERE guild_id = ? ORDER BY name`, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.GuildID, &t.Name, &t.Content, &t.CreatedBy, &t.UseCount); err != nil {
			return nil, err
		}
		t.Content = d.Decrypt(t.Content)
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (d *DB) IncrementTagUse(guildID, name string) error {
	_, err := d.Exec(`UPDATE tags SET use_count = use_count + 1 WHERE guild_id = ? AND name = ?`,
		guildID, name)
	return err
}

// Keyword Notifications
func (d *DB) AddKeywordNotification(userID, guildID, keyword string) error {
	_, err := d.Exec(`INSERT INTO keyword_notifications (user_id, guild_id, keyword) VALUES (?, ?, ?)`,
		userID, guildID, keyword)
	return err
}

func (d *DB) RemoveKeywordNotification(userID, keyword string) error {
	_, err := d.Exec(`DELETE FROM keyword_notifications WHERE user_id = ? AND keyword = ?`, userID, keyword)
	return err
}

func (d *DB) GetKeywordNotifications(userID string) ([]KeywordNotification, error) {
	rows, err := d.Query(`SELECT id, user_id, guild_id, keyword FROM keyword_notifications WHERE user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []KeywordNotification
	for rows.Next() {
		var kn KeywordNotification
		if err := rows.Scan(&kn.ID, &kn.UserID, &kn.GuildID, &kn.Keyword); err != nil {
			return nil, err
		}
		notifications = append(notifications, kn)
	}
	return notifications, rows.Err()
}

func (d *DB) GetAllKeywordNotifications() ([]KeywordNotification, error) {
	rows, err := d.Query(`SELECT id, user_id, guild_id, keyword FROM keyword_notifications`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []KeywordNotification
	for rows.Next() {
		var kn KeywordNotification
		if err := rows.Scan(&kn.ID, &kn.UserID, &kn.GuildID, &kn.Keyword); err != nil {
			return nil, err
		}
		notifications = append(notifications, kn)
	}
	return notifications, rows.Err()
}

// ============ XP/Leveling System ============

func (d *DB) GetUserXP(guildID, userID string) (*UserXP, error) {
	var ux UserXP
	err := d.QueryRow(`SELECT guild_id, user_id, xp, level, updated_at FROM user_xp WHERE guild_id = ? AND user_id = ?`,
		guildID, userID).Scan(&ux.GuildID, &ux.UserID, &ux.XP, &ux.Level, &ux.UpdatedAt)
	if err == sql.ErrNoRows {
		return &UserXP{GuildID: guildID, UserID: userID, XP: 0, Level: 0}, nil
	}
	return &ux, err
}

func (d *DB) SetUserXP(guildID, userID string, xp int64, level int) error {
	_, err := d.Exec(`INSERT INTO user_xp (guild_id, user_id, xp, level, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(guild_id, user_id) DO UPDATE SET
		xp = excluded.xp, level = excluded.level, updated_at = CURRENT_TIMESTAMP`,
		guildID, userID, xp, level)
	return err
}

func (d *DB) AddUserXP(guildID, userID string, amount int64) (*UserXP, error) {
	ux, err := d.GetUserXP(guildID, userID)
	if err != nil {
		return nil, err
	}
	ux.XP += amount
	// Calculate new level
	ux.Level = CalculateLevel(ux.XP)
	err = d.SetUserXP(guildID, userID, ux.XP, ux.Level)
	return ux, err
}

func (d *DB) GetGuildLeaderboard(guildID string, limit int) ([]UserXP, error) {
	rows, err := d.Query(`SELECT guild_id, user_id, xp, level, updated_at FROM user_xp
		WHERE guild_id = ? ORDER BY xp DESC LIMIT ?`, guildID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	leaderboard := make([]UserXP, 0, limit)
	for rows.Next() {
		var ux UserXP
		if err := rows.Scan(&ux.GuildID, &ux.UserID, &ux.XP, &ux.Level, &ux.UpdatedAt); err != nil {
			return nil, err
		}
		leaderboard = append(leaderboard, ux)
	}
	return leaderboard, rows.Err()
}

func (d *DB) GetUserRank(guildID, userID string) (int, error) {
	var rank int
	err := d.QueryRow(`SELECT COUNT(*) + 1 FROM user_xp WHERE guild_id = ? AND xp > (
		SELECT COALESCE(xp, 0) FROM user_xp WHERE guild_id = ? AND user_id = ?
	)`, guildID, guildID, userID).Scan(&rank)
	return rank, err
}

// CalculateLevel calculates level from XP using formula: level = floor((sqrt(1 + 8*xp/50) - 1) / 2)
func CalculateLevel(xp int64) int {
	if xp <= 0 {
		return 0
	}
	// level = floor((sqrt(1 + 8*xp/50) - 1) / 2)
	import_val := float64(1 + 8*xp/50)
	level := int((sqrt(import_val) - 1) / 2)
	if level < 0 {
		return 0
	}
	return level
}

// XPForLevel returns XP needed for a specific level
func XPForLevel(level int) int64 {
	// XP = 5*level^2 + 50*level + 100
	return int64(5*level*level + 50*level + 100)
}

// ============ Regex Filters ============

func (d *DB) AddRegexFilter(guildID, pattern, action, reason, createdBy string) error {
	_, err := d.Exec(`INSERT INTO regex_filters (guild_id, pattern, action, reason, created_by) VALUES (?, ?, ?, ?, ?)`,
		guildID, pattern, action, d.Encrypt(reason), createdBy)
	return err
}

func (d *DB) RemoveRegexFilter(guildID string, id int64) error {
	_, err := d.Exec(`DELETE FROM regex_filters WHERE guild_id = ? AND id = ?`, guildID, id)
	return err
}

func (d *DB) GetRegexFilters(guildID string) ([]RegexFilter, error) {
	rows, err := d.Query(`SELECT id, guild_id, pattern, action, reason, created_by, created_at
		FROM regex_filters WHERE guild_id = ? ORDER BY id`, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var filters []RegexFilter
	for rows.Next() {
		var f RegexFilter
		if err := rows.Scan(&f.ID, &f.GuildID, &f.Pattern, &f.Action, &f.Reason, &f.CreatedBy, &f.CreatedAt); err != nil {
			return nil, err
		}
		f.Reason = d.Decrypt(f.Reason)
		filters = append(filters, f)
	}
	return filters, rows.Err()
}

// ============ Auto-Clean Channels ============

func (d *DB) AddAutoCleanChannel(guildID, channelID, createdBy string, intervalHours, warningMinutes int) error {
	nextRun := time.Now().Add(time.Duration(intervalHours) * time.Hour)
	_, err := d.Exec(`INSERT INTO autoclean_channels (guild_id, channel_id, interval_hours, warning_minutes, next_run, created_by)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(guild_id, channel_id) DO UPDATE SET
		interval_hours = excluded.interval_hours, warning_minutes = excluded.warning_minutes, next_run = excluded.next_run`,
		guildID, channelID, intervalHours, warningMinutes, nextRun, createdBy)
	return err
}

func (d *DB) RemoveAutoCleanChannel(guildID, channelID string) error {
	_, err := d.Exec(`DELETE FROM autoclean_channels WHERE guild_id = ? AND channel_id = ?`, guildID, channelID)
	return err
}

func (d *DB) GetAutoCleanChannels(guildID string) ([]AutoCleanChannel, error) {
	rows, err := d.Query(`SELECT id, guild_id, channel_id, interval_hours, warning_minutes, next_run, clean_message, clean_image, created_by, created_at
		FROM autoclean_channels WHERE guild_id = ? ORDER BY id`, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []AutoCleanChannel
	for rows.Next() {
		var c AutoCleanChannel
		if err := rows.Scan(&c.ID, &c.GuildID, &c.ChannelID, &c.IntervalHours, &c.WarningMinutes, &c.NextRun, &c.CleanMessage, &c.CleanImage, &c.CreatedBy, &c.CreatedAt); err != nil {
			return nil, err
		}
		channels = append(channels, c)
	}
	return channels, rows.Err()
}

func (d *DB) GetPendingAutoCleanChannels() ([]AutoCleanChannel, error) {
	rows, err := d.Query(`SELECT id, guild_id, channel_id, interval_hours, warning_minutes, next_run, clean_message, clean_image, created_by, created_at
		FROM autoclean_channels WHERE next_run <= ? ORDER BY next_run`, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []AutoCleanChannel
	for rows.Next() {
		var c AutoCleanChannel
		if err := rows.Scan(&c.ID, &c.GuildID, &c.ChannelID, &c.IntervalHours, &c.WarningMinutes, &c.NextRun, &c.CleanMessage, &c.CleanImage, &c.CreatedBy, &c.CreatedAt); err != nil {
			return nil, err
		}
		channels = append(channels, c)
	}
	return channels, rows.Err()
}

func (d *DB) UpdateAutoCleanNextRun(id int64, nextRun time.Time) error {
	_, err := d.Exec(`UPDATE autoclean_channels SET next_run = ? WHERE id = ?`, nextRun, id)
	return err
}

func (d *DB) SetAutoCleanMessage(guildID, channelID string, enabled bool) error {
	val := 0
	if enabled {
		val = 1
	}
	_, err := d.Exec(`UPDATE autoclean_channels SET clean_message = ? WHERE guild_id = ? AND channel_id = ?`,
		val, guildID, channelID)
	return err
}

func (d *DB) SetAutoCleanImage(guildID, channelID string, enabled bool) error {
	val := 0
	if enabled {
		val = 1
	}
	_, err := d.Exec(`UPDATE autoclean_channels SET clean_image = ? WHERE guild_id = ? AND channel_id = ?`,
		val, guildID, channelID)
	return err
}

// ============ Logging Configuration ============

func (d *DB) GetLoggingConfig(guildID string) (*LoggingConfig, error) {
	var lc LoggingConfig
	err := d.QueryRow(`SELECT guild_id, log_channel_id, enabled, message_delete, message_edit,
		voice_join, voice_leave, nickname_change, avatar_change, presence_change, presence_batch_mins
		FROM logging_config WHERE guild_id = ?`, guildID).Scan(
		&lc.GuildID, &lc.LogChannelID, &lc.Enabled, &lc.MessageDelete, &lc.MessageEdit,
		&lc.VoiceJoin, &lc.VoiceLeave, &lc.NicknameChange, &lc.AvatarChange, &lc.PresenceChange, &lc.PresenceBatchMins)
	if err == sql.ErrNoRows {
		return &LoggingConfig{GuildID: guildID}, nil
	}
	return &lc, err
}

func (d *DB) SetLoggingConfig(lc *LoggingConfig) error {
	_, err := d.Exec(`INSERT INTO logging_config (guild_id, log_channel_id, enabled, message_delete, message_edit,
		voice_join, voice_leave, nickname_change, avatar_change, presence_change, presence_batch_mins)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(guild_id) DO UPDATE SET
		log_channel_id = excluded.log_channel_id, enabled = excluded.enabled,
		message_delete = excluded.message_delete, message_edit = excluded.message_edit,
		voice_join = excluded.voice_join, voice_leave = excluded.voice_leave,
		nickname_change = excluded.nickname_change, avatar_change = excluded.avatar_change,
		presence_change = excluded.presence_change, presence_batch_mins = excluded.presence_batch_mins`,
		lc.GuildID, lc.LogChannelID, lc.Enabled, lc.MessageDelete, lc.MessageEdit,
		lc.VoiceJoin, lc.VoiceLeave, lc.NicknameChange, lc.AvatarChange, lc.PresenceChange, lc.PresenceBatchMins)
	return err
}

func (d *DB) SetLogChannel(guildID, channelID string) error {
	_, err := d.Exec(`INSERT INTO logging_config (guild_id, log_channel_id, enabled)
		VALUES (?, ?, 1)
		ON CONFLICT(guild_id) DO UPDATE SET log_channel_id = excluded.log_channel_id, enabled = 1`,
		guildID, channelID)
	return err
}

func (d *DB) ToggleLogging(guildID string, enabled bool) error {
	val := 0
	if enabled {
		val = 1
	}
	_, err := d.Exec(`UPDATE logging_config SET enabled = ? WHERE guild_id = ?`, val, guildID)
	return err
}

func (d *DB) AddDisabledLogChannel(guildID, channelID string) error {
	_, err := d.Exec(`INSERT OR IGNORE INTO disabled_log_channels (guild_id, channel_id) VALUES (?, ?)`,
		guildID, channelID)
	return err
}

func (d *DB) RemoveDisabledLogChannel(guildID, channelID string) error {
	_, err := d.Exec(`DELETE FROM disabled_log_channels WHERE guild_id = ? AND channel_id = ?`, guildID, channelID)
	return err
}

func (d *DB) IsLogChannelDisabled(guildID, channelID string) (bool, error) {
	var count int
	err := d.QueryRow(`SELECT COUNT(*) FROM disabled_log_channels WHERE guild_id = ? AND channel_id = ?`,
		guildID, channelID).Scan(&count)
	return count > 0, err
}

func (d *DB) GetDisabledLogChannels(guildID string) ([]string, error) {
	rows, err := d.Query(`SELECT channel_id FROM disabled_log_channels WHERE guild_id = ?`, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []string
	for rows.Next() {
		var channelID string
		if err := rows.Scan(&channelID); err != nil {
			return nil, err
		}
		channels = append(channels, channelID)
	}
	return channels, rows.Err()
}

// ============ Voice XP Configuration ============

func (d *DB) GetVoiceXPConfig(guildID string) (*VoiceXPConfig, error) {
	var vc VoiceXPConfig
	err := d.QueryRow(`SELECT guild_id, enabled, xp_rate, interval_mins, ignore_afk
		FROM voice_xp_config WHERE guild_id = ?`, guildID).Scan(
		&vc.GuildID, &vc.Enabled, &vc.XPRate, &vc.IntervalMins, &vc.IgnoreAFK)
	if err == sql.ErrNoRows {
		return &VoiceXPConfig{GuildID: guildID, Enabled: false, XPRate: 10, IntervalMins: 5, IgnoreAFK: true}, nil
	}
	return &vc, err
}

func (d *DB) SetVoiceXPConfig(vc *VoiceXPConfig) error {
	_, err := d.Exec(`INSERT INTO voice_xp_config (guild_id, enabled, xp_rate, interval_mins, ignore_afk)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(guild_id) DO UPDATE SET
		enabled = excluded.enabled, xp_rate = excluded.xp_rate,
		interval_mins = excluded.interval_mins, ignore_afk = excluded.ignore_afk`,
		vc.GuildID, vc.Enabled, vc.XPRate, vc.IntervalMins, vc.IgnoreAFK)
	return err
}

// ============ Level Ranks ============

func (d *DB) AddLevelRank(guildID, roleID string, level int) error {
	_, err := d.Exec(`INSERT INTO level_ranks (guild_id, role_id, level)
		VALUES (?, ?, ?)
		ON CONFLICT(guild_id, role_id) DO UPDATE SET level = excluded.level`,
		guildID, roleID, level)
	return err
}

func (d *DB) RemoveLevelRank(guildID, roleID string) error {
	_, err := d.Exec(`DELETE FROM level_ranks WHERE guild_id = ? AND role_id = ?`, guildID, roleID)
	return err
}

func (d *DB) GetLevelRanks(guildID string) ([]LevelRank, error) {
	rows, err := d.Query(`SELECT id, guild_id, role_id, level, created_at
		FROM level_ranks WHERE guild_id = ? ORDER BY level`, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ranks []LevelRank
	for rows.Next() {
		var r LevelRank
		if err := rows.Scan(&r.ID, &r.GuildID, &r.RoleID, &r.Level, &r.CreatedAt); err != nil {
			return nil, err
		}
		ranks = append(ranks, r)
	}
	return ranks, rows.Err()
}

func (d *DB) GetRanksForLevel(guildID string, level int) ([]LevelRank, error) {
	rows, err := d.Query(`SELECT id, guild_id, role_id, level, created_at
		FROM level_ranks WHERE guild_id = ? AND level <= ? ORDER BY level`, guildID, level)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ranks []LevelRank
	for rows.Next() {
		var r LevelRank
		if err := rows.Scan(&r.ID, &r.GuildID, &r.RoleID, &r.Level, &r.CreatedAt); err != nil {
			return nil, err
		}
		ranks = append(ranks, r)
	}
	return ranks, rows.Err()
}

// ============ DM Forwarding ============

func (d *DB) GetDMConfig(guildID string) (*DMConfig, error) {
	var dc DMConfig
	err := d.QueryRow(`SELECT guild_id, channel_id, enabled FROM dm_config WHERE guild_id = ?`, guildID).Scan(
		&dc.GuildID, &dc.ChannelID, &dc.Enabled)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &dc, err
}

func (d *DB) SetDMConfig(guildID, channelID string, enabled bool) error {
	_, err := d.Exec(`INSERT INTO dm_config (guild_id, channel_id, enabled)
		VALUES (?, ?, ?)
		ON CONFLICT(guild_id) DO UPDATE SET channel_id = excluded.channel_id, enabled = excluded.enabled`,
		guildID, channelID, enabled)
	return err
}

func (d *DB) GetAllDMConfigs() ([]DMConfig, error) {
	rows, err := d.Query(`SELECT guild_id, channel_id, enabled FROM dm_config WHERE enabled = 1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []DMConfig
	for rows.Next() {
		var dc DMConfig
		if err := rows.Scan(&dc.GuildID, &dc.ChannelID, &dc.Enabled); err != nil {
			return nil, err
		}
		configs = append(configs, dc)
	}
	return configs, rows.Err()
}

// ============ Bot Bans ============

func (d *DB) AddBotBan(targetID, banType, reason, bannedBy string) error {
	_, err := d.Exec(`INSERT INTO bot_bans (target_id, ban_type, reason, banned_by)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(target_id) DO UPDATE SET ban_type = excluded.ban_type, reason = excluded.reason`,
		targetID, banType, d.Encrypt(reason), bannedBy)
	return err
}

func (d *DB) RemoveBotBan(targetID string) error {
	_, err := d.Exec(`DELETE FROM bot_bans WHERE target_id = ?`, targetID)
	return err
}

func (d *DB) IsBotBanned(targetID string) (bool, error) {
	var count int
	err := d.QueryRow(`SELECT COUNT(*) FROM bot_bans WHERE target_id = ?`, targetID).Scan(&count)
	return count > 0, err
}

func (d *DB) GetBotBan(targetID string) (*BotBan, error) {
	var bb BotBan
	err := d.QueryRow(`SELECT id, target_id, ban_type, reason, banned_by, created_at FROM bot_bans WHERE target_id = ?`,
		targetID).Scan(&bb.ID, &bb.TargetID, &bb.BanType, &bb.Reason, &bb.BannedBy, &bb.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err == nil {
		bb.Reason = d.Decrypt(bb.Reason)
	}
	return &bb, err
}

func (d *DB) GetBotBans(banType string) ([]BotBan, error) {
	var rows *sql.Rows
	var err error
	if banType != "" {
		rows, err = d.Query(`SELECT id, target_id, ban_type, reason, banned_by, created_at FROM bot_bans WHERE ban_type = ? ORDER BY created_at DESC`, banType)
	} else {
		rows, err = d.Query(`SELECT id, target_id, ban_type, reason, banned_by, created_at FROM bot_bans ORDER BY created_at DESC`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bans []BotBan
	for rows.Next() {
		var bb BotBan
		if err := rows.Scan(&bb.ID, &bb.TargetID, &bb.BanType, &bb.Reason, &bb.BannedBy, &bb.CreatedAt); err != nil {
			return nil, err
		}
		bb.Reason = d.Decrypt(bb.Reason)
		bans = append(bans, bb)
	}
	return bans, rows.Err()
}

// Helper function for sqrt
func sqrt(x float64) float64 {
	if x < 0 {
		return 0
	}
	z := x / 2
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

// ============ Moderation Actions ============

func (d *DB) AddModAction(guildID, moderatorID, targetID, action string, reason *string, timestamp int64) error {
	encReason := d.EncryptNullable(reason)
	_, err := d.Exec(`INSERT INTO mod_actions (guild_id, moderator_id, target_id, action, reason, timestamp) VALUES (?, ?, ?, ?, ?, ?)`,
		guildID, moderatorID, targetID, action, encReason, timestamp)
	return err
}

func (d *DB) ModActionExists(guildID, targetID, action string, timestamp int64) (bool, error) {
	var count int
	err := d.QueryRow(`SELECT COUNT(*) FROM mod_actions WHERE guild_id = ? AND target_id = ? AND action = ? AND timestamp = ?`,
		guildID, targetID, action, timestamp).Scan(&count)
	return count > 0, err
}

func (d *DB) GetModActionsCount(guildID string) (int, error) {
	var count int
	err := d.QueryRow(`SELECT COUNT(*) FROM mod_actions WHERE guild_id = ?`, guildID).Scan(&count)
	return count, err
}

func (d *DB) GetModStats(guildID string) (*ModStats, error) {
	stats := &ModStats{
		ActionCounts: make(map[string]int),
		TopMods:      []ModeratorCount{},
	}

	// Get total count
	d.QueryRow(`SELECT COUNT(*) FROM mod_actions WHERE guild_id = ?`, guildID).Scan(&stats.TotalActions)

	// Get action counts
	rows, err := d.Query(`SELECT action, COUNT(*) as count FROM mod_actions WHERE guild_id = ? GROUP BY action`, guildID)
	if err != nil {
		return stats, err
	}
	defer rows.Close()

	for rows.Next() {
		var action string
		var count int
		if err := rows.Scan(&action, &count); err == nil {
			stats.ActionCounts[action] = count
		}
	}

	// Get top moderators
	rows, err = d.Query(`SELECT moderator_id, COUNT(*) as count FROM mod_actions WHERE guild_id = ? GROUP BY moderator_id ORDER BY count DESC LIMIT 10`, guildID)
	if err != nil {
		return stats, err
	}
	defer rows.Close()

	modMap := make(map[string]*ModeratorCount)
	for rows.Next() {
		var modID string
		var count int
		if err := rows.Scan(&modID, &count); err == nil {
			modMap[modID] = &ModeratorCount{
				ModeratorID: modID,
				Count:       count,
				Actions:     make(map[string]int),
			}
		}
	}

	// Get per-moderator action breakdown
	for modID := range modMap {
		rows, err := d.Query(`SELECT action, COUNT(*) as count FROM mod_actions WHERE guild_id = ? AND moderator_id = ? GROUP BY action`, guildID, modID)
		if err != nil {
			continue
		}
		for rows.Next() {
			var action string
			var count int
			if err := rows.Scan(&action, &count); err == nil {
				modMap[modID].Actions[action] = count
			}
		}
		rows.Close()
		stats.TopMods = append(stats.TopMods, *modMap[modID])
	}

	return stats, nil
}

func (d *DB) GetModActionsForTarget(guildID, targetID string) ([]ModAction, error) {
	rows, err := d.Query(`SELECT id, guild_id, moderator_id, target_id, action, reason, timestamp, created_at
		FROM mod_actions WHERE guild_id = ? AND target_id = ? ORDER BY timestamp DESC`, guildID, targetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []ModAction
	for rows.Next() {
		var ma ModAction
		if err := rows.Scan(&ma.ID, &ma.GuildID, &ma.ModeratorID, &ma.TargetID, &ma.Action, &ma.Reason, &ma.Timestamp, &ma.CreatedAt); err != nil {
			return nil, err
		}
		ma.Reason = d.DecryptNullable(ma.Reason)
		actions = append(actions, ma)
	}
	return actions, rows.Err()
}

// ============ Mention Responses ============

func (d *DB) AddMentionResponse(guildID, trigger, response string, imageURL *string, createdBy string) error {
	_, err := d.Exec(`INSERT INTO mention_responses (guild_id, trigger_text, response, image_url, created_by)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(guild_id, trigger_text) DO UPDATE SET response = excluded.response, image_url = excluded.image_url`,
		guildID, d.Encrypt(trigger), d.Encrypt(response), d.EncryptNullable(imageURL), createdBy)
	return err
}

func (d *DB) RemoveMentionResponse(guildID, trigger string) error {
	// Need to encrypt trigger for lookup since it's stored encrypted
	_, err := d.Exec(`DELETE FROM mention_responses WHERE guild_id = ? AND trigger_text = ?`, guildID, d.Encrypt(trigger))
	return err
}

func (d *DB) GetMentionResponses(guildID string) ([]MentionResponse, error) {
	rows, err := d.Query(`SELECT id, guild_id, trigger_text, response, image_url, created_by, created_at
		FROM mention_responses WHERE guild_id = ? ORDER BY trigger_text`, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var responses []MentionResponse
	for rows.Next() {
		var mr MentionResponse
		if err := rows.Scan(&mr.ID, &mr.GuildID, &mr.TriggerText, &mr.Response, &mr.ImageURL, &mr.CreatedBy, &mr.CreatedAt); err != nil {
			return nil, err
		}
		mr.TriggerText = d.Decrypt(mr.TriggerText)
		mr.Response = d.Decrypt(mr.Response)
		mr.ImageURL = d.DecryptNullable(mr.ImageURL)
		responses = append(responses, mr)
	}
	return responses, rows.Err()
}

func (d *DB) GetMentionResponse(guildID, trigger string) (*MentionResponse, error) {
	var mr MentionResponse
	// Need to encrypt trigger for lookup since it's stored encrypted
	err := d.QueryRow(`SELECT id, guild_id, trigger_text, response, image_url, created_by, created_at
		FROM mention_responses WHERE guild_id = ? AND trigger_text = ?`, guildID, d.Encrypt(trigger)).Scan(
		&mr.ID, &mr.GuildID, &mr.TriggerText, &mr.Response, &mr.ImageURL, &mr.CreatedBy, &mr.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err == nil {
		mr.TriggerText = d.Decrypt(mr.TriggerText)
		mr.Response = d.Decrypt(mr.Response)
		mr.ImageURL = d.DecryptNullable(mr.ImageURL)
	}
	return &mr, err
}

// ============ Spam Filter ============

func (d *DB) GetSpamFilterConfig(guildID string) (*SpamFilterConfig, error) {
	var sf SpamFilterConfig
	err := d.QueryRow(`SELECT guild_id, enabled, max_mentions, max_links, max_emojis, action
		FROM spam_filter_config WHERE guild_id = ?`, guildID).Scan(
		&sf.GuildID, &sf.Enabled, &sf.MaxMentions, &sf.MaxLinks, &sf.MaxEmojis, &sf.Action)
	if err == sql.ErrNoRows {
		return &SpamFilterConfig{GuildID: guildID, Enabled: false, MaxMentions: 5, MaxLinks: 3, MaxEmojis: 10, Action: "delete"}, nil
	}
	return &sf, err
}

func (d *DB) SetSpamFilterConfig(sf *SpamFilterConfig) error {
	_, err := d.Exec(`INSERT INTO spam_filter_config (guild_id, enabled, max_mentions, max_links, max_emojis, action)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(guild_id) DO UPDATE SET
		enabled = excluded.enabled, max_mentions = excluded.max_mentions,
		max_links = excluded.max_links, max_emojis = excluded.max_emojis, action = excluded.action`,
		sf.GuildID, sf.Enabled, sf.MaxMentions, sf.MaxLinks, sf.MaxEmojis, sf.Action)
	return err
}

// ============ Ticket System ============

func (d *DB) GetTicketConfig(guildID string) (*TicketConfig, error) {
	var tc TicketConfig
	err := d.QueryRow(`SELECT guild_id, channel_id, enabled FROM ticket_config WHERE guild_id = ?`, guildID).Scan(
		&tc.GuildID, &tc.ChannelID, &tc.Enabled)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &tc, err
}

func (d *DB) SetTicketConfig(guildID, channelID string, enabled bool) error {
	_, err := d.Exec(`INSERT INTO ticket_config (guild_id, channel_id, enabled)
		VALUES (?, ?, ?)
		ON CONFLICT(guild_id) DO UPDATE SET
		channel_id = excluded.channel_id, enabled = excluded.enabled`,
		guildID, channelID, enabled)
	return err
}

func (d *DB) DeleteTicketConfig(guildID string) error {
	_, err := d.Exec(`DELETE FROM ticket_config WHERE guild_id = ?`, guildID)
	return err
}

// ============ Anti-Raid System ============

func (d *DB) GetAntiRaidConfig(guildID string) (*AntiRaidConfig, error) {
	var cfg AntiRaidConfig
	var silentRole, alertRole, logChannel sql.NullString
	err := d.QueryRow(`SELECT guild_id, enabled, raid_time, raid_size, auto_silence,
		lockdown_duration, silent_role_id, alert_role_id, log_channel_id, action
		FROM antiraid_config WHERE guild_id = ?`, guildID).Scan(
		&cfg.GuildID, &cfg.Enabled, &cfg.RaidTime, &cfg.RaidSize, &cfg.AutoSilence,
		&cfg.LockdownDuration, &silentRole, &alertRole, &logChannel, &cfg.Action)
	if err == sql.ErrNoRows {
		return &AntiRaidConfig{
			GuildID:          guildID,
			Enabled:          false,
			RaidTime:         300,
			RaidSize:         5,
			AutoSilence:      0,
			LockdownDuration: 120,
			Action:           "silence",
		}, nil
	}
	if silentRole.Valid {
		cfg.SilentRoleID = silentRole.String
	}
	if alertRole.Valid {
		cfg.AlertRoleID = alertRole.String
	}
	if logChannel.Valid {
		cfg.LogChannelID = logChannel.String
	}
	return &cfg, err
}

func (d *DB) SetAntiRaidConfig(cfg *AntiRaidConfig) error {
	_, err := d.Exec(`INSERT INTO antiraid_config (guild_id, enabled, raid_time, raid_size, auto_silence,
		lockdown_duration, silent_role_id, alert_role_id, log_channel_id, action)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(guild_id) DO UPDATE SET
		enabled = excluded.enabled, raid_time = excluded.raid_time, raid_size = excluded.raid_size,
		auto_silence = excluded.auto_silence, lockdown_duration = excluded.lockdown_duration,
		silent_role_id = excluded.silent_role_id, alert_role_id = excluded.alert_role_id,
		log_channel_id = excluded.log_channel_id, action = excluded.action`,
		cfg.GuildID, cfg.Enabled, cfg.RaidTime, cfg.RaidSize, cfg.AutoSilence,
		cfg.LockdownDuration, cfg.SilentRoleID, cfg.AlertRoleID, cfg.LogChannelID, cfg.Action)
	return err
}

func (d *DB) RecordMemberJoin(guildID, userID string, joinedAt, accountCreatedAt int64) error {
	_, err := d.Exec(`INSERT INTO member_joins (guild_id, user_id, joined_at, account_created_at)
		VALUES (?, ?, ?, ?)`, guildID, userID, joinedAt, accountCreatedAt)
	return err
}

func (d *DB) CountRecentJoins(guildID string, sinceTimestamp int64) (int, error) {
	var count int
	err := d.QueryRow(`SELECT COUNT(*) FROM member_joins WHERE guild_id = ? AND joined_at >= ?`,
		guildID, sinceTimestamp).Scan(&count)
	return count, err
}

func (d *DB) GetRecentJoins(guildID string, sinceTimestamp int64) ([]MemberJoin, error) {
	rows, err := d.Query(`SELECT id, guild_id, user_id, joined_at, account_created_at
		FROM member_joins WHERE guild_id = ? AND joined_at >= ? ORDER BY joined_at DESC`,
		guildID, sinceTimestamp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var joins []MemberJoin
	for rows.Next() {
		var mj MemberJoin
		if err := rows.Scan(&mj.ID, &mj.GuildID, &mj.UserID, &mj.JoinedAt, &mj.AccountCreatedAt); err != nil {
			return nil, err
		}
		joins = append(joins, mj)
	}
	return joins, rows.Err()
}

func (d *DB) CleanOldJoins(guildID string, beforeTimestamp int64) error {
	_, err := d.Exec(`DELETE FROM member_joins WHERE guild_id = ? AND joined_at < ?`,
		guildID, beforeTimestamp)
	return err
}

// ============ Anti-Spam System ============

func (d *DB) GetAntiSpamConfig(guildID string) (*AntiSpamConfig, error) {
	var cfg AntiSpamConfig
	var silentRole sql.NullString
	err := d.QueryRow(`SELECT guild_id, enabled, base_pressure, image_pressure, link_pressure,
		ping_pressure, length_pressure, line_pressure, repeat_pressure, max_pressure,
		pressure_decay, action, silent_role_id FROM antispam_config WHERE guild_id = ?`, guildID).Scan(
		&cfg.GuildID, &cfg.Enabled, &cfg.BasePressure, &cfg.ImagePressure, &cfg.LinkPressure,
		&cfg.PingPressure, &cfg.LengthPressure, &cfg.LinePressure, &cfg.RepeatPressure,
		&cfg.MaxPressure, &cfg.PressureDecay, &cfg.Action, &silentRole)
	if err == sql.ErrNoRows {
		return &AntiSpamConfig{
			GuildID:        guildID,
			Enabled:        false,
			BasePressure:   10.0,
			ImagePressure:  8.33,
			LinkPressure:   8.33,
			PingPressure:   2.5,
			LengthPressure: 0.00625,
			LinePressure:   0.71,
			RepeatPressure: 10.0,
			MaxPressure:    60.0,
			PressureDecay:  2.5,
			Action:         "delete",
		}, nil
	}
	if silentRole.Valid {
		cfg.SilentRoleID = silentRole.String
	}
	return &cfg, err
}

func (d *DB) SetAntiSpamConfig(cfg *AntiSpamConfig) error {
	_, err := d.Exec(`INSERT INTO antispam_config (guild_id, enabled, base_pressure, image_pressure,
		link_pressure, ping_pressure, length_pressure, line_pressure, repeat_pressure,
		max_pressure, pressure_decay, action, silent_role_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(guild_id) DO UPDATE SET
		enabled = excluded.enabled, base_pressure = excluded.base_pressure,
		image_pressure = excluded.image_pressure, link_pressure = excluded.link_pressure,
		ping_pressure = excluded.ping_pressure, length_pressure = excluded.length_pressure,
		line_pressure = excluded.line_pressure, repeat_pressure = excluded.repeat_pressure,
		max_pressure = excluded.max_pressure, pressure_decay = excluded.pressure_decay,
		action = excluded.action, silent_role_id = excluded.silent_role_id`,
		cfg.GuildID, cfg.Enabled, cfg.BasePressure, cfg.ImagePressure, cfg.LinkPressure,
		cfg.PingPressure, cfg.LengthPressure, cfg.LinePressure, cfg.RepeatPressure,
		cfg.MaxPressure, cfg.PressureDecay, cfg.Action, cfg.SilentRoleID)
	return err
}

// ============ Scheduled Events ============

func (d *DB) AddScheduledEvent(guildID, eventType, targetID string, executeAt int64) error {
	_, err := d.Exec(`INSERT INTO scheduled_events (guild_id, event_type, target_id, execute_at)
		VALUES (?, ?, ?, ?)`, guildID, eventType, targetID, executeAt)
	return err
}

func (d *DB) GetDueEvents(beforeTimestamp int64) ([]ScheduledEvent, error) {
	rows, err := d.Query(`SELECT id, guild_id, event_type, target_id, execute_at
		FROM scheduled_events WHERE execute_at <= ?`, beforeTimestamp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []ScheduledEvent
	for rows.Next() {
		var ev ScheduledEvent
		if err := rows.Scan(&ev.ID, &ev.GuildID, &ev.EventType, &ev.TargetID, &ev.ExecuteAt); err != nil {
			return nil, err
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}

func (d *DB) DeleteScheduledEvent(id int64) error {
	_, err := d.Exec(`DELETE FROM scheduled_events WHERE id = ?`, id)
	return err
}

func (d *DB) DeleteScheduledEventByTarget(guildID, eventType, targetID string) error {
	_, err := d.Exec(`DELETE FROM scheduled_events WHERE guild_id = ? AND event_type = ? AND target_id = ?`,
		guildID, eventType, targetID)
	return err
}

// ============ User Aliases ============

func (d *DB) RecordAlias(userID, alias, aliasType string) error {
	_, err := d.Exec(`INSERT INTO user_aliases (user_id, alias, alias_type, last_seen, use_count)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, 1)
		ON CONFLICT(user_id, alias, alias_type) DO UPDATE SET
		last_seen = CURRENT_TIMESTAMP, use_count = use_count + 1`,
		userID, alias, aliasType)
	return err
}

func (d *DB) GetUserAliases(userID string, limit int) ([]UserAlias, error) {
	rows, err := d.Query(`SELECT id, user_id, alias, alias_type, first_seen, last_seen, use_count
		FROM user_aliases WHERE user_id = ? ORDER BY use_count DESC, last_seen DESC LIMIT ?`,
		userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	aliases := make([]UserAlias, 0, limit)
	for rows.Next() {
		var a UserAlias
		if err := rows.Scan(&a.ID, &a.UserID, &a.Alias, &a.AliasType, &a.FirstSeen, &a.LastSeen, &a.UseCount); err != nil {
			return nil, err
		}
		aliases = append(aliases, a)
	}
	return aliases, rows.Err()
}

func (d *DB) SearchUserByAlias(alias string) ([]string, error) {
	rows, err := d.Query(`SELECT DISTINCT user_id FROM user_aliases WHERE alias LIKE ? LIMIT 10`,
		"%"+alias+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userIDs := make([]string, 0, 10)
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}
	return userIDs, rows.Err()
}

// ============ User Activity ============

func (d *DB) UpdateUserActivity(guildID, userID string, isMessage bool) error {
	now := time.Now()

	if isMessage {
		_, err := d.Exec(`INSERT INTO user_activity (guild_id, user_id, first_seen, first_message, last_seen, message_count)
			VALUES (?, ?, ?, ?, ?, 1)
			ON CONFLICT(guild_id, user_id) DO UPDATE SET
			last_seen = ?,
			message_count = message_count + 1,
			first_message = COALESCE(first_message, ?)`,
			guildID, userID, now, now, now, now, now)
		return err
	}

	_, err := d.Exec(`INSERT INTO user_activity (guild_id, user_id, first_seen, last_seen, message_count)
		VALUES (?, ?, ?, ?, 0)
		ON CONFLICT(guild_id, user_id) DO UPDATE SET last_seen = ?`,
		guildID, userID, now, now, now)
	return err
}

func (d *DB) GetUserActivity(guildID, userID string) (*UserActivity, error) {
	var ua UserActivity
	err := d.QueryRow(`SELECT guild_id, user_id, first_seen, first_message, last_seen, message_count
		FROM user_activity WHERE guild_id = ? AND user_id = ?`, guildID, userID).Scan(
		&ua.GuildID, &ua.UserID, &ua.FirstSeen, &ua.FirstMessage, &ua.LastSeen, &ua.MessageCount)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &ua, err
}

func (d *DB) GetNewestMembers(guildID string, limit int) ([]UserActivity, error) {
	rows, err := d.Query(`SELECT guild_id, user_id, first_seen, first_message, last_seen, message_count
		FROM user_activity WHERE guild_id = ? ORDER BY first_seen DESC LIMIT ?`,
		guildID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	activities := make([]UserActivity, 0, limit)
	for rows.Next() {
		var ua UserActivity
		if err := rows.Scan(&ua.GuildID, &ua.UserID, &ua.FirstSeen, &ua.FirstMessage, &ua.LastSeen, &ua.MessageCount); err != nil {
			return nil, err
		}
		activities = append(activities, ua)
	}
	return activities, rows.Err()
}

// ============ User Timezones ============

func (d *DB) SetUserTimezone(userID, timezone string) error {
	_, err := d.Exec(`INSERT INTO user_timezones (user_id, timezone, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(user_id) DO UPDATE SET timezone = excluded.timezone, updated_at = CURRENT_TIMESTAMP`,
		userID, timezone)
	return err
}

func (d *DB) GetUserTimezone(userID string) (string, error) {
	var tz string
	err := d.QueryRow(`SELECT timezone FROM user_timezones WHERE user_id = ?`, userID).Scan(&tz)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return tz, err
}

func (d *DB) DeleteUserTimezone(userID string) error {
	_, err := d.Exec(`DELETE FROM user_timezones WHERE user_id = ?`, userID)
	return err
}

// ============ Music Settings ============

func (d *DB) GetMusicSettings(guildID string) (*MusicSettings, error) {
	var ms MusicSettings
	err := d.QueryRow(`SELECT guild_id, dj_role_id, mod_role_id, volume, music_folder
		FROM music_settings WHERE guild_id = ?`, guildID).Scan(
		&ms.GuildID, &ms.DJRoleID, &ms.ModRoleID, &ms.Volume, &ms.MusicFolder)
	if err == sql.ErrNoRows {
		return &MusicSettings{GuildID: guildID, Volume: 50}, nil
	}
	return &ms, err
}

func (d *DB) SetMusicSettings(ms *MusicSettings) error {
	_, err := d.Exec(`INSERT INTO music_settings (guild_id, dj_role_id, mod_role_id, volume, music_folder, updated_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(guild_id) DO UPDATE SET
		dj_role_id = excluded.dj_role_id, mod_role_id = excluded.mod_role_id,
		volume = excluded.volume, music_folder = excluded.music_folder,
		updated_at = CURRENT_TIMESTAMP`,
		ms.GuildID, ms.DJRoleID, ms.ModRoleID, ms.Volume, ms.MusicFolder)
	return err
}

func (d *DB) UpdateMusicVolume(guildID string, volume int) error {
	_, err := d.Exec(`INSERT INTO music_settings (guild_id, volume, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(guild_id) DO UPDATE SET volume = excluded.volume, updated_at = CURRENT_TIMESTAMP`,
		guildID, volume)
	return err
}

func (d *DB) UpdateMusicRoles(guildID string, djRoleID, modRoleID *string) error {
	_, err := d.Exec(`INSERT INTO music_settings (guild_id, dj_role_id, mod_role_id, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(guild_id) DO UPDATE SET
		dj_role_id = excluded.dj_role_id, mod_role_id = excluded.mod_role_id,
		updated_at = CURRENT_TIMESTAMP`,
		guildID, djRoleID, modRoleID)
	return err
}

// ============ Music Queue ============

func (d *DB) AddToMusicQueue(item *MusicQueueItem) error {
	_, err := d.Exec(`INSERT INTO music_queue (guild_id, channel_id, user_id, title, url, duration, thumbnail, is_local, position)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?,
			COALESCE((SELECT MAX(position) + 1 FROM music_queue WHERE guild_id = ?), 0))`,
		item.GuildID, item.ChannelID, item.UserID, item.Title, item.URL, item.Duration, item.Thumbnail, item.IsLocal, item.GuildID)
	return err
}

func (d *DB) GetMusicQueue(guildID string) ([]MusicQueueItem, error) {
	rows, err := d.Query(`SELECT id, guild_id, channel_id, user_id, title, url, duration, thumbnail, is_local, position, added_at
		FROM music_queue WHERE guild_id = ? ORDER BY position ASC`, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []MusicQueueItem
	for rows.Next() {
		var item MusicQueueItem
		if err := rows.Scan(&item.ID, &item.GuildID, &item.ChannelID, &item.UserID, &item.Title, &item.URL,
			&item.Duration, &item.Thumbnail, &item.IsLocal, &item.Position, &item.AddedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (d *DB) RemoveFromMusicQueue(guildID string, position int) error {
	tx, err := d.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`DELETE FROM music_queue WHERE guild_id = ? AND position = ?`, guildID, position)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`UPDATE music_queue SET position = position - 1 WHERE guild_id = ? AND position > ?`, guildID, position)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (d *DB) ClearMusicQueue(guildID string) error {
	_, err := d.Exec(`DELETE FROM music_queue WHERE guild_id = ?`, guildID)
	return err
}

func (d *DB) MoveToTopMusicQueue(guildID string, position int) error {
	tx, err := d.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Temporarily set position to -1
	_, err = tx.Exec(`UPDATE music_queue SET position = -1 WHERE guild_id = ? AND position = ?`, guildID, position)
	if err != nil {
		return err
	}

	// Increment all positions that were less than the moved item
	_, err = tx.Exec(`UPDATE music_queue SET position = position + 1 WHERE guild_id = ? AND position >= 0 AND position < ?`, guildID, position)
	if err != nil {
		return err
	}

	// Set the moved item to position 0
	_, err = tx.Exec(`UPDATE music_queue SET position = 0 WHERE guild_id = ? AND position = -1`, guildID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// ============ Music History ============

func (d *DB) AddToMusicHistory(guildID, userID, title, url string) error {
	_, err := d.Exec(`INSERT INTO music_history (guild_id, user_id, title, url) VALUES (?, ?, ?, ?)`,
		guildID, userID, title, url)
	return err
}

func (d *DB) GetMusicHistory(guildID string, limit int) ([]MusicHistory, error) {
	rows, err := d.Query(`SELECT id, guild_id, user_id, title, url, played_at
		FROM music_history WHERE guild_id = ? ORDER BY played_at DESC LIMIT ?`, guildID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]MusicHistory, 0, limit)
	for rows.Next() {
		var item MusicHistory
		if err := rows.Scan(&item.ID, &item.GuildID, &item.UserID, &item.Title, &item.URL, &item.PlayedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// ============ Disabled Commands/Categories ============

// IsCommandDisabled checks if a specific command is disabled for a guild
func (d *DB) IsCommandDisabled(guildID, commandName string) bool {
	var count int
	err := d.QueryRow(`SELECT COUNT(*) FROM guild_disabled_commands WHERE guild_id = ? AND command_name = ?`,
		guildID, commandName).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// IsCategoryDisabled checks if a category is disabled for a guild
func (d *DB) IsCategoryDisabled(guildID, category string) bool {
	var count int
	err := d.QueryRow(`SELECT COUNT(*) FROM guild_disabled_commands WHERE guild_id = ? AND category = ?`,
		guildID, category).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// DisableCommand disables a specific command for a guild
func (d *DB) DisableCommand(guildID, commandName string) error {
	_, err := d.Exec(`INSERT OR IGNORE INTO guild_disabled_commands (guild_id, command_name) VALUES (?, ?)`,
		guildID, commandName)
	return err
}

// EnableCommand enables a previously disabled command for a guild
func (d *DB) EnableCommand(guildID, commandName string) error {
	_, err := d.Exec(`DELETE FROM guild_disabled_commands WHERE guild_id = ? AND command_name = ?`,
		guildID, commandName)
	return err
}

// DisableCategory disables an entire category for a guild
func (d *DB) DisableCategory(guildID, category string) error {
	_, err := d.Exec(`INSERT OR IGNORE INTO guild_disabled_commands (guild_id, category) VALUES (?, ?)`,
		guildID, category)
	return err
}

// EnableCategory enables a previously disabled category for a guild
func (d *DB) EnableCategory(guildID, category string) error {
	_, err := d.Exec(`DELETE FROM guild_disabled_commands WHERE guild_id = ? AND category = ?`,
		guildID, category)
	return err
}

// GetDisabledCommands returns all disabled command names for a guild
func (d *DB) GetDisabledCommands(guildID string) ([]string, error) {
	rows, err := d.Query(`SELECT command_name FROM guild_disabled_commands WHERE guild_id = ? AND command_name IS NOT NULL`,
		guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commands []string
	for rows.Next() {
		var cmd string
		if err := rows.Scan(&cmd); err != nil {
			return nil, err
		}
		commands = append(commands, cmd)
	}
	return commands, rows.Err()
}

// GetDisabledCategories returns all disabled category names for a guild
func (d *DB) GetDisabledCategories(guildID string) ([]string, error) {
	rows, err := d.Query(`SELECT category FROM guild_disabled_commands WHERE guild_id = ? AND category IS NOT NULL`,
		guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var cat string
		if err := rows.Scan(&cat); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, rows.Err()
}

// SetDisabledCommands replaces all disabled commands for a guild
func (d *DB) SetDisabledCommands(guildID string, commands []string) error {
	tx, err := d.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete all existing disabled commands (not categories)
	_, err = tx.Exec(`DELETE FROM guild_disabled_commands WHERE guild_id = ? AND command_name IS NOT NULL`, guildID)
	if err != nil {
		return err
	}

	// Insert new disabled commands
	for _, cmd := range commands {
		_, err = tx.Exec(`INSERT INTO guild_disabled_commands (guild_id, command_name) VALUES (?, ?)`, guildID, cmd)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// SetDisabledCategories replaces all disabled categories for a guild
func (d *DB) SetDisabledCategories(guildID string, categories []string) error {
	tx, err := d.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete all existing disabled categories (not commands)
	_, err = tx.Exec(`DELETE FROM guild_disabled_commands WHERE guild_id = ? AND category IS NOT NULL`, guildID)
	if err != nil {
		return err
	}

	// Insert new disabled categories
	for _, cat := range categories {
		_, err = tx.Exec(`INSERT INTO guild_disabled_commands (guild_id, category) VALUES (?, ?)`, guildID, cat)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
