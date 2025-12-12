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
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func New(path string) (*DB, error) {
	db, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	d := &DB{db}
	if err := d.migrate(); err != nil {
		return nil, err
	}

	return d, nil
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

	CREATE INDEX IF NOT EXISTS idx_user_xp_guild ON user_xp(guild_id);
	CREATE INDEX IF NOT EXISTS idx_regex_filters_guild ON regex_filters(guild_id);
	CREATE INDEX IF NOT EXISTS idx_level_ranks_guild ON level_ranks(guild_id);
	CREATE INDEX IF NOT EXISTS idx_mod_actions_guild ON mod_actions(guild_id);
	CREATE INDEX IF NOT EXISTS idx_mod_actions_moderator ON mod_actions(guild_id, moderator_id);
	CREATE INDEX IF NOT EXISTS idx_mod_actions_target ON mod_actions(guild_id, target_id);
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

// Guild Settings
func (d *DB) GetGuildSettings(guildID string) (*GuildSettings, error) {
	var gs GuildSettings
	err := d.QueryRow(`SELECT guild_id, prefix, mod_log_channel, welcome_channel, welcome_message, join_dm_title, join_dm_message
		FROM guild_settings WHERE guild_id = ?`, guildID).Scan(
		&gs.GuildID, &gs.Prefix, &gs.ModLogChannel, &gs.WelcomeChannel, &gs.WelcomeMessage, &gs.JoinDMTitle, &gs.JoinDMMessage)
	if err == sql.ErrNoRows {
		return &GuildSettings{GuildID: guildID, Prefix: "/"}, nil
	}
	return &gs, err
}

func (d *DB) SetGuildSettings(gs *GuildSettings) error {
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
		gs.GuildID, gs.Prefix, gs.ModLogChannel, gs.WelcomeChannel, gs.WelcomeMessage, gs.JoinDMTitle, gs.JoinDMMessage)
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
	return &cc, err
}

func (d *DB) CreateCustomCommand(guildID, name, response, createdBy string) error {
	_, err := d.Exec(`INSERT INTO custom_commands (guild_id, name, response, created_by) VALUES (?, ?, ?, ?)`,
		guildID, name, response, createdBy)
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

	var history []CommandHistory
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
		guildID, userID, moderatorID, reason)
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
		guildID, channelID, userID, content)
	return err
}

func (d *DB) GetDeletedMessages(channelID string, limit int) ([]DeletedMessage, error) {
	rows, err := d.Query(`SELECT id, guild_id, channel_id, user_id, content, deleted_at
		FROM deleted_messages WHERE channel_id = ? ORDER BY deleted_at DESC LIMIT ?`, channelID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []DeletedMessage
	for rows.Next() {
		var dm DeletedMessage
		if err := rows.Scan(&dm.ID, &dm.GuildID, &dm.ChannelID, &dm.UserID, &dm.Content, &dm.DeletedAt); err != nil {
			return nil, err
		}
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
		guildID, channelID, userID, message, scheduledFor)
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
		userID, message)
	return err
}

func (d *DB) GetAFK(userID string) (*AFKStatus, error) {
	var afk AFKStatus
	err := d.QueryRow(`SELECT user_id, message, set_at FROM afk_status WHERE user_id = ?`, userID).Scan(
		&afk.UserID, &afk.Message, &afk.SetAt)
	if err == sql.ErrNoRows {
		return nil, nil
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
		userID, channelID, message, remindAt)
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
	return &t, err
}

func (d *DB) CreateTag(guildID, name, content, createdBy string) error {
	_, err := d.Exec(`INSERT INTO tags (guild_id, name, content, created_by) VALUES (?, ?, ?, ?)`,
		guildID, name, content, createdBy)
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

	var leaderboard []UserXP
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
		guildID, pattern, action, reason, createdBy)
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
		targetID, banType, reason, bannedBy)
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
	_, err := d.Exec(`INSERT INTO mod_actions (guild_id, moderator_id, target_id, action, reason, timestamp) VALUES (?, ?, ?, ?, ?, ?)`,
		guildID, moderatorID, targetID, action, reason, timestamp)
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
		actions = append(actions, ma)
	}
	return actions, rows.Err()
}

// ============ Mention Responses ============

func (d *DB) AddMentionResponse(guildID, trigger, response string, imageURL *string, createdBy string) error {
	_, err := d.Exec(`INSERT INTO mention_responses (guild_id, trigger_text, response, image_url, created_by)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(guild_id, trigger_text) DO UPDATE SET response = excluded.response, image_url = excluded.image_url`,
		guildID, trigger, response, imageURL, createdBy)
	return err
}

func (d *DB) RemoveMentionResponse(guildID, trigger string) error {
	_, err := d.Exec(`DELETE FROM mention_responses WHERE guild_id = ? AND trigger_text = ?`, guildID, trigger)
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
		responses = append(responses, mr)
	}
	return responses, rows.Err()
}

func (d *DB) GetMentionResponse(guildID, trigger string) (*MentionResponse, error) {
	var mr MentionResponse
	err := d.QueryRow(`SELECT id, guild_id, trigger_text, response, image_url, created_by, created_at
		FROM mention_responses WHERE guild_id = ? AND trigger_text = ?`, guildID, trigger).Scan(
		&mr.ID, &mr.GuildID, &mr.TriggerText, &mr.Response, &mr.ImageURL, &mr.CreatedBy, &mr.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
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
