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

func (ch *CommandHandler) registerSpamCommands() {
	// Spam filter configuration
	ch.Register(&Command{
		Name:        "spamfilter",
		Description: "Configure the spam filter settings",
		Category:    "Moderation",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "status",
				Description: "View current spam filter configuration",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "enable",
				Description: "Enable the spam filter",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "disable",
				Description: "Disable the spam filter",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "set",
				Description: "Configure spam filter limits",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "max_mentions",
						Description: "Maximum mentions allowed per message",
						Required:    false,
						MinValue:    floatPtr(1),
						MaxValue:    50,
					},
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "max_links",
						Description: "Maximum links allowed per message",
						Required:    false,
						MinValue:    floatPtr(1),
						MaxValue:    20,
					},
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "max_emojis",
						Description: "Maximum emojis allowed per message",
						Required:    false,
						MinValue:    floatPtr(1),
						MaxValue:    100,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "action",
						Description: "Action to take when spam is detected",
						Required:    false,
						Choices: []*discordgo.ApplicationCommandOptionChoice{
							{Name: "Delete message", Value: "delete"},
							{Name: "Warn user", Value: "warn"},
							{Name: "Kick user", Value: "kick"},
							{Name: "Ban user", Value: "ban"},
						},
					},
				},
			},
		},
		Handler: ch.spamFilterHandler,
	})
}

func (ch *CommandHandler) spamFilterHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to manage spam filter settings.")
		return
	}

	subCmd := i.ApplicationCommandData().Options[0].Name

	switch subCmd {
	case "status":
		ch.spamFilterStatusHandler(s, i)
	case "enable":
		ch.spamFilterEnableHandler(s, i)
	case "disable":
		ch.spamFilterDisableHandler(s, i)
	case "set":
		ch.spamFilterSetHandler(s, i)
	}
}

func (ch *CommandHandler) spamFilterStatusHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config, err := ch.bot.DB.GetSpamFilterConfig(i.GuildID)
	if err != nil {
		respondEphemeral(s, i, "Failed to get spam filter configuration.")
		return
	}

	status := "Disabled"
	if config.Enabled {
		status = "Enabled"
	}

	actionLabels := map[string]string{
		"delete": "Delete message",
		"warn":   "Warn user",
		"kick":   "Kick user",
		"ban":    "Ban user",
	}

	actionLabel := actionLabels[config.Action]
	if actionLabel == "" {
		actionLabel = config.Action
	}

	embed := &discordgo.MessageEmbed{
		Title: "Spam Filter Configuration",
		Color: 0xFF69B4,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Status", Value: status, Inline: true},
			{Name: "Action", Value: actionLabel, Inline: true},
			{Name: "Max Mentions", Value: fmt.Sprintf("%d", config.MaxMentions), Inline: true},
			{Name: "Max Links", Value: fmt.Sprintf("%d", config.MaxLinks), Inline: true},
			{Name: "Max Emojis", Value: fmt.Sprintf("%d", config.MaxEmojis), Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Use /spamfilter set to modify these settings",
		},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) spamFilterEnableHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config, _ := ch.bot.DB.GetSpamFilterConfig(i.GuildID)
	config.Enabled = true

	err := ch.bot.DB.SetSpamFilterConfig(config)
	if err != nil {
		respondEphemeral(s, i, "Failed to enable spam filter.")
		return
	}

	embed := successEmbed("Spam Filter Enabled", "The spam filter is now active.")
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) spamFilterDisableHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config, _ := ch.bot.DB.GetSpamFilterConfig(i.GuildID)
	config.Enabled = false

	err := ch.bot.DB.SetSpamFilterConfig(config)
	if err != nil {
		respondEphemeral(s, i, "Failed to disable spam filter.")
		return
	}

	embed := successEmbed("Spam Filter Disabled", "The spam filter has been disabled.")
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) spamFilterSetHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config, _ := ch.bot.DB.GetSpamFilterConfig(i.GuildID)

	opts := i.ApplicationCommandData().Options[0].Options
	changes := []string{}

	for _, opt := range opts {
		switch opt.Name {
		case "max_mentions":
			config.MaxMentions = int(opt.IntValue())
			changes = append(changes, fmt.Sprintf("Max mentions: %d", config.MaxMentions))
		case "max_links":
			config.MaxLinks = int(opt.IntValue())
			changes = append(changes, fmt.Sprintf("Max links: %d", config.MaxLinks))
		case "max_emojis":
			config.MaxEmojis = int(opt.IntValue())
			changes = append(changes, fmt.Sprintf("Max emojis: %d", config.MaxEmojis))
		case "action":
			config.Action = opt.StringValue()
			changes = append(changes, fmt.Sprintf("Action: %s", config.Action))
		}
	}

	if len(changes) == 0 {
		respondEphemeral(s, i, "No settings provided to update.")
		return
	}

	err := ch.bot.DB.SetSpamFilterConfig(config)
	if err != nil {
		respondEphemeral(s, i, "Failed to update spam filter settings.")
		return
	}

	description := "Updated settings:\n"
	for _, change := range changes {
		description += "- " + change + "\n"
	}

	embed := successEmbed("Spam Filter Updated", description)
	respondEmbed(s, i, embed)
}
