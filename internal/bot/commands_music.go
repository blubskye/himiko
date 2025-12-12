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
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/blubskye/himiko/internal/database"
	"github.com/bwmarrin/discordgo"
)

func (ch *CommandHandler) registerMusicCommands() {
	// Play command
	ch.Register(&Command{
		Name:        "play",
		Description: "Play a song from URL or search query",
		Category:    "Music",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "query",
				Description: "URL or search query",
				Required:    true,
			},
		},
		Handler: ch.playHandler,
	})

	// Skip command
	ch.Register(&Command{
		Name:        "skip",
		Description: "Skip the current track",
		Category:    "Music",
		Handler:     ch.skipHandler,
	})

	// Stop command
	ch.Register(&Command{
		Name:        "stop",
		Description: "Stop playback and clear the queue",
		Category:    "Music",
		Handler:     ch.stopHandler,
	})

	// Pause command
	ch.Register(&Command{
		Name:        "pause",
		Description: "Pause the current track",
		Category:    "Music",
		Handler:     ch.pauseHandler,
	})

	// Resume command
	ch.Register(&Command{
		Name:        "resume",
		Description: "Resume the paused track",
		Category:    "Music",
		Handler:     ch.resumeHandler,
	})

	// Queue command
	ch.Register(&Command{
		Name:        "queue",
		Description: "Show the current queue",
		Category:    "Music",
		Handler:     ch.queueHandler,
	})

	// Now Playing command
	ch.Register(&Command{
		Name:        "nowplaying",
		Description: "Show the currently playing track",
		Category:    "Music",
		Handler:     ch.nowPlayingHandler,
	})

	// Remove command
	ch.Register(&Command{
		Name:        "remove",
		Description: "Remove a track from the queue",
		Category:    "Music",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "position",
				Description: "Position in queue (1-based)",
				Required:    true,
				MinValue:    floatPtr(1),
			},
		},
		Handler: ch.removeHandler,
	})

	// Clear command
	ch.Register(&Command{
		Name:        "clear",
		Description: "Clear the queue",
		Category:    "Music",
		Handler:     ch.clearHandler,
	})

	// Move to Top command
	ch.Register(&Command{
		Name:        "movetop",
		Description: "Move a track to the top of the queue",
		Category:    "Music",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "position",
				Description: "Position in queue (1-based)",
				Required:    true,
				MinValue:    floatPtr(1),
			},
		},
		Handler: ch.moveTopHandler,
	})

	// Volume command
	ch.Register(&Command{
		Name:        "volume",
		Description: "Set the playback volume",
		Category:    "Music",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "level",
				Description: "Volume level (0-100)",
				Required:    true,
				MinValue:    floatPtr(0),
				MaxValue:    100,
			},
		},
		Handler: ch.volumeHandler,
	})

	// Join command
	ch.Register(&Command{
		Name:        "join",
		Description: "Join your voice channel",
		Category:    "Music",
		Handler:     ch.joinHandler,
	})

	// Leave command
	ch.Register(&Command{
		Name:        "leave",
		Description: "Leave the voice channel",
		Category:    "Music",
		Handler:     ch.leaveHandler,
	})

	// Music role settings
	ch.Register(&Command{
		Name:        "musicrole",
		Description: "Configure DJ and Mod roles for music commands",
		Category:    "Music",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "dj",
				Description: "Set the DJ role",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionRole,
						Name:        "role",
						Description: "The DJ role (leave empty to clear)",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "mod",
				Description: "Set the Mod role for music",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionRole,
						Name:        "role",
						Description: "The Mod role (leave empty to clear)",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "show",
				Description: "Show current music role settings",
			},
		},
		Handler: ch.musicRoleHandler,
	})

	// Local file library commands
	ch.Register(&Command{
		Name:        "folders",
		Description: "List available music folders",
		Category:    "Music",
		Handler:     ch.foldersHandler,
	})

	ch.Register(&Command{
		Name:        "files",
		Description: "List files in a folder",
		Category:    "Music",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "folder",
				Description: "Folder name",
				Required:    true,
			},
		},
		Handler:      ch.filesHandler,
		Autocomplete: ch.filesAutocomplete,
	})

	ch.Register(&Command{
		Name:        "local",
		Description: "Play a local file",
		Category:    "Music",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "file",
				Description: "File path (folder/filename)",
				Required:    true,
			},
		},
		Handler:      ch.localHandler,
		Autocomplete: ch.localAutocomplete,
	})

	ch.Register(&Command{
		Name:        "search",
		Description: "Search local music library",
		Category:    "Music",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "query",
				Description: "Search query",
				Required:    true,
			},
		},
		Handler: ch.searchLocalHandler,
	})

	ch.Register(&Command{
		Name:        "musicfolder",
		Description: "Set the music folder path",
		Category:    "Music",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "path",
				Description: "Absolute path to the music folder",
				Required:    true,
			},
		},
		Handler: ch.musicFolderHandler,
	})

	ch.Register(&Command{
		Name:        "musichistory",
		Description: "Show recently played tracks",
		Category:    "Music",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "count",
				Description: "Number of tracks to show (default 10)",
				Required:    false,
				MinValue:    floatPtr(1),
				MaxValue:    25,
			},
		},
		Handler: ch.musicHistoryHandler,
	})
}


