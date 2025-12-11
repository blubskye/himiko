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

func (ch *CommandHandler) registerLoggingCommands() {
	// Set log channel
	ch.Register(&Command{
		Name:        "setlogchannel",
		Description: "Set the channel for server logs",
		Category:    "Logging",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "Channel to send logs to",
				Required:    true,
			},
		},
		Handler: ch.setLogChannelHandler,
	})

	// Toggle logging
	ch.Register(&Command{
		Name:        "togglelogging",
		Description: "Enable or disable logging",
		Category:    "Logging",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "enabled",
				Description: "Enable or disable logging",
				Required:    true,
			},
		},
		Handler: ch.toggleLoggingHandler,
	})

	// Configure log types
	ch.Register(&Command{
		Name:        "logconfig",
		Description: "Configure which events to log",
		Category:    "Logging",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "type",
				Description: "Log type to configure",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "Message Delete", Value: "message_delete"},
					{Name: "Message Edit", Value: "message_edit"},
					{Name: "Voice Join", Value: "voice_join"},
					{Name: "Voice Leave", Value: "voice_leave"},
					{Name: "Nickname Change", Value: "nickname"},
					{Name: "Avatar Change", Value: "avatar"},
					{Name: "Presence Change", Value: "presence"},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "enabled",
				Description: "Enable or disable this log type",
				Required:    true,
			},
		},
		Handler: ch.logConfigHandler,
	})

	// Disable logging for channel
	ch.Register(&Command{
		Name:        "disablechannellog",
		Description: "Disable logging for a specific channel",
		Category:    "Logging",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "Channel to disable logging for",
				Required:    true,
			},
		},
		Handler: ch.disableChannelLogHandler,
	})

	// Enable logging for channel
	ch.Register(&Command{
		Name:        "enablechannellog",
		Description: "Re-enable logging for a channel",
		Category:    "Logging",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "Channel to re-enable logging for",
				Required:    true,
			},
		},
		Handler: ch.enableChannelLogHandler,
	})

	// Log status
	ch.Register(&Command{
		Name:        "logstatus",
		Description: "View current logging configuration",
		Category:    "Logging",
		Handler:     ch.logStatusHandler,
	})
}

