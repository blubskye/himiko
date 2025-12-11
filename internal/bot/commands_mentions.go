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

func (ch *CommandHandler) registerMentionCommands() {
	// Mention response management
	ch.Register(&Command{
		Name:        "mention",
		Description: "Manage custom mention responses",
		Category:    "Configuration",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add",
				Description: "Add a custom mention response",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "trigger",
						Description: "Trigger text (when someone mentions the bot with this text)",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "response",
						Description: "Response message",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "image",
						Description: "Optional image URL to include",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remove",
				Description: "Remove a custom mention response",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "trigger",
						Description: "Trigger text to remove",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "list",
				Description: "List all custom mention responses",
			},
		},
		Handler: ch.mentionHandler,
	})
}

func (ch *CommandHandler) mentionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to manage mention responses.")
		return
	}

	subCmd := i.ApplicationCommandData().Options[0].Name

	switch subCmd {
	case "add":
		ch.mentionAddHandler(s, i)
	case "remove":
		ch.mentionRemoveHandler(s, i)
	case "list":
		ch.mentionListHandler(s, i)
	}
}

func (ch *CommandHandler) mentionAddHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := i.ApplicationCommandData().Options[0].Options

	var trigger, response, imageURL string
	for _, opt := range opts {
		switch opt.Name {
		case "trigger":
			trigger = strings.ToLower(opt.StringValue())
		case "response":
			response = opt.StringValue()
		case "image":
			imageURL = opt.StringValue()
		}
	}

	var imgPtr *string
	if imageURL != "" {
		imgPtr = &imageURL
	}

	err := ch.bot.DB.AddMentionResponse(i.GuildID, trigger, response, imgPtr, i.Member.User.ID)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			respondEphemeral(s, i, "A mention response with that trigger already exists.")
			return
		}
		respondEphemeral(s, i, "Failed to add mention response.")
		return
	}

	description := fmt.Sprintf("**Trigger:** %s\n**Response:** %s", trigger, truncate(response, 100))
	if imageURL != "" {
		description += "\n**Image:** Attached"
	}

	embed := successEmbed("Mention Response Added", description)
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) mentionRemoveHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	trigger := strings.ToLower(i.ApplicationCommandData().Options[0].Options[0].StringValue())

	err := ch.bot.DB.RemoveMentionResponse(i.GuildID, trigger)
	if err != nil {
		respondEphemeral(s, i, "Failed to remove mention response or it doesn't exist.")
		return
	}

	embed := successEmbed("Mention Response Removed", fmt.Sprintf("Removed response for trigger: **%s**", trigger))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) mentionListHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	responses, err := ch.bot.DB.GetMentionResponses(i.GuildID)
	if err != nil {
		respondEphemeral(s, i, "Failed to get mention responses.")
		return
	}

	if len(responses) == 0 {
		respondEphemeral(s, i, "No custom mention responses configured.")
		return
	}

	var description strings.Builder
	for idx, resp := range responses {
		if idx >= 15 {
			description.WriteString(fmt.Sprintf("\n... and %d more", len(responses)-15))
			break
		}

		hasImage := ""
		if resp.ImageURL != nil && *resp.ImageURL != "" {
			hasImage = " [IMG]"
		}
		description.WriteString(fmt.Sprintf("**%s**%s\n└ %s\n\n", resp.TriggerText, hasImage, truncate(resp.Response, 50)))
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Custom Mention Responses",
		Description: description.String(),
		Color:       0xFF69B4,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("%d responses configured", len(responses)),
		},
	}

	respondEmbed(s, i, embed)
}
