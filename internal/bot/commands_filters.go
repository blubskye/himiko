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
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (ch *CommandHandler) registerFiltersCommands() {
	// Add filter
	ch.Register(&Command{
		Name:        "addfilter",
		Description: "Add a regex filter for auto-moderation",
		Category:    "Filters",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "pattern",
				Description: "Regex pattern to match",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "action",
				Description: "Action to take when matched",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "Delete Message", Value: "delete"},
					{Name: "Warn User", Value: "warn"},
					{Name: "Ban User", Value: "ban"},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "Reason for this filter",
				Required:    false,
			},
		},
		Handler: ch.addFilterHandler,
	})

	// Remove filter
	ch.Register(&Command{
		Name:        "removefilter",
		Description: "Remove a regex filter",
		Category:    "Filters",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "id",
				Description: "Filter ID to remove",
				Required:    true,
			},
		},
		Handler: ch.removeFilterHandler,
	})

	// List filters
	ch.Register(&Command{
		Name:        "listfilters",
		Description: "List all regex filters for this server",
		Category:    "Filters",
		Handler:     ch.listFiltersHandler,
	})

	// Test filter
	ch.Register(&Command{
		Name:        "testfilter",
		Description: "Test a regex pattern against text",
		Category:    "Filters",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "pattern",
				Description: "Regex pattern to test",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "text",
				Description: "Text to test against",
				Required:    true,
			},
		},
		Handler: ch.testFilterHandler,
	})
}

func (ch *CommandHandler) addFilterHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to manage filters.")
		return
	}

	pattern := getStringOption(i, "pattern")
	action := getStringOption(i, "action")
	reason := getStringOption(i, "reason")

	if reason == "" {
		reason = "No reason provided"
	}

	// Validate regex
	_, err := regexp.Compile(pattern)
	if err != nil {
		respondEphemeral(s, i, fmt.Sprintf("Invalid regex pattern: %v", err))
		return
	}

	err = ch.bot.DB.AddRegexFilter(i.GuildID, pattern, action, reason, i.Member.User.ID)
	if err != nil {
		respondEphemeral(s, i, "Failed to add filter.")
		return
	}

	actionText := map[string]string{
		"delete": "Delete message",
		"warn":   "Warn user",
		"ban":    "Ban user",
	}

	embed := successEmbed("Filter Added",
		fmt.Sprintf("**Pattern:** `%s`\n**Action:** %s\n**Reason:** %s",
			pattern, actionText[action], reason))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) removeFilterHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to manage filters.")
		return
	}

	id := getIntOption(i, "id")

	err := ch.bot.DB.RemoveRegexFilter(i.GuildID, id)
	if err != nil {
		respondEphemeral(s, i, "Failed to remove filter.")
		return
	}

	embed := successEmbed("Filter Removed",
		fmt.Sprintf("Removed filter #%d", id))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) listFiltersHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	filters, err := ch.bot.DB.GetRegexFilters(i.GuildID)
	if err != nil {
		respondEphemeral(s, i, "Failed to get filters.")
		return
	}

	if len(filters) == 0 {
		respondEphemeral(s, i, "No filters configured for this server.")
		return
	}

	var description strings.Builder
	actionEmoji := map[string]string{
		"delete": ":wastebasket:",
		"warn":   ":warning:",
		"ban":    ":hammer:",
	}

	for _, f := range filters {
		emoji := actionEmoji[f.Action]
		if emoji == "" {
			emoji = ":question:"
		}
		description.WriteString(fmt.Sprintf("**#%d** %s `%s`\n└ %s\n",
			f.ID, emoji, truncate(f.Pattern, 40), f.Reason))
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Regex Filters (%d)", len(filters)),
		Description: description.String(),
		Color:       0x5865F2,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Use /removefilter <id> to remove a filter",
		},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) testFilterHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	pattern := getStringOption(i, "pattern")
	text := getStringOption(i, "text")

	re, err := regexp.Compile(pattern)
	if err != nil {
		respondEphemeral(s, i, fmt.Sprintf("Invalid regex pattern: %v", err))
		return
	}

	matches := re.FindAllString(text, -1)
	matched := re.MatchString(text)

	var result string
	if matched {
		result = fmt.Sprintf(":white_check_mark: **Pattern matches!**\n\n**Matches found:** %d", len(matches))
		if len(matches) > 0 && len(matches) <= 10 {
			result += "\n**Matched text:**\n"
			for _, m := range matches {
				result += fmt.Sprintf("• `%s`\n", truncate(m, 50))
			}
		}
	} else {
		result = ":x: **Pattern does not match**"
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Filter Test",
		Description: result,
		Color:       0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Pattern", Value: fmt.Sprintf("`%s`", pattern), Inline: false},
			{Name: "Test Text", Value: truncate(text, 200), Inline: false},
		},
	}

	respondEmbedEphemeral(s, i, embed)
}