func (ch *CommandHandler) setLogChannelHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to configure logging.")
		return
	}

	channel := getChannelOption(i, "channel")
	if channel == nil {
		respondEphemeral(s, i, "Please specify a channel.")
		return
	}

	err := ch.bot.DB.SetLogChannel(i.GuildID, channel.ID)
	if err != nil {
		respondEphemeral(s, i, "Failed to set log channel.")
		return
	}

	embed := successEmbed("Log Channel Set",
		fmt.Sprintf("Server logs will be sent to <#%s>", channel.ID))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) toggleLoggingHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to configure logging.")
		return
	}

	enabled := getBoolOption(i, "enabled")

	err := ch.bot.DB.ToggleLogging(i.GuildID, enabled)
	if err != nil {
		respondEphemeral(s, i, "Failed to update logging setting.")
		return
	}

	status := "disabled"
	if enabled {
		status = "enabled"
	}

	embed := successEmbed("Logging Updated",
		fmt.Sprintf("Logging has been **%s**", status))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) logConfigHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to configure logging.")
		return
	}

	logType := getStringOption(i, "type")
	enabled := getBoolOption(i, "enabled")

	config, err := ch.bot.DB.GetLoggingConfig(i.GuildID)
	if err != nil {
		respondEphemeral(s, i, "Failed to get logging config.")
		return
	}

	switch logType {
	case "message_delete":
		config.MessageDelete = enabled
	case "message_edit":
		config.MessageEdit = enabled
	case "voice_join":
		config.VoiceJoin = enabled
	case "voice_leave":
		config.VoiceLeave = enabled
	case "nickname":
		config.NicknameChange = enabled
	case "avatar":
		config.AvatarChange = enabled
	case "presence":
		config.PresenceChange = enabled
	}

	err = ch.bot.DB.SetLoggingConfig(config)
	if err != nil {
		respondEphemeral(s, i, "Failed to update logging config.")
		return
	}

	status := "disabled"
	if enabled {
		status = "enabled"
	}

	typeNames := map[string]string{
		"message_delete": "Message Delete",
		"message_edit":   "Message Edit",
		"voice_join":     "Voice Join",
		"voice_leave":    "Voice Leave",
		"nickname":       "Nickname Change",
		"avatar":         "Avatar Change",
		"presence":       "Presence Change",
	}

	embed := successEmbed("Log Config Updated",
		fmt.Sprintf("**%s** logging has been **%s**", typeNames[logType], status))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) disableChannelLogHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to configure logging.")
		return
	}

	channel := getChannelOption(i, "channel")
	if channel == nil {
		respondEphemeral(s, i, "Please specify a channel.")
		return
	}

	err := ch.bot.DB.AddDisabledLogChannel(i.GuildID, channel.ID)
	if err != nil {
		respondEphemeral(s, i, "Failed to disable logging for channel.")
		return
	}

	embed := successEmbed("Channel Logging Disabled",
		fmt.Sprintf("Logging is now disabled for <#%s>", channel.ID))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) enableChannelLogHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to configure logging.")
		return
	}

	channel := getChannelOption(i, "channel")
	if channel == nil {
		respondEphemeral(s, i, "Please specify a channel.")
		return
	}

	err := ch.bot.DB.RemoveDisabledLogChannel(i.GuildID, channel.ID)
	if err != nil {
		respondEphemeral(s, i, "Failed to enable logging for channel.")
		return
	}

	embed := successEmbed("Channel Logging Enabled",
		fmt.Sprintf("Logging is now enabled for <#%s>", channel.ID))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) logStatusHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config, err := ch.bot.DB.GetLoggingConfig(i.GuildID)
	if err != nil {
		respondEphemeral(s, i, "Failed to get logging config.")
		return
	}

	disabledChannels, _ := ch.bot.DB.GetDisabledLogChannels(i.GuildID)

	statusEmoji := func(enabled bool) string {
		if enabled {
			return ":white_check_mark:"
		}
		return ":x:"
	}

	logChannel := "Not set"
	if config.LogChannelID != nil {
		logChannel = fmt.Sprintf("<#%s>", *config.LogChannelID)
	}

	var disabledList string
	if len(disabledChannels) > 0 {
		channels := make([]string, len(disabledChannels))
		for i, ch := range disabledChannels {
			channels[i] = fmt.Sprintf("<#%s>", ch)
		}
		disabledList = strings.Join(channels, ", ")
	} else {
		disabledList = "None"
	}

	embed := &discordgo.MessageEmbed{
		Title: "Logging Configuration",
		Color: 0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Log Channel", Value: logChannel, Inline: true},
			{Name: "Enabled", Value: statusEmoji(config.Enabled), Inline: true},
			{Name: "\u200b", Value: "\u200b", Inline: true},
			{Name: "Message Delete", Value: statusEmoji(config.MessageDelete), Inline: true},
			{Name: "Message Edit", Value: statusEmoji(config.MessageEdit), Inline: true},
			{Name: "Voice Join", Value: statusEmoji(config.VoiceJoin), Inline: true},
			{Name: "Voice Leave", Value: statusEmoji(config.VoiceLeave), Inline: true},
			{Name: "Nickname Change", Value: statusEmoji(config.NicknameChange), Inline: true},
			{Name: "Avatar Change", Value: statusEmoji(config.AvatarChange), Inline: true},
			{Name: "Presence Change", Value: statusEmoji(config.PresenceChange), Inline: true},
			{Name: "Disabled Channels", Value: disabledList, Inline: false},
		},
	}

	respondEmbed(s, i, embed)
}
