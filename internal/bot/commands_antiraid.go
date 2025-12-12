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
	"time"

	"github.com/bwmarrin/discordgo"
)

func (ch *CommandHandler) registerAntiRaidCommands() {
	// Anti-raid configuration
	ch.Register(&Command{
		Name:        "antiraid",
		Description: "Configure anti-raid protection",
		Category:    "Anti-Raid",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "status",
				Description: "View anti-raid configuration",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "enable",
				Description: "Enable anti-raid protection",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "disable",
				Description: "Disable anti-raid protection",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "set",
				Description: "Configure anti-raid settings",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "raid_time",
						Description: "Time window in seconds to detect raid (default: 300)",
						Required:    false,
						MinValue:    floatPtr(10),
						MaxValue:    3600,
					},
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "raid_size",
						Description: "Number of joins to trigger raid (default: 5)",
						Required:    false,
						MinValue:    floatPtr(2),
						MaxValue:    100,
					},
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "lockdown_duration",
						Description: "Lockdown duration in seconds (default: 120)",
						Required:    false,
						MinValue:    floatPtr(0),
						MaxValue:    3600,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "action",
						Description: "Action to take on raid",
						Required:    false,
						Choices: []*discordgo.ApplicationCommandOptionChoice{
							{Name: "Silence (assign role)", Value: "silence"},
							{Name: "Kick", Value: "kick"},
							{Name: "Ban", Value: "ban"},
						},
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "setrole",
				Description: "Set the silent role for silenced users",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionRole,
						Name:        "role",
						Description: "Role to assign to silenced users",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "setalert",
				Description: "Set the alert role and channel",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionChannel,
						Name:        "channel",
						Description: "Channel for raid alerts",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionRole,
						Name:        "role",
						Description: "Role to ping on raid detection",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "autosilence",
				Description: "Configure auto-silence mode",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "mode",
						Description: "Auto-silence mode",
						Required:    true,
						Choices: []*discordgo.ApplicationCommandOptionChoice{
							{Name: "Off - No automatic silencing", Value: "off"},
							{Name: "Log - Only log joins", Value: "log"},
							{Name: "Alert - Alert on all joins", Value: "alert"},
							{Name: "Raid - Silence only during raid", Value: "raid"},
							{Name: "All - Silence all new members", Value: "all"},
						},
					},
				},
			},
		},
		Handler: ch.antiRaidHandler,
	})

	// Silence user
	ch.Register(&Command{
		Name:        "silence",
		Description: "Silence a user by assigning the silent role",
		Category:    "Anti-Raid",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "User to silence",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "duration",
				Description: "Duration (e.g., 1h, 30m, 1d)",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "Reason for silencing",
				Required:    false,
			},
		},
		Handler: ch.silenceHandler,
	})

	// Unsilence user
	ch.Register(&Command{
		Name:        "unsilence",
		Description: "Remove silence from a user",
		Category:    "Anti-Raid",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "User to unsilence",
				Required:    true,
			},
		},
		Handler: ch.unsilenceHandler,
	})

	// Get raid users
	ch.Register(&Command{
		Name:        "getraid",
		Description: "Get list of users from recent raid",
		Category:    "Anti-Raid",
		Handler:     ch.getRaidHandler,
	})

	// Ban raid users
	ch.Register(&Command{
		Name:        "banraid",
		Description: "Ban all users from recent raid",
		Category:    "Anti-Raid",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "Reason for banning",
				Required:    false,
			},
		},
		Handler: ch.banRaidHandler,
	})

	// Lockdown
	ch.Register(&Command{
		Name:        "lockdown",
		Description: "Toggle server lockdown (raises verification level)",
		Category:    "Anti-Raid",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "enable",
				Description: "Enable or disable lockdown",
				Required:    true,
			},
		},
		Handler: ch.lockdownHandler,
	})
}

func (ch *CommandHandler) antiRaidHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to configure anti-raid.")
		return
	}

	subCmd := i.ApplicationCommandData().Options[0].Name

	switch subCmd {
	case "status":
		ch.antiRaidStatusHandler(s, i)
	case "enable":
		ch.antiRaidEnableHandler(s, i)
	case "disable":
		ch.antiRaidDisableHandler(s, i)
	case "set":
		ch.antiRaidSetHandler(s, i)
	case "setrole":
		ch.antiRaidSetRoleHandler(s, i)
	case "setalert":
		ch.antiRaidSetAlertHandler(s, i)
	case "autosilence":
		ch.antiRaidAutoSilenceHandler(s, i)
	}
}

