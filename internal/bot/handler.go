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
	Name        string
	Description string
	Category    string
	Options     []*discordgo.ApplicationCommandOption
	Handler     func(s *discordgo.Session, i *discordgo.InteractionCreate)
	Autocomplete func(s *discordgo.Session, i *discordgo.InteractionCreate)
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

	return ch
}

func (ch *CommandHandler) Register(cmd *Command) {
	ch.commands[cmd.Name] = cmd
}

func (ch *CommandHandler) RegisterCommands() error {
	var appCommands []*discordgo.ApplicationCommand

	for _, cmd := range ch.commands {
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

	log.Printf("Registered %d slash commands", len(appCommands))
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
