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

package bot

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/blubskye/himiko/internal/database"
	"github.com/bwmarrin/discordgo"
)

// UserPressure tracks spam pressure for a user
type UserPressure struct {
	Pressure    float64
	LastMessage string
	LastUpdate  time.Time
}

// SpamTracker manages spam pressure for all users
type SpamTracker struct {
	mu       sync.RWMutex
	pressure map[string]*UserPressure // key: guildID:userID
}

// NewSpamTracker creates a new spam tracker
func NewSpamTracker() *SpamTracker {
	return &SpamTracker{
		pressure: make(map[string]*UserPressure),
	}
}

// Global spam tracker
var spamTracker = NewSpamTracker()

// URL regex for detecting links
var urlRegex = regexp.MustCompile(`https?://[^\s]+`)

// GetPressure calculates pressure for a message
func (st *SpamTracker) GetPressure(guildID, userID string, msg *discordgo.Message, cfg *database.AntiSpamConfig) float64 {
	st.mu.Lock()
	defer st.mu.Unlock()

	key := guildID + ":" + userID
	up, exists := st.pressure[key]
	if !exists {
		up = &UserPressure{}
		st.pressure[key] = up
	}

	// Decay pressure based on time
	if !up.LastUpdate.IsZero() {
		elapsed := time.Since(up.LastUpdate).Seconds()
		decay := cfg.BasePressure * (elapsed / cfg.PressureDecay)
		up.Pressure -= decay
		if up.Pressure < 0 {
			up.Pressure = 0
		}
	}

	// Calculate new pressure
	pressure := cfg.BasePressure

	// Image/attachment pressure
	pressure += float64(len(msg.Attachments)) * cfg.ImagePressure
	pressure += float64(len(msg.Embeds)) * cfg.ImagePressure

	// Link pressure
	links := urlRegex.FindAllString(msg.Content, -1)
	pressure += float64(len(links)) * cfg.LinkPressure

	// Ping pressure
	pressure += float64(len(msg.Mentions)) * cfg.PingPressure
	if msg.MentionEveryone {
		pressure += cfg.PingPressure * 10 // Heavy penalty for @everyone
	}

	// Length pressure
	pressure += float64(len(msg.Content)) * cfg.LengthPressure

	// Line pressure
	lines := strings.Count(msg.Content, "\n")
	pressure += float64(lines) * cfg.LinePressure

	// Repeat pressure
	if msg.Content == up.LastMessage && msg.Content != "" {
		pressure += cfg.RepeatPressure
	}

	up.Pressure += pressure
	up.LastMessage = msg.Content
	up.LastUpdate = time.Now()

	return up.Pressure
}

// ResetPressure resets a user's pressure
func (st *SpamTracker) ResetPressure(guildID, userID string) {
	st.mu.Lock()
	defer st.mu.Unlock()

	key := guildID + ":" + userID
	delete(st.pressure, key)
}

// CheckSpam checks if a message should be flagged as spam
func (b *Bot) CheckSpam(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore bots
	if m.Author.Bot {
		return
	}

	// Ignore DMs
	if m.GuildID == "" {
		return
	}

	cfg, err := b.DB.GetAntiSpamConfig(m.GuildID)
	if err != nil || !cfg.Enabled {
		return
	}

	// Check if user is moderator (exempt from spam detection)
	if isModerator(s, m.GuildID, m.Author.ID) {
		return
	}

	pressure := spamTracker.GetPressure(m.GuildID, m.Author.ID, m.Message, cfg)

	if pressure >= cfg.MaxPressure {
		b.HandleSpamAction(s, m, cfg, pressure)
		spamTracker.ResetPressure(m.GuildID, m.Author.ID)
	}
}

// HandleSpamAction takes action against a spammer
func (b *Bot) HandleSpamAction(s *discordgo.Session, m *discordgo.MessageCreate, cfg *database.AntiSpamConfig, pressure float64) {
	switch cfg.Action {
	case "delete":
		s.ChannelMessageDelete(m.ChannelID, m.ID)

	case "warn":
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		// Send warning to user
		channel, err := s.UserChannelCreate(m.Author.ID)
		if err == nil {
			s.ChannelMessageSend(channel.ID, fmt.Sprintf("You are sending messages too quickly in **%s**. Please slow down.", m.GuildID))
		}

	case "silence":
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		if cfg.SilentRoleID != "" {
			s.GuildMemberRoleAdd(m.GuildID, m.Author.ID, cfg.SilentRoleID)
			// Log to mod channel
			b.LogSpamAction(s, m.GuildID, m.Author, "silenced", pressure)
		}

	case "kick":
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		s.GuildMemberDeleteWithReason(m.GuildID, m.Author.ID, "Spam detected")
		b.LogSpamAction(s, m.GuildID, m.Author, "kicked", pressure)

	case "ban":
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		s.GuildBanCreateWithReason(m.GuildID, m.Author.ID, "Spam detected", 1)
		b.LogSpamAction(s, m.GuildID, m.Author, "banned", pressure)
	}
}

// LogSpamAction logs a spam action to the mod log
func (b *Bot) LogSpamAction(s *discordgo.Session, guildID string, user *discordgo.User, action string, pressure float64) {
	settings, err := b.DB.GetGuildSettings(guildID)
	if err != nil || settings.ModLogChannel == nil {
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Anti-Spam Action",
		Description: fmt.Sprintf("User %s was %s for spam", user.Mention(), action),
		Color:       0xFF0000,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "User", Value: fmt.Sprintf("%s (%s)", user.Username, user.ID), Inline: true},
			{Name: "Pressure", Value: fmt.Sprintf("%.1f", pressure), Inline: true},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	s.ChannelMessageSendEmbed(*settings.ModLogChannel, embed)
}
