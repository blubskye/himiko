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

	"github.com/blubskye/himiko/internal/database"
	"github.com/bwmarrin/discordgo"
)

func (ch *CommandHandler) registerXPCommands() {
	// XP/Level check
	ch.Register(&Command{
		Name:        "xp",
		Description: "Check your or another user's XP and level",
		Category:    "XP",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "User to check (defaults to yourself)",
				Required:    false,
			},
		},
		Handler: ch.xpHandler,
	})

	// Leaderboard
	ch.Register(&Command{
		Name:        "leaderboard",
		Description: "View the server XP leaderboard",
		Category:    "XP",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "page",
				Description: "Page number",
				Required:    false,
			},
		},
		Handler: ch.leaderboardHandler,
	})

	// Rank (same as XP but with different styling)
	ch.Register(&Command{
		Name:        "rank",
		Description: "Check your rank on the server",
		Category:    "XP",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "User to check (defaults to yourself)",
				Required:    false,
			},
		},
		Handler: ch.rankHandler,
	})

	// Set Level (Admin)
	ch.Register(&Command{
		Name:        "setlevel",
		Description: "Set a user's level",
		Category:    "XP",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "User to set level for",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "level",
				Description: "Level to set",
				Required:    true,
				MinValue:    floatPtr(0),
				MaxValue:    1000,
			},
		},
		Handler: ch.setLevelHandler,
	})

	// Set XP (Admin)
	ch.Register(&Command{
		Name:        "setxp",
		Description: "Set a user's XP",
		Category:    "XP",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "User to set XP for",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "xp",
				Description: "XP amount to set",
				Required:    true,
				MinValue:    floatPtr(0),
			},
		},
		Handler: ch.setXPHandler,
	})

	// Add XP (Admin)
	ch.Register(&Command{
		Name:        "addxp",
		Description: "Add XP to a user",
		Category:    "XP",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "User to add XP to",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "amount",
				Description: "Amount of XP to add",
				Required:    true,
			},
		},
		Handler: ch.addXPHandler,
	})

	// Mass Add XP (Admin)
	ch.Register(&Command{
		Name:        "massaddxp",
		Description: "Add XP to all members with a specific role",
		Category:    "XP",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionRole,
				Name:        "role",
				Description: "Role to target",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "amount",
				Description: "Amount of XP to add",
				Required:    true,
			},
		},
		Handler: ch.massAddXPHandler,
	})
}

