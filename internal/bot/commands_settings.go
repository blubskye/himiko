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

	"github.com/bwmarrin/discordgo"
)

func (ch *CommandHandler) registerSettingsCommands() {
	// Set prefix (for message commands if implemented)
	ch.Register(&Command{
		Name:        "setprefix",
		Description: "Set the bot prefix for this server",
		Category:    "Settings",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "prefix",
				Description: "New prefix",
				Required:    true,
			},
		},
		Handler: ch.setPrefixHandler,
	})

	// Set mod log channel
	ch.Register(&Command{
		Name:        "setmodlog",
		Description: "Set the moderation log channel",
		Category:    "Settings",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "Mod log channel",
				Required:    true,
			},
		},
		Handler: ch.setModLogHandler,
	})

	// Set welcome channel and message
	ch.Register(&Command{
		Name:        "setwelcome",
		Description: "Configure welcome messages",
		Category:    "Settings",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "Welcome channel",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "message",
				Description: "Welcome message ({user}, {username}, {server} placeholders)",
				Required:    true,
			},
		},
		Handler: ch.setWelcomeHandler,
	})

	// Disable welcome
	ch.Register(&Command{
		Name:        "disablewelcome",
		Description: "Disable welcome messages",
		Category:    "Settings",
		Handler:     ch.disableWelcomeHandler,
	})

	// View settings
	ch.Register(&Command{
		Name:        "settings",
		Description: "View server settings",
		Category:    "Settings",
		Handler:     ch.viewSettingsHandler,
	})
}

func (ch *CommandHandler) setPrefixHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to change settings.")
		return
	}

	prefix := getStringOption(i, "prefix")

	settings, _ := ch.bot.DB.GetGuildSettings(i.GuildID)
	settings.Prefix = prefix

	err := ch.bot.DB.SetGuildSettings(settings)
	if err != nil {
		respondEphemeral(s, i, "Failed to update prefix.")
		return
	}

	embed := successEmbed("Prefix Updated",
		fmt.Sprintf("The bot prefix has been set to `%s`", prefix))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) setModLogHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to change settings.")
		return
	}

	channel := getChannelOption(i, "channel")
	if channel == nil {
		respondEphemeral(s, i, "Please specify a channel.")
		return
	}

	settings, _ := ch.bot.DB.GetGuildSettings(i.GuildID)
	settings.ModLogChannel = &channel.ID

	err := ch.bot.DB.SetGuildSettings(settings)
	if err != nil {
		respondEphemeral(s, i, "Failed to update mod log channel.")
		return
	}

	embed := successEmbed("Mod Log Channel Set",
		fmt.Sprintf("Moderation logs will be sent to <#%s>", channel.ID))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) setWelcomeHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to change settings.")
		return
	}

	channel := getChannelOption(i, "channel")
	message := getStringOption(i, "message")

	if channel == nil {
		respondEphemeral(s, i, "Please specify a channel.")
		return
	}

	settings, _ := ch.bot.DB.GetGuildSettings(i.GuildID)
	settings.WelcomeChannel = &channel.ID
	settings.WelcomeMessage = &message

	err := ch.bot.DB.SetGuildSettings(settings)
	if err != nil {
		respondEphemeral(s, i, "Failed to update welcome settings.")
		return
	}

	embed := successEmbed("Welcome Message Configured",
		fmt.Sprintf("Welcome messages will be sent to <#%s>\n\n**Preview:**\n%s",
			channel.ID, replacePlaceholders(message, i.Member.User, i.GuildID)))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) disableWelcomeHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to change settings.")
		return
	}

	settings, _ := ch.bot.DB.GetGuildSettings(i.GuildID)
	settings.WelcomeChannel = nil
	settings.WelcomeMessage = nil

	err := ch.bot.DB.SetGuildSettings(settings)
	if err != nil {
		respondEphemeral(s, i, "Failed to update settings.")
		return
	}

	embed := successEmbed("Welcome Messages Disabled",
		"Welcome messages have been disabled for this server.")
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) viewSettingsHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	settings, err := ch.bot.DB.GetGuildSettings(i.GuildID)
	if err != nil {
		respondEphemeral(s, i, "Failed to fetch settings.")
		return
	}

	modLog := "Not set"
	if settings.ModLogChannel != nil {
		modLog = fmt.Sprintf("<#%s>", *settings.ModLogChannel)
	}

	welcomeChannel := "Disabled"
	welcomeMessage := "N/A"
	if settings.WelcomeChannel != nil {
		welcomeChannel = fmt.Sprintf("<#%s>", *settings.WelcomeChannel)
		if settings.WelcomeMessage != nil {
			welcomeMessage = truncate(*settings.WelcomeMessage, 100)
		}
	}

	embed := &discordgo.MessageEmbed{
		Title: "Server Settings",
		Color: 0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Prefix", Value: fmt.Sprintf("`%s`", settings.Prefix), Inline: true},
			{Name: "Mod Log Channel", Value: modLog, Inline: true},
			{Name: "Welcome Channel", Value: welcomeChannel, Inline: true},
			{Name: "Welcome Message", Value: welcomeMessage, Inline: false},
		},
	}

	respondEmbed(s, i, embed)
}
