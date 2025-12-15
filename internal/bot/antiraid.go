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
	"strings"
	"sync"
	"time"

	"github.com/blubskye/himiko/internal/database"
	"github.com/bwmarrin/discordgo"
)

// RaidTracker manages raid detection state
type RaidTracker struct {
	mu            sync.RWMutex
	lastRaidAlert map[string]time.Time // Last raid alert per guild
	inLockdown    map[string]time.Time // Lockdown start time per guild
}

// NewRaidTracker creates a new raid tracker
func NewRaidTracker() *RaidTracker {
	return &RaidTracker{
		lastRaidAlert: make(map[string]time.Time),
		inLockdown:    make(map[string]time.Time),
	}
}

// Global raid tracker
var raidTracker = NewRaidTracker()

// CheckRaid checks if a raid is occurring and takes action
func (b *Bot) CheckRaid(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	cfg, err := b.DB.GetAntiRaidConfig(m.GuildID)
	if err != nil || !cfg.Enabled {
		return
	}

	now := time.Now()
	nowMs := now.UnixMilli()

	// Get account creation time from snowflake
	accountCreated := snowflakeToTimestamp(m.User.ID)

	// Record the join
	b.DB.RecordMemberJoin(m.GuildID, m.User.ID, nowMs, accountCreated)

	// Clean up old join records
	cleanupTime := now.Add(-time.Duration(cfg.RaidTime*3) * time.Second).UnixMilli()
	b.DB.CleanOldJoins(m.GuildID, cleanupTime)

	// Handle based on auto-silence mode
	switch cfg.AutoSilence {
	case -2: // Log only
		b.LogMemberJoin(s, m, cfg, accountCreated)

	case -1: // Alert on all joins
		b.AlertMemberJoin(s, m, cfg, accountCreated)

	case 0: // Off - just check for raid
		b.checkForRaid(s, m.GuildID, cfg, now)

	case 1: // Raid mode - silence if raid detected
		if b.checkForRaid(s, m.GuildID, cfg, now) {
			b.SilenceMember(s, m.GuildID, m.User.ID, cfg)
		}

	case 2: // All mode - silence everyone
		b.SilenceMember(s, m.GuildID, m.User.ID, cfg)
		b.AlertMemberJoin(s, m, cfg, accountCreated)
	}
}

// checkForRaid checks if a raid is occurring
func (b *Bot) checkForRaid(s *discordgo.Session, guildID string, cfg *database.AntiRaidConfig, now time.Time) bool {
	sinceTime := now.Add(-time.Duration(cfg.RaidTime) * time.Second).UnixMilli()
	count, err := b.DB.CountRecentJoins(guildID, sinceTime)
	if err != nil {
		return false
	}

	if count >= cfg.RaidSize {
		// Check rate limit on alerts
		raidTracker.mu.Lock()
		lastAlert := raidTracker.lastRaidAlert[guildID]
		alertCooldown := time.Duration(cfg.RaidTime*2) * time.Second

		if time.Since(lastAlert) > alertCooldown {
			raidTracker.lastRaidAlert[guildID] = now
			raidTracker.mu.Unlock()

			// Raid detected!
			b.HandleRaidDetected(s, guildID, cfg, count)
			return true
		}
		raidTracker.mu.Unlock()
		return true // Still in raid window
	}

	return false
}

// HandleRaidDetected handles a detected raid
func (b *Bot) HandleRaidDetected(s *discordgo.Session, guildID string, cfg *database.AntiRaidConfig, joinCount int) {
	// Alert moderators
	if cfg.LogChannelID != "" {
		sinceTime := time.Now().Add(-time.Duration(cfg.RaidTime) * time.Second).UnixMilli()
		joins, _ := b.DB.GetRecentJoins(guildID, sinceTime)

		var userList strings.Builder
		for i, join := range joins {
			if i >= 10 {
				fmt.Fprintf(&userList, "\n... and %d more", len(joins)-10)
				break
			}
			accountAge := time.Since(time.UnixMilli(join.AccountCreatedAt))
			fmt.Fprintf(&userList, "<@%s> (Account: %s)\n", join.UserID, formatDuration(accountAge))
		}

		alertText := ""
		if cfg.AlertRoleID != "" {
			alertText = fmt.Sprintf("<@&%s> ", cfg.AlertRoleID)
		}

		embed := &discordgo.MessageEmbed{
			Title:       "RAID DETECTED",
			Description: fmt.Sprintf("%d users joined in the past %d seconds!", joinCount, cfg.RaidTime),
			Color:       0xFF0000,
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Recent Joins", Value: userList.String(), Inline: false},
				{Name: "Action", Value: cfg.Action, Inline: true},
				{Name: "Auto-Silence", Value: autoSilenceModeLabel(cfg.AutoSilence), Inline: true},
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}

		s.ChannelMessageSendComplex(cfg.LogChannelID, &discordgo.MessageSend{
			Content: alertText,
			Embed:   embed,
		})
	}

	// Auto-lockdown if configured
	if cfg.LockdownDuration > 0 {
		b.EnableLockdown(s, guildID, cfg)
	}

	// Take action on raid users if auto-silence is in raid mode
	if cfg.AutoSilence == 1 {
		sinceTime := time.Now().Add(-time.Duration(cfg.RaidTime) * time.Second).UnixMilli()
		joins, _ := b.DB.GetRecentJoins(guildID, sinceTime)

		for _, join := range joins {
			b.SilenceMember(s, guildID, join.UserID, cfg)
		}
	}
}

