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

func (ch *CommandHandler) registerWebServerCommands() {
	ch.Register(&Command{
		Name:        "webserver",
		Description: "Manage the web dashboard server (Owner only)",
		Category:    "Admin",
		PrefixOnly:  true, // Owner-only command, no need for slash command
		PrefixHandler: func(ctx *PrefixContext) {
			ch.webserverPrefixHandler(ctx)
		},
	})
}

// webserverPrefixHandler handles prefix-based webserver commands
func (ch *CommandHandler) webserverPrefixHandler(ctx *PrefixContext) {
	// Owner only
	if !ch.bot.Config.IsOwner(ctx.Author.ID) {
		ctx.Reply("This command is only available to bot owners.")
		return
	}

	if len(ctx.Args) == 0 {
		ctx.Reply("Usage: `" + ctx.Prefix + "webserver <on|off|status|config>`")
		return
	}

	subCmd := ctx.Args[0]

	switch subCmd {
	case "on":
		ch.webserverOnPrefix(ctx)
	case "off":
		ch.webserverOffPrefix(ctx)
	case "status":
		ch.webserverStatusPrefix(ctx)
	case "config":
		ch.webserverConfigPrefix(ctx)
	default:
		ctx.Reply("Unknown subcommand. Usage: `" + ctx.Prefix + "webserver <on|off|status|config>`")
	}
}

func (ch *CommandHandler) webserverOnHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if ch.bot.WebServer.IsRunning() {
		respondEphemeral(s, i, "The web server is already running.")
		return
	}

	if err := ch.bot.WebServer.Start(); err != nil {
		respondEphemeral(s, i, fmt.Sprintf("Failed to start web server: %v", err))
		return
	}

	// Update config to persist the setting
	ch.bot.Config.WebServer.Enabled = true
	ch.bot.Config.Save("config.json")

	addr := fmt.Sprintf("http://%s:%d", ch.bot.Config.WebServer.Host, ch.bot.Config.WebServer.Port)

	embed := &discordgo.MessageEmbed{
		Title:       "Web Server Started",
		Description: fmt.Sprintf("The dashboard is now available at:\n**%s**", addr),
		Color:       0x57F287,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Host",
				Value:  ch.bot.Config.WebServer.Host,
				Inline: true,
			},
			{
				Name:   "Port",
				Value:  fmt.Sprintf("%d", ch.bot.Config.WebServer.Port),
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Configure NGINX to proxy this for external access",
		},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) webserverOffHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !ch.bot.WebServer.IsRunning() {
		respondEphemeral(s, i, "The web server is not running.")
		return
	}

	if err := ch.bot.WebServer.Stop(); err != nil {
		respondEphemeral(s, i, fmt.Sprintf("Failed to stop web server: %v", err))
		return
	}

	// Update config to persist the setting
	ch.bot.Config.WebServer.Enabled = false
	ch.bot.Config.Save("config.json")

	embed := &discordgo.MessageEmbed{
		Title:       "Web Server Stopped",
		Description: "The dashboard has been shut down.",
		Color:       0xED4245,
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) webserverStatusHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	running := ch.bot.WebServer.IsRunning()

	var status, statusEmoji string
	var color int
	if running {
		status = "Running"
		statusEmoji = "🟢"
		color = 0x57F287
	} else {
		status = "Stopped"
		statusEmoji = "🔴"
		color = 0xED4245
	}

	addr := fmt.Sprintf("http://%s:%d", ch.bot.Config.WebServer.Host, ch.bot.Config.WebServer.Port)

	embed := &discordgo.MessageEmbed{
		Title: "Web Server Status",
		Color: color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Status",
				Value:  fmt.Sprintf("%s %s", statusEmoji, status),
				Inline: true,
			},
			{
				Name:   "Auto-Start",
				Value:  boolToEnabled(ch.bot.Config.WebServer.Enabled),
				Inline: true,
			},
			{
				Name:   "Address",
				Value:  addr,
				Inline: false,
			},
			{
				Name:   "Allow Remote",
				Value:  boolToEnabled(ch.bot.Config.WebServer.AllowRemote),
				Inline: true,
			},
		},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) webserverConfigHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := getSubcommandOptions(i)

	// Check if any options were provided
	if len(opts) == 0 {
		// Just show current config
		ch.webserverStatusHandler(s, i)
		return
	}

	// Update configuration
	changed := false
	for _, opt := range opts {
		switch opt.Name {
		case "port":
			port := int(opt.IntValue())
			if port >= 1 && port <= 65535 {
				ch.bot.Config.WebServer.Port = port
				changed = true
			}
		case "allow_remote":
			ch.bot.Config.WebServer.AllowRemote = opt.BoolValue()
			changed = true
		}
	}

	if changed {
		// Save config
		if err := ch.bot.Config.Save("config.json"); err != nil {
			respondEphemeral(s, i, fmt.Sprintf("Failed to save config: %v", err))
			return
		}

		// If server is running, it needs to be restarted for port changes
		needsRestart := ch.bot.WebServer.IsRunning()

		embed := &discordgo.MessageEmbed{
			Title:       "Web Server Configuration Updated",
			Description: "The configuration has been saved.",
			Color:       0x57F287,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Port",
					Value:  fmt.Sprintf("%d", ch.bot.Config.WebServer.Port),
					Inline: true,
				},
				{
					Name:   "Allow Remote",
					Value:  boolToEnabled(ch.bot.Config.WebServer.AllowRemote),
					Inline: true,
				},
			},
		}

		if needsRestart {
			embed.Footer = &discordgo.MessageEmbedFooter{
				Text: "Restart the web server for changes to take effect",
			}
		}

		respondEmbed(s, i, embed)
	}
}

