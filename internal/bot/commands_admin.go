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
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

func (ch *CommandHandler) registerAdminCommands() {
	// Kick command
	ch.Register(&Command{
		Name:        "kick",
		Description: "Kick a member from the server",
		Category:    "Administration",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "member",
				Description: "The member to kick",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "Reason for kick",
				Required:    false,
			},
		},
		Handler: ch.kickHandler,
	})

	// Ban command
	ch.Register(&Command{
		Name:        "ban",
		Description: "Ban a member from the server",
		Category:    "Administration",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "member",
				Description: "The member to ban",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "Reason for ban",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "delete_days",
				Description: "Days of messages to delete (0-7)",
				Required:    false,
				MinValue:    floatPtr(0),
				MaxValue:    7,
			},
		},
		Handler: ch.banHandler,
	})

	// Unban command
	ch.Register(&Command{
		Name:        "unban",
		Description: "Unban a user from the server",
		Category:    "Administration",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "user_id",
				Description: "The user ID to unban",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "Reason for unban",
				Required:    false,
			},
		},
		Handler: ch.unbanHandler,
	})

	// Timeout command
	ch.Register(&Command{
		Name:        "timeout",
		Description: "Timeout a member",
		Category:    "Administration",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "member",
				Description: "The member to timeout",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "minutes",
				Description: "Duration in minutes",
				Required:    true,
				MinValue:    floatPtr(1),
				MaxValue:    40320, // 28 days max
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "Reason for timeout",
				Required:    false,
			},
		},
		Handler: ch.timeoutHandler,
	})

	// Remove timeout command
	ch.Register(&Command{
		Name:        "untimeout",
		Description: "Remove timeout from a member",
		Category:    "Administration",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "member",
				Description: "The member to remove timeout from",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "Reason for removing timeout",
				Required:    false,
			},
		},
		Handler: ch.untimeoutHandler,
	})

	// Purge command
	ch.Register(&Command{
		Name:        "purge",
		Description: "Delete messages from the channel",
		Category:    "Administration",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "amount",
				Description: "Number of messages to delete (1-100)",
				Required:    true,
				MinValue:    floatPtr(1),
				MaxValue:    100,
			},
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "Only delete messages from this user",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "contains",
				Description: "Only delete messages containing this text",
				Required:    false,
			},
		},
		Handler: ch.purgeHandler,
	})

	// Slowmode command
	ch.Register(&Command{
		Name:        "slowmode",
		Description: "Set channel slowmode",
		Category:    "Administration",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "seconds",
				Description: "Slowmode delay in seconds (0 to disable)",
				Required:    true,
				MinValue:    floatPtr(0),
				MaxValue:    21600,
			},
		},
		Handler: ch.slowmodeHandler,
	})

	// Warn command
	ch.Register(&Command{
		Name:        "warn",
		Description: "Warn a member",
		Category:    "Administration",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "member",
				Description: "The member to warn",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "Reason for warning",
				Required:    true,
			},
		},
		Handler: ch.warnHandler,
	})

	// Warnings command
	ch.Register(&Command{
		Name:        "warnings",
		Description: "View warnings for a member",
		Category:    "Administration",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "member",
				Description: "The member to check",
				Required:    true,
			},
		},
		Handler: ch.warningsHandler,
	})

	// Clear warnings command
	ch.Register(&Command{
		Name:        "clearwarnings",
		Description: "Clear all warnings for a member",
		Category:    "Administration",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "member",
				Description: "The member to clear warnings for",
				Required:    true,
			},
		},
		Handler: ch.clearWarningsHandler,
	})

	// Lock channel
	ch.Register(&Command{
		Name:        "lock",
		Description: "Lock a channel (prevent @everyone from sending messages)",
		Category:    "Administration",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "Channel to lock (defaults to current)",
				Required:    false,
			},
		},
		Handler: ch.lockHandler,
	})

	// Unlock channel
	ch.Register(&Command{
		Name:        "unlock",
		Description: "Unlock a channel",
		Category:    "Administration",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "Channel to unlock (defaults to current)",
				Required:    false,
			},
		},
		Handler: ch.unlockHandler,
	})

	// Bans list
	ch.Register(&Command{
		Name:        "bans",
		Description: "List banned users",
		Category:    "Administration",
		Handler:     ch.bansHandler,
	})

	// Hackban (ban by ID)
	ch.Register(&Command{
		Name:        "hackban",
		Description: "Ban a user by ID (doesn't need to be in server)",
		Category:    "Administration",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "user_id",
				Description: "The user ID to ban",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "Reason for ban",
				Required:    false,
			},
		},
		Handler: ch.hackbanHandler,
	})

	// Softban
	ch.Register(&Command{
		Name:        "softban",
		Description: "Ban and immediately unban to delete messages",
		Category:    "Administration",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "member",
				Description: "The member to softban",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "Reason for softban",
				Required:    false,
			},
		},
		Handler: ch.softbanHandler,
	})

	// Mass add role
	ch.Register(&Command{
		Name:        "massrole",
		Description: "Add or remove a role from all members or members with a specific role",
		Category:    "Administration",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add",
				Description: "Add a role to members",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionRole,
						Name:        "role",
						Description: "The role to add",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionRole,
						Name:        "filter",
						Description: "Only add to members who have this role (optional)",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remove",
				Description: "Remove a role from members",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionRole,
						Name:        "role",
						Description: "The role to remove",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionRole,
						Name:        "filter",
						Description: "Only remove from members who have this role (optional)",
						Required:    false,
					},
				},
			},
		},
		Handler: ch.massRoleHandler,
	})

	// Channel lockdown (more restrictive than lock)
	ch.Register(&Command{
		Name:        "chanlockdown",
		Description: "Lockdown a channel (block messages AND reactions)",
		Category:    "Administration",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "Channel to lockdown (defaults to current)",
				Required:    false,
			},
		},
		Handler: ch.chanLockdownHandler,
	})

	// Channel unlock (restore from lockdown)
	ch.Register(&Command{
		Name:        "chanunlock",
		Description: "Remove lockdown from a channel (restore messages and reactions)",
		Category:    "Administration",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "Channel to unlock (defaults to current)",
				Required:    false,
			},
		},
		Handler: ch.chanUnlockHandler,
	})

	// Sync permissions command
	ch.Register(&Command{
		Name:        "syncperms",
		Description: "Sync permissions from one channel to other channels",
		Category:    "Administration",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "source",
				Description: "The channel to copy permissions from",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "target",
				Description: "Target channel (or specify category for all channels in it)",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "all_channels",
				Description: "Apply to all text channels in the server",
				Required:    false,
			},
		},
		Handler: ch.syncPermsHandler,
	})
}

