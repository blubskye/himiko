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
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (ch *CommandHandler) registerMiscCommands() {
	// Help command
	ch.Register(&Command{
		Name:        "help",
		Description: "Get help with commands",
		Category:    "Misc",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "category",
				Description: "Command category",
				Required:    false,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "Administration", Value: "Administration"},
					{Name: "Fun", Value: "Fun"},
					{Name: "Text", Value: "Text"},
					{Name: "Images", Value: "Images"},
					{Name: "Utility", Value: "Utility"},
					{Name: "Info", Value: "Info"},
					{Name: "Lookup", Value: "Lookup"},
					{Name: "Random", Value: "Random"},
					{Name: "Tools", Value: "Tools"},
					{Name: "Settings", Value: "Settings"},
					{Name: "Misc", Value: "Misc"},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "command",
				Description: "Specific command to get help for",
				Required:    false,
			},
		},
		Handler: ch.helpHandler,
	})

	// Create custom command
	ch.Register(&Command{
		Name:        "customcommand",
		Description: "Manage custom commands",
		Category:    "Misc",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add",
				Description: "Create a new custom command",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "name",
						Description: "Command name",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "response",
						Description: "Command response",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remove",
				Description: "Remove a custom command",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "name",
						Description: "Command name",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "list",
				Description: "List all custom commands",
			},
		},
		Handler: ch.customCommandHandler,
	})

	// Tag system
	ch.Register(&Command{
		Name:        "tag",
		Description: "Manage tags/snippets",
		Category:    "Misc",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "get",
				Description: "Get a tag",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "name",
						Description: "Tag name",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add",
				Description: "Create a new tag",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "name",
						Description: "Tag name",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "content",
						Description: "Tag content",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remove",
				Description: "Remove a tag",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "name",
						Description: "Tag name",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "list",
				Description: "List all tags",
			},
		},
		Handler: ch.tagHandler,
	})

	// Keyword notifier
	ch.Register(&Command{
		Name:        "keyword",
		Description: "Get notified when keywords are mentioned",
		Category:    "Misc",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add",
				Description: "Add a keyword to track",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "keyword",
						Description: "Keyword to track",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remove",
				Description: "Remove a tracked keyword",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "keyword",
						Description: "Keyword to remove",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "list",
				Description: "List tracked keywords",
			},
		},
		Handler: ch.keywordHandler,
	})

	// Command history
	ch.Register(&Command{
		Name:        "history",
		Description: "View command history",
		Category:    "Misc",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "limit",
				Description: "Number of commands to show",
				Required:    false,
				MinValue:    floatPtr(1),
				MaxValue:    25,
			},
		},
		Handler: ch.historyHandler,
	})

	// About/Credits
	ch.Register(&Command{
		Name:        "about",
		Description: "About this bot",
		Category:    "Misc",
		Handler:     ch.aboutHandler,
	})

	// Invite
	ch.Register(&Command{
		Name:        "invite",
		Description: "Get the bot invite link",
		Category:    "Misc",
		Handler:     ch.inviteHandler,
	})

	// Source code (AGPL compliance)
	ch.Register(&Command{
		Name:        "source",
		Description: "Get the source code link (AGPL-3.0)",
		Category:    "Misc",
		Handler:     ch.sourceHandler,
	})
}

