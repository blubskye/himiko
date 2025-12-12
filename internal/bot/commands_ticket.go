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

func (ch *CommandHandler) registerTicketCommands() {
	// Set ticket channel
	ch.Register(&Command{
		Name:        "setticket",
		Description: "Set the channel where tickets will be forwarded",
		Category:    "Ticket",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "Channel to receive tickets",
				Required:    true,
			},
		},
		Handler: ch.setTicketHandler,
	})

	// Disable ticket system
	ch.Register(&Command{
		Name:        "disableticket",
		Description: "Disable the ticket system",
		Category:    "Ticket",
		Handler:     ch.disableTicketHandler,
	})

	// View ticket status
	ch.Register(&Command{
		Name:        "ticketstatus",
		Description: "View ticket system status",
		Category:    "Ticket",
		Handler:     ch.ticketStatusHandler,
	})

	// Submit a ticket
	ch.Register(&Command{
		Name:        "ticket",
		Description: "Submit a ticket/issue to the server staff",
		Category:    "Ticket",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "issue",
				Description: "Describe your issue or problem",
				Required:    true,
			},
		},
		Handler: ch.ticketSubmitHandler,
	})
}

func (ch *CommandHandler) setTicketHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to configure tickets.")
		return
	}

	channel := getChannelOption(i, "channel")
	if channel == nil {
		respondEphemeral(s, i, "Please specify a channel.")
		return
	}

	err := ch.bot.DB.SetTicketConfig(i.GuildID, channel.ID, true)
	if err != nil {
		respondEphemeral(s, i, "Failed to set ticket channel.")
		return
	}

	embed := successEmbed("Ticket System Enabled",
		fmt.Sprintf("Tickets will be forwarded to <#%s>\n\nUsers can now use `/ticket` to submit issues.", channel.ID))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) disableTicketHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to configure tickets.")
		return
	}

	err := ch.bot.DB.DeleteTicketConfig(i.GuildID)
	if err != nil {
		respondEphemeral(s, i, "Failed to disable ticket system.")
		return
	}

	embed := successEmbed("Ticket System Disabled",
		"The ticket system has been disabled for this server.")
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) ticketStatusHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config, err := ch.bot.DB.GetTicketConfig(i.GuildID)
	if err != nil || config == nil {
		respondEphemeral(s, i, "The ticket system is not configured for this server.")
		return
	}

	status := "Disabled"
	if config.Enabled {
		status = "Enabled"
	}

	embed := &discordgo.MessageEmbed{
		Title: "Ticket System Status",
		Color: 0xFF69B4,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Status", Value: status, Inline: true},
			{Name: "Channel", Value: fmt.Sprintf("<#%s>", config.ChannelID), Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Users can use /ticket to submit issues",
		},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) ticketSubmitHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config, err := ch.bot.DB.GetTicketConfig(i.GuildID)
	if err != nil || config == nil || !config.Enabled {
		respondEphemeral(s, i, "The ticket system is not enabled on this server.")
		return
	}

	issue := getStringOption(i, "issue")
	if issue == "" {
		respondEphemeral(s, i, "Please describe your issue.")
		return
	}

	// Create embed for the ticket channel
	ticketEmbed := &discordgo.MessageEmbed{
		Title: "New Ticket",
		Color: 0xFF69B4,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "From", Value: fmt.Sprintf("%s (%s)", i.Member.User.Username, i.Member.User.Mention()), Inline: true},
			{Name: "User ID", Value: i.Member.User.ID, Inline: true},
			{Name: "Issue", Value: issue, Inline: false},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: avatarURL(i.Member.User),
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Submitted from #%s", i.ChannelID),
		},
	}

	// Send to ticket channel
	_, err = s.ChannelMessageSendEmbed(config.ChannelID, ticketEmbed)
	if err != nil {
		respondEphemeral(s, i, "Failed to submit ticket. Please try again later.")
		return
	}

	// Respond to user (ephemeral so only they see it)
	embed := successEmbed("Ticket Submitted",
		"Your ticket has been submitted to the server staff. They will review it shortly.")
	respondEmbedEphemeral(s, i, embed)
}