func (ch *CommandHandler) kickHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !hasPermission(s, i.GuildID, i.Member.User.ID, discordgo.PermissionKickMembers) {
		respondEphemeral(s, i, "You don't have permission to kick members.")
		return
	}

	user := getUserOption(i, "member")
	reason := getStringOption(i, "reason")
	if reason == "" {
		reason = "No reason provided"
	}

	if user == nil {
		respondEphemeral(s, i, "Please specify a member to kick.")
		return
	}

	err := s.GuildMemberDeleteWithReason(i.GuildID, user.ID, reason)
	if err != nil {
		respondEphemeral(s, i, "Failed to kick member: "+err.Error())
		return
	}

	embed := successEmbed("Member Kicked",
		fmt.Sprintf("**%s** has been kicked.\n**Reason:** %s", user.Username, reason))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) banHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !hasPermission(s, i.GuildID, i.Member.User.ID, discordgo.PermissionBanMembers) {
		respondEphemeral(s, i, "You don't have permission to ban members.")
		return
	}

	user := getUserOption(i, "member")
	reason := getStringOption(i, "reason")
	deleteDays := int(getIntOption(i, "delete_days"))

	if reason == "" {
		reason = "No reason provided"
	}

	if user == nil {
		respondEphemeral(s, i, "Please specify a member to ban.")
		return
	}

	err := s.GuildBanCreateWithReason(i.GuildID, user.ID, reason, deleteDays)
	if err != nil {
		respondEphemeral(s, i, "Failed to ban member: "+err.Error())
		return
	}

	embed := successEmbed("Member Banned",
		fmt.Sprintf("**%s** has been banned.\n**Reason:** %s", user.Username, reason))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) unbanHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !hasPermission(s, i.GuildID, i.Member.User.ID, discordgo.PermissionBanMembers) {
		respondEphemeral(s, i, "You don't have permission to unban members.")
		return
	}

	userID := getStringOption(i, "user_id")

	err := s.GuildBanDelete(i.GuildID, userID)
	if err != nil {
		respondEphemeral(s, i, "Failed to unban user: "+err.Error())
		return
	}

	embed := successEmbed("User Unbanned", fmt.Sprintf("User <@%s> has been unbanned.", userID))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) timeoutHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !hasPermission(s, i.GuildID, i.Member.User.ID, discordgo.PermissionModerateMembers) {
		respondEphemeral(s, i, "You don't have permission to timeout members.")
		return
	}

	user := getUserOption(i, "member")
	minutes := getIntOption(i, "minutes")
	reason := getStringOption(i, "reason")

	if user == nil {
		respondEphemeral(s, i, "Please specify a member to timeout.")
		return
	}

	until := time.Now().Add(time.Duration(minutes) * time.Minute)

	err := s.GuildMemberTimeout(i.GuildID, user.ID, &until)
	if err != nil {
		respondEphemeral(s, i, "Failed to timeout member: "+err.Error())
		return
	}

	desc := fmt.Sprintf("**%s** has been timed out for %d minutes.", user.Username, minutes)
	if reason != "" {
		desc += fmt.Sprintf("\n**Reason:** %s", reason)
	}

	embed := successEmbed("Member Timed Out", desc)
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) untimeoutHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !hasPermission(s, i.GuildID, i.Member.User.ID, discordgo.PermissionModerateMembers) {
		respondEphemeral(s, i, "You don't have permission to remove timeouts.")
		return
	}

	user := getUserOption(i, "member")

	if user == nil {
		respondEphemeral(s, i, "Please specify a member.")
		return
	}

	err := s.GuildMemberTimeout(i.GuildID, user.ID, nil)
	if err != nil {
		respondEphemeral(s, i, "Failed to remove timeout: "+err.Error())
		return
	}

	embed := successEmbed("Timeout Removed",
		fmt.Sprintf("Timeout has been removed from **%s**.", user.Username))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) purgeHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !hasPermission(s, i.GuildID, i.Member.User.ID, discordgo.PermissionManageMessages) {
		respondEphemeral(s, i, "You don't have permission to manage messages.")
		return
	}

	amount := int(getIntOption(i, "amount"))
	filterUser := getUserOption(i, "user")
	contains := getStringOption(i, "contains")

	respondDeferredEphemeral(s, i)

	messages, err := s.ChannelMessages(i.ChannelID, amount+1, "", "", "")
	if err != nil {
		followUp(s, i, "Failed to fetch messages: "+err.Error())
		return
	}

	var toDelete []string
	for _, msg := range messages {
		if msg.ID == i.ID {
			continue
		}

		// Apply filters
		if filterUser != nil && msg.Author.ID != filterUser.ID {
			continue
		}
		if contains != "" && !containsWord(msg.Content, contains) {
			continue
		}

		// Can only bulk delete messages less than 14 days old
		msgTime, _ := discordgo.SnowflakeTimestamp(msg.ID)
		if time.Since(msgTime) > 14*24*time.Hour {
			continue
		}

		toDelete = append(toDelete, msg.ID)
		if len(toDelete) >= amount {
			break
		}
	}

	if len(toDelete) == 0 {
		followUp(s, i, "No messages found matching the criteria.")
		return
	}

	if len(toDelete) == 1 {
		err = s.ChannelMessageDelete(i.ChannelID, toDelete[0])
	} else {
		err = s.ChannelMessagesBulkDelete(i.ChannelID, toDelete)
	}

	if err != nil {
		followUp(s, i, "Failed to delete messages: "+err.Error())
		return
	}

	followUp(s, i, fmt.Sprintf("Successfully deleted %d messages.", len(toDelete)))
}