func (ch *CommandHandler) helpHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	category := getStringOption(i, "category")
	command := getStringOption(i, "command")

	// Show specific command help
	if command != "" {
		if cmd, ok := ch.commands[command]; ok {
			embed := &discordgo.MessageEmbed{
				Title:       "/" + cmd.Name,
				Description: cmd.Description,
				Color:       0x5865F2,
				Fields: []*discordgo.MessageEmbedField{
					{Name: "Category", Value: cmd.Category, Inline: true},
				},
			}

			if len(cmd.Options) > 0 {
				var options []string
				for _, opt := range cmd.Options {
					required := ""
					if opt.Required {
						required = " (required)"
					}
					options = append(options, fmt.Sprintf("`%s`%s - %s", opt.Name, required, opt.Description))
				}
				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					Name:  "Options",
					Value: strings.Join(options, "\n"),
				})
			}

			respondEmbed(s, i, embed)
			return
		}
		respondEphemeral(s, i, "Command not found.")
		return
	}

	// Show category commands
	if category != "" {
		cmds := ch.GetCommandsByCategory(category)
		if len(cmds) == 0 {
			respondEphemeral(s, i, "No commands found in that category.")
			return
		}

		var cmdList []string
		for _, cmd := range cmds {
			cmdList = append(cmdList, fmt.Sprintf("`/%s` - %s", cmd.Name, cmd.Description))
		}

		// Sort alphabetically
		sort.Strings(cmdList)

		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("%s Commands", category),
			Description: strings.Join(cmdList, "\n"),
			Color:       0x5865F2,
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("%d commands", len(cmds)),
			},
		}

		respondEmbed(s, i, embed)
		return
	}

	// Show all categories
	categories := ch.GetCategories()
	sort.Strings(categories)

	var fields []*discordgo.MessageEmbedField
	for _, cat := range categories {
		cmds := ch.GetCommandsByCategory(cat)
		var cmdNames []string
		for _, cmd := range cmds {
			cmdNames = append(cmdNames, "`"+cmd.Name+"`")
		}
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s (%d)", cat, len(cmds)),
			Value:  strings.Join(cmdNames, ", "),
			Inline: false,
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Himiko Bot Help",
		Description: "*\"Let me help you... I promise I won't bite~ Much.\"*\n\nUse `/help category:<name>` to see commands in a category\nUse `/help command:<name>` for detailed command help",
		Color:       0xFF69B4,
		Fields:      fields,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://raw.githubusercontent.com/blubskye/himiko/main/himiko.png",
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("%d total commands", len(ch.commands)),
		},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) customCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	subcommand := getSubcommandName(i)

	switch subcommand {
	case "add":
		name := getStringOption(i, "name")
		response := getStringOption(i, "response")

		// Check if command already exists
		existing, _ := ch.bot.DB.GetCustomCommand(i.GuildID, name)
		if existing != nil {
			respondEphemeral(s, i, "A custom command with that name already exists.")
			return
		}

		err := ch.bot.DB.CreateCustomCommand(i.GuildID, name, response, i.Member.User.ID)
		if err != nil {
			respondEphemeral(s, i, "Failed to create custom command.")
			return
		}

		embed := successEmbed("Custom Command Created",
			fmt.Sprintf("Command `%s` has been created.", name))
		respondEmbed(s, i, embed)

	case "remove":
		name := getStringOption(i, "name")

		existing, _ := ch.bot.DB.GetCustomCommand(i.GuildID, name)
		if existing == nil {
			respondEphemeral(s, i, "Custom command not found.")
			return
		}

		err := ch.bot.DB.DeleteCustomCommand(i.GuildID, name)
		if err != nil {
			respondEphemeral(s, i, "Failed to delete custom command.")
			return
		}

		embed := successEmbed("Custom Command Deleted",
			fmt.Sprintf("Command `%s` has been deleted.", name))
		respondEmbed(s, i, embed)

	case "list":
		commands, err := ch.bot.DB.ListCustomCommands(i.GuildID)
		if err != nil || len(commands) == 0 {
			respondEphemeral(s, i, "No custom commands found.")
			return
		}

		var list []string
		for _, cmd := range commands {
			list = append(list, fmt.Sprintf("`%s` - %s (used %d times)",
				cmd.Name, truncate(cmd.Response, 50), cmd.UseCount))
		}

		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Custom Commands (%d)", len(commands)),
			Description: strings.Join(list, "\n"),
			Color:       0x5865F2,
		}

		respondEmbed(s, i, embed)
	}
}

func (ch *CommandHandler) tagHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	subcommand := getSubcommandName(i)

	switch subcommand {
	case "get":
		name := getStringOption(i, "name")

		tag, err := ch.bot.DB.GetTag(i.GuildID, name)
		if err != nil || tag == nil {
			respondEphemeral(s, i, "Tag not found.")
			return
		}

		ch.bot.DB.IncrementTagUse(i.GuildID, name)
		respond(s, i, tag.Content)

	case "add":
		name := getStringOption(i, "name")
		content := getStringOption(i, "content")

		existing, _ := ch.bot.DB.GetTag(i.GuildID, name)
		if existing != nil {
			respondEphemeral(s, i, "A tag with that name already exists.")
			return
		}

		err := ch.bot.DB.CreateTag(i.GuildID, name, content, i.Member.User.ID)
		if err != nil {
			respondEphemeral(s, i, "Failed to create tag.")
			return
		}

		embed := successEmbed("Tag Created", fmt.Sprintf("Tag `%s` has been created.", name))
		respondEmbed(s, i, embed)

	case "remove":
		name := getStringOption(i, "name")

		existing, _ := ch.bot.DB.GetTag(i.GuildID, name)
		if existing == nil {
			respondEphemeral(s, i, "Tag not found.")
			return
		}

		err := ch.bot.DB.DeleteTag(i.GuildID, name)
		if err != nil {
			respondEphemeral(s, i, "Failed to delete tag.")
			return
		}

		embed := successEmbed("Tag Deleted", fmt.Sprintf("Tag `%s` has been deleted.", name))
		respondEmbed(s, i, embed)

	case "list":
		tags, err := ch.bot.DB.ListTags(i.GuildID)
		if err != nil || len(tags) == 0 {
			respondEphemeral(s, i, "No tags found.")
			return
		}

		var list []string
		for _, tag := range tags {
			list = append(list, fmt.Sprintf("`%s` (used %d times)", tag.Name, tag.UseCount))
		}

		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Tags (%d)", len(tags)),
			Description: strings.Join(list, ", "),
			Color:       0x5865F2,
		}

		respondEmbed(s, i, embed)
	}
}

