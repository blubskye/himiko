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