func (ch *CommandHandler) slowmodeHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !hasPermission(s, i.GuildID, i.Member.User.ID, discordgo.PermissionManageChannels) {
		respondEphemeral(s, i, "You don't have permission to manage channels.")
		return
	}

	seconds := int(getIntOption(i, "seconds"))

	_, err := s.ChannelEdit(i.ChannelID, &discordgo.ChannelEdit{
		RateLimitPerUser: &seconds,
	})
	if err != nil {
		respondEphemeral(s, i, "Failed to set slowmode: "+err.Error())
		return
	}

	if seconds == 0 {
		respond(s, i, "Slowmode has been disabled.")
	} else {
		respond(s, i, fmt.Sprintf("Slowmode set to %d seconds.", seconds))
	}
}

func (ch *CommandHandler) warnHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isModerator(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You don't have permission to warn members.")
		return
	}

	user := getUserOption(i, "member")
	reason := getStringOption(i, "reason")

	if user == nil {
		respondEphemeral(s, i, "Please specify a member to warn.")
		return
	}

	err := ch.bot.DB.AddWarning(i.GuildID, user.ID, i.Member.User.ID, reason)
	if err != nil {
		respondEphemeral(s, i, "Failed to add warning: "+err.Error())
		return
	}

	// Get total warnings
	warnings, _ := ch.bot.DB.GetWarnings(i.GuildID, user.ID)

	embed := successEmbed("Warning Issued",
		fmt.Sprintf("**%s** has been warned.\n**Reason:** %s\n**Total Warnings:** %d",
			user.Username, reason, len(warnings)))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) warningsHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := getUserOption(i, "member")

	if user == nil {
		respondEphemeral(s, i, "Please specify a member.")
		return
	}

	warnings, err := ch.bot.DB.GetWarnings(i.GuildID, user.ID)
	if err != nil {
		respondEphemeral(s, i, "Failed to fetch warnings: "+err.Error())
		return
	}

	if len(warnings) == 0 {
		respond(s, i, fmt.Sprintf("**%s** has no warnings.", user.Username))
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("Warnings for %s", user.Username),
		Color: 0xFEE75C,
	}

	for i, w := range warnings {
		reason := "No reason"
		if w.Reason != nil {
			reason = *w.Reason
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("#%d - <t:%s:R>", i+1, formatUnixTime(w.CreatedAt)),
			Value: fmt.Sprintf("**Reason:** %s\n**By:** <@%s>", reason, w.ModeratorID),
		})
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) clearWarningsHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isModerator(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You don't have permission to clear warnings.")
		return
	}

	user := getUserOption(i, "member")

	if user == nil {
		respondEphemeral(s, i, "Please specify a member.")
		return
	}

	err := ch.bot.DB.ClearWarnings(i.GuildID, user.ID)
	if err != nil {
		respondEphemeral(s, i, "Failed to clear warnings: "+err.Error())
		return
	}

	embed := successEmbed("Warnings Cleared",
		fmt.Sprintf("All warnings for **%s** have been cleared.", user.Username))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) lockHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !hasPermission(s, i.GuildID, i.Member.User.ID, discordgo.PermissionManageChannels) {
		respondEphemeral(s, i, "You don't have permission to manage channels.")
		return
	}

	channel := getChannelOption(i, "channel")
	channelID := i.ChannelID
	if channel != nil {
		channelID = channel.ID
	}

	// Get @everyone role (same ID as guild)
	err := s.ChannelPermissionSet(channelID, i.GuildID, discordgo.PermissionOverwriteTypeRole,
		0, discordgo.PermissionSendMessages)
	if err != nil {
		respondEphemeral(s, i, "Failed to lock channel: "+err.Error())
		return
	}

	respond(s, i, fmt.Sprintf("Channel <#%s> has been locked.", channelID))
}

