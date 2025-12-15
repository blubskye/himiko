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
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func (ch *CommandHandler) registerToolsCommands() {
	// TinyURL shortener
	ch.Register(&Command{
		Name:        "tinyurl",
		Description: "Shorten a URL",
		Category:    "Tools",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "url",
				Description: "URL to shorten",
				Required:    true,
			},
		},
		Handler: ch.tinyurlHandler,
	})

	// QR Code generator
	ch.Register(&Command{
		Name:        "qrcode",
		Description: "Generate a QR code",
		Category:    "Tools",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "text",
				Description: "Text or URL to encode",
				Required:    true,
			},
		},
		Handler: ch.qrcodeHandler,
	})

	// Timestamp generator
	ch.Register(&Command{
		Name:        "timestamp",
		Description: "Generate Discord timestamps",
		Category:    "Tools",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "time",
				Description: "Time offset from now (e.g., 1h, 2d, -30m)",
				Required:    false,
			},
		},
		Handler: ch.timestampHandler,
	})

	// Character count
	ch.Register(&Command{
		Name:        "charcount",
		Description: "Count characters in text",
		Category:    "Tools",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "text",
				Description: "Text to count",
				Required:    true,
			},
		},
		Handler: ch.charCountHandler,
	})

	// Snowflake decoder
	ch.Register(&Command{
		Name:        "snowflake",
		Description: "Decode a Discord snowflake ID",
		Category:    "Tools",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "id",
				Description: "Snowflake ID to decode",
				Required:    true,
			},
		},
		Handler: ch.snowflakeHandler,
	})

	// Server list
	ch.Register(&Command{
		Name:        "servers",
		Description: "List servers the bot is in",
		Category:    "Tools",
		Handler:     ch.serversHandler,
	})

	// Permissions calculator
	ch.Register(&Command{
		Name:        "permissions",
		Description: "View your permissions in this channel",
		Category:    "Tools",
		Handler:     ch.permissionsHandler,
	})

	// Raw message
	ch.Register(&Command{
		Name:        "raw",
		Description: "Get raw message content (for copying formatting)",
		Category:    "Tools",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "message_id",
				Description: "Message ID to get raw content of",
				Required:    true,
			},
		},
		Handler: ch.rawHandler,
	})

	// Message link
	ch.Register(&Command{
		Name:        "messagelink",
		Description: "Create a jump link to a message",
		Category:    "Tools",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "message_id",
				Description: "Message ID",
				Required:    true,
			},
		},
		Handler: ch.messageLinkHandler,
	})
}

func (ch *CommandHandler) tinyurlHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	urlStr := getStringOption(i, "url")

	// Generate TinyURL
	shortURL := fmt.Sprintf("https://tinyurl.com/api-create.php?url=%s", url.QueryEscape(urlStr))

	resp, err := httpClient.Get(shortURL)
	if err != nil {
		respondEphemeral(s, i, "Failed to shorten URL.")
		return
	}
	defer resp.Body.Close()

	buf := make([]byte, 1024)
	n, _ := resp.Body.Read(buf)
	result := string(buf[:n])

	embed := &discordgo.MessageEmbed{
		Title: "URL Shortened",
		Color: 0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Original", Value: truncate(urlStr, 100), Inline: false},
			{Name: "Shortened", Value: result, Inline: false},
		},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) qrcodeHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	text := getStringOption(i, "text")

	qrURL := fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=%s", url.QueryEscape(text))

	embed := &discordgo.MessageEmbed{
		Title: "QR Code",
		Image: &discordgo.MessageEmbedImage{URL: qrURL},
		Color: 0x5865F2,
		Footer: &discordgo.MessageEmbedFooter{
			Text: truncate(text, 100),
		},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) timestampHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	timeStr := getStringOption(i, "time")

	// Get current time
	now := time.Now()
	if timeStr != "" {
		duration, err := parseDuration(timeStr)
		if err == nil {
			now = now.Add(duration)
		}
	}

	unix := now.Unix()

	formats := map[string]string{
		"Short Time":     fmt.Sprintf("<t:%d:t>", unix),
		"Long Time":      fmt.Sprintf("<t:%d:T>", unix),
		"Short Date":     fmt.Sprintf("<t:%d:d>", unix),
		"Long Date":      fmt.Sprintf("<t:%d:D>", unix),
		"Short DateTime": fmt.Sprintf("<t:%d:f>", unix),
		"Long DateTime":  fmt.Sprintf("<t:%d:F>", unix),
		"Relative":       fmt.Sprintf("<t:%d:R>", unix),
	}

	var fields []*discordgo.MessageEmbedField
	for name, format := range formats {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   name,
			Value:  fmt.Sprintf("%s\n`%s`", format, format),
			Inline: true,
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:  "Discord Timestamps",
		Fields: fields,
		Color:  0x5865F2,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Unix: %d", unix),
		},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) charCountHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	text := getStringOption(i, "text")

	chars := len(text)
	words := len(strings.Fields(text))
	lines := len(strings.Split(text, "\n"))

	embed := &discordgo.MessageEmbed{
		Title: "Character Count",
		Color: 0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Characters", Value: strconv.Itoa(chars), Inline: true},
			{Name: "Words", Value: strconv.Itoa(words), Inline: true},
			{Name: "Lines", Value: strconv.Itoa(lines), Inline: true},
		},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) snowflakeHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	id := getStringOption(i, "id")

	timestamp, err := discordgo.SnowflakeTimestamp(id)
	if err != nil {
		respondEphemeral(s, i, "Invalid snowflake ID.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: "Snowflake Decoded",
		Color: 0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "ID", Value: id, Inline: true},
			{Name: "Created", Value: fmt.Sprintf("<t:%d:F>", timestamp.Unix()), Inline: true},
			{Name: "Relative", Value: fmt.Sprintf("<t:%d:R>", timestamp.Unix()), Inline: true},
			{Name: "Unix Timestamp", Value: strconv.FormatInt(timestamp.Unix(), 10), Inline: true},
		},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) serversHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guilds := s.State.Guilds

	var serverList strings.Builder
	for idx, guild := range guilds {
		if idx >= 25 {
			serverList.WriteString(fmt.Sprintf("\n... and %d more", len(guilds)-25))
			break
		}
		serverList.WriteString(fmt.Sprintf("**%s** (`%s`) - %d members\n", guild.Name, guild.ID, guild.MemberCount))
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Servers (%d)", len(guilds)),
		Description: serverList.String(),
		Color:       0x5865F2,
	}

	respondEmbedEphemeral(s, i, embed)
}