// getSubcommandOptions returns the options for the current subcommand
func getSubcommandOptions(i *discordgo.InteractionCreate) []*discordgo.ApplicationCommandInteractionDataOption {
	if len(i.ApplicationCommandData().Options) == 0 {
		return nil
	}
	subCmd := i.ApplicationCommandData().Options[0]
	return subCmd.Options
}

// Prefix command handlers

func (ch *CommandHandler) webserverOnPrefix(ctx *PrefixContext) {
	if ch.bot.WebServer.IsRunning() {
		ctx.Reply("The web server is already running.")
		return
	}

	if err := ch.bot.WebServer.Start(); err != nil {
		ctx.Reply(fmt.Sprintf("Failed to start web server: %v", err))
		return
	}

	// Update config to persist the setting
	ch.bot.Config.WebServer.Enabled = true
	ch.bot.Config.Save("config.json")

	addr := fmt.Sprintf("http://%s:%d", ch.bot.Config.WebServer.Host, ch.bot.Config.WebServer.Port)

	embed := &discordgo.MessageEmbed{
		Title:       "Web Server Started",
		Description: fmt.Sprintf("The dashboard is now available at:\n**%s**", addr),
		Color:       0x57F287,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Host",
				Value:  ch.bot.Config.WebServer.Host,
				Inline: true,
			},
			{
				Name:   "Port",
				Value:  fmt.Sprintf("%d", ch.bot.Config.WebServer.Port),
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Configure NGINX to proxy this for external access",
		},
	}

	ctx.ReplyEmbed(embed)
}

func (ch *CommandHandler) webserverOffPrefix(ctx *PrefixContext) {
	if !ch.bot.WebServer.IsRunning() {
		ctx.Reply("The web server is not running.")
		return
	}

	if err := ch.bot.WebServer.Stop(); err != nil {
		ctx.Reply(fmt.Sprintf("Failed to stop web server: %v", err))
		return
	}

	// Update config to persist the setting
	ch.bot.Config.WebServer.Enabled = false
	ch.bot.Config.Save("config.json")

	embed := &discordgo.MessageEmbed{
		Title:       "Web Server Stopped",
		Description: "The dashboard has been shut down.",
		Color:       0xED4245,
	}

	ctx.ReplyEmbed(embed)
}

func (ch *CommandHandler) webserverStatusPrefix(ctx *PrefixContext) {
	running := ch.bot.WebServer.IsRunning()

	var status, statusEmoji string
	var color int
	if running {
		status = "Running"
		statusEmoji = "🟢"
		color = 0x57F287
	} else {
		status = "Stopped"
		statusEmoji = "🔴"
		color = 0xED4245
	}

	addr := fmt.Sprintf("http://%s:%d", ch.bot.Config.WebServer.Host, ch.bot.Config.WebServer.Port)

	embed := &discordgo.MessageEmbed{
		Title: "Web Server Status",
		Color: color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Status",
				Value:  fmt.Sprintf("%s %s", statusEmoji, status),
				Inline: true,
			},
			{
				Name:   "Auto-Start",
				Value:  boolToEnabled(ch.bot.Config.WebServer.Enabled),
				Inline: true,
			},
			{
				Name:   "Address",
				Value:  addr,
				Inline: false,
			},
			{
				Name:   "Allow Remote",
				Value:  boolToEnabled(ch.bot.Config.WebServer.AllowRemote),
				Inline: true,
			},
		},
	}

	ctx.ReplyEmbed(embed)
}

func (ch *CommandHandler) webserverConfigPrefix(ctx *PrefixContext) {
	// Parse args: config [port <num>] [allow_remote <true|false>]
	if len(ctx.Args) < 2 {
		// Just show current config
		ch.webserverStatusPrefix(ctx)
		return
	}

	changed := false
	for i := 1; i < len(ctx.Args); i += 2 {
		if i+1 >= len(ctx.Args) {
			break
		}
		key := ctx.Args[i]
		value := ctx.Args[i+1]

		switch key {
		case "port":
			var port int
			fmt.Sscanf(value, "%d", &port)
			if port >= 1 && port <= 65535 {
				ch.bot.Config.WebServer.Port = port
				changed = true
			}
		case "allow_remote":
			ch.bot.Config.WebServer.AllowRemote = value == "true" || value == "1" || value == "yes"
			changed = true
		}
	}

	if changed {
		// Save config
		if err := ch.bot.Config.Save("config.json"); err != nil {
			ctx.Reply(fmt.Sprintf("Failed to save config: %v", err))
			return
		}

		needsRestart := ch.bot.WebServer.IsRunning()

		embed := &discordgo.MessageEmbed{
			Title:       "Web Server Configuration Updated",
			Description: "The configuration has been saved.",
			Color:       0x57F287,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Port",
					Value:  fmt.Sprintf("%d", ch.bot.Config.WebServer.Port),
					Inline: true,
				},
				{
					Name:   "Allow Remote",
					Value:  boolToEnabled(ch.bot.Config.WebServer.AllowRemote),
					Inline: true,
				},
			},
		}

		if needsRestart {
			embed.Footer = &discordgo.MessageEmbedFooter{
				Text: "Restart the web server for changes to take effect",
			}
		}

		ctx.ReplyEmbed(embed)
	} else {
		ctx.Reply("Usage: `" + ctx.Prefix + "webserver config port <number>` or `" + ctx.Prefix + "webserver config allow_remote <true|false>`")
	}
}