func (ch *CommandHandler) unlockHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !hasPermission(s, i.GuildID, i.Member.User.ID, discordgo.PermissionManageChannels) {
		respondEphemeral(s, i, "You don't have permission to manage channels.")
		return
	}

	channel := getChannelOption(i, "channel")
	channelID := i.ChannelID
	if channel != nil {
		channelID = channel.ID
	}

	err := s.ChannelPermissionDelete(channelID, i.GuildID)
	if err != nil {
		respondEphemeral(s, i, "Failed to unlock channel: "+err.Error())
		return
	}

	respond(s, i, fmt.Sprintf("Channel <#%s> has been unlocked.", channelID))
}

func (ch *CommandHandler) bansHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !hasPermission(s, i.GuildID, i.Member.User.ID, discordgo.PermissionBanMembers) {
		respondEphemeral(s, i, "You don't have permission to view bans.")
		return
	}

	respondDeferred(s, i)

	bans, err := s.GuildBans(i.GuildID, 100, "", "")
	if err != nil {
		followUp(s, i, "Failed to fetch bans: "+err.Error())
		return
	}

	if len(bans) == 0 {
		followUp(s, i, "No banned users found.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("Banned Users (%d)", len(bans)),
		Color: 0xED4245,
	}

	var desc string
	for i, ban := range bans {
		if i >= 25 {
			desc += fmt.Sprintf("\n... and %d more", len(bans)-25)
			break
		}
		reason := "No reason"
		if ban.Reason != "" {
			reason = truncate(ban.Reason, 50)
		}
		desc += fmt.Sprintf("**%s** (`%s`) - %s\n", ban.User.Username, ban.User.ID, reason)
	}
	embed.Description = desc

	followUpEmbed(s, i, embed)
}

