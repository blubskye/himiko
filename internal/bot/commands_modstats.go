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
	"time"

	"github.com/bwmarrin/discordgo"
)

func (ch *CommandHandler) registerModStatsCommands() {
	// Mod stats
	ch.Register(&Command{
		Name:        "modstats",
		Description: "View moderation statistics for this server",
		Category:    "Moderation",
		Handler:     ch.modStatsHandler,
	})

	// Import mod history (from audit log/ban list)
	ch.Register(&Command{
		Name:        "importmodhistory",
		Description: "Import moderation history from audit logs or ban list",
		Category:    "Moderation",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "mode",
				Description: "Scan mode",
				Required:    false,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "Ban List (full history)", Value: "bans"},
					{Name: "Audit Log (includes moderator info)", Value: "audit"},
				},
			},
		},
		Handler: ch.importModHistoryHandler,
	})

	// User mod history
	ch.Register(&Command{
		Name:        "modhistory",
		Description: "View moderation history for a user",
		Category:    "Moderation",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "User to check history for",
				Required:    true,
			},
		},
		Handler: ch.modHistoryHandler,
	})
}

func (ch *CommandHandler) modStatsHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to view mod stats.")
		return
	}

	respondDeferred(s, i)

	// Get guild info
	guild, err := s.Guild(i.GuildID)
	if err != nil {
		followUp(s, i, "Failed to get guild info.")
		return
	}

	// Get ban count
	bans, _ := s.GuildBans(i.GuildID, 100, "", "")
	banCount := len(bans)

	// Get mod stats from database
	stats, err := ch.bot.DB.GetModStats(i.GuildID)
	if err != nil {
		followUp(s, i, "Failed to get mod stats.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:     fmt.Sprintf("Moderator Statistics for %s", guild.Name),
		Color:     0xFF69B4,
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: guild.IconURL("128")},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Server overview
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name: "Server Overview",
		Value: fmt.Sprintf("**Total Members:** %d\n**Total Bans:** %d+\n**Tracked Actions:** %d",
			guild.MemberCount, banCount, stats.TotalActions),
		Inline: false,
	})

	// Action breakdown
	if stats.TotalActions > 0 {
		actionText := fmt.Sprintf("**Bans:** %d\n**Unbans:** %d\n**Kicks:** %d\n**Timeouts:** %d",
			stats.ActionCounts["ban"],
			stats.ActionCounts["unban"],
			stats.ActionCounts["kick"],
			stats.ActionCounts["timeout"])
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Action Breakdown",
			Value:  actionText,
			Inline: true,
		})
	}

	// Top moderators
	if len(stats.TopMods) > 0 {
		var topModsText strings.Builder
		for idx, mod := range stats.TopMods {
			if idx >= 5 {
				break
			}

			modName := mod.ModeratorID
			// Try to resolve username
			member, err := s.GuildMember(i.GuildID, mod.ModeratorID)
			if err == nil && member.User != nil {
				modName = member.User.Username
			} else {
				user, err := s.User(mod.ModeratorID)
				if err == nil {
					modName = user.Username
				}
			}

			breakdown := []string{}
			if mod.Actions["ban"] > 0 {
				breakdown = append(breakdown, fmt.Sprintf("%d bans", mod.Actions["ban"]))
			}
			if mod.Actions["kick"] > 0 {
				breakdown = append(breakdown, fmt.Sprintf("%d kicks", mod.Actions["kick"]))
			}
			if mod.Actions["timeout"] > 0 {
				breakdown = append(breakdown, fmt.Sprintf("%d timeouts", mod.Actions["timeout"]))
			}
			if mod.Actions["unban"] > 0 {
				breakdown = append(breakdown, fmt.Sprintf("%d unbans", mod.Actions["unban"]))
			}

			topModsText.WriteString(fmt.Sprintf("**%d. %s** - %d actions\n   %s\n\n",
				idx+1, modName, mod.Count, strings.Join(breakdown, ", ")))
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Top Moderators",
			Value:  topModsText.String(),
			Inline: false,
		})
	} else {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "No Tracked Actions",
			Value:  "Run `/scanbans` to import moderation history from audit logs.",
			Inline: false,
		})
	}

	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "Use /scanbans to import audit log history",
	}

	followUpEmbed(s, i, embed)
}