func (ch *CommandHandler) playHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "This command can only be used in a server.")
		return
	}

	query := getStringOption(i, "query")
	if query == "" {
		respondEphemeral(s, i, "Please provide a URL or search query.")
		return
	}

	// Get user's voice channel
	channelID, err := GetUserVoiceChannel(s, i.GuildID, i.Member.User.ID)
	if err != nil {
		respondEphemeral(s, i, "You need to be in a voice channel to use this command.")
		return
	}

	respondDeferred(s, i)

	player := ch.bot.MusicManager.GetPlayer(i.GuildID)

	// Connect if not already connected
	if !player.IsConnected() {
		if err := player.Connect(s, channelID); err != nil {
			editResponse(s, i, "Failed to join voice channel: "+err.Error())
			return
		}
	}

	// Extract video info
	info, err := ExtractInfo(query)
	if err != nil {
		editResponse(s, i, "Failed to get track info: "+err.Error())
		return
	}

	track := &Track{
		Title:     info.Title,
		URL:       info.URL,
		Duration:  info.Duration,
		Thumbnail: info.Thumbnail,
		Requester: i.Member.User.Username,
		IsLocal:   false,
	}

	player.AddTrack(track)

	// Save to database queue
	var thumbnail *string
	if info.Thumbnail != "" {
		thumbnail = &info.Thumbnail
	}
	ch.bot.DB.AddToMusicQueue(&database.MusicQueueItem{
		GuildID:   i.GuildID,
		ChannelID: channelID,
		UserID:    i.Member.User.ID,
		Title:     info.Title,
		URL:       info.URL,
		Duration:  info.Duration,
		Thumbnail: thumbnail,
		IsLocal:   false,
	})

	// Start playing if not already
	if !player.IsPlaying() {
		if err := player.Play(); err != nil {
			editResponse(s, i, "Failed to start playback: "+err.Error())
			return
		}

		// Add to history
		ch.bot.DB.AddToMusicHistory(i.GuildID, i.Member.User.ID, info.Title, info.URL)

		embed := &discordgo.MessageEmbed{
			Title:       "Now Playing",
			Description: info.Title,
			Color:       0xFF69B4,
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Duration", Value: formatMusicDuration(info.Duration), Inline: true},
				{Name: "Requested by", Value: i.Member.User.Username, Inline: true},
			},
		}
		if info.Thumbnail != "" {
			embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: info.Thumbnail}
		}
		editResponseEmbed(s, i, embed)
	} else {
		queue := player.GetQueue()
		embed := &discordgo.MessageEmbed{
			Title:       "Added to Queue",
			Description: info.Title,
			Color:       0x5865F2,
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Position", Value: fmt.Sprintf("#%d", len(queue)), Inline: true},
				{Name: "Duration", Value: formatMusicDuration(info.Duration), Inline: true},
			},
		}
		if info.Thumbnail != "" {
			embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: info.Thumbnail}
		}
		editResponseEmbed(s, i, embed)
	}
}