func (ch *CommandHandler) antiRaidStatusHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cfg, err := ch.bot.DB.GetAntiRaidConfig(i.GuildID)
	if err != nil {
		respondEphemeral(s, i, "Failed to get anti-raid configuration.")
		return
	}

	status := "Disabled"
	if cfg.Enabled {
		status = "Enabled"
	}

	autoSilenceMode := map[int]string{
		-2: "Log only",
		-1: "Alert on joins",
		0:  "Off",
		1:  "Raid mode",
		2:  "All joins",
	}[cfg.AutoSilence]

	silentRole := "Not set"
	if cfg.SilentRoleID != "" {
		silentRole = fmt.Sprintf("<@&%s>", cfg.SilentRoleID)
	}

	alertChannel := "Not set"
	if cfg.LogChannelID != "" {
		alertChannel = fmt.Sprintf("<#%s>", cfg.LogChannelID)
	}

	alertRole := "Not set"
	if cfg.AlertRoleID != "" {
		alertRole = fmt.Sprintf("<@&%s>", cfg.AlertRoleID)
	}

	embed := &discordgo.MessageEmbed{
		Title: "Anti-Raid Configuration",
		Color: 0xFF69B4,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Status", Value: status, Inline: true},
			{Name: "Action", Value: cfg.Action, Inline: true},
			{Name: "Auto-Silence", Value: autoSilenceMode, Inline: true},
			{Name: "Raid Time", Value: fmt.Sprintf("%d seconds", cfg.RaidTime), Inline: true},
			{Name: "Raid Size", Value: fmt.Sprintf("%d users", cfg.RaidSize), Inline: true},
			{Name: "Lockdown Duration", Value: fmt.Sprintf("%d seconds", cfg.LockdownDuration), Inline: true},
			{Name: "Silent Role", Value: silentRole, Inline: true},
			{Name: "Alert Channel", Value: alertChannel, Inline: true},
			{Name: "Alert Role", Value: alertRole, Inline: true},
		},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) antiRaidEnableHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cfg, _ := ch.bot.DB.GetAntiRaidConfig(i.GuildID)
	cfg.Enabled = true

	if err := ch.bot.DB.SetAntiRaidConfig(cfg); err != nil {
		respondEphemeral(s, i, "Failed to enable anti-raid.")
		return
	}

	embed := successEmbed("Anti-Raid Enabled", "Anti-raid protection is now active.")
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) antiRaidDisableHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cfg, _ := ch.bot.DB.GetAntiRaidConfig(i.GuildID)
	cfg.Enabled = false

	if err := ch.bot.DB.SetAntiRaidConfig(cfg); err != nil {
		respondEphemeral(s, i, "Failed to disable anti-raid.")
		return
	}

	embed := successEmbed("Anti-Raid Disabled", "Anti-raid protection has been disabled.")
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) antiRaidSetHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cfg, _ := ch.bot.DB.GetAntiRaidConfig(i.GuildID)
	opts := i.ApplicationCommandData().Options[0].Options

	changes := []string{}

	for _, opt := range opts {
		switch opt.Name {
		case "raid_time":
			cfg.RaidTime = int(opt.IntValue())
			changes = append(changes, fmt.Sprintf("Raid time: %d seconds", cfg.RaidTime))
		case "raid_size":
			cfg.RaidSize = int(opt.IntValue())
			changes = append(changes, fmt.Sprintf("Raid size: %d users", cfg.RaidSize))
		case "lockdown_duration":
			cfg.LockdownDuration = int(opt.IntValue())
			changes = append(changes, fmt.Sprintf("Lockdown duration: %d seconds", cfg.LockdownDuration))
		case "action":
			cfg.Action = opt.StringValue()
			changes = append(changes, fmt.Sprintf("Action: %s", cfg.Action))
		}
	}

	if len(changes) == 0 {
		respondEphemeral(s, i, "No settings provided.")
		return
	}

	if err := ch.bot.DB.SetAntiRaidConfig(cfg); err != nil {
		respondEphemeral(s, i, "Failed to update settings.")
		return
	}

	embed := successEmbed("Anti-Raid Updated", "Updated settings:\n- "+strings.Join(changes, "\n- "))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) antiRaidSetRoleHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	role := getRoleOption(i, "role")
	if role == nil {
		respondEphemeral(s, i, "Please specify a role.")
		return
	}

	cfg, _ := ch.bot.DB.GetAntiRaidConfig(i.GuildID)
	cfg.SilentRoleID = role.ID

	if err := ch.bot.DB.SetAntiRaidConfig(cfg); err != nil {
		respondEphemeral(s, i, "Failed to set silent role.")
		return
	}

	embed := successEmbed("Silent Role Set", fmt.Sprintf("Silent role set to <@&%s>", role.ID))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) antiRaidSetAlertHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	channel := getChannelOption(i, "channel")
	if channel == nil {
		respondEphemeral(s, i, "Please specify a channel.")
		return
	}

	cfg, _ := ch.bot.DB.GetAntiRaidConfig(i.GuildID)
	cfg.LogChannelID = channel.ID

	role := getRoleOption(i, "role")
	if role != nil {
		cfg.AlertRoleID = role.ID
	}

	if err := ch.bot.DB.SetAntiRaidConfig(cfg); err != nil {
		respondEphemeral(s, i, "Failed to set alert settings.")
		return
	}

	msg := fmt.Sprintf("Alert channel set to <#%s>", channel.ID)
	if role != nil {
		msg += fmt.Sprintf("\nAlert role set to <@&%s>", role.ID)
	}

	embed := successEmbed("Alert Settings Updated", msg)
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) antiRaidAutoSilenceHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	mode := getStringOption(i, "mode")

	modeMap := map[string]int{
		"off":   0,
		"log":   -2,
		"alert": -1,
		"raid":  1,
		"all":   2,
	}

	cfg, _ := ch.bot.DB.GetAntiRaidConfig(i.GuildID)
	cfg.AutoSilence = modeMap[mode]

	if err := ch.bot.DB.SetAntiRaidConfig(cfg); err != nil {
		respondEphemeral(s, i, "Failed to set auto-silence mode.")
		return
	}

	modeLabels := map[string]string{
		"off":   "Off - No automatic silencing",
		"log":   "Log - Only logging joins",
		"alert": "Alert - Alerting on all joins",
		"raid":  "Raid - Silencing during raids only",
		"all":   "All - Silencing all new members",
	}

	embed := successEmbed("Auto-Silence Mode Set", fmt.Sprintf("Auto-silence mode: **%s**", modeLabels[mode]))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) silenceHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isModerator(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need moderator permissions to silence users.")
		return
	}

	cfg, err := ch.bot.DB.GetAntiRaidConfig(i.GuildID)
	if err != nil || cfg.SilentRoleID == "" {
		respondEphemeral(s, i, "Silent role not configured. Use `/antiraid setrole` first.")
		return
	}

	user := getUserOption(i, "user")
	if user == nil {
		respondEphemeral(s, i, "Please specify a user.")
		return
	}

	durationStr := getStringOption(i, "duration")
	reason := getStringOption(i, "reason")

	// Add silent role
	err = s.GuildMemberRoleAdd(i.GuildID, user.ID, cfg.SilentRoleID)
	if err != nil {
		respondEphemeral(s, i, "Failed to silence user: "+err.Error())
		return
	}

	// Schedule unsilence if duration provided
	if durationStr != "" {
		duration, err := parseDuration(durationStr)
		if err == nil && duration > 0 {
			executeAt := time.Now().Add(duration).UnixMilli()
			ch.bot.DB.AddScheduledEvent(i.GuildID, "unsilence", user.ID, executeAt)
		}
	}

	msg := fmt.Sprintf("Silenced %s", user.Mention())
	if durationStr != "" {
		msg += fmt.Sprintf(" for %s", durationStr)
	}
	if reason != "" {
		msg += fmt.Sprintf("\nReason: %s", reason)
	}

	embed := successEmbed("User Silenced", msg)
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) unsilenceHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isModerator(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need moderator permissions to unsilence users.")
		return
	}

	cfg, err := ch.bot.DB.GetAntiRaidConfig(i.GuildID)
	if err != nil || cfg.SilentRoleID == "" {
		respondEphemeral(s, i, "Silent role not configured.")
		return
	}

	user := getUserOption(i, "user")
	if user == nil {
		respondEphemeral(s, i, "Please specify a user.")
		return
	}

	// Remove silent role
	err = s.GuildMemberRoleRemove(i.GuildID, user.ID, cfg.SilentRoleID)
	if err != nil {
		respondEphemeral(s, i, "Failed to unsilence user: "+err.Error())
		return
	}

	// Remove scheduled unsilence
	ch.bot.DB.DeleteScheduledEventByTarget(i.GuildID, "unsilence", user.ID)

	embed := successEmbed("User Unsilenced", fmt.Sprintf("Unsilenced %s", user.Mention()))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) getRaidHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isModerator(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need moderator permissions to view raid users.")
		return
	}

	cfg, _ := ch.bot.DB.GetAntiRaidConfig(i.GuildID)
	sinceTimestamp := time.Now().Add(-time.Duration(cfg.RaidTime*2) * time.Second).UnixMilli()

	joins, err := ch.bot.DB.GetRecentJoins(i.GuildID, sinceTimestamp)
	if err != nil || len(joins) == 0 {
		respondEphemeral(s, i, "No recent joins found.")
		return
	}

	var userList strings.Builder
	for idx, join := range joins {
		if idx >= 25 {
			userList.WriteString(fmt.Sprintf("\n... and %d more", len(joins)-25))
			break
		}
		accountAge := time.Since(time.UnixMilli(join.AccountCreatedAt))
		userList.WriteString(fmt.Sprintf("<@%s> (Account: %s old)\n", join.UserID, formatDuration(accountAge)))
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Recent Joins (Potential Raid)",
		Description: userList.String(),
		Color:       0xFF0000,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("%d users in the past %d seconds", len(joins), cfg.RaidTime*2),
		},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) banRaidHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to ban raid users.")
		return
	}

	respondDeferred(s, i)

	cfg, _ := ch.bot.DB.GetAntiRaidConfig(i.GuildID)
	sinceTimestamp := time.Now().Add(-time.Duration(cfg.RaidTime*2) * time.Second).UnixMilli()

	joins, err := ch.bot.DB.GetRecentJoins(i.GuildID, sinceTimestamp)
	if err != nil || len(joins) == 0 {
		followUp(s, i, "No recent joins found to ban.")
		return
	}

	reason := getStringOption(i, "reason")
	if reason == "" {
		reason = "Raid - banned by " + i.Member.User.Username
	}

	banned := 0
	failed := 0

	for _, join := range joins {
		err := s.GuildBanCreateWithReason(i.GuildID, join.UserID, reason, 1)
		if err != nil {
			failed++
		} else {
			banned++
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Raid Users Banned",
		Description: fmt.Sprintf("Banned **%d** users\nFailed: **%d**", banned, failed),
		Color:       0xFF0000,
	}

	followUpEmbed(s, i, embed)
}