func (ch *CommandHandler) importModHistoryHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to scan bans.")
		return
	}

	mode := getStringOption(i, "mode")
	if mode == "" {
		mode = "bans"
	}

	respondDeferred(s, i)

	if mode == "audit" {
		ch.scanAuditLogs(s, i)
	} else {
		ch.scanBanList(s, i)
	}
}

func (ch *CommandHandler) scanBanList(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// First build audit log cache for moderator info
	auditBanMap := make(map[string]struct {
		ModeratorID string
		Reason      string
		Timestamp   int64
	})

	// Try to get audit log info
	auditLogs, err := s.GuildAuditLog(i.GuildID, "", "", int(discordgo.AuditLogActionMemberBanAdd), 100)
	if err == nil {
		for _, entry := range auditLogs.AuditLogEntries {
			if entry.TargetID != "" {
				modID := "unknown"
				if entry.UserID != "" {
					modID = entry.UserID
				}
				auditBanMap[entry.TargetID] = struct {
					ModeratorID string
					Reason      string
					Timestamp   int64
				}{
					ModeratorID: modID,
					Reason:      entry.Reason,
					Timestamp:   snowflakeToTimestamp(entry.ID),
				}
			}
		}
	}

	// Fetch all bans
	bans, err := s.GuildBans(i.GuildID, 0, "", "")
	if err != nil {
		followUp(s, i, "Failed to get ban list.")
		return
	}

	if len(bans) == 0 {
		followUp(s, i, "No bans found in this server.")
		return
	}

	imported := 0
	skipped := 0

	for _, ban := range bans {
		// Check if already exists
		exists, _ := ch.bot.DB.ModActionExists(i.GuildID, ban.User.ID, "ban", 0)
		if exists {
			skipped++
			continue
		}

		// Get moderator info from audit cache
		auditInfo, hasAudit := auditBanMap[ban.User.ID]
		modID := "unknown"
		var reason *string
		timestamp := time.Now().UnixMilli()

		if hasAudit {
			modID = auditInfo.ModeratorID
			if auditInfo.Reason != "" {
				reason = &auditInfo.Reason
			}
			timestamp = auditInfo.Timestamp
		} else if ban.Reason != "" {
			reason = &ban.Reason
		}

		err := ch.bot.DB.AddModAction(i.GuildID, modID, ban.User.ID, "ban", reason, timestamp)
		if err == nil {
			imported++
		}
	}

	embed := &discordgo.MessageEmbed{
		Title: "Ban List Scan Complete",
		Color: 0x00FF00,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Total Processed", Value: fmt.Sprintf("%d", len(bans)), Inline: true},
			{Name: "Imported", Value: fmt.Sprintf("%d", imported), Inline: true},
			{Name: "Skipped (existing)", Value: fmt.Sprintf("%d", skipped), Inline: true},
			{Name: "With Moderator Info", Value: fmt.Sprintf("%d", len(auditBanMap)), Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Use /modstats to view moderator statistics",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	followUpEmbed(s, i, embed)
}

func (ch *CommandHandler) scanAuditLogs(s *discordgo.Session, i *discordgo.InteractionCreate) {
	imported := 0
	skipped := 0

	// Scan bans
	auditLogs, err := s.GuildAuditLog(i.GuildID, "", "", int(discordgo.AuditLogActionMemberBanAdd), 100)
	if err == nil {
		for _, entry := range auditLogs.AuditLogEntries {
			if entry.TargetID == "" {
				continue
			}

			timestamp := snowflakeToTimestamp(entry.ID)
			exists, _ := ch.bot.DB.ModActionExists(i.GuildID, entry.TargetID, "ban", timestamp)
			if exists {
				skipped++
				continue
			}

			modID := entry.UserID
			if modID == "" {
				modID = "unknown"
			}
			var reason *string
			if entry.Reason != "" {
				reason = &entry.Reason
			}

			err := ch.bot.DB.AddModAction(i.GuildID, modID, entry.TargetID, "ban", reason, timestamp)
			if err == nil {
				imported++
			}
		}
	}

	// Scan kicks
	auditLogs, err = s.GuildAuditLog(i.GuildID, "", "", int(discordgo.AuditLogActionMemberKick), 100)
	if err == nil {
		for _, entry := range auditLogs.AuditLogEntries {
			if entry.TargetID == "" {
				continue
			}

			timestamp := snowflakeToTimestamp(entry.ID)
			exists, _ := ch.bot.DB.ModActionExists(i.GuildID, entry.TargetID, "kick", timestamp)
			if exists {
				skipped++
				continue
			}

			modID := entry.UserID
			if modID == "" {
				modID = "unknown"
			}
			var reason *string
			if entry.Reason != "" {
				reason = &entry.Reason
			}

			err := ch.bot.DB.AddModAction(i.GuildID, modID, entry.TargetID, "kick", reason, timestamp)
			if err == nil {
				imported++
			}
		}
	}

	// Scan unbans
	auditLogs, err = s.GuildAuditLog(i.GuildID, "", "", int(discordgo.AuditLogActionMemberBanRemove), 100)
	if err == nil {
		for _, entry := range auditLogs.AuditLogEntries {
			if entry.TargetID == "" {
				continue
			}

			timestamp := snowflakeToTimestamp(entry.ID)
			exists, _ := ch.bot.DB.ModActionExists(i.GuildID, entry.TargetID, "unban", timestamp)
			if exists {
				skipped++
				continue
			}

			modID := entry.UserID
			if modID == "" {
				modID = "unknown"
			}
			var reason *string
			if entry.Reason != "" {
				reason = &entry.Reason
			}

			err := ch.bot.DB.AddModAction(i.GuildID, modID, entry.TargetID, "unban", reason, timestamp)
			if err == nil {
				imported++
			}
		}
	}

	embed := &discordgo.MessageEmbed{
		Title: "Audit Log Scan Complete",
		Color: 0x00FF00,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Imported", Value: fmt.Sprintf("%d", imported), Inline: true},
			{Name: "Skipped (duplicates)", Value: fmt.Sprintf("%d", skipped), Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Note: Audit logs only go back ~45 days. Use /scanbans bans for full ban list.",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	followUpEmbed(s, i, embed)
}

func (ch *CommandHandler) modHistoryHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to view mod history.")
		return
	}

	user := getUserOption(i, "user")
	if user == nil {
		respondEphemeral(s, i, "Please specify a user.")
		return
	}

	actions, err := ch.bot.DB.GetModActionsForTarget(i.GuildID, user.ID)
	if err != nil {
		respondEphemeral(s, i, "Failed to get mod history.")
		return
	}

	if len(actions) == 0 {
		respondEphemeral(s, i, fmt.Sprintf("No moderation history found for %s.", user.Username))
		return
	}

	var description strings.Builder
	for idx, action := range actions {
		if idx >= 10 {
			description.WriteString(fmt.Sprintf("\n... and %d more", len(actions)-10))
			break
		}

		reason := "No reason"
		if action.Reason != nil && *action.Reason != "" {
			reason = *action.Reason
		}

		description.WriteString(fmt.Sprintf("**%s** by <@%s>\n└ %s\n└ <t:%d:R>\n\n",
			strings.Title(action.Action), action.ModeratorID, truncate(reason, 50), action.Timestamp/1000))
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Mod History for %s", user.Username),
		Description: description.String(),
		Color:       0xFF69B4,
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: user.AvatarURL("64")},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Total: %d actions", len(actions)),
		},
	}

	respondEmbed(s, i, embed)
}
