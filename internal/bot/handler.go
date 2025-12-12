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
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type CommandHandler struct {
	bot      *Bot
	commands map[string]*Command
}

type Command struct {
	Name          string
	Description   string
	Category      string
	Options       []*discordgo.ApplicationCommandOption
	Handler       func(s *discordgo.Session, i *discordgo.InteractionCreate)
	Autocomplete  func(s *discordgo.Session, i *discordgo.InteractionCreate)
	PrefixHandler func(ctx *PrefixContext) // Handler for prefix-based commands
	SlashOnly     bool                     // If true, only register as slash command (default behavior for essential commands)
	PrefixOnly    bool                     // If true, only available via prefix (not registered as slash command)
}

// PrefixContext holds context for prefix-based command execution
type PrefixContext struct {
	Session   *discordgo.Session
	Message   *discordgo.MessageCreate
	Command   *Command
	Args      []string
	Bot       *Bot
	ChannelID string
	GuildID   string
	Author    *discordgo.User
	Prefix    string
}

// Reply sends a message reply
func (ctx *PrefixContext) Reply(content string) {
	ctx.Session.ChannelMessageSend(ctx.ChannelID, content)
}

// ReplyEmbed sends an embed reply
func (ctx *PrefixContext) ReplyEmbed(embed *discordgo.MessageEmbed) {
	ctx.Session.ChannelMessageSendEmbed(ctx.ChannelID, embed)
}

// GetArg returns the argument at the given index, or empty string if not found
func (ctx *PrefixContext) GetArg(index int) string {
	if index < len(ctx.Args) {
		return ctx.Args[index]
	}
	return ""
}

// GetArgRest returns all arguments from the given index joined by space
func (ctx *PrefixContext) GetArgRest(index int) string {
	if index < len(ctx.Args) {
		return strings.Join(ctx.Args[index:], " ")
	}
	return ""
}

// Categories that should be prefix-only to stay under Discord's 100 slash command limit
var prefixOnlyCategories = map[string]bool{
	"Fun":           true,
	"Text":          true,
	"Random":        true,
	"Images":        true,
	"Lookup":        true,
	"Music":         true,
	"Tools":         true,
	"Utility":       true, // 13 commands
	"Configuration": true, // 1 command
}

func NewCommandHandler(b *Bot) *CommandHandler {
	ch := &CommandHandler{
		bot:      b,
		commands: make(map[string]*Command),
	}

	// Register all commands
	ch.registerAdminCommands()
	ch.registerFunCommands()
	ch.registerTextCommands()
	ch.registerImageCommands()
	ch.registerUtilityCommands()
	ch.registerMiscCommands()
	ch.registerInfoCommands()
	ch.registerLookupCommands()
	ch.registerRandomCommands()
	ch.registerToolsCommands()
	ch.registerSettingsCommands()
	ch.registerAICommands()

	// New Yuno-ported commands
	ch.registerXPCommands()
	ch.registerFiltersCommands()
	ch.registerLoggingCommands()
	ch.registerAutoCleanCommands()
	ch.registerVoiceXPCommands()
	ch.registerRanksCommands()
	ch.registerDMCommands()
	ch.registerBotBanCommands()
	ch.registerBanExportCommands()
	ch.registerModStatsCommands()
	ch.registerSpamCommands()
	ch.registerMentionCommands()
	ch.registerTicketCommands()
	ch.registerAntiRaidCommands()
	ch.registerAntiSpamCommands()
	ch.registerMusicCommands()
	ch.registerUpdateCommands()

	return ch
}

func (ch *CommandHandler) Register(cmd *Command) {
	ch.commands[cmd.Name] = cmd
}

func (ch *CommandHandler) RegisterCommands() error {
	var appCommands []*discordgo.ApplicationCommand
	var prefixOnlyCount int

	for _, cmd := range ch.commands {
		// Skip prefix-only commands
		if cmd.PrefixOnly {
			prefixOnlyCount++
			continue
		}

		// Skip commands in prefix-only categories (unless explicitly marked as slash-only)
		if !cmd.SlashOnly && prefixOnlyCategories[cmd.Category] {
			prefixOnlyCount++
			continue
		}

		appCommands = append(appCommands, &discordgo.ApplicationCommand{
			Name:        cmd.Name,
			Description: cmd.Description,
			Options:     cmd.Options,
		})
	}

	// Register commands globally
	_, err := ch.bot.Session.ApplicationCommandBulkOverwrite(ch.bot.Session.State.User.ID, "", appCommands)
	if err != nil {
		return err
	}

	log.Printf("Registered %d slash commands (%d prefix-only)", len(appCommands), prefixOnlyCount)
	return nil
}

func (ch *CommandHandler) UnregisterCommands() {
	commands, err := ch.bot.Session.ApplicationCommands(ch.bot.Session.State.User.ID, "")
	if err != nil {
		return
	}

	for _, cmd := range commands {
		ch.bot.Session.ApplicationCommandDelete(ch.bot.Session.State.User.ID, "", cmd.ID)
	}
}

func (ch *CommandHandler) HandleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cmdName := i.ApplicationCommandData().Name

	// Handle subcommands
	if len(i.ApplicationCommandData().Options) > 0 {
		opt := i.ApplicationCommandData().Options[0]
		if opt.Type == discordgo.ApplicationCommandOptionSubCommand ||
			opt.Type == discordgo.ApplicationCommandOptionSubCommandGroup {
			cmdName = cmdName + "_" + opt.Name

			// Handle nested subcommands
			if opt.Type == discordgo.ApplicationCommandOptionSubCommandGroup && len(opt.Options) > 0 {
				cmdName = cmdName + "_" + opt.Options[0].Name
			}
		}
	}

	cmd, exists := ch.commands[cmdName]
	if !exists {
		// Try base command
		cmd, exists = ch.commands[i.ApplicationCommandData().Name]
	}

	if exists && cmd.Handler != nil {
		// Log command usage
		guildID := ""
		if i.GuildID != "" {
			guildID = i.GuildID
		}

		var args string
		for _, opt := range i.ApplicationCommandData().Options {
			args += opt.Name + " "
		}

		ch.bot.DB.LogCommand(guildID, i.ChannelID, i.Member.User.ID, cmdName, strings.TrimSpace(args))

		cmd.Handler(s, i)
	} else {
		respond(s, i, "Unknown command")
	}
}

func (ch *CommandHandler) HandleAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cmdName := i.ApplicationCommandData().Name

	cmd, exists := ch.commands[cmdName]
	if exists && cmd.Autocomplete != nil {
		cmd.Autocomplete(s, i)
	}
}

func (ch *CommandHandler) GetCommands() map[string]*Command {
	return ch.commands
}

func (ch *CommandHandler) GetCommandsByCategory(category string) []*Command {
	var cmds []*Command
	for _, cmd := range ch.commands {
		if cmd.Category == category {
			cmds = append(cmds, cmd)
		}
	}
	return cmds
}

func (ch *CommandHandler) GetCategories() []string {
	categoryMap := make(map[string]bool)
	for _, cmd := range ch.commands {
		if cmd.Category != "" {
			categoryMap[cmd.Category] = true
		}
	}

	var categories []string
	for cat := range categoryMap {
		categories = append(categories, cat)
	}
	return categories
}
