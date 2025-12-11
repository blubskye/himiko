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

func (ch *CommandHandler) registerInfoCommands() {
	// User info
	ch.Register(&Command{
		Name:        "userinfo",
		Description: "Get information about a user",
		Category:    "Info",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "User to get info about",
				Required:    false,
			},
		},
		Handler: ch.userInfoHandler,
	})

	// Server info
	ch.Register(&Command{
		Name:        "serverinfo",
		Description: "Get information about the server",
		Category:    "Info",
		Handler:     ch.serverInfoHandler,
	})

	// Channel info
	ch.Register(&Command{
		Name:        "channelinfo",
		Description: "Get information about a channel",
		Category:    "Info",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "Channel to get info about",
				Required:    false,
			},
		},
		Handler: ch.channelInfoHandler,
	})

	// Role info
	ch.Register(&Command{
		Name:        "roleinfo",
		Description: "Get information about a role",
		Category:    "Info",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionRole,
				Name:        "role",
				Description: "Role to get info about",
				Required:    true,
			},
		},
		Handler: ch.roleInfoHandler,
	})

	// Emoji info
	ch.Register(&Command{
		Name:        "emojiinfo",
		Description: "Get information about an emoji",
		Category:    "Info",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "emoji",
				Description: "Emoji to get info about",
				Required:    true,
			},
		},
		Handler: ch.emojiInfoHandler,
	})

	// Bot info
	ch.Register(&Command{
		Name:        "botinfo",
		Description: "Get information about the bot",
		Category:    "Info",
		Handler:     ch.botInfoHandler,
	})

	// Invite info
	ch.Register(&Command{
		Name:        "inviteinfo",
		Description: "Get information about an invite",
		Category:    "Info",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "invite",
				Description: "Invite code or URL",
				Required:    true,
			},
		},
		Handler: ch.inviteInfoHandler,
	})

	// Roles list
	ch.Register(&Command{
		Name:        "roles",
		Description: "List all server roles",
		Category:    "Info",
		Handler:     ch.rolesHandler,
	})

	// Member count
	ch.Register(&Command{
		Name:        "membercount",
		Description: "Get server member count",
		Category:    "Info",
		Handler:     ch.memberCountHandler,
	})
}

