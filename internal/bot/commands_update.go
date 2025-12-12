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
	"time"

	"github.com/blubskye/himiko/internal/updater"
	"github.com/bwmarrin/discordgo"
)

func (ch *CommandHandler) registerUpdateCommands() {
	ch.Register(&Command{
		Name:        "update",
		Description: "Check for and apply bot updates (Owner only)",
		Category:    "Admin",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "check",
				Description: "Check if a new version is available",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "apply",
				Description: "Download and apply the latest update",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "version",
				Description: "Show current version",
			},
		},
		Handler: ch.updateHandler,
	})
}

func (ch *CommandHandler) updateHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Owner only
	if i.Member == nil || !ch.bot.Config.IsOwner(i.Member.User.ID) {
		respondEphemeral(s, i, "This command is only available to bot owners.")
		return
	}

	subCmd := getSubcommandName(i)

	switch subCmd {
	case "check":
		ch.updateCheckHandler(s, i)
	case "apply":
		ch.updateApplyHandler(s, i)
	case "version":
		ch.updateVersionHandler(s, i)
	default:
		respondEphemeral(s, i, "Unknown subcommand.")
	}
}

func (ch *CommandHandler) updateCheckHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	respondDeferred(s, i)

	info, err := updater.CheckForUpdateByPattern()
	if err != nil {
		editResponse(s, i, "Failed to check for updates: "+err.Error())
		return
	}

	if !info.Available {
		embed := &discordgo.MessageEmbed{
			Title:       "No Updates Available",
			Description: fmt.Sprintf("You are running the latest version (**v%s**).", info.CurrentVersion),
			Color:       0x57F287,
		}
		editResponseEmbed(s, i, embed)
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Update Available!",
		Description: fmt.Sprintf("A new version is available: **v%s** (current: v%s)", info.NewVersion, info.CurrentVersion),
		Color:       0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Download Size",
				Value: formatBytes(info.Size),
			},
		},
	}

	if info.ReleaseNotes != "" {
		notes := info.ReleaseNotes
		if len(notes) > 1000 {
			notes = notes[:1000] + "..."
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "Release Notes",
			Value: notes,
		})
	}

	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "Use /update apply to download and install",
	}

	editResponseEmbed(s, i, embed)
}

func (ch *CommandHandler) updateApplyHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	respondDeferred(s, i)

	// Check for update
	info, err := updater.CheckForUpdateByPattern()
	if err != nil {
		editResponse(s, i, "Failed to check for updates: "+err.Error())
		return
	}

	if !info.Available {
		editResponse(s, i, "No updates available. You are running the latest version.")
		return
	}

	// Update status message
	editResponse(s, i, fmt.Sprintf("Downloading update v%s (%s)...", info.NewVersion, formatBytes(info.Size)))

	// Download update
	var lastUpdate time.Time
	zipPath, err := updater.DownloadUpdate(info, func(downloaded, total int64) {
		// Update progress every 2 seconds
		if time.Since(lastUpdate) > 2*time.Second {
			lastUpdate = time.Now()
			percent := float64(downloaded) / float64(total) * 100
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: strPtr(fmt.Sprintf("Downloading update v%s... %.1f%% (%s / %s)",
					info.NewVersion, percent, formatBytes(downloaded), formatBytes(total))),
			})
		}
	})
	if err != nil {
		editResponse(s, i, "Failed to download update: "+err.Error())
		return
	}

	editResponse(s, i, "Download complete. Applying update...")

	// Apply update
	if err := updater.ApplyUpdate(zipPath); err != nil {
		editResponse(s, i, "Failed to apply update: "+err.Error())
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Update Applied Successfully!",
		Description: fmt.Sprintf("Updated from v%s to v%s\n\n**Relaunching bot with new version...**", info.CurrentVersion, info.NewVersion),
		Color:       0x57F287,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Bot is relaunching automatically",
		},
	}

	editResponseEmbed(s, i, embed)

	// Give Discord a moment to receive the response
	time.Sleep(2 * time.Second)

	// Relaunch the bot with the new executable
	if err := updater.RelaunchAfterUpdate(); err != nil {
		// If relaunch fails, log it and notify the user
		fmt.Printf("[Update] Failed to relaunch: %v\n", err)
		followUp(s, i, fmt.Sprintf("Failed to auto-relaunch: %v\nPlease restart the bot manually.", err))
	}
}

func (ch *CommandHandler) updateVersionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	embed := &discordgo.MessageEmbed{
		Title: "Himiko Version Info",
		Color: 0xFF69B4,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Current Version",
				Value:  "v" + updater.GetCurrentVersion(),
				Inline: true,
			},
			{
				Name:   "Auto-Update Check",
				Value:  boolToEnabled(ch.bot.Config.Features.AutoUpdate),
				Inline: true,
			},
			{
				Name:   "Auto-Apply Updates",
				Value:  boolToEnabled(ch.bot.Config.Features.AutoUpdateApply),
				Inline: true,
			},
		},
	}

	respondEmbed(s, i, embed)
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func boolToEnabled(b bool) string {
	if b {
		return "Enabled"
	}
	return "Disabled"
}


