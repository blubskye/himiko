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

	"github.com/bwmarrin/discordgo"
)

func (ch *CommandHandler) registerBotBanCommands() {
	// Bot ban
	ch.Register(&Command{
		Name:        "botban",
		Description: "Ban a user or server from using the bot (owner only)",
		Category:    "BotBan",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "type",
				Description: "Type of ban",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "User", Value: "user"},
					{Name: "Server", Value: "server"},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "id",
				Description: "User or Server ID to ban",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "Reason for the ban",
				Required:    false,
			},
		},
		Handler: ch.botBanHandler,
	})

	// Bot unban
	ch.Register(&Command{
		Name:        "botunban",
		Description: "Remove a bot-level ban (owner only)",
		Category:    "BotBan",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "id",
				Description: "User or Server ID to unban",
				Required:    true,
			},
		},
		Handler: ch.botUnbanHandler,
	})

	// Bot ban list
	ch.Register(&Command{
		Name:        "botbanlist",
		Description: "List all bot-level bans (owner only)",
		Category:    "BotBan",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "type",
				Description: "Filter by type",
				Required:    false,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "Users", Value: "user"},
					{Name: "Servers", Value: "server"},
				},
			},
		},
		Handler: ch.botBanListHandler,
	})
}

func (ch *CommandHandler) botBanHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isOwner(ch.bot.Config.OwnerID, i.Member.User.ID) {
		respondEphemeral(s, i, "Only the bot owner can use bot-level bans.")
		return
	}

	banType := getStringOption(i, "type")
	targetID := getStringOption(i, "id")
	reason := getStringOption(i, "reason")

	if reason == "" {
		reason = "No reason provided"
	}

	err := ch.bot.DB.AddBotBan(targetID, banType, reason, i.Member.User.ID)
	if err != nil {
		respondEphemeral(s, i, "Failed to add bot ban.")
		return
	}

	emoji := ":bust_in_silhouette:"
	if banType == "server" {
		emoji = ":homes:"
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s Bot Ban Added", emoji),
		Description: fmt.Sprintf("**Type:** %s\n**ID:** `%s`\n**Reason:** %s", banType, targetID, reason),
		Color:       0xFF0000,
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) botUnbanHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isOwner(ch.bot.Config.OwnerID, i.Member.User.ID) {
		respondEphemeral(s, i, "Only the bot owner can manage bot-level bans.")
		return
	}

	targetID := getStringOption(i, "id")

	err := ch.bot.DB.RemoveBotBan(targetID)
	if err != nil {
		respondEphemeral(s, i, "Failed to remove bot ban.")
		return
	}

	embed := successEmbed("Bot Ban Removed",
		fmt.Sprintf("Removed bot-level ban for `%s`", targetID))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) botBanListHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isOwner(ch.bot.Config.OwnerID, i.Member.User.ID) {
		respondEphemeral(s, i, "Only the bot owner can view bot-level bans.")
		return
	}

	filterType := getStringOption(i, "type")

	bans, err := ch.bot.DB.GetBotBans(filterType)
	if err != nil {
		respondEphemeral(s, i, "Failed to get bot bans.")
		return
	}

	if len(bans) == 0 {
		respondEphemeral(s, i, "No bot-level bans found.")
		return
	}

	var userBans, serverBans []string
	for _, ban := range bans {
		entry := fmt.Sprintf("`%s` - %s", ban.TargetID, truncate(ban.Reason, 50))
		if ban.BanType == "user" {
			userBans = append(userBans, entry)
		} else {
			serverBans = append(serverBans, entry)
		}
	}

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("Bot-Level Bans (%d)", len(bans)),
		Color: 0xFF0000,
	}

	if len(userBans) > 0 && (filterType == "" || filterType == "user") {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf(":bust_in_silhouette: Users (%d)", len(userBans)),
			Value: truncate(strings.Join(userBans, "\n"), 1024),
		})
	}

	if len(serverBans) > 0 && (filterType == "" || filterType == "server") {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf(":homes: Servers (%d)", len(serverBans)),
			Value: truncate(strings.Join(serverBans, "\n"), 1024),
		})
	}

	respondEmbedEphemeral(s, i, embed)
}
