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
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Response helpers
func respond(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

func respondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func respondEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func respondEmbedEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

func respondDeferred(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
}

func respondDeferredEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

func followUp(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: content,
	})
}

func followUpEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed) {
	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed},
	})
}

func respondAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate, choices []*discordgo.ApplicationCommandOptionChoice) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
}

// Option helpers
func getStringOption(i *discordgo.InteractionCreate, name string) string {
	options := getOptions(i)
	for _, opt := range options {
		if opt.Name == name {
			return opt.StringValue()
		}
	}
	return ""
}

func getIntOption(i *discordgo.InteractionCreate, name string) int64 {
	options := getOptions(i)
	for _, opt := range options {
		if opt.Name == name {
			return opt.IntValue()
		}
	}
	return 0
}

func getBoolOption(i *discordgo.InteractionCreate, name string) bool {
	options := getOptions(i)
	for _, opt := range options {
		if opt.Name == name {
			return opt.BoolValue()
		}
	}
	return false
}

func getUserOption(i *discordgo.InteractionCreate, name string) *discordgo.User {
	options := getOptions(i)
	for _, opt := range options {
		if opt.Name == name {
			return opt.UserValue(nil)
		}
	}
	return nil
}

func getChannelOption(i *discordgo.InteractionCreate, name string) *discordgo.Channel {
	options := getOptions(i)
	for _, opt := range options {
		if opt.Name == name {
			return opt.ChannelValue(nil)
		}
	}
	return nil
}

func getRoleOption(i *discordgo.InteractionCreate, name string) *discordgo.Role {
	options := getOptions(i)
	for _, opt := range options {
		if opt.Name == name {
			return opt.RoleValue(nil, "")
		}
	}
	return nil
}

func getOptions(i *discordgo.InteractionCreate) []*discordgo.ApplicationCommandInteractionDataOption {
	options := i.ApplicationCommandData().Options

	// Handle subcommands
	if len(options) > 0 {
		if options[0].Type == discordgo.ApplicationCommandOptionSubCommand {
			return options[0].Options
		}
		if options[0].Type == discordgo.ApplicationCommandOptionSubCommandGroup {
			if len(options[0].Options) > 0 {
				return options[0].Options[0].Options
			}
		}
	}

	return options
}

func getSubcommandName(i *discordgo.InteractionCreate) string {
	options := i.ApplicationCommandData().Options
	if len(options) > 0 && options[0].Type == discordgo.ApplicationCommandOptionSubCommand {
		return options[0].Name
	}
	return ""
}

// Permission helpers
func hasPermission(s *discordgo.Session, guildID, userID string, permission int64) bool {
	member, err := s.GuildMember(guildID, userID)
	if err != nil {
		return false
	}

	guild, err := s.State.Guild(guildID)
	if err != nil {
		guild, err = s.Guild(guildID)
		if err != nil {
			return false
		}
	}

	// Check if owner
	if guild.OwnerID == userID {
		return true
	}

	// Check role permissions
	for _, roleID := range member.Roles {
		for _, role := range guild.Roles {
			if role.ID == roleID {
				if role.Permissions&permission != 0 || role.Permissions&discordgo.PermissionAdministrator != 0 {
					return true
				}
			}
		}
	}

	return false
}

func isAdmin(s *discordgo.Session, guildID, userID string) bool {
	return hasPermission(s, guildID, userID, discordgo.PermissionAdministrator)
}

func isModerator(s *discordgo.Session, guildID, userID string) bool {
	return hasPermission(s, guildID, userID, discordgo.PermissionKickMembers|discordgo.PermissionBanMembers)
}

// String helpers
func replacePlaceholders(text string, user *discordgo.User, guildID string) string {
	text = strings.ReplaceAll(text, "{user}", user.Mention())
	text = strings.ReplaceAll(text, "{username}", user.Username)
	text = strings.ReplaceAll(text, "{userid}", user.ID)
	text = strings.ReplaceAll(text, "{server}", guildID)
	return text
}

func formatUnixTime(t time.Time) string {
	return strconv.FormatInt(t.Unix(), 10)
}

func containsWord(text, word string) bool {
	text = strings.ToLower(text)
	word = strings.ToLower(word)
	return strings.Contains(text, word)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Time parsing
func parseDuration(input string) (time.Duration, error) {
	input = strings.ToLower(strings.TrimSpace(input))

	var total time.Duration
	var num string

	for _, c := range input {
		if c >= '0' && c <= '9' {
			num += string(c)
		} else if c == ' ' {
			continue
		} else {
			if num == "" {
				continue
			}
			n, err := strconv.Atoi(num)
			if err != nil {
				return 0, err
			}

			switch c {
			case 's':
				total += time.Duration(n) * time.Second
			case 'm':
				total += time.Duration(n) * time.Minute
			case 'h':
				total += time.Duration(n) * time.Hour
			case 'd':
				total += time.Duration(n) * 24 * time.Hour
			case 'w':
				total += time.Duration(n) * 7 * 24 * time.Hour
			}
			num = ""
		}
	}

	return total, nil
}

// Embed helpers
func errorEmbed(title, description string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       0xED4245, // Red
	}
}

func successEmbed(title, description string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       0x57F287, // Green
	}
}

func infoEmbed(title, description string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       0x5865F2, // Blurple
	}
}

// Avatar URL helper
func avatarURL(user *discordgo.User) string {
	if user.Avatar == "" {
		return fmt.Sprintf("https://cdn.discordapp.com/embed/avatars/%d.png", (user.ID[len(user.ID)-1]-'0')%5)
	}
	ext := "png"
	if strings.HasPrefix(user.Avatar, "a_") {
		ext = "gif"
	}
	return fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.%s?size=1024", user.ID, user.Avatar, ext)
}

func bannerURL(user *discordgo.User) string {
	if user.Banner == "" {
		return ""
	}
	ext := "png"
	if strings.HasPrefix(user.Banner, "a_") {
		ext = "gif"
	}
	return fmt.Sprintf("https://cdn.discordapp.com/banners/%s/%s.%s?size=1024", user.ID, user.Banner, ext)
}

func guildIconURL(guild *discordgo.Guild) string {
	if guild.Icon == "" {
		return ""
	}
	ext := "png"
	if strings.HasPrefix(guild.Icon, "a_") {
		ext = "gif"
	}
	return fmt.Sprintf("https://cdn.discordapp.com/icons/%s/%s.%s?size=1024", guild.ID, guild.Icon, ext)
}

func guildBannerURL(guild *discordgo.Guild) string {
	if guild.Banner == "" {
		return ""
	}
	return fmt.Sprintf("https://cdn.discordapp.com/banners/%s/%s.png?size=1024", guild.ID, guild.Banner)
}
