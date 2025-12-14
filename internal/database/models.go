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

import "time"

type GuildSettings struct {
	GuildID        string
	Prefix         string
	ModLogChannel  *string
	WelcomeChannel *string
	WelcomeMessage *string
	JoinDMTitle    *string
	JoinDMMessage  *string
}

type CustomCommand struct {
	ID        int64
	GuildID   string
	Name      string
	Response  string
	CreatedBy string
	UseCount  int
}

type CommandHistory struct {
	ID         int64
	GuildID    *string
	ChannelID  string
	UserID     string
	Command    string
	Args       *string
	ExecutedAt time.Time
}

type Warning struct {
	ID          int64
	GuildID     string
	UserID      string
	ModeratorID string
	Reason      *string
	CreatedAt   time.Time
}

type DeletedMessage struct {
	ID        int64
	GuildID   *string
	ChannelID string
	UserID    string
	Content   string
	DeletedAt time.Time
}

type ScheduledMessage struct {
	ID           int64
	GuildID      *string
	ChannelID    string
	UserID       string
	Message      string
	ScheduledFor time.Time
}

type AFKStatus struct {
	UserID  string
	Message *string
	SetAt   time.Time
}

type Reminder struct {
	ID        int64
	UserID    string
	ChannelID string
	Message   string
	RemindAt  time.Time
}

type Tag struct {
	ID        int64
	GuildID   string
	Name      string
	Content   string
	CreatedBy string
	UseCount  int
}

type KeywordNotification struct {
	ID      int64
	UserID  string
	GuildID *string
	Keyword string
}

// XP/Leveling
type UserXP struct {
	GuildID   string
	UserID    string
	XP        int64
	Level     int
	UpdatedAt time.Time
}

// Regex Filters
type RegexFilter struct {
	ID        int64
	GuildID   string
	Pattern   string
	Action    string // warn, delete, ban
	Reason    string
	CreatedBy string
	CreatedAt time.Time
}

// Auto-Clean Channels
type AutoCleanChannel struct {
	ID              int64
	GuildID         string
	ChannelID       string
	IntervalHours   int
	WarningMinutes  int
	NextRun         time.Time
	CleanMessage    bool
	CleanImage      bool
	CreatedBy       string
	CreatedAt       time.Time
}

// Logging Configuration
type LoggingConfig struct {
	GuildID           string
	LogChannelID      *string
	Enabled           bool
	MessageDelete     bool
	MessageEdit       bool
	VoiceJoin         bool
	VoiceLeave        bool
	NicknameChange    bool
	AvatarChange      bool
	PresenceChange    bool
	PresenceBatchMins int
}

// Disabled Log Channels
type DisabledLogChannel struct {
	GuildID   string
	ChannelID string
}

// Voice XP Configuration
type VoiceXPConfig struct {
	GuildID      string
	Enabled      bool
	XPRate       int
	IntervalMins int
	IgnoreAFK    bool
}

// Level Ranks
type LevelRank struct {
	ID        int64
	GuildID   string
	RoleID    string
	Level     int
	CreatedAt time.Time
}

// DM Forwarding Configuration
type DMConfig struct {
	GuildID   string
	ChannelID string
	Enabled   bool
}

// Bot Bans
type BotBan struct {
	ID        int64
	TargetID  string
	BanType   string // user, server
	Reason    string
	BannedBy  string
	CreatedAt time.Time
}

// Moderation Actions
type ModAction struct {
	ID          int64
	GuildID     string
	ModeratorID string
	TargetID    string
	Action      string // ban, unban, kick, timeout
	Reason      *string
	Timestamp   int64
	CreatedAt   time.Time
}

// Mod Stats
type ModStats struct {
	TotalActions int
	ActionCounts map[string]int
	TopMods      []ModeratorCount
}

type ModeratorCount struct {
	ModeratorID string
	Count       int
	Actions     map[string]int
}

// Mention Responses
type MentionResponse struct {
	ID          int64
	GuildID     string
	TriggerText string
	Response    string
	ImageURL    *string
	CreatedBy   string
	CreatedAt   time.Time
}

// Spam Filter Config
type SpamFilterConfig struct {
	GuildID     string
	Enabled     bool
	MaxMentions int
	MaxLinks    int
	MaxEmojis   int
	Action      string // delete, warn, kick, ban
}

// Ticket System Config
type TicketConfig struct {
	GuildID   string
	ChannelID string
	Enabled   bool
}

// Anti-Raid Config
type AntiRaidConfig struct {
	GuildID          string
	Enabled          bool
	RaidTime         int    // Time window in seconds
	RaidSize         int    // Number of joins to trigger raid
	AutoSilence      int    // -2=log, -1=alert, 0=off, 1=raid, 2=all
	LockdownDuration int    // Seconds to lockdown
	SilentRoleID     string // Role to assign to silenced users
	AlertRoleID      string // Role to ping on raid
	LogChannelID     string // Channel for raid alerts
	Action           string // silence, kick, ban
}

// Member Join record
type MemberJoin struct {
	ID               int64
	GuildID          string
	UserID           string
	JoinedAt         int64
	AccountCreatedAt int64
}

// Anti-Spam Config
type AntiSpamConfig struct {
	GuildID        string
	Enabled        bool
	BasePressure   float64
	ImagePressure  float64
	LinkPressure   float64
	PingPressure   float64
	LengthPressure float64
	LinePressure   float64
	RepeatPressure float64
	MaxPressure    float64
	PressureDecay  float64 // Seconds to decay BasePressure
	Action         string  // delete, warn, silence, kick, ban
	SilentRoleID   string
}

// Scheduled Event
type ScheduledEvent struct {
	ID        int64
	GuildID   string
	EventType string // unsilence, unmute, etc
	TargetID  string
	ExecuteAt int64
}

// User Alias - tracks username/nickname history
type UserAlias struct {
	ID        int64
	UserID    string
	Alias     string
	AliasType string // username, nickname
	FirstSeen time.Time
	LastSeen  time.Time
	UseCount  int
}

// User Activity - tracks user activity per guild
type UserActivity struct {
	GuildID      string
	UserID       string
	FirstSeen    *time.Time
	FirstMessage *time.Time
	LastSeen     *time.Time
	MessageCount int
}

// Music Settings - per-guild music configuration
type MusicSettings struct {
	GuildID     string
	DJRoleID    *string
	ModRoleID   *string
	Volume      int
	MusicFolder *string
}

// Music Queue Item
type MusicQueueItem struct {
	ID        int64
	GuildID   string
	ChannelID string
	UserID    string
	Title     string
	URL       string
	Duration  int
	Thumbnail *string
	IsLocal   bool
	Position  int
	AddedAt   time.Time
}

// Music History
type MusicHistory struct {
	ID       int64
	GuildID  string
	UserID   string
	Title    string
	URL      string
	PlayedAt time.Time
}

// Disabled Commands/Categories - for per-guild command enable/disable
type DisabledCommand struct {
	ID          int64
	GuildID     string
	CommandName *string // nil for category disable
	Category    *string // nil for individual command disable
	CreatedAt   time.Time
}