func (ch *CommandHandler) xpHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := getUserOption(i, "user")
	if user == nil {
		user = i.Member.User
	}

	xpData, err := ch.bot.DB.GetUserXP(i.GuildID, user.ID)
	if err != nil {
		respondEphemeral(s, i, "Failed to get XP data.")
		return
	}

	rank, _ := ch.bot.DB.GetUserRank(i.GuildID, user.ID)
	nextLevelXP := database.XPForLevel(xpData.Level + 1)
	currentLevelXP := database.XPForLevel(xpData.Level)
	progress := xpData.XP - currentLevelXP
	needed := nextLevelXP - currentLevelXP

	// Create progress bar
	progressPercent := float64(progress) / float64(needed) * 100
	progressBar := createProgressBar(progressPercent, 20)

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s's XP", user.Username),
		Color: 0x5865F2,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: user.AvatarURL("128"),
		},
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Level", Value: strconv.Itoa(xpData.Level), Inline: true},
			{Name: "XP", Value: strconv.FormatInt(xpData.XP, 10), Inline: true},
			{Name: "Rank", Value: fmt.Sprintf("#%d", rank), Inline: true},
			{Name: "Progress to Next Level", Value: fmt.Sprintf("%s\n%d / %d XP (%.1f%%)", progressBar, progress, needed, progressPercent), Inline: false},
		},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) leaderboardHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	page := max(int(getIntOption(i, "page")), 1)

	perPage := 10
	offset := (page - 1) * perPage

	leaderboard, err := ch.bot.DB.GetGuildLeaderboard(i.GuildID, 100) // Get top 100
	if err != nil {
		respondEphemeral(s, i, "Failed to get leaderboard.")
		return
	}

	if len(leaderboard) == 0 {
		respondEphemeral(s, i, "No XP data yet! Start chatting to earn XP.")
		return
	}

	totalPages := (len(leaderboard) + perPage - 1) / perPage
	page = min(page, totalPages)

	start := offset
	end := offset + perPage
	if start >= len(leaderboard) {
		start = max(len(leaderboard)-perPage, 0)
	}
	end = min(end, len(leaderboard))

	var description strings.Builder
	for idx, entry := range leaderboard[start:end] {
		rank := start + idx + 1
		medal := ""
		switch rank {
		case 1:
			medal = " :first_place:"
		case 2:
			medal = " :second_place:"
		case 3:
			medal = " :third_place:"
		}
		description.WriteString(fmt.Sprintf("**#%d**%s <@%s> - Level %d (%d XP)\n",
			rank, medal, entry.UserID, entry.Level, entry.XP))
	}

	embed := &discordgo.MessageEmbed{
		Title:       "XP Leaderboard",
		Description: description.String(),
		Color:       0xFFD700,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Page %d/%d", page, totalPages),
		},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) rankHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := getUserOption(i, "user")
	if user == nil {
		user = i.Member.User
	}

	xpData, err := ch.bot.DB.GetUserXP(i.GuildID, user.ID)
	if err != nil {
		respondEphemeral(s, i, "Failed to get rank data.")
		return
	}

	rank, _ := ch.bot.DB.GetUserRank(i.GuildID, user.ID)

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s's Rank", user.Username),
		Color: 0x5865F2,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: user.AvatarURL("128"),
		},
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Rank", Value: fmt.Sprintf("#%d", rank), Inline: true},
			{Name: "Level", Value: strconv.Itoa(xpData.Level), Inline: true},
			{Name: "Total XP", Value: strconv.FormatInt(xpData.XP, 10), Inline: true},
		},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) setLevelHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to use this command.")
		return
	}

	user := getUserOption(i, "user")
	level := int(getIntOption(i, "level"))

	if user == nil {
		respondEphemeral(s, i, "Please specify a user.")
		return
	}

	// Calculate XP for this level
	xp := database.XPForLevel(level)

	err := ch.bot.DB.SetUserXP(i.GuildID, user.ID, xp, level)
	if err != nil {
		respondEphemeral(s, i, "Failed to set level.")
		return
	}

	embed := successEmbed("Level Set",
		fmt.Sprintf("Set %s's level to **%d** (%d XP)", user.Mention(), level, xp))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) setXPHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to use this command.")
		return
	}

	user := getUserOption(i, "user")
	xp := getIntOption(i, "xp")

	if user == nil {
		respondEphemeral(s, i, "Please specify a user.")
		return
	}

	level := database.CalculateLevel(xp)
	err := ch.bot.DB.SetUserXP(i.GuildID, user.ID, xp, level)
	if err != nil {
		respondEphemeral(s, i, "Failed to set XP.")
		return
	}

	embed := successEmbed("XP Set",
		fmt.Sprintf("Set %s's XP to **%d** (Level %d)", user.Mention(), xp, level))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) addXPHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to use this command.")
		return
	}

	user := getUserOption(i, "user")
	amount := getIntOption(i, "amount")

	if user == nil {
		respondEphemeral(s, i, "Please specify a user.")
		return
	}

	xpData, err := ch.bot.DB.AddUserXP(i.GuildID, user.ID, amount)
	if err != nil {
		respondEphemeral(s, i, "Failed to add XP.")
		return
	}

	embed := successEmbed("XP Added",
		fmt.Sprintf("Added **%d XP** to %s\nNew total: %d XP (Level %d)", amount, user.Mention(), xpData.XP, xpData.Level))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) massAddXPHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to use this command.")
		return
	}

	role := getRoleOption(i, "role")
	amount := getIntOption(i, "amount")

	if role == nil {
		respondEphemeral(s, i, "Please specify a role.")
		return
	}

	respondDeferred(s, i)

	// Get all members with this role
	members, err := s.GuildMembers(i.GuildID, "", 1000)
	if err != nil {
		followUp(s, i, "Failed to get server members.")
		return
	}

	count := 0
	for _, member := range members {
		for _, roleID := range member.Roles {
			if roleID == role.ID {
				_, err := ch.bot.DB.AddUserXP(i.GuildID, member.User.ID, amount)
				if err == nil {
					count++
				}
				break
			}
		}
	}

	embed := successEmbed("Mass XP Added",
		fmt.Sprintf("Added **%d XP** to **%d members** with role %s", amount, count, role.Mention()))
	followUpEmbed(s, i, embed)
}

// Helper function to create a progress bar
func createProgressBar(percent float64, length int) string {
	filled := int(percent / 100 * float64(length))
	if filled > length {
		filled = length
	}
	if filled < 0 {
		filled = 0
	}
	empty := length - filled
	return "[" + strings.Repeat("=", filled) + strings.Repeat("-", empty) + "]"
}