func (ch *CommandHandler) keywordHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	subcommand := getSubcommandName(i)

	switch subcommand {
	case "add":
		keyword := getStringOption(i, "keyword")

		err := ch.bot.DB.AddKeywordNotification(i.Member.User.ID, i.GuildID, keyword)
		if err != nil {
			respondEphemeral(s, i, "Failed to add keyword. It may already be tracked.")
			return
		}

		embed := successEmbed("Keyword Added",
			fmt.Sprintf("You will be notified when `%s` is mentioned.", keyword))
		respondEmbedEphemeral(s, i, embed)

	case "remove":
		keyword := getStringOption(i, "keyword")

		err := ch.bot.DB.RemoveKeywordNotification(i.Member.User.ID, keyword)
		if err != nil {
			respondEphemeral(s, i, "Failed to remove keyword.")
			return
		}

		embed := successEmbed("Keyword Removed",
			fmt.Sprintf("You will no longer be notified for `%s`.", keyword))
		respondEmbedEphemeral(s, i, embed)

	case "list":
		keywords, err := ch.bot.DB.GetKeywordNotifications(i.Member.User.ID)
		if err != nil || len(keywords) == 0 {
			respondEphemeral(s, i, "You have no tracked keywords.")
			return
		}

		var list []string
		for _, kw := range keywords {
			list = append(list, fmt.Sprintf("`%s`", kw.Keyword))
		}

		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Your Keywords (%d)", len(keywords)),
			Description: strings.Join(list, ", "),
			Color:       0x5865F2,
		}

		respondEmbedEphemeral(s, i, embed)
	}
}

func (ch *CommandHandler) historyHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	limit := int(getIntOption(i, "limit"))
	if limit == 0 {
		limit = 10
	}

	history, err := ch.bot.DB.GetCommandHistory(i.GuildID, limit)
	if err != nil || len(history) == 0 {
		respondEphemeral(s, i, "No command history found.")
		return
	}

	var list []string
	for _, h := range history {
		args := ""
		if h.Args != nil {
			args = " " + *h.Args
		}
		list = append(list, fmt.Sprintf("<t:%d:R> - <@%s> used `%s%s`",
			h.ExecutedAt.Unix(), h.UserID, h.Command, args))
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Command History",
		Description: strings.Join(list, "\n"),
		Color:       0x5865F2,
	}

	respondEmbedEphemeral(s, i, embed)
}

func (ch *CommandHandler) aboutHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	embed := &discordgo.MessageEmbed{
		Title:       "About Himiko",
		Description: "*\"I just wanna love you, wanna be loved~\"*\n\nA feature-rich Discord bot written in Go, named after everyone's favorite blood-obsessed villain! She's cute, she's crazy, and she'll manage your server with deadly efficiency~ 💕",
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://raw.githubusercontent.com/blubskye/himiko/main/himiko.png",
		},
		Color: 0xFF69B4,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Language", Value: "Go", Inline: true},
			{Name: "Library", Value: "discordgo", Inline: true},
			{Name: "Database", Value: "SQLite", Inline: true},
			{Name: "License", Value: "AGPL-3.0", Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Made with 💉 and obsessive love",
			IconURL: avatarURL(s.State.User),
		},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) inviteHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	clientID := s.State.User.ID

	// Permissions for a typical moderation/utility bot
	permissions := discordgo.PermissionAdministrator

	inviteURL := fmt.Sprintf("https://discord.com/api/oauth2/authorize?client_id=%s&permissions=%d&scope=bot%%20applications.commands",
		clientID, permissions)

	embed := &discordgo.MessageEmbed{
		Title:       "Invite Himiko",
		Description: fmt.Sprintf("[Click here to invite the bot](%s)", inviteURL),
		Color:       0x5865F2,
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) sourceHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	embed := &discordgo.MessageEmbed{
		Title: "Source Code",
		Description: "Himiko is free software licensed under the **GNU Affero General Public License v3.0** (AGPL-3.0).\n\n" +
			"This means you have the right to:\n" +
			"- Use this software for any purpose\n" +
			"- Study how the program works and modify it\n" +
			"- Redistribute copies\n" +
			"- Distribute your modified versions\n\n" +
			"If you run a modified version of this bot as a network service, you must make the source code available to users.",
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://raw.githubusercontent.com/blubskye/himiko/main/himiko.png",
		},
		Color: 0xFF69B4,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Repository",
				Value:  "https://github.com/blubskye/himiko",
				Inline: false,
			},
			{
				Name:   "License",
				Value:  "[AGPL-3.0](https://www.gnu.org/licenses/agpl-3.0.html)",
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Copyright (C) 2025 - Himiko Contributors",
			IconURL: avatarURL(s.State.User),
		},
	}

	respondEmbed(s, i, embed)
}