func (ch *CommandHandler) skipHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "This command can only be used in a server.")
		return
	}

	player := ch.bot.MusicManager.GetPlayer(i.GuildID)

	settings, _ := ch.bot.DB.GetMusicSettings(i.GuildID)
	permLevel := GetMusicPermLevel(s, i.GuildID, i.Member.User.ID, settings.DJRoleID, settings.ModRoleID)

	// Check if user is in voice channel
	_, err := GetUserVoiceChannel(s, i.GuildID, i.Member.User.ID)
	if err != nil {
		respondEphemeral(s, i, "You need to be in a voice channel to use this command.")
		return
	}

	// Only DJ+ can skip others' tracks
	nowPlaying := player.NowPlaying()
	if nowPlaying != nil && nowPlaying.Requester != i.Member.User.Username && permLevel < MusicPermDJ {
		respondEphemeral(s, i, "You need DJ role to skip other users' tracks.")
		return
	}

	if err := player.Skip(); err != nil {
		respondEphemeral(s, i, "Failed to skip: "+err.Error())
		return
	}

	respond(s, i, "⏭️ Skipped the current track.")
}

func (ch *CommandHandler) stopHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "This command can only be used in a server.")
		return
	}

	settings, _ := ch.bot.DB.GetMusicSettings(i.GuildID)
	permLevel := GetMusicPermLevel(s, i.GuildID, i.Member.User.ID, settings.DJRoleID, settings.ModRoleID)

	if permLevel < MusicPermDJ {
		respondEphemeral(s, i, "You need DJ role to stop playback.")
		return
	}

	player := ch.bot.MusicManager.GetPlayer(i.GuildID)
	player.Stop()
	player.ClearQueue()

	// Clear database queue
	ch.bot.DB.ClearMusicQueue(i.GuildID)

	respond(s, i, "⏹️ Stopped playback and cleared the queue.")
}

func (ch *CommandHandler) pauseHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "This command can only be used in a server.")
		return
	}

	player := ch.bot.MusicManager.GetPlayer(i.GuildID)

	if err := player.Pause(); err != nil {
		respondEphemeral(s, i, "Failed to pause: "+err.Error())
		return
	}

	respond(s, i, "⏸️ Paused playback.")
}

func (ch *CommandHandler) resumeHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "This command can only be used in a server.")
		return
	}

	player := ch.bot.MusicManager.GetPlayer(i.GuildID)

	if err := player.Resume(); err != nil {
		respondEphemeral(s, i, "Failed to resume: "+err.Error())
		return
	}

	respond(s, i, "▶️ Resumed playback.")
}

func (ch *CommandHandler) queueHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "This command can only be used in a server.")
		return
	}

	player := ch.bot.MusicManager.GetPlayer(i.GuildID)
	queue := player.GetQueue()
	nowPlaying := player.NowPlaying()

	if nowPlaying == nil && len(queue) == 0 {
		respondEphemeral(s, i, "The queue is empty.")
		return
	}

	var description strings.Builder

	if nowPlaying != nil {
		description.WriteString(fmt.Sprintf("**Now Playing:**\n🎵 %s [%s]\n\n", nowPlaying.Title, formatMusicDuration(nowPlaying.Duration)))
	}

	if len(queue) > 0 {
		description.WriteString("**Up Next:**\n")
		for idx, track := range queue {
			if idx >= 10 {
				description.WriteString(fmt.Sprintf("\n*...and %d more tracks*", len(queue)-10))
				break
			}
			description.WriteString(fmt.Sprintf("%d. %s [%s]\n", idx+1, track.Title, formatMusicDuration(track.Duration)))
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Music Queue",
		Description: description.String(),
		Color:       0x5865F2,
		Footer:      &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("%d tracks in queue", len(queue))},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) nowPlayingHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "This command can only be used in a server.")
		return
	}

	player := ch.bot.MusicManager.GetPlayer(i.GuildID)
	nowPlaying := player.NowPlaying()

	if nowPlaying == nil {
		respondEphemeral(s, i, "Nothing is currently playing.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Now Playing",
		Description: nowPlaying.Title,
		Color:       0xFF69B4,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Duration", Value: formatMusicDuration(nowPlaying.Duration), Inline: true},
			{Name: "Requested by", Value: nowPlaying.Requester, Inline: true},
		},
	}

	if nowPlaying.Thumbnail != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: nowPlaying.Thumbnail}
	}

	if player.IsPaused() {
		embed.Footer = &discordgo.MessageEmbedFooter{Text: "⏸️ Paused"}
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) removeHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "This command can only be used in a server.")
		return
	}

	position := int(getIntOption(i, "position"))

	player := ch.bot.MusicManager.GetPlayer(i.GuildID)
	queue := player.GetQueue()

	if position < 1 || position > len(queue) {
		respondEphemeral(s, i, fmt.Sprintf("Invalid position. Queue has %d tracks.", len(queue)))
		return
	}

	track := queue[position-1]

	settings, _ := ch.bot.DB.GetMusicSettings(i.GuildID)
	permLevel := GetMusicPermLevel(s, i.GuildID, i.Member.User.ID, settings.DJRoleID, settings.ModRoleID)

	// Check permission
	if track.Requester != i.Member.User.Username && permLevel < MusicPermDJ {
		respondEphemeral(s, i, "You can only remove your own tracks unless you have DJ role.")
		return
	}

	if err := player.RemoveTrack(position - 1); err != nil {
		respondEphemeral(s, i, "Failed to remove track: "+err.Error())
		return
	}

	respond(s, i, fmt.Sprintf("🗑️ Removed **%s** from the queue.", track.Title))
}