func (ch *CommandHandler) hackbanHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !hasPermission(s, i.GuildID, i.Member.User.ID, discordgo.PermissionBanMembers) {
		respondEphemeral(s, i, "You don't have permission to ban members.")
		return
	}

	userID := getStringOption(i, "user_id")
	reason := getStringOption(i, "reason")
	if reason == "" {
		reason = "No reason provided"
	}

	// Validate user ID
	if _, err := strconv.ParseInt(userID, 10, 64); err != nil {
		respondEphemeral(s, i, "Invalid user ID.")
		return
	}

	err := s.GuildBanCreateWithReason(i.GuildID, userID, reason, 0)
	if err != nil {
		respondEphemeral(s, i, "Failed to ban user: "+err.Error())
		return
	}

	embed := successEmbed("User Banned",
		fmt.Sprintf("User `%s` has been banned.\n**Reason:** %s", userID, reason))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) softbanHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !hasPermission(s, i.GuildID, i.Member.User.ID, discordgo.PermissionBanMembers) {
		respondEphemeral(s, i, "You don't have permission to ban members.")
		return
	}

	user := getUserOption(i, "member")
	reason := getStringOption(i, "reason")
	if reason == "" {
		reason = "Softban"
	}

	if user == nil {
		respondEphemeral(s, i, "Please specify a member.")
		return
	}

	// Ban with message deletion
	err := s.GuildBanCreateWithReason(i.GuildID, user.ID, reason, 7)
	if err != nil {
		respondEphemeral(s, i, "Failed to ban user: "+err.Error())
		return
	}

	// Immediately unban
	err = s.GuildBanDelete(i.GuildID, user.ID)
	if err != nil {
		respondEphemeral(s, i, "Failed to unban user after softban: "+err.Error())
		return
	}

	embed := successEmbed("Member Softbanned",
		fmt.Sprintf("**%s** has been softbanned (messages deleted).\n**Reason:** %s", user.Username, reason))
	respondEmbed(s, i, embed)
}