func (ch *CommandHandler) userInfoHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := getUserOption(i, "user")
	if user == nil {
		user = i.Member.User
	}

	// Get full user data
	fullUser, err := s.User(user.ID)
	if err != nil {
		fullUser = user
	}

	// Get member data if in guild
	var member *discordgo.Member
	if i.GuildID != "" {
		member, _ = s.GuildMember(i.GuildID, user.ID)
	}

	// Parse timestamps
	createdAt, _ := discordgo.SnowflakeTimestamp(user.ID)

	embed := &discordgo.MessageEmbed{
		Title:     user.Username,
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: avatarURL(fullUser)},
		Color:     0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "ID", Value: user.ID, Inline: true},
			{Name: "Bot", Value: fmt.Sprintf("%t", user.Bot), Inline: true},
			{Name: "Created", Value: fmt.Sprintf("<t:%d:F>\n(<t:%d:R>)", createdAt.Unix(), createdAt.Unix()), Inline: false},
		},
	}

	if member != nil {
		joinedAt := member.JoinedAt

		// Roles
		var roleNames []string
		for _, roleID := range member.Roles {
			roleNames = append(roleNames, fmt.Sprintf("<@&%s>", roleID))
		}
		rolesStr := "None"
		if len(roleNames) > 0 {
			rolesStr = strings.Join(roleNames, ", ")
			if len(rolesStr) > 1024 {
				rolesStr = fmt.Sprintf("%d roles", len(roleNames))
			}
		}

		embed.Fields = append(embed.Fields,
			&discordgo.MessageEmbedField{Name: "Joined Server", Value: fmt.Sprintf("<t:%d:F>\n(<t:%d:R>)", joinedAt.Unix(), joinedAt.Unix()), Inline: false},
			&discordgo.MessageEmbedField{Name: fmt.Sprintf("Roles [%d]", len(member.Roles)), Value: rolesStr, Inline: false},
		)

		if member.Nick != "" {
			embed.Fields = append([]*discordgo.MessageEmbedField{
				{Name: "Nickname", Value: member.Nick, Inline: true},
			}, embed.Fields...)
		}
	}

	// Add banner if exists
	if bannerURL := bannerURL(fullUser); bannerURL != "" {
		embed.Image = &discordgo.MessageEmbedImage{URL: bannerURL}
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) serverInfoHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guild, err := s.Guild(i.GuildID)
	if err != nil {
		respondEphemeral(s, i, "Failed to fetch server info.")
		return
	}

	// Get owner
	owner, _ := s.User(guild.OwnerID)
	ownerStr := guild.OwnerID
	if owner != nil {
		ownerStr = owner.Username
	}

	// Parse creation time
	createdAt, _ := discordgo.SnowflakeTimestamp(guild.ID)

	// Count channels by type
	var textChannels, voiceChannels, categories int
	for _, ch := range guild.Channels {
		switch ch.Type {
		case discordgo.ChannelTypeGuildText, discordgo.ChannelTypeGuildNews:
			textChannels++
		case discordgo.ChannelTypeGuildVoice, discordgo.ChannelTypeGuildStageVoice:
			voiceChannels++
		case discordgo.ChannelTypeGuildCategory:
			categories++
		}
	}

	// Verification level
	verificationLevels := map[discordgo.VerificationLevel]string{
		discordgo.VerificationLevelNone:    "None",
		discordgo.VerificationLevelLow:     "Low",
		discordgo.VerificationLevelMedium:  "Medium",
		discordgo.VerificationLevelHigh:    "High",
		discordgo.VerificationLevelVeryHigh: "Highest",
	}

	embed := &discordgo.MessageEmbed{
		Title:     guild.Name,
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: guildIconURL(guild)},
		Color:     0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "ID", Value: guild.ID, Inline: true},
			{Name: "Owner", Value: ownerStr, Inline: true},
			{Name: "Created", Value: fmt.Sprintf("<t:%d:F>", createdAt.Unix()), Inline: true},
			{Name: "Members", Value: fmt.Sprintf("%d", guild.MemberCount), Inline: true},
			{Name: "Roles", Value: fmt.Sprintf("%d", len(guild.Roles)), Inline: true},
			{Name: "Emojis", Value: fmt.Sprintf("%d", len(guild.Emojis)), Inline: true},
			{Name: "Channels", Value: fmt.Sprintf("Text: %d\nVoice: %d\nCategories: %d", textChannels, voiceChannels, categories), Inline: true},
			{Name: "Verification", Value: verificationLevels[guild.VerificationLevel], Inline: true},
			{Name: "Boost Level", Value: fmt.Sprintf("Level %d (%d boosts)", guild.PremiumTier, guild.PremiumSubscriptionCount), Inline: true},
		},
	}

	if bannerURL := guildBannerURL(guild); bannerURL != "" {
		embed.Image = &discordgo.MessageEmbedImage{URL: bannerURL}
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) channelInfoHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	channel := getChannelOption(i, "channel")
	if channel == nil {
		var err error
		channel, err = s.Channel(i.ChannelID)
		if err != nil {
			respondEphemeral(s, i, "Failed to fetch channel info.")
			return
		}
	}

	createdAt, _ := discordgo.SnowflakeTimestamp(channel.ID)

	channelTypes := map[discordgo.ChannelType]string{
		discordgo.ChannelTypeGuildText:           "Text",
		discordgo.ChannelTypeDM:                  "DM",
		discordgo.ChannelTypeGuildVoice:          "Voice",
		discordgo.ChannelTypeGroupDM:             "Group DM",
		discordgo.ChannelTypeGuildCategory:       "Category",
		discordgo.ChannelTypeGuildNews:           "News",
		discordgo.ChannelTypeGuildStore:          "Store",
		discordgo.ChannelTypeGuildNewsThread:     "News Thread",
		discordgo.ChannelTypeGuildPublicThread:   "Public Thread",
		discordgo.ChannelTypeGuildPrivateThread:  "Private Thread",
		discordgo.ChannelTypeGuildStageVoice:     "Stage",
		discordgo.ChannelTypeGuildForum:          "Forum",
	}

	embed := &discordgo.MessageEmbed{
		Title: "#" + channel.Name,
		Color: 0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "ID", Value: channel.ID, Inline: true},
			{Name: "Type", Value: channelTypes[channel.Type], Inline: true},
			{Name: "Created", Value: fmt.Sprintf("<t:%d:F>", createdAt.Unix()), Inline: true},
		},
	}

	if channel.Topic != "" {
		embed.Description = channel.Topic
	}

	if channel.ParentID != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name: "Category", Value: fmt.Sprintf("<#%s>", channel.ParentID), Inline: true,
		})
	}

	if channel.NSFW {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name: "NSFW", Value: "Yes", Inline: true,
		})
	}

	if channel.RateLimitPerUser > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name: "Slowmode", Value: fmt.Sprintf("%ds", channel.RateLimitPerUser), Inline: true,
		})
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) roleInfoHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	role := getRoleOption(i, "role")
	if role == nil {
		respondEphemeral(s, i, "Please specify a role.")
		return
	}

	createdAt, _ := discordgo.SnowflakeTimestamp(role.ID)

	// Count members with this role
	members, _ := s.GuildMembers(i.GuildID, "", 1000)
	memberCount := 0
	for _, m := range members {
		for _, r := range m.Roles {
			if r == role.ID {
				memberCount++
				break
			}
		}
	}

	embed := &discordgo.MessageEmbed{
		Title: role.Name,
		Color: role.Color,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "ID", Value: role.ID, Inline: true},
			{Name: "Color", Value: fmt.Sprintf("#%06X", role.Color), Inline: true},
			{Name: "Position", Value: fmt.Sprintf("%d", role.Position), Inline: true},
			{Name: "Members", Value: fmt.Sprintf("%d", memberCount), Inline: true},
			{Name: "Mentionable", Value: fmt.Sprintf("%t", role.Mentionable), Inline: true},
			{Name: "Hoisted", Value: fmt.Sprintf("%t", role.Hoist), Inline: true},
			{Name: "Created", Value: fmt.Sprintf("<t:%d:F>", createdAt.Unix()), Inline: false},
		},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) emojiInfoHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	emojiStr := getStringOption(i, "emoji")

	// Parse emoji
	var emojiID, emojiName string
	var animated bool

	if strings.HasPrefix(emojiStr, "<a:") {
		animated = true
		parts := strings.Split(strings.Trim(emojiStr, "<>"), ":")
		if len(parts) >= 3 {
			emojiName = parts[1]
			emojiID = parts[2]
		}
	} else if strings.HasPrefix(emojiStr, "<:") {
		parts := strings.Split(strings.Trim(emojiStr, "<>"), ":")
		if len(parts) >= 3 {
			emojiName = parts[1]
			emojiID = parts[2]
		}
	} else {
		respondEphemeral(s, i, "Please provide a custom emoji.")
		return
	}

	if emojiID == "" {
		respondEphemeral(s, i, "Could not parse emoji.")
		return
	}

	createdAt, _ := discordgo.SnowflakeTimestamp(emojiID)

	ext := "png"
	if animated {
		ext = "gif"
	}
	url := fmt.Sprintf("https://cdn.discordapp.com/emojis/%s.%s", emojiID, ext)

	embed := &discordgo.MessageEmbed{
		Title:     emojiName,
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: url},
		Color:     0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "ID", Value: emojiID, Inline: true},
			{Name: "Animated", Value: fmt.Sprintf("%t", animated), Inline: true},
			{Name: "Created", Value: fmt.Sprintf("<t:%d:F>", createdAt.Unix()), Inline: true},
			{Name: "URL", Value: fmt.Sprintf("[Link](%s)", url), Inline: false},
		},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) botInfoHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guilds := len(s.State.Guilds)

	embed := &discordgo.MessageEmbed{
		Title:     "Himiko Bot",
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: avatarURL(s.State.User)},
		Color:     0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Servers", Value: fmt.Sprintf("%d", guilds), Inline: true},
			{Name: "Commands", Value: fmt.Sprintf("%d", len(ch.commands)), Inline: true},
			{Name: "Uptime", Value: formatDuration(time.Since(botStartTime)), Inline: true},
			{Name: "Library", Value: "discordgo", Inline: true},
			{Name: "Go Version", Value: "1.21+", Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Made with Go",
		},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) inviteInfoHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	inviteStr := getStringOption(i, "invite")

	// Extract invite code
	code := inviteStr
	code = strings.TrimPrefix(code, "https://discord.gg/")
	code = strings.TrimPrefix(code, "https://discord.com/invite/")
	code = strings.TrimPrefix(code, "discord.gg/")

	invite, err := s.InviteWithCounts(code)
	if err != nil {
		respondEphemeral(s, i, "Failed to fetch invite info.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: invite.Guild.Name,
		Color: 0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Server ID", Value: invite.Guild.ID, Inline: true},
			{Name: "Channel", Value: fmt.Sprintf("#%s", invite.Channel.Name), Inline: true},
			{Name: "Members", Value: fmt.Sprintf("Online: %d\nTotal: %d", invite.ApproximatePresenceCount, invite.ApproximateMemberCount), Inline: true},
		},
	}

	if invite.Inviter != nil {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name: "Inviter", Value: invite.Inviter.Username, Inline: true,
		})
	}

	if invite.Guild.Icon != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: fmt.Sprintf("https://cdn.discordapp.com/icons/%s/%s.png", invite.Guild.ID, invite.Guild.Icon),
		}
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) rolesHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guild, err := s.Guild(i.GuildID)
	if err != nil {
		respondEphemeral(s, i, "Failed to fetch server info.")
		return
	}

	// Sort roles by position (highest first)
	var roleList []string
	for _, role := range guild.Roles {
		if role.Name != "@everyone" {
			roleList = append(roleList, fmt.Sprintf("<@&%s>", role.ID))
		}
	}

	desc := strings.Join(roleList, ", ")
	if len(desc) > 4096 {
		desc = desc[:4093] + "..."
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Roles [%d]", len(guild.Roles)-1),
		Description: desc,
		Color:       0x5865F2,
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) memberCountHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guild, err := s.Guild(i.GuildID)
	if err != nil {
		respondEphemeral(s, i, "Failed to fetch server info.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: guild.Name + " Member Count",
		Color: 0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Total Members", Value: fmt.Sprintf("%d", guild.MemberCount), Inline: true},
		},
	}

	if guild.ApproximatePresenceCount > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name: "Online", Value: fmt.Sprintf("%d", guild.ApproximatePresenceCount), Inline: true,
		})
	}

	respondEmbed(s, i, embed)
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