func (ch *CommandHandler) permissionsHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Get member permissions
	perms, err := s.State.UserChannelPermissions(i.Member.User.ID, i.ChannelID)
	if err != nil {
		perms, _ = s.UserChannelPermissions(i.Member.User.ID, i.ChannelID)
	}

	permNames := map[int64]string{
		discordgo.PermissionViewChannel:        "View Channel",
		discordgo.PermissionSendMessages:       "Send Messages",
		discordgo.PermissionManageMessages:     "Manage Messages",
		discordgo.PermissionEmbedLinks:         "Embed Links",
		discordgo.PermissionAttachFiles:        "Attach Files",
		discordgo.PermissionReadMessageHistory: "Read Message History",
		discordgo.PermissionMentionEveryone:    "Mention Everyone",
		discordgo.PermissionUseExternalEmojis:  "Use External Emojis",
		discordgo.PermissionAddReactions:       "Add Reactions",
		discordgo.PermissionManageChannels:     "Manage Channels",
		discordgo.PermissionKickMembers:        "Kick Members",
		discordgo.PermissionBanMembers:         "Ban Members",
		discordgo.PermissionAdministrator:      "Administrator",
		discordgo.PermissionManageGuild:        "Manage Server",
		discordgo.PermissionManageRoles:        "Manage Roles",
		discordgo.PermissionModerateMembers:    "Timeout Members",
	}

	var has, hasNot []string
	for perm, name := range permNames {
		if perms&perm != 0 {
			has = append(has, "✅ "+name)
		} else {
			hasNot = append(hasNot, "❌ "+name)
		}
	}

	embed := &discordgo.MessageEmbed{
		Title: "Your Permissions",
		Color: 0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Allowed", Value: strings.Join(has, "\n"), Inline: true},
			{Name: "Denied", Value: strings.Join(hasNot, "\n"), Inline: true},
		},
	}

	respondEmbedEphemeral(s, i, embed)
}

func (ch *CommandHandler) rawHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	messageID := getStringOption(i, "message_id")

	msg, err := s.ChannelMessage(i.ChannelID, messageID)
	if err != nil {
		respondEphemeral(s, i, "Message not found.")
		return
	}

	content := msg.Content
	if content == "" {
		content = "*No text content*"
	}

	// Escape markdown
	escaped := strings.ReplaceAll(content, "`", "\\`")

	embed := &discordgo.MessageEmbed{
		Title:       "Raw Message Content",
		Description: fmt.Sprintf("```\n%s\n```", escaped),
		Color:       0x5865F2,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Message by %s", msg.Author.Username),
		},
	}

	respondEmbedEphemeral(s, i, embed)
}

func (ch *CommandHandler) messageLinkHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	messageID := getStringOption(i, "message_id")

	link := fmt.Sprintf("https://discord.com/channels/%s/%s/%s", i.GuildID, i.ChannelID, messageID)

	respond(s, i, link)
}