func floatPtr(f float64) *float64 {
	return &f
}

func (ch *CommandHandler) massRoleHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !hasPermission(s, i.GuildID, i.Member.User.ID, discordgo.PermissionManageRoles) {
		respondEphemeral(s, i, "You don't have permission to manage roles.")
		return
	}

	subCmd := i.ApplicationCommandData().Options[0].Name
	opts := i.ApplicationCommandData().Options[0].Options

	var roleID, filterRoleID string
	for _, opt := range opts {
		switch opt.Name {
		case "role":
			roleID = opt.RoleValue(s, i.GuildID).ID
		case "filter":
			filterRoleID = opt.RoleValue(s, i.GuildID).ID
		}
	}

	// Defer response since this can take a while
	respondDeferred(s, i)

	// Get all guild members
	members, err := s.GuildMembers(i.GuildID, "", 1000)
	if err != nil {
		editResponse(s, i, "Failed to get guild members: "+err.Error())
		return
	}

	var affected int
	var errors int

	for _, member := range members {
		// Skip bots
		if member.User.Bot {
			continue
		}

		// If filter role specified, check if member has it
		if filterRoleID != "" {
			hasFilter := false
			for _, r := range member.Roles {
				if r == filterRoleID {
					hasFilter = true
					break
				}
			}
			if !hasFilter {
				continue
			}
		}

		switch subCmd {
		case "add":
			// Check if they already have the role
			hasRole := false
			for _, r := range member.Roles {
				if r == roleID {
					hasRole = true
					break
				}
			}
			if !hasRole {
				err := s.GuildMemberRoleAdd(i.GuildID, member.User.ID, roleID)
				if err != nil {
					errors++
				} else {
					affected++
				}
			}

		case "remove":
			// Check if they have the role
			hasRole := false
			for _, r := range member.Roles {
				if r == roleID {
					hasRole = true
					break
				}
			}
			if hasRole {
				err := s.GuildMemberRoleRemove(i.GuildID, member.User.ID, roleID)
				if err != nil {
					errors++
				} else {
					affected++
				}
			}
		}
	}

	action := "added to"
	if subCmd == "remove" {
		action = "removed from"
	}

	msg := fmt.Sprintf("Role <@&%s> %s **%d** members.", roleID, action, affected)
	if errors > 0 {
		msg += fmt.Sprintf(" (%d errors)", errors)
	}

	embed := successEmbed("Mass Role Complete", msg)
	editResponseEmbed(s, i, embed)
}