func (ch *CommandHandler) clearHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "This command can only be used in a server.")
		return
	}

	settings, _ := ch.bot.DB.GetMusicSettings(i.GuildID)
	permLevel := GetMusicPermLevel(s, i.GuildID, i.Member.User.ID, settings.DJRoleID, settings.ModRoleID)

	if permLevel < MusicPermDJ {
		respondEphemeral(s, i, "You need DJ role to clear the queue.")
		return
	}

	player := ch.bot.MusicManager.GetPlayer(i.GuildID)
	player.ClearQueue()
	ch.bot.DB.ClearMusicQueue(i.GuildID)

	respond(s, i, "🧹 Cleared the queue.")
}

func (ch *CommandHandler) moveTopHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "This command can only be used in a server.")
		return
	}

	position := int(getIntOption(i, "position"))

	player := ch.bot.MusicManager.GetPlayer(i.GuildID)
	queue := player.GetQueue()

	if position < 1 || position > len(queue) {
		respondEphemeral(s, i, fmt.Sprintf("Invalid position. Queue has %d tracks.", len(queue)))
		return
	}

	track := queue[position-1]

	settings, _ := ch.bot.DB.GetMusicSettings(i.GuildID)
	permLevel := GetMusicPermLevel(s, i.GuildID, i.Member.User.ID, settings.DJRoleID, settings.ModRoleID)

	if track.Requester != i.Member.User.Username && permLevel < MusicPermDJ {
		respondEphemeral(s, i, "You can only move your own tracks unless you have DJ role.")
		return
	}

	if err := player.MoveToTop(position - 1); err != nil {
		respondEphemeral(s, i, "Failed to move track: "+err.Error())
		return
	}

	respond(s, i, fmt.Sprintf("⬆️ Moved **%s** to the top of the queue.", track.Title))
}

func (ch *CommandHandler) volumeHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "This command can only be used in a server.")
		return
	}

	level := int(getIntOption(i, "level"))

	settings, _ := ch.bot.DB.GetMusicSettings(i.GuildID)
	permLevel := GetMusicPermLevel(s, i.GuildID, i.Member.User.ID, settings.DJRoleID, settings.ModRoleID)

	if permLevel < MusicPermDJ {
		respondEphemeral(s, i, "You need DJ role to change the volume.")
		return
	}

	player := ch.bot.MusicManager.GetPlayer(i.GuildID)
	if err := player.SetVolume(level); err != nil {
		respondEphemeral(s, i, "Failed to set volume: "+err.Error())
		return
	}

	ch.bot.DB.UpdateMusicVolume(i.GuildID, level)

	respond(s, i, fmt.Sprintf("🔊 Volume set to %d%%", level))
}

func (ch *CommandHandler) joinHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "This command can only be used in a server.")
		return
	}

	channelID, err := GetUserVoiceChannel(s, i.GuildID, i.Member.User.ID)
	if err != nil {
		respondEphemeral(s, i, "You need to be in a voice channel to use this command.")
		return
	}

	player := ch.bot.MusicManager.GetPlayer(i.GuildID)
	if err := player.Connect(s, channelID); err != nil {
		respondEphemeral(s, i, "Failed to join voice channel: "+err.Error())
		return
	}

	respond(s, i, "🔊 Joined your voice channel!")
}

