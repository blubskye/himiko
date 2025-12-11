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

	"github.com/blubskye/himiko/internal/database"
	"github.com/bwmarrin/discordgo"
)

func (ch *CommandHandler) registerVoiceXPCommands() {
	// Voice XP configuration
	ch.Register(&Command{
		Name:        "voicexp",
		Description: "Configure voice channel XP settings",
		Category:    "VoiceXP",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "enable",
				Description: "Enable voice XP",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "disable",
				Description: "Disable voice XP",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "rate",
				Description: "Set XP rate per interval",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "amount",
						Description: "XP to give per interval",
						Required:    true,
						MinValue:    floatPtr(1),
						MaxValue:    100,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "interval",
				Description: "Set interval between XP gains",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "minutes",
						Description: "Minutes between XP gains",
						Required:    true,
						MinValue:    floatPtr(1),
						MaxValue:    60,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "ignoreafk",
				Description: "Toggle ignoring AFK channel",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionBoolean,
						Name:        "ignore",
						Description: "Ignore AFK channel for voice XP",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "status",
				Description: "View current voice XP settings",
			},
		},
		Handler: ch.voiceXPHandler,
	})
}

func (ch *CommandHandler) voiceXPHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to configure voice XP.")
		return
	}

	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		respondEphemeral(s, i, "Please specify a subcommand.")
		return
	}

	subCmd := options[0]
	config, err := ch.bot.DB.GetVoiceXPConfig(i.GuildID)
	if err != nil {
		respondEphemeral(s, i, "Failed to get voice XP config.")
		return
	}

	switch subCmd.Name {
	case "enable":
		config.Enabled = true
		if err := ch.bot.DB.SetVoiceXPConfig(config); err != nil {
			respondEphemeral(s, i, "Failed to enable voice XP.")
			return
		}
		embed := successEmbed("Voice XP Enabled",
			"Users will now earn XP while in voice channels.")
		respondEmbed(s, i, embed)

	case "disable":
		config.Enabled = false
		if err := ch.bot.DB.SetVoiceXPConfig(config); err != nil {
			respondEphemeral(s, i, "Failed to disable voice XP.")
			return
		}
		embed := successEmbed("Voice XP Disabled",
			"Users will no longer earn XP in voice channels.")
		respondEmbed(s, i, embed)

	case "rate":
		amount := int(subCmd.Options[0].IntValue())
		config.XPRate = amount
		if err := ch.bot.DB.SetVoiceXPConfig(config); err != nil {
			respondEphemeral(s, i, "Failed to update XP rate.")
			return
		}
		embed := successEmbed("Voice XP Rate Updated",
			fmt.Sprintf("Users will earn **%d XP** per interval in voice channels.", amount))
		respondEmbed(s, i, embed)

	case "interval":
		minutes := int(subCmd.Options[0].IntValue())
		config.IntervalMins = minutes
		if err := ch.bot.DB.SetVoiceXPConfig(config); err != nil {
			respondEphemeral(s, i, "Failed to update interval.")
			return
		}
		embed := successEmbed("Voice XP Interval Updated",
			fmt.Sprintf("Users will earn XP every **%d minutes** in voice channels.", minutes))
		respondEmbed(s, i, embed)

	case "ignoreafk":
		ignore := subCmd.Options[0].BoolValue()
		config.IgnoreAFK = ignore
		if err := ch.bot.DB.SetVoiceXPConfig(config); err != nil {
			respondEphemeral(s, i, "Failed to update AFK setting.")
			return
		}
		status := "will be ignored"
		if !ignore {
			status = "will earn XP too"
		}
		embed := successEmbed("AFK Setting Updated",
			fmt.Sprintf("Users in AFK channel %s.", status))
		respondEmbed(s, i, embed)

	case "status":
		ch.voiceXPStatus(s, i, config)
	}
}

func (ch *CommandHandler) voiceXPStatus(s *discordgo.Session, i *discordgo.InteractionCreate, config *database.VoiceXPConfig) {
	statusEmoji := func(enabled bool) string {
		if enabled {
			return ":white_check_mark:"
		}
		return ":x:"
	}

	embed := &discordgo.MessageEmbed{
		Title: "Voice XP Configuration",
		Color: 0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Enabled", Value: statusEmoji(config.Enabled), Inline: true},
			{Name: "XP Rate", Value: fmt.Sprintf("%d XP", config.XPRate), Inline: true},
			{Name: "Interval", Value: fmt.Sprintf("%d minutes", config.IntervalMins), Inline: true},
			{Name: "Ignore AFK", Value: statusEmoji(config.IgnoreAFK), Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Users earn %d XP every %d minutes in voice", config.XPRate, config.IntervalMins),
		},
	}

	respondEmbed(s, i, embed)
}