func (ch *CommandHandler) lockdownHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to toggle lockdown.")
		return
	}

	enable := getBoolOption(i, "enable")

	guild, err := s.Guild(i.GuildID)
	if err != nil {
		respondEphemeral(s, i, "Failed to get server information.")
		return
	}

	if enable {
		// Set verification level to High
		_, err = s.GuildEdit(i.GuildID, &discordgo.GuildParams{
			VerificationLevel: &[]discordgo.VerificationLevel{discordgo.VerificationLevelHigh}[0],
		})
		if err != nil {
			respondEphemeral(s, i, "Failed to enable lockdown: "+err.Error())
			return
		}

		embed := &discordgo.MessageEmbed{
			Title:       "Server Lockdown Enabled",
			Description: "Verification level raised to **High**\nNew members must wait 10 minutes before chatting.",
			Color:       0xFF0000,
		}
		respondEmbed(s, i, embed)
	} else {
		// Restore to medium or previous level
		newLevel := discordgo.VerificationLevelMedium
		if guild.VerificationLevel < discordgo.VerificationLevelHigh {
			respondEphemeral(s, i, "Server is not in lockdown.")
			return
		}

		_, err = s.GuildEdit(i.GuildID, &discordgo.GuildParams{
			VerificationLevel: &newLevel,
		})
		if err != nil {
			respondEphemeral(s, i, "Failed to disable lockdown: "+err.Error())
			return
		}

		embed := successEmbed("Lockdown Disabled", "Verification level restored to **Medium**")
		respondEmbed(s, i, embed)
	}
}