func (ch *CommandHandler) leaveHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "This command can only be used in a server.")
		return
	}

	settings, _ := ch.bot.DB.GetMusicSettings(i.GuildID)
	permLevel := GetMusicPermLevel(s, i.GuildID, i.Member.User.ID, settings.DJRoleID, settings.ModRoleID)

	if permLevel < MusicPermDJ {
		respondEphemeral(s, i, "You need DJ role to make the bot leave.")
		return
	}

	ch.bot.MusicManager.RemovePlayer(i.GuildID)
	ch.bot.DB.ClearMusicQueue(i.GuildID)

	respond(s, i, "👋 Left the voice channel.")
}

func (ch *CommandHandler) musicRoleHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "This command can only be used in a server.")
		return
	}

	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to configure music roles.")
		return
	}

	subCmd := getSubcommandName(i)

	switch subCmd {
	case "dj":
		role := getRoleOption(i, "role")
		var roleID *string
		if role != nil {
			roleID = &role.ID
		}
		ch.bot.DB.UpdateMusicRoles(i.GuildID, roleID, nil)
		if role != nil {
			respond(s, i, fmt.Sprintf("✅ DJ role set to %s", role.Mention()))
		} else {
			respond(s, i, "✅ DJ role cleared.")
		}

	case "mod":
		role := getRoleOption(i, "role")
		var roleID *string
		if role != nil {
			roleID = &role.ID
		}
		ch.bot.DB.UpdateMusicRoles(i.GuildID, nil, roleID)
		if role != nil {
			respond(s, i, fmt.Sprintf("✅ Mod role set to %s", role.Mention()))
		} else {
			respond(s, i, "✅ Mod role cleared.")
		}

	case "show":
		settings, _ := ch.bot.DB.GetMusicSettings(i.GuildID)
		djRole := "Not set"
		modRole := "Not set"
		if settings.DJRoleID != nil && *settings.DJRoleID != "" {
			djRole = fmt.Sprintf("<@&%s>", *settings.DJRoleID)
		}
		if settings.ModRoleID != nil && *settings.ModRoleID != "" {
			modRole = fmt.Sprintf("<@&%s>", *settings.ModRoleID)
		}

		embed := &discordgo.MessageEmbed{
			Title: "Music Role Settings",
			Color: 0x5865F2,
			Fields: []*discordgo.MessageEmbedField{
				{Name: "DJ Role", Value: djRole, Inline: true},
				{Name: "Mod Role", Value: modRole, Inline: true},
				{Name: "Volume", Value: fmt.Sprintf("%d%%", settings.Volume), Inline: true},
			},
		}
		respondEmbed(s, i, embed)
	}
}

