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

func (ch *CommandHandler) registerAutoCleanCommands() {
	// Auto-clean management
	ch.Register(&Command{
		Name:        "autoclean",
		Description: "Manage auto-clean channels",
		Category:    "AutoClean",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add",
				Description: "Add a channel to auto-clean",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionChannel,
						Name:        "channel",
						Description: "Channel to auto-clean",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "interval",
						Description: "Interval in hours (default: 24)",
						Required:    false,
						MinValue:    floatPtr(1),
						MaxValue:    168, // 1 week
					},
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "warning",
						Description: "Warning before clean in minutes (default: 5)",
						Required:    false,
						MinValue:    floatPtr(1),
						MaxValue:    60,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remove",
				Description: "Remove a channel from auto-clean",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionChannel,
						Name:        "channel",
						Description: "Channel to remove",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "list",
				Description: "List all auto-clean channels",
			},
		},
		Handler: ch.autoCleanHandler,
	})

	// Set clean message (whether to post warning)
	ch.Register(&Command{
		Name:        "setcleanmessage",
		Description: "Toggle warning message before auto-clean",
		Category:    "AutoClean",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "Channel to configure",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "enabled",
				Description: "Enable or disable warning message",
				Required:    true,
			},
		},
		Handler: ch.setCleanMessageHandler,
	})

	// Set clean image (whether to preserve images)
	ch.Register(&Command{
		Name:        "setcleanimage",
		Description: "Toggle whether to preserve images during clean",
		Category:    "AutoClean",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "Channel to configure",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "preserve",
				Description: "Preserve images (true) or delete all (false)",
				Required:    true,
			},
		},
		Handler: ch.setCleanImageHandler,
	})
}

func (ch *CommandHandler) autoCleanHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to manage auto-clean.")
		return
	}

	// Get subcommand
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		respondEphemeral(s, i, "Please specify a subcommand.")
		return
	}

	subCmd := options[0]
	switch subCmd.Name {
	case "add":
		ch.autoCleanAdd(s, i, subCmd.Options)
	case "remove":
		ch.autoCleanRemove(s, i, subCmd.Options)
	case "list":
		ch.autoCleanList(s, i)
	}
}

func (ch *CommandHandler) autoCleanAdd(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
	var channelID string
	interval := 24
	warning := 5

	for _, opt := range options {
		switch opt.Name {
		case "channel":
			channelID = opt.ChannelValue(s).ID
		case "interval":
			interval = int(opt.IntValue())
		case "warning":
			warning = int(opt.IntValue())
		}
	}

	if channelID == "" {
		respondEphemeral(s, i, "Please specify a channel.")
		return
	}

	err := ch.bot.DB.AddAutoCleanChannel(i.GuildID, channelID, i.Member.User.ID, interval, warning)
	if err != nil {
		respondEphemeral(s, i, "Failed to add auto-clean channel.")
		return
	}

	embed := successEmbed("Auto-Clean Added",
		fmt.Sprintf("<#%s> will be cleaned every **%d hours** with a **%d minute** warning.",
			channelID, interval, warning))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) autoCleanRemove(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
	var channelID string
	for _, opt := range options {
		if opt.Name == "channel" {
			channelID = opt.ChannelValue(s).ID
		}
	}

	if channelID == "" {
		respondEphemeral(s, i, "Please specify a channel.")
		return
	}

	err := ch.bot.DB.RemoveAutoCleanChannel(i.GuildID, channelID)
	if err != nil {
		respondEphemeral(s, i, "Failed to remove auto-clean channel.")
		return
	}

	embed := successEmbed("Auto-Clean Removed",
		fmt.Sprintf("<#%s> has been removed from auto-clean.", channelID))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) autoCleanList(s *discordgo.Session, i *discordgo.InteractionCreate) {
	channels, err := ch.bot.DB.GetAutoCleanChannels(i.GuildID)
	if err != nil {
		respondEphemeral(s, i, "Failed to get auto-clean channels.")
		return
	}

	if len(channels) == 0 {
		respondEphemeral(s, i, "No auto-clean channels configured.")
		return
	}

	var description strings.Builder
	for _, c := range channels {
		description.WriteString(fmt.Sprintf("<#%s>\n", c.ChannelID))
		description.WriteString(fmt.Sprintf("├ Interval: %d hours\n", c.IntervalHours))
		description.WriteString(fmt.Sprintf("├ Warning: %d minutes\n", c.WarningMinutes))
		description.WriteString(fmt.Sprintf("├ Next run: <t:%d:R>\n", c.NextRun.Unix()))

		msgStatus := ":x:"
		if c.CleanMessage {
			msgStatus = ":white_check_mark:"
		}
		imgStatus := ":x:"
		if c.CleanImage {
			imgStatus = ":white_check_mark:"
		}
		description.WriteString(fmt.Sprintf("└ Warning msg: %s | Clean images: %s\n\n", msgStatus, imgStatus))
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Auto-Clean Channels (%d)", len(channels)),
		Description: description.String(),
		Color:       0x5865F2,
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) setCleanMessageHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to configure auto-clean.")
		return
	}

	channel := getChannelOption(i, "channel")
	enabled := getBoolOption(i, "enabled")

	if channel == nil {
		respondEphemeral(s, i, "Please specify a channel.")
		return
	}

	err := ch.bot.DB.SetAutoCleanMessage(i.GuildID, channel.ID, enabled)
	if err != nil {
		respondEphemeral(s, i, "Failed to update setting.")
		return
	}

	status := "disabled"
	if enabled {
		status = "enabled"
	}

	embed := successEmbed("Setting Updated",
		fmt.Sprintf("Warning messages %s for <#%s>", status, channel.ID))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) setCleanImageHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to configure auto-clean.")
		return
	}

	channel := getChannelOption(i, "channel")
	preserve := getBoolOption(i, "preserve")

	if channel == nil {
		respondEphemeral(s, i, "Please specify a channel.")
		return
	}

	// Note: CleanImage = true means images ARE deleted, preserve = true means DON'T delete
	err := ch.bot.DB.SetAutoCleanImage(i.GuildID, channel.ID, !preserve)
	if err != nil {
		respondEphemeral(s, i, "Failed to update setting.")
		return
	}

	status := "will be deleted"
	if preserve {
		status = "will be preserved"
	}

	embed := successEmbed("Setting Updated",
		fmt.Sprintf("Images in <#%s> %s during clean", channel.ID, status))
	respondEmbed(s, i, embed)
}
