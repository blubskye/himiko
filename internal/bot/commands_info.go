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
	"runtime"
	"strings"
	"time"

	"github.com/blubskye/himiko/internal/updater"
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

	// Stats command
	ch.Register(&Command{
		Name:        "stats",
		Description: "Show detailed bot statistics",
		Category:    "Info",
		Handler:     ch.statsHandler,
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

	// New users (from sweetiebot)
	ch.Register(&Command{
		Name:        "newusers",
		Description: "List the newest users to join the server",
		Category:    "Info",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "count",
				Description: "Number of users to show (max 30)",
				Required:    false,
			},
		},
		Handler: ch.newUsersHandler,
	})

	// AKA - alias lookup (from sweetiebot)
	ch.Register(&Command{
		Name:        "aka",
		Description: "List all known aliases for a user",
		Category:    "Info",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "User to look up aliases for",
				Required:    true,
			},
		},
		Handler: ch.akaHandler,
	})

	// Set timezone
	ch.Register(&Command{
		Name:        "settimezone",
		Description: "Set your timezone for time displays",
		Category:    "Info",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "timezone",
				Description: "Your timezone (e.g. America/New_York, Europe/London, Asia/Tokyo)",
				Required:    true,
			},
		},
		Handler: ch.setTimezoneHandler,
	})

	// Time command - show user's time
	ch.Register(&Command{
		Name:        "time",
		Description: "Show the current time for a user (if they have set their timezone)",
		Category:    "Info",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "User to show time for",
				Required:    false,
			},
		},
		Handler: ch.timeHandler,
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

	// Build username with bot tag
	usernameDisplay := user.Username
	if user.Bot {
		usernameDisplay += " [BOT]"
	}

	embed := &discordgo.MessageEmbed{
		Title:     usernameDisplay,
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: avatarURL(fullUser)},
		Color:     0xFF69B4,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "ID", Value: user.ID, Inline: true},
			{Name: "Created", Value: fmt.Sprintf("<t:%d:F>\n(<t:%d:R>)", createdAt.Unix(), createdAt.Unix()), Inline: false},
		},
	}

	// Add nickname if present
	if member != nil && member.Nick != "" {
		embed.Fields = append([]*discordgo.MessageEmbedField{
			{Name: "Nickname", Value: member.Nick, Inline: true},
		}, embed.Fields...)
	}

	// Get known aliases
	aliases, _ := ch.bot.DB.GetUserAliases(user.ID, 10)
	if len(aliases) > 0 {
		var aliasNames []string
		for _, a := range aliases {
			aliasNames = append(aliasNames, a.Alias)
		}
		aliasStr := strings.Join(aliasNames, ", ")
		if len(aliasStr) > 1024 {
			aliasStr = aliasStr[:1020] + "..."
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("Known Aliases [%d]", len(aliases)),
			Value: aliasStr,
		})
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
	}

	// Get user timezone
	tz, _ := ch.bot.DB.GetUserTimezone(user.ID)
	if tz != "" {
		loc, err := time.LoadLocation(tz)
		if err == nil {
			localTime := time.Now().In(loc).Format("Mon, 02 Jan 2006 15:04 MST")
			embed.Fields = append(embed.Fields,
				&discordgo.MessageEmbedField{Name: "Timezone", Value: tz, Inline: true},
				&discordgo.MessageEmbedField{Name: "Local Time", Value: localTime, Inline: true},
			)
		}
	}

	// Get user activity (last seen, first message)
	if i.GuildID != "" {
		activity, _ := ch.bot.DB.GetUserActivity(i.GuildID, user.ID)
		if activity != nil {
			if activity.LastSeen != nil {
				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					Name:   "Last Seen",
					Value:  fmt.Sprintf("<t:%d:R>", activity.LastSeen.Unix()),
					Inline: true,
				})
			}
			if activity.FirstMessage != nil {
				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					Name:   "First Message",
					Value:  fmt.Sprintf("<t:%d:F>", activity.FirstMessage.Unix()),
					Inline: true,
				})
			}
			if activity.MessageCount > 0 {
				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					Name:   "Messages",
					Value:  fmt.Sprintf("%d", activity.MessageCount),
					Inline: true,
				})
			}
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
		Title: "Himiko Bot",
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://raw.githubusercontent.com/blubskye/himiko/main/himiko.png",
		},
		Color: 0xFF69B4,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Servers", Value: fmt.Sprintf("%d", guilds), Inline: true},
			{Name: "Commands", Value: fmt.Sprintf("%d", len(ch.commands)), Inline: true},
			{Name: "Uptime", Value: formatDuration(time.Since(botStartTime)), Inline: true},
			{Name: "Library", Value: "discordgo", Inline: true},
			{Name: "Go Version", Value: "1.21+", Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Made with 💉 and obsessive love",
			IconURL: avatarURL(s.State.User),
		},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) statsHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Gather statistics
	guilds := len(s.State.Guilds)
	var totalMembers int
	var totalChannels int
	for _, guild := range s.State.Guilds {
		totalMembers += guild.MemberCount
		totalChannels += len(guild.Channels)
	}

	// Get database stats
	commandCount := len(ch.commands)

	// Memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	memUsed := float64(memStats.Alloc) / 1024 / 1024

	embed := &discordgo.MessageEmbed{
		Title:       "Himiko Statistics",
		Description: "*\"Let me show you what I can do~\"*",
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://raw.githubusercontent.com/blubskye/himiko/main/himiko.png",
		},
		Color: 0xFF69B4,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Servers", Value: fmt.Sprintf("%d", guilds), Inline: true},
			{Name: "Users", Value: fmt.Sprintf("%d", totalMembers), Inline: true},
			{Name: "Channels", Value: fmt.Sprintf("%d", totalChannels), Inline: true},
			{Name: "Commands", Value: fmt.Sprintf("%d", commandCount), Inline: true},
			{Name: "Uptime", Value: formatDuration(time.Since(botStartTime)), Inline: true},
			{Name: "Memory", Value: fmt.Sprintf("%.2f MB", memUsed), Inline: true},
			{Name: "Version", Value: "v" + updater.GetCurrentVersion(), Inline: true},
			{Name: "Go Version", Value: runtime.Version(), Inline: true},
			{Name: "Library", Value: "discordgo", Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    fmt.Sprintf("Started %s • Made with 💉 and obsessive love", botStartTime.Format("Jan 2, 2006 at 3:04 PM")),
			IconURL: avatarURL(s.State.User),
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

func (ch *CommandHandler) newUsersHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	count := int64(5)
	if opt := getIntOption(i, "count"); opt > 0 {
		count = opt
		if count > 30 {
			count = 30
		}
	}

	activities, err := ch.bot.DB.GetNewestMembers(i.GuildID, int(count))
	if err != nil || len(activities) == 0 {
		respondEphemeral(s, i, "No user activity data found. Users need to send messages first to be tracked.")
		return
	}

	var lines []string
	for _, a := range activities {
		user, err := s.User(a.UserID)
		username := a.UserID
		if err == nil {
			username = user.Username
		}

		joined := "Unknown"
		if a.FirstSeen != nil {
			joined = fmt.Sprintf("<t:%d:R>", a.FirstSeen.Unix())
		}

		lines = append(lines, fmt.Sprintf("**%s** - Joined: %s", username, joined))
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Newest %d Members", len(activities)),
		Description: strings.Join(lines, "\n"),
		Color:       0xFF69B4,
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) akaHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := getUserOption(i, "user")
	if user == nil {
		respondEphemeral(s, i, "Please specify a user.")
		return
	}

	aliases, err := ch.bot.DB.GetUserAliases(user.ID, 20)
	if err != nil || len(aliases) == 0 {
		respondEphemeral(s, i, fmt.Sprintf("No known aliases for **%s**.", user.Username))
		return
	}

	var usernameAliases, nicknameAliases []string
	for _, a := range aliases {
		if a.AliasType == "username" {
			usernameAliases = append(usernameAliases, a.Alias)
		} else {
			nicknameAliases = append(nicknameAliases, a.Alias)
		}
	}

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("Known Aliases for %s", user.Username),
		Color: 0xFF69B4,
	}

	if len(usernameAliases) > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "Usernames",
			Value: strings.Join(usernameAliases, ", "),
		})
	}

	if len(nicknameAliases) > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "Nicknames",
			Value: strings.Join(nicknameAliases, ", "),
		})
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) setTimezoneHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	tz := getStringOption(i, "timezone")
	if tz == "" {
		respondEphemeral(s, i, "Please provide a timezone.")
		return
	}

	// Validate timezone
	loc, err := time.LoadLocation(tz)
	if err != nil {
		respondEphemeral(s, i, fmt.Sprintf("Invalid timezone: `%s`\n\nExamples: `America/New_York`, `Europe/London`, `Asia/Tokyo`, `UTC`", tz))
		return
	}

	err = ch.bot.DB.SetUserTimezone(i.Member.User.ID, tz)
	if err != nil {
		respondEphemeral(s, i, "Failed to save timezone.")
		return
	}

	currentTime := time.Now().In(loc).Format("Mon, 02 Jan 2006 15:04 MST")
	respond(s, i, fmt.Sprintf("Your timezone has been set to **%s**.\nCurrent time: **%s**", tz, currentTime))
}

func (ch *CommandHandler) timeHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := getUserOption(i, "user")
	if user == nil {
		user = i.Member.User
	}

	tz, err := ch.bot.DB.GetUserTimezone(user.ID)
	if err != nil || tz == "" {
		if user.ID == i.Member.User.ID {
			respondEphemeral(s, i, "You haven't set your timezone. Use `/settimezone` to set it.")
		} else {
			respondEphemeral(s, i, fmt.Sprintf("**%s** hasn't set their timezone.", user.Username))
		}
		return
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		respondEphemeral(s, i, "Invalid timezone stored.")
		return
	}

	currentTime := time.Now().In(loc).Format("Monday, 02 Jan 2006 15:04:05 MST")
	respond(s, i, fmt.Sprintf("**%s**'s current time (%s):\n**%s**", user.Username, tz, currentTime))
}