func (ch *CommandHandler) chanLockdownHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !hasPermission(s, i.GuildID, i.Member.User.ID, discordgo.PermissionManageChannels) {
		respondEphemeral(s, i, "You don't have permission to manage channels.")
		return
	}

	channel := getChannelOption(i, "channel")
	channelID := i.ChannelID
	if channel != nil {
		channelID = channel.ID
	}

	// Deny send messages AND add reactions for @everyone
	err := s.ChannelPermissionSet(channelID, i.GuildID, discordgo.PermissionOverwriteTypeRole,
		0, discordgo.PermissionSendMessages|discordgo.PermissionAddReactions)
	if err != nil {
		respondEphemeral(s, i, "Failed to lockdown channel: "+err.Error())
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Channel Locked Down",
		Description: fmt.Sprintf("Channel <#%s> has been locked down.\n\n**Blocked:** Send Messages, Add Reactions", channelID),
		Color:       0xFF0000,
	}
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) chanUnlockHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !hasPermission(s, i.GuildID, i.Member.User.ID, discordgo.PermissionManageChannels) {
		respondEphemeral(s, i, "You don't have permission to manage channels.")
		return
	}

	channel := getChannelOption(i, "channel")
	channelID := i.ChannelID
	if channel != nil {
		channelID = channel.ID
	}

	// Remove the permission override entirely
	err := s.ChannelPermissionDelete(channelID, i.GuildID)
	if err != nil {
		respondEphemeral(s, i, "Failed to unlock channel: "+err.Error())
		return
	}

	embed := successEmbed("Channel Unlocked",
		fmt.Sprintf("Channel <#%s> lockdown has been lifted.\n\n**Restored:** Send Messages, Add Reactions", channelID))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) syncPermsHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !hasPermission(s, i.GuildID, i.Member.User.ID, discordgo.PermissionManageChannels) {
		respondEphemeral(s, i, "You don't have permission to manage channels.")
		return
	}

	sourceChannel := getChannelOption(i, "source")
	if sourceChannel == nil {
		respondEphemeral(s, i, "Please specify a source channel.")
		return
	}

	targetChannel := getChannelOption(i, "target")
	allChannels := getBoolOption(i, "all_channels")

	// Defer response since this may take a while
	respondDeferred(s, i)

	// Get the source channel with full details
	source, err := s.Channel(sourceChannel.ID)
	if err != nil {
		editResponse(s, i, "Failed to get source channel: "+err.Error())
		return
	}

	var targetChannels []*discordgo.Channel

	if allChannels {
		// Get all channels in the guild
		channels, err := s.GuildChannels(i.GuildID)
		if err != nil {
			editResponse(s, i, "Failed to get guild channels: "+err.Error())
			return
		}

		for _, ch := range channels {
			// Only sync to text channels, skip the source
			if (ch.Type == discordgo.ChannelTypeGuildText || ch.Type == discordgo.ChannelTypeGuildNews) && ch.ID != source.ID {
				targetChannels = append(targetChannels, ch)
			}
		}
	} else if targetChannel != nil {
		// Check if target is a category
		target, err := s.Channel(targetChannel.ID)
		if err != nil {
			editResponse(s, i, "Failed to get target channel: "+err.Error())
			return
		}

		if target.Type == discordgo.ChannelTypeGuildCategory {
			// Get all channels in this category
			channels, err := s.GuildChannels(i.GuildID)
			if err != nil {
				editResponse(s, i, "Failed to get guild channels: "+err.Error())
				return
			}

			for _, ch := range channels {
				if ch.ParentID == target.ID && ch.ID != source.ID {
					targetChannels = append(targetChannels, ch)
				}
			}
		} else {
			// Single channel
			targetChannels = append(targetChannels, target)
		}
	} else {
		editResponse(s, i, "Please specify a target channel, category, or use all_channels=true.")
		return
	}

	if len(targetChannels) == 0 {
		editResponse(s, i, "No target channels found to sync permissions to.")
		return
	}

	// Sync permissions
	var synced, errors int
	for _, target := range targetChannels {
		// Clear existing overwrites on target
		for _, overwrite := range target.PermissionOverwrites {
			err := s.ChannelPermissionDelete(target.ID, overwrite.ID)
			if err != nil {
				errors++
			}
		}

		// Copy overwrites from source
		for _, overwrite := range source.PermissionOverwrites {
			err := s.ChannelPermissionSet(target.ID, overwrite.ID, overwrite.Type, overwrite.Allow, overwrite.Deny)
			if err != nil {
				errors++
			}
		}
		synced++
	}

	msg := fmt.Sprintf("Synced permissions from <#%s> to **%d** channels.", source.ID, synced)
	if errors > 0 {
		msg += fmt.Sprintf(" (%d errors)", errors)
	}

	embed := successEmbed("Permissions Synced", msg)
	editResponseEmbed(s, i, embed)
}
