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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// BanEntry represents a ban record for export/import
type BanEntry struct {
	UserID   string    `json:"user_id"`
	Username string    `json:"username"`
	Reason   string    `json:"reason"`
	BannedAt time.Time `json:"banned_at,omitempty"`
}

func (ch *CommandHandler) registerBanExportCommands() {
	// Export bans
	ch.Register(&Command{
		Name:        "exportbans",
		Description: "Export the server's ban list to JSON",
		Category:    "BanExport",
		Handler:     ch.exportBansHandler,
	})

	// Import bans
	ch.Register(&Command{
		Name:        "importbans",
		Description: "Import bans from a JSON file",
		Category:    "BanExport",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Name:        "file",
				Description: "JSON file containing ban list",
				Required:    true,
			},
		},
		Handler: ch.importBansHandler,
	})

	// Scan bans
	ch.Register(&Command{
		Name:        "scanbans",
		Description: "Scan and show statistics about the ban list",
		Category:    "BanExport",
		Handler:     ch.scanBansHandler,
	})
}

func (ch *CommandHandler) exportBansHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Check ban permission
	perms, err := s.UserChannelPermissions(i.Member.User.ID, i.ChannelID)
	if err != nil || perms&discordgo.PermissionBanMembers == 0 {
		respondEphemeral(s, i, "You need Ban Members permission to export bans.")
		return
	}

	respondDeferred(s, i)

	// Get all bans
	bans, err := s.GuildBans(i.GuildID, 0, "", "")
	if err != nil {
		followUp(s, i, "Failed to get ban list.")
		return
	}

	if len(bans) == 0 {
		followUp(s, i, "No bans found in this server.")
		return
	}

	// Convert to export format
	entries := make([]BanEntry, 0, len(bans))
	for _, ban := range bans {
		entries = append(entries, BanEntry{
			UserID:   ban.User.ID,
			Username: ban.User.Username,
			Reason:   ban.Reason,
			BannedAt: time.Now(),
		})
	}

	// Create JSON
	jsonData, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		followUp(s, i, "Failed to create JSON export.")
		return
	}

	// Get guild name for filename
	guild, _ := s.Guild(i.GuildID)
	guildName := "server"
	if guild != nil {
		guildName = strings.ReplaceAll(guild.Name, " ", "_")
	}

	filename := fmt.Sprintf("%s_bans_%s.json", guildName, time.Now().Format("2006-01-02"))

	// Send as file
	reader := bytes.NewReader(jsonData)
	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: fmt.Sprintf(":white_check_mark: Exported **%d** bans to `%s`", len(entries), filename),
		Files: []*discordgo.File{
			{
				Name:        filename,
				ContentType: "application/json",
				Reader:      reader,
			},
		},
	})

	if err != nil {
		followUp(s, i, "Failed to send ban export file.")
	}
}

func (ch *CommandHandler) importBansHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Check ban permission
	perms, err := s.UserChannelPermissions(i.Member.User.ID, i.ChannelID)
	if err != nil || perms&discordgo.PermissionBanMembers == 0 {
		respondEphemeral(s, i, "You need Ban Members permission to import bans.")
		return
	}

	// Get the attachment
	data := i.ApplicationCommandData()
	attachmentID := ""
	for _, opt := range data.Options {
		if opt.Name == "file" {
			attachmentID = opt.Value.(string)
			break
		}
	}

	attachment, ok := data.Resolved.Attachments[attachmentID]
	if !ok {
		respondEphemeral(s, i, "Please attach a JSON file.")
		return
	}

	if !strings.HasSuffix(strings.ToLower(attachment.Filename), ".json") {
		respondEphemeral(s, i, "Please attach a JSON file.")
		return
	}

	respondDeferred(s, i)

	// Download the file
	resp, err := httpClient.Get(attachment.URL)
	if err != nil {
		followUp(s, i, "Failed to download file.")
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		followUp(s, i, "Failed to read file.")
		return
	}

	var entries []BanEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		followUp(s, i, "Failed to parse JSON. Ensure the file is in the correct format.")
		return
	}

	if len(entries) == 0 {
		followUp(s, i, "No ban entries found in the file.")
		return
	}

	imported := 0
	skipped := 0
	errors := 0

	for _, entry := range entries {
		if entry.UserID == "" {
			skipped++
			continue
		}

		reason := entry.Reason
		if reason == "" {
			reason = "Imported ban"
		}
		reason = fmt.Sprintf("[Import] %s | Imported by %s", reason, i.Member.User.Username)

		err := s.GuildBanCreateWithReason(i.GuildID, entry.UserID, reason, 0)
		if err != nil {
			if strings.Contains(err.Error(), "already banned") {
				skipped++
			} else {
				errors++
			}
			continue
		}
		imported++
	}

	embed := &discordgo.MessageEmbed{
		Title: ":white_check_mark: Bans Imported",
		Color: 0x00FF00,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Imported", Value: fmt.Sprintf("%d", imported), Inline: true},
			{Name: "Skipped", Value: fmt.Sprintf("%d", skipped), Inline: true},
			{Name: "Errors", Value: fmt.Sprintf("%d", errors), Inline: true},
		},
	}

	followUpEmbed(s, i, embed)
}

func (ch *CommandHandler) scanBansHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Check ban permission
	perms, err := s.UserChannelPermissions(i.Member.User.ID, i.ChannelID)
	if err != nil || perms&discordgo.PermissionBanMembers == 0 {
		respondEphemeral(s, i, "You need Ban Members permission to scan bans.")
		return
	}

	respondDeferred(s, i)

	bans, err := s.GuildBans(i.GuildID, 0, "", "")
	if err != nil {
		followUp(s, i, "Failed to get ban list.")
		return
	}

	if len(bans) == 0 {
		followUp(s, i, "No bans found in this server.")
		return
	}

	// Analyze bans
	withReason := 0
	withoutReason := 0
	deletedUsers := 0

	for _, ban := range bans {
		if ban.Reason != "" {
			withReason++
		} else {
			withoutReason++
		}
		if strings.HasPrefix(ban.User.Username, "Deleted User") {
			deletedUsers++
		}
	}

	embed := &discordgo.MessageEmbed{
		Title: "Ban List Analysis",
		Color: 0x3498DB,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Total Bans", Value: fmt.Sprintf("%d", len(bans)), Inline: true},
			{Name: "With Reason", Value: fmt.Sprintf("%d", withReason), Inline: true},
			{Name: "Without Reason", Value: fmt.Sprintf("%d", withoutReason), Inline: true},
			{Name: "Deleted Accounts", Value: fmt.Sprintf("%d", deletedUsers), Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Use /exportbans to export the full list",
		},
	}

	followUpEmbed(s, i, embed)
}