// CheckAndNotifyUpdate checks for updates and notifies the owner via DM
func (b *Bot) CheckAndNotifyUpdate() {
	b.checkForUpdates(true)
}

// StartPeriodicUpdateCheck starts a background goroutine that periodically checks for updates
func (b *Bot) StartPeriodicUpdateCheck() {
	if b.Config.Features.UpdateCheckHours <= 0 {
		return
	}

	go func() {
		ticker := time.NewTicker(time.Duration(b.Config.Features.UpdateCheckHours) * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-b.stopChan:
				return
			case <-ticker.C:
				b.checkForUpdates(false)
			}
		}
	}()
}

// checkForUpdates performs the actual update check
func (b *Bot) checkForUpdates(isStartup bool) {
	if !b.Config.Features.AutoUpdate {
		return
	}

	info, err := updater.CheckForUpdateByPattern()
	if err != nil {
		fmt.Printf("[Update] Failed to check for updates: %v\n", err)
		return
	}

	if !info.Available {
		if isStartup {
			fmt.Printf("[Update] Running latest version (v%s)\n", info.CurrentVersion)
		}
		return
	}

	fmt.Printf("[Update] Update available: v%s -> v%s\n", info.CurrentVersion, info.NewVersion)

	// If auto-apply is enabled, download and apply
	if b.Config.Features.AutoUpdateApply {
		fmt.Println("[Update] Auto-applying update...")
		zipPath, err := updater.DownloadUpdate(info, nil)
		if err != nil {
			fmt.Printf("[Update] Failed to download update: %v\n", err)
			return
		}

		if err := updater.ApplyUpdate(zipPath); err != nil {
			fmt.Printf("[Update] Failed to apply update: %v\n", err)
			return
		}

		fmt.Println("[Update] Update applied! Relaunching with new version...")

		// Notify via channel if configured
		b.sendUpdateNotification(info, true)

		// Notify owners via DM
		b.notifyOwnersDM(&discordgo.MessageEmbed{
			Title:       "Himiko Auto-Updated!",
			Description: fmt.Sprintf("Updated from v%s to v%s\n\n**Bot is relaunching automatically...**", info.CurrentVersion, info.NewVersion),
			Color:       0x57F287,
		})

		// Give Discord a moment to send notifications
		time.Sleep(2 * time.Second)

		// Relaunch the bot with the new executable
		fmt.Println("[Update] Relaunching...")
		if err := updater.RelaunchAfterUpdate(); err != nil {
			fmt.Printf("[Update] Failed to relaunch: %v\n", err)
			fmt.Println("[Update] Exiting for manual restart...")
			os.Exit(0)
		}
	} else {
		// Notify via channel if configured
		b.sendUpdateNotification(info, false)

		// Notify owners via DM
		b.notifyOwnersDM(&discordgo.MessageEmbed{
			Title:       "Himiko Update Available!",
			Description: fmt.Sprintf("A new version is available: **v%s** (current: v%s)\n\nUse `/update apply` to download and install.", info.NewVersion, info.CurrentVersion),
			Color:       0x5865F2,
		})
	}
}

// notifyOwnersDM sends a DM to all configured bot owners
func (b *Bot) notifyOwnersDM(embed *discordgo.MessageEmbed) {
	// Collect all owner IDs
	ownerIDs := make(map[string]bool)
	if b.Config.OwnerID != "" {
		ownerIDs[b.Config.OwnerID] = true
	}
	for _, id := range b.Config.OwnerIDs {
		ownerIDs[id] = true
	}

	// Send DM to each owner
	for ownerID := range ownerIDs {
		channel, err := b.Session.UserChannelCreate(ownerID)
		if err == nil {
			b.Session.ChannelMessageSendEmbed(channel.ID, embed)
		}
	}
}

// sendUpdateNotification sends an update notification to the configured channel
func (b *Bot) sendUpdateNotification(info *updater.UpdateInfo, applied bool) {
	if b.Config.Features.UpdateNotifyChannel == "" {
		return
	}

	var embed *discordgo.MessageEmbed
	if applied {
		embed = &discordgo.MessageEmbed{
			Title:       "Himiko Update Applied!",
			Description: fmt.Sprintf("Himiko has been updated from **v%s** to **v%s**.\n\nThe bot will use the new version after restart.", info.CurrentVersion, info.NewVersion),
			Color:       0x57F287,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "https://raw.githubusercontent.com/blubskye/himiko/main/himiko.png",
			},
		}
	} else {
		embed = &discordgo.MessageEmbed{
			Title:       "Himiko Update Available!",
			Description: fmt.Sprintf("A new version of Himiko is available!\n\n**Current:** v%s\n**New:** v%s", info.CurrentVersion, info.NewVersion),
			Color:       0x5865F2,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "https://raw.githubusercontent.com/blubskye/himiko/main/himiko.png",
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Bot owner can use /update apply to install",
			},
		}
	}

	if info.ReleaseNotes != "" {
		notes := info.ReleaseNotes
		if len(notes) > 500 {
			notes = notes[:500] + "..."
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "What's New",
			Value: notes,
		})
	}

	b.Session.ChannelMessageSendEmbed(b.Config.Features.UpdateNotifyChannel, embed)
}