func (ch *CommandHandler) foldersHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "This command can only be used in a server.")
		return
	}

	settings, err := ch.bot.DB.GetMusicSettings(i.GuildID)
	if err != nil || settings.MusicFolder == nil || *settings.MusicFolder == "" {
		respondEphemeral(s, i, "No music folder configured. Use `/musicfolder` to set one.")
		return
	}

	entries, err := os.ReadDir(*settings.MusicFolder)
	if err != nil {
		respondEphemeral(s, i, "Failed to read music folder: "+err.Error())
		return
	}

	var folders []string
	for _, entry := range entries {
		if entry.IsDir() {
			folders = append(folders, entry.Name())
		}
	}

	if len(folders) == 0 {
		respondEphemeral(s, i, "No folders found in the music directory.")
		return
	}

	var description strings.Builder
	for _, folder := range folders {
		description.WriteString(fmt.Sprintf("📁 %s\n", folder))
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Music Folders",
		Description: description.String(),
		Color:       0x5865F2,
		Footer:      &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("%d folders", len(folders))},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) filesHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "This command can only be used in a server.")
		return
	}

	folderName := getStringOption(i, "folder")

	settings, err := ch.bot.DB.GetMusicSettings(i.GuildID)
	if err != nil || settings.MusicFolder == nil || *settings.MusicFolder == "" {
		respondEphemeral(s, i, "No music folder configured.")
		return
	}

	folderPath := filepath.Join(*settings.MusicFolder, folderName)
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		respondEphemeral(s, i, "Failed to read folder: "+err.Error())
		return
	}

	var files []string
	audioExts := map[string]bool{".mp3": true, ".wav": true, ".ogg": true, ".flac": true, ".m4a": true, ".opus": true}

	for _, entry := range entries {
		if !entry.IsDir() {
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if audioExts[ext] {
				files = append(files, entry.Name())
			}
		}
	}

	if len(files) == 0 {
		respondEphemeral(s, i, "No audio files found in this folder.")
		return
	}

	var description strings.Builder
	for idx, file := range files {
		if idx >= 20 {
			description.WriteString(fmt.Sprintf("\n*...and %d more files*", len(files)-20))
			break
		}
		description.WriteString(fmt.Sprintf("🎵 %s\n", file))
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Files in %s", folderName),
		Description: description.String(),
		Color:       0x5865F2,
		Footer:      &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("%d audio files", len(files))},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) filesAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	settings, err := ch.bot.DB.GetMusicSettings(i.GuildID)
	if err != nil || settings.MusicFolder == nil || *settings.MusicFolder == "" {
		respondAutocomplete(s, i, nil)
		return
	}

	entries, err := os.ReadDir(*settings.MusicFolder)
	if err != nil {
		respondAutocomplete(s, i, nil)
		return
	}

	var choices []*discordgo.ApplicationCommandOptionChoice
	for _, entry := range entries {
		if entry.IsDir() && len(choices) < 25 {
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  entry.Name(),
				Value: entry.Name(),
			})
		}
	}

	respondAutocomplete(s, i, choices)
}

func (ch *CommandHandler) localHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "This command can only be used in a server.")
		return
	}

	filePath := getStringOption(i, "file")

	settings, err := ch.bot.DB.GetMusicSettings(i.GuildID)
	if err != nil || settings.MusicFolder == nil || *settings.MusicFolder == "" {
		respondEphemeral(s, i, "No music folder configured.")
		return
	}

	channelID, err := GetUserVoiceChannel(s, i.GuildID, i.Member.User.ID)
	if err != nil {
		respondEphemeral(s, i, "You need to be in a voice channel to use this command.")
		return
	}

	fullPath := filepath.Join(*settings.MusicFolder, filePath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		respondEphemeral(s, i, "File not found: "+filePath)
		return
	}

	respondDeferred(s, i)

	player := ch.bot.MusicManager.GetPlayer(i.GuildID)

	if !player.IsConnected() {
		if err := player.Connect(s, channelID); err != nil {
			editResponse(s, i, "Failed to join voice channel: "+err.Error())
			return
		}
	}

	fileName := filepath.Base(fullPath)
	title := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	track := &Track{
		Title:     title,
		URL:       fullPath,
		Duration:  0,
		Thumbnail: "",
		Requester: i.Member.User.Username,
		IsLocal:   true,
	}

	player.AddTrack(track)
	ch.bot.DB.AddToMusicQueue(&database.MusicQueueItem{
		GuildID:   i.GuildID,
		ChannelID: channelID,
		UserID:    i.Member.User.ID,
		Title:     title,
		URL:       fullPath,
		Duration:  0,
		Thumbnail: nil,
		IsLocal:   true,
	})

	if !player.IsPlaying() {
		if err := player.Play(); err != nil {
			editResponse(s, i, "Failed to start playback: "+err.Error())
			return
		}

		ch.bot.DB.AddToMusicHistory(i.GuildID, i.Member.User.ID, title, fullPath)

		embed := &discordgo.MessageEmbed{
			Title:       "Now Playing (Local)",
			Description: title,
			Color:       0xFF69B4,
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Requested by", Value: i.Member.User.Username, Inline: true},
			},
		}
		editResponseEmbed(s, i, embed)
	} else {
		queue := player.GetQueue()
		embed := &discordgo.MessageEmbed{
			Title:       "Added to Queue (Local)",
			Description: title,
			Color:       0x5865F2,
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Position", Value: fmt.Sprintf("#%d", len(queue)), Inline: true},
			},
		}
		editResponseEmbed(s, i, embed)
	}
}