// SilenceMember silences a member
func (b *Bot) SilenceMember(s *discordgo.Session, guildID, userID string, cfg *database.AntiRaidConfig) {
	if cfg.SilentRoleID == "" {
		return
	}

	switch cfg.Action {
	case "silence":
		s.GuildMemberRoleAdd(guildID, userID, cfg.SilentRoleID)

	case "kick":
		s.GuildMemberDeleteWithReason(guildID, userID, "Raid protection")

	case "ban":
		s.GuildBanCreateWithReason(guildID, userID, "Raid protection", 1)
	}
}

// LogMemberJoin logs a member join (log mode)
func (b *Bot) LogMemberJoin(s *discordgo.Session, m *discordgo.GuildMemberAdd, cfg *database.AntiRaidConfig, accountCreated int64) {
	if cfg.LogChannelID == "" {
		return
	}

	accountAge := time.Since(time.UnixMilli(accountCreated))

	embed := &discordgo.MessageEmbed{
		Title:       "Member Joined",
		Description: fmt.Sprintf("%s joined the server", m.User.Mention()),
		Color:       0x00FF00,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Account Age", Value: formatDuration(accountAge), Inline: true},
			{Name: "User ID", Value: m.User.ID, Inline: true},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: avatarURL(m.User),
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	s.ChannelMessageSendEmbed(cfg.LogChannelID, embed)
}

// AlertMemberJoin alerts on a member join (alert mode)
func (b *Bot) AlertMemberJoin(s *discordgo.Session, m *discordgo.GuildMemberAdd, cfg *database.AntiRaidConfig, accountCreated int64) {
	if cfg.LogChannelID == "" {
		return
	}

	accountAge := time.Since(time.UnixMilli(accountCreated))

	alertText := ""
	if cfg.AlertRoleID != "" {
		alertText = fmt.Sprintf("<@&%s> ", cfg.AlertRoleID)
	}

	// Flag suspicious account ages
	color := 0xFFFF00 // Yellow for new joins
	warning := ""
	if accountAge < 24*time.Hour {
		color = 0xFF0000 // Red for very new accounts
		warning = " **[NEW ACCOUNT]**"
	} else if accountAge < 7*24*time.Hour {
		color = 0xFFA500 // Orange for week-old accounts
		warning = " [Recent Account]"
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Member Joined" + warning,
		Description: fmt.Sprintf("%s joined the server", m.User.Mention()),
		Color:       color,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Account Age", Value: formatDuration(accountAge), Inline: true},
			{Name: "User ID", Value: m.User.ID, Inline: true},
			{Name: "Username", Value: m.User.Username, Inline: true},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: avatarURL(m.User),
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	s.ChannelMessageSendComplex(cfg.LogChannelID, &discordgo.MessageSend{
		Content: alertText,
		Embed:   embed,
	})
}

// EnableLockdown enables server lockdown
func (b *Bot) EnableLockdown(s *discordgo.Session, guildID string, cfg *database.AntiRaidConfig) {
	raidTracker.mu.Lock()
	if _, ok := raidTracker.inLockdown[guildID]; ok {
		raidTracker.mu.Unlock()
		return // Already in lockdown
	}
	raidTracker.inLockdown[guildID] = time.Now()
	raidTracker.mu.Unlock()

	// Raise verification level
	highLevel := discordgo.VerificationLevelHigh
	_, err := s.GuildEdit(guildID, &discordgo.GuildParams{
		VerificationLevel: &highLevel,
	})

	if err == nil && cfg.LogChannelID != "" {
		embed := &discordgo.MessageEmbed{
			Title:       "Server Lockdown Enabled",
			Description: fmt.Sprintf("Verification level raised to **High** for %d seconds", cfg.LockdownDuration),
			Color:       0xFF0000,
			Timestamp:   time.Now().Format(time.RFC3339),
		}
		s.ChannelMessageSendEmbed(cfg.LogChannelID, embed)
	}
}

// DisableLockdown disables server lockdown
func (b *Bot) DisableLockdown(s *discordgo.Session, guildID string) {
	raidTracker.mu.Lock()
	delete(raidTracker.inLockdown, guildID)
	raidTracker.mu.Unlock()

	// Lower verification level
	mediumLevel := discordgo.VerificationLevelMedium
	s.GuildEdit(guildID, &discordgo.GuildParams{
		VerificationLevel: &mediumLevel,
	})
}

// CheckLockdownExpiry checks if any lockdowns have expired
func (b *Bot) CheckLockdownExpiry(s *discordgo.Session) {
	raidTracker.mu.Lock()
	defer raidTracker.mu.Unlock()

	for guildID, startTime := range raidTracker.inLockdown {
		cfg, err := b.DB.GetAntiRaidConfig(guildID)
		if err != nil {
			continue
		}

		if time.Since(startTime) > time.Duration(cfg.LockdownDuration)*time.Second {
			delete(raidTracker.inLockdown, guildID)

			// Lower verification level
			mediumLevel := discordgo.VerificationLevelMedium
			s.GuildEdit(guildID, &discordgo.GuildParams{
				VerificationLevel: &mediumLevel,
			})

			if cfg.LogChannelID != "" {
				embed := &discordgo.MessageEmbed{
					Title:       "Lockdown Expired",
					Description: "Server verification level restored to **Medium**",
					Color:       0x00FF00,
					Timestamp:   time.Now().Format(time.RFC3339),
				}
				s.ChannelMessageSendEmbed(cfg.LogChannelID, embed)
			}
		}
	}
}

// Helper function for auto-silence mode labels
func autoSilenceModeLabel(mode int) string {
	labels := map[int]string{
		-2: "Log only",
		-1: "Alert",
		0:  "Off",
		1:  "Raid mode",
		2:  "All",
	}
	if label, ok := labels[mode]; ok {
		return label
	}
	return "Unknown"
}
