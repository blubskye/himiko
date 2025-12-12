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

func (ch *CommandHandler) registerDMCommands() {
	// Set DM channel
	ch.Register(&Command{
		Name:        "setdmchannel",
		Description: "Set a channel to forward DMs to",
		Category:    "DM",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "Channel to forward DMs to",
				Required:    true,
			},
		},
		Handler: ch.setDMChannelHandler,
	})

	// Disable DM forwarding
	ch.Register(&Command{
		Name:        "disabledm",
		Description: "Disable DM forwarding for this server",
		Category:    "DM",
		Handler:     ch.disableDMHandler,
	})

	// DM status
	ch.Register(&Command{
		Name:        "dmstatus",
		Description: "View DM forwarding status",
		Category:    "DM",
		Handler:     ch.dmStatusHandler,
	})
}

func (ch *CommandHandler) setDMChannelHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !ch.bot.Config.IsOwner(i.Member.User.ID) {
		respondEphemeral(s, i, "Only bot owners can configure DM forwarding.")
		return
	}

	channel := getChannelOption(i, "channel")
	if channel == nil {
		respondEphemeral(s, i, "Please specify a channel.")
		return
	}

	err := ch.bot.DB.SetDMConfig(i.GuildID, channel.ID, true)
	if err != nil {
		respondEphemeral(s, i, "Failed to set DM channel.")
		return
	}

	embed := successEmbed("DM Forwarding Enabled",
		fmt.Sprintf("DMs will be forwarded to <#%s>", channel.ID))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) disableDMHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !ch.bot.Config.IsOwner(i.Member.User.ID) {
		respondEphemeral(s, i, "Only bot owners can configure DM forwarding.")
		return
	}

	config, err := ch.bot.DB.GetDMConfig(i.GuildID)
	if err != nil || config == nil {
		respondEphemeral(s, i, "DM forwarding is not configured for this server.")
		return
	}

	err = ch.bot.DB.SetDMConfig(i.GuildID, config.ChannelID, false)
	if err != nil {
		respondEphemeral(s, i, "Failed to disable DM forwarding.")
		return
	}

	embed := successEmbed("DM Forwarding Disabled",
		"DMs will no longer be forwarded to this server.")
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) dmStatusHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config, err := ch.bot.DB.GetDMConfig(i.GuildID)
	if err != nil || config == nil {
		respondEphemeral(s, i, "DM forwarding is not configured for this server.")
		return
	}

	status := ":x: Disabled"
	if config.Enabled {
		status = ":white_check_mark: Enabled"
	}

	embed := &discordgo.MessageEmbed{
		Title: "DM Forwarding Status",
		Color: 0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Status", Value: status, Inline: true},
			{Name: "Channel", Value: fmt.Sprintf("<#%s>", config.ChannelID), Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "DMs sent to the bot will be forwarded to this channel",
		},
	}

	respondEmbed(s, i, embed)
}