func (ch *CommandHandler) localAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	settings, err := ch.bot.DB.GetMusicSettings(i.GuildID)
	if err != nil || settings.MusicFolder == nil || *settings.MusicFolder == "" {
		respondAutocomplete(s, i, nil)
		return
	}

	input := ""
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Focused {
			input = opt.StringValue()
		}
	}

	var choices []*discordgo.ApplicationCommandOptionChoice
	audioExts := map[string]bool{".mp3": true, ".wav": true, ".ogg": true, ".flac": true, ".m4a": true, ".opus": true}

	// Walk through folders
	filepath.Walk(*settings.MusicFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !audioExts[ext] {
			return nil
		}

		relPath, _ := filepath.Rel(*settings.MusicFolder, path)
		if strings.Contains(strings.ToLower(relPath), strings.ToLower(input)) && len(choices) < 25 {
			displayName := relPath
			if len(displayName) > 100 {
				displayName = "..." + displayName[len(displayName)-97:]
			}
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  displayName,
				Value: relPath,
			})
		}
		return nil
	})

	respondAutocomplete(s, i, choices)
}

func (ch *CommandHandler) searchLocalHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "This command can only be used in a server.")
		return
	}

	query := strings.ToLower(getStringOption(i, "query"))

	settings, err := ch.bot.DB.GetMusicSettings(i.GuildID)
	if err != nil || settings.MusicFolder == nil || *settings.MusicFolder == "" {
		respondEphemeral(s, i, "No music folder configured.")
		return
	}

	var results []string
	audioExts := map[string]bool{".mp3": true, ".wav": true, ".ogg": true, ".flac": true, ".m4a": true, ".opus": true}

	filepath.Walk(*settings.MusicFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !audioExts[ext] {
			return nil
		}

		relPath, _ := filepath.Rel(*settings.MusicFolder, path)
		if strings.Contains(strings.ToLower(relPath), query) {
			results = append(results, relPath)
		}
		return nil
	})

	if len(results) == 0 {
		respondEphemeral(s, i, "No results found for: "+query)
		return
	}

	var description strings.Builder
	for idx, result := range results {
		if idx >= 15 {
			description.WriteString(fmt.Sprintf("\n*...and %d more results*", len(results)-15))
			break
		}
		description.WriteString(fmt.Sprintf("🎵 `%s`\n", result))
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Search Results",
		Description: description.String(),
		Color:       0x5865F2,
		Footer:      &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("%d results found • Use /local <path> to play", len(results))},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) musicFolderHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "This command can only be used in a server.")
		return
	}

	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to configure the music folder.")
		return
	}

	path := getStringOption(i, "path")

	// Verify path exists
	info, err := os.Stat(path)
	if err != nil {
		respondEphemeral(s, i, "Path does not exist: "+path)
		return
	}

	if !info.IsDir() {
		respondEphemeral(s, i, "Path is not a directory: "+path)
		return
	}

	settings, _ := ch.bot.DB.GetMusicSettings(i.GuildID)
	settings.MusicFolder = &path
	ch.bot.DB.SetMusicSettings(settings)

	respond(s, i, fmt.Sprintf("✅ Music folder set to: `%s`", path))
}

func (ch *CommandHandler) musicHistoryHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.GuildID == "" {
		respondEphemeral(s, i, "This command can only be used in a server.")
		return
	}

	count := int(getIntOption(i, "count"))
	if count == 0 {
		count = 10
	}

	history, err := ch.bot.DB.GetMusicHistory(i.GuildID, count)
	if err != nil || len(history) == 0 {
		respondEphemeral(s, i, "No playback history found.")
		return
	}

	var description strings.Builder
	for _, h := range history {
		description.WriteString(fmt.Sprintf("🎵 **%s**\n   <t:%d:R> by <@%s>\n", h.Title, h.PlayedAt.Unix(), h.UserID))
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Recently Played",
		Description: description.String(),
		Color:       0x5865F2,
		Footer:      &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("Showing last %d tracks", len(history))},
	}

	respondEmbed(s, i, embed)
}

func formatMusicDuration(seconds int) string {
	if seconds == 0 {
		return "Unknown"
	}

	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
	}
	return strconv.Itoa(minutes) + ":" + fmt.Sprintf("%02d", secs)
}
