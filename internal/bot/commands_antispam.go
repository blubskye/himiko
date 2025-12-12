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

func (ch *CommandHandler) registerAntiSpamCommands() {
	// Anti-spam configuration (pressure-based)
	ch.Register(&Command{
		Name:        "antispam",
		Description: "Configure pressure-based anti-spam protection",
		Category:    "Anti-Spam",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "status",
				Description: "View anti-spam configuration",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "enable",
				Description: "Enable anti-spam protection",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "disable",
				Description: "Disable anti-spam protection",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "set",
				Description: "Configure anti-spam thresholds",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionNumber,
						Name:        "max_pressure",
						Description: "Maximum pressure before action (default: 60)",
						Required:    false,
						MinValue:    floatPtr(10),
						MaxValue:    200,
					},
					{
						Type:        discordgo.ApplicationCommandOptionNumber,
						Name:        "base_pressure",
						Description: "Base pressure per message (default: 10)",
						Required:    false,
						MinValue:    floatPtr(1),
						MaxValue:    50,
					},
					{
						Type:        discordgo.ApplicationCommandOptionNumber,
						Name:        "decay",
						Description: "Seconds to decay base pressure (default: 2.5)",
						Required:    false,
						MinValue:    floatPtr(0.5),
						MaxValue:    30,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "action",
						Description: "Action when threshold exceeded",
						Required:    false,
						Choices: []*discordgo.ApplicationCommandOptionChoice{
							{Name: "Delete message", Value: "delete"},
							{Name: "Warn user", Value: "warn"},
							{Name: "Silence user", Value: "silence"},
							{Name: "Kick user", Value: "kick"},
							{Name: "Ban user", Value: "ban"},
						},
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "penalties",
				Description: "Configure spam penalties",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionNumber,
						Name:        "image",
						Description: "Pressure per image/embed (default: 8.33)",
						Required:    false,
						MinValue:    floatPtr(0),
						MaxValue:    50,
					},
					{
						Type:        discordgo.ApplicationCommandOptionNumber,
						Name:        "link",
						Description: "Pressure per link (default: 8.33)",
						Required:    false,
						MinValue:    floatPtr(0),
						MaxValue:    50,
					},
					{
						Type:        discordgo.ApplicationCommandOptionNumber,
						Name:        "ping",
						Description: "Pressure per mention (default: 2.5)",
						Required:    false,
						MinValue:    floatPtr(0),
						MaxValue:    20,
					},
					{
						Type:        discordgo.ApplicationCommandOptionNumber,
						Name:        "repeat",
						Description: "Extra pressure for repeated message (default: 10)",
						Required:    false,
						MinValue:    floatPtr(0),
						MaxValue:    50,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "setrole",
				Description: "Set the silent role for spammers",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionRole,
						Name:        "role",
						Description: "Role to assign to silenced users",
						Required:    true,
					},
				},
			},
		},
		Handler: ch.antiSpamHandler,
	})
}

func (ch *CommandHandler) antiSpamHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to configure anti-spam.")
		return
	}

	subCmd := i.ApplicationCommandData().Options[0].Name

	switch subCmd {
	case "status":
		ch.antiSpamStatusHandler(s, i)
	case "enable":
		ch.antiSpamEnableHandler(s, i)
	case "disable":
		ch.antiSpamDisableHandler(s, i)
	case "set":
		ch.antiSpamSetHandler(s, i)
	case "penalties":
		ch.antiSpamPenaltiesHandler(s, i)
	case "setrole":
		ch.antiSpamSetRoleHandler(s, i)
	}
}

func (ch *CommandHandler) antiSpamStatusHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cfg, err := ch.bot.DB.GetAntiSpamConfig(i.GuildID)
	if err != nil {
		respondEphemeral(s, i, "Failed to get anti-spam configuration.")
		return
	}

	status := "Disabled"
	if cfg.Enabled {
		status = "Enabled"
	}

	silentRole := "Not set"
	if cfg.SilentRoleID != "" {
		silentRole = fmt.Sprintf("<@&%s>", cfg.SilentRoleID)
	}

	embed := &discordgo.MessageEmbed{
		Title: "Anti-Spam Configuration (Pressure System)",
		Color: 0xFF69B4,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Status", Value: status, Inline: true},
			{Name: "Action", Value: cfg.Action, Inline: true},
			{Name: "Silent Role", Value: silentRole, Inline: true},
			{Name: "Max Pressure", Value: fmt.Sprintf("%.1f", cfg.MaxPressure), Inline: true},
			{Name: "Base Pressure", Value: fmt.Sprintf("%.1f", cfg.BasePressure), Inline: true},
			{Name: "Decay Rate", Value: fmt.Sprintf("%.1fs", cfg.PressureDecay), Inline: true},
			{Name: "Image Penalty", Value: fmt.Sprintf("%.2f", cfg.ImagePressure), Inline: true},
			{Name: "Link Penalty", Value: fmt.Sprintf("%.2f", cfg.LinkPressure), Inline: true},
			{Name: "Ping Penalty", Value: fmt.Sprintf("%.2f", cfg.PingPressure), Inline: true},
			{Name: "Repeat Penalty", Value: fmt.Sprintf("%.2f", cfg.RepeatPressure), Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Pressure accumulates from rapid messaging and decays over time",
		},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) antiSpamEnableHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cfg, _ := ch.bot.DB.GetAntiSpamConfig(i.GuildID)
	cfg.Enabled = true

	if err := ch.bot.DB.SetAntiSpamConfig(cfg); err != nil {
		respondEphemeral(s, i, "Failed to enable anti-spam.")
		return
	}

	embed := successEmbed("Anti-Spam Enabled", "Pressure-based spam detection is now active.")
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) antiSpamDisableHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cfg, _ := ch.bot.DB.GetAntiSpamConfig(i.GuildID)
	cfg.Enabled = false

	if err := ch.bot.DB.SetAntiSpamConfig(cfg); err != nil {
		respondEphemeral(s, i, "Failed to disable anti-spam.")
		return
	}

	embed := successEmbed("Anti-Spam Disabled", "Pressure-based spam detection has been disabled.")
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) antiSpamSetHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cfg, _ := ch.bot.DB.GetAntiSpamConfig(i.GuildID)
	opts := i.ApplicationCommandData().Options[0].Options

	changes := []string{}

	for _, opt := range opts {
		switch opt.Name {
		case "max_pressure":
			cfg.MaxPressure = opt.FloatValue()
			changes = append(changes, fmt.Sprintf("Max pressure: %.1f", cfg.MaxPressure))
		case "base_pressure":
			cfg.BasePressure = opt.FloatValue()
			changes = append(changes, fmt.Sprintf("Base pressure: %.1f", cfg.BasePressure))
		case "decay":
			cfg.PressureDecay = opt.FloatValue()
			changes = append(changes, fmt.Sprintf("Decay rate: %.1fs", cfg.PressureDecay))
		case "action":
			cfg.Action = opt.StringValue()
			changes = append(changes, fmt.Sprintf("Action: %s", cfg.Action))
		}
	}

	if len(changes) == 0 {
		respondEphemeral(s, i, "No settings provided.")
		return
	}

	if err := ch.bot.DB.SetAntiSpamConfig(cfg); err != nil {
		respondEphemeral(s, i, "Failed to update settings.")
		return
	}

	description := "Updated settings:\n"
	for _, change := range changes {
		description += "- " + change + "\n"
	}

	embed := successEmbed("Anti-Spam Updated", description)
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) antiSpamPenaltiesHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cfg, _ := ch.bot.DB.GetAntiSpamConfig(i.GuildID)
	opts := i.ApplicationCommandData().Options[0].Options

	changes := []string{}

	for _, opt := range opts {
		switch opt.Name {
		case "image":
			cfg.ImagePressure = opt.FloatValue()
			changes = append(changes, fmt.Sprintf("Image penalty: %.2f", cfg.ImagePressure))
		case "link":
			cfg.LinkPressure = opt.FloatValue()
			changes = append(changes, fmt.Sprintf("Link penalty: %.2f", cfg.LinkPressure))
		case "ping":
			cfg.PingPressure = opt.FloatValue()
			changes = append(changes, fmt.Sprintf("Ping penalty: %.2f", cfg.PingPressure))
		case "repeat":
			cfg.RepeatPressure = opt.FloatValue()
			changes = append(changes, fmt.Sprintf("Repeat penalty: %.2f", cfg.RepeatPressure))
		}
	}

	if len(changes) == 0 {
		respondEphemeral(s, i, "No penalties provided.")
		return
	}

	if err := ch.bot.DB.SetAntiSpamConfig(cfg); err != nil {
		respondEphemeral(s, i, "Failed to update penalties.")
		return
	}

	description := "Updated penalties:\n"
	for _, change := range changes {
		description += "- " + change + "\n"
	}

	embed := successEmbed("Spam Penalties Updated", description)
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) antiSpamSetRoleHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	role := getRoleOption(i, "role")
	if role == nil {
		respondEphemeral(s, i, "Please specify a role.")
		return
	}

	cfg, _ := ch.bot.DB.GetAntiSpamConfig(i.GuildID)
	cfg.SilentRoleID = role.ID

	if err := ch.bot.DB.SetAntiSpamConfig(cfg); err != nil {
		respondEphemeral(s, i, "Failed to set silent role.")
		return
	}

	embed := successEmbed("Silent Role Set", fmt.Sprintf("Silent role for spammers set to <@&%s>", role.ID))
	respondEmbed(s, i, embed)
}
