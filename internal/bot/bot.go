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
	"time"

	"github.com/blubskye/himiko/internal/config"
	"github.com/blubskye/himiko/internal/database"
	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	Session  *discordgo.Session
	Config   *config.Config
	DB       *database.DB
	Commands *CommandHandler
	stopChan chan struct{}
}

func New(cfg *config.Config, db *database.DB) (*Bot, error) {
	session, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		return nil, err
	}

	// Set intents
	session.Identify.Intents = discordgo.IntentsAll

	b := &Bot{
		Session:  session,
		Config:   cfg,
		DB:       db,
		stopChan: make(chan struct{}),
	}

	// Initialize command handler
	b.Commands = NewCommandHandler(b)

	// Register event handlers
	session.AddHandler(b.onReady)
	session.AddHandler(b.onInteractionCreate)
	session.AddHandler(b.onMessageCreate)
	session.AddHandler(b.onMessageDelete)
	session.AddHandler(b.onGuildMemberAdd)

	return b, nil
}

func (b *Bot) Start() error {
	if err := b.Session.Open(); err != nil {
		return err
	}

	// Register slash commands
	if err := b.Commands.RegisterCommands(); err != nil {
		log.Printf("Warning: Failed to register some commands: %v", err)
	}

	// Start background tasks
	go b.runScheduledTasks()

	return nil
}

func (b *Bot) Stop() {
	close(b.stopChan)

	// Unregister commands on shutdown (optional)
	// b.Commands.UnregisterCommands()

	b.Session.Close()
}

func (b *Bot) onReady(s *discordgo.Session, r *discordgo.Ready) {
	log.Printf("Logged in as %s#%s", r.User.Username, r.User.Discriminator)
	log.Printf("Connected to %d guilds", len(r.Guilds))

	// Set status
	s.UpdateGameStatus(0, "Use /help for commands")
}

func (b *Bot) onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		b.Commands.HandleSlashCommand(s, i)
	} else if i.Type == discordgo.InteractionApplicationCommandAutocomplete {
		b.Commands.HandleAutocomplete(s, i)
	}
}

func (b *Bot) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore bot messages
	if m.Author.Bot {
		return
	}

	// Check for AFK mentions
	b.checkAFKMentions(s, m)

	// Check if user is AFK and remove status
	b.checkAFKReturn(s, m)

	// Check keyword notifications
	b.checkKeywordNotifications(s, m)
}

func (b *Bot) onMessageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
	// Log deleted message for snipe command
	if m.BeforeDelete != nil && m.BeforeDelete.Content != "" {
		guildID := ""
		if m.GuildID != "" {
			guildID = m.GuildID
		}
		b.DB.LogDeletedMessage(guildID, m.ChannelID, m.BeforeDelete.Author.ID, m.BeforeDelete.Content)
	}
}

func (b *Bot) onGuildMemberAdd(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	// Send welcome message if configured
	settings, err := b.DB.GetGuildSettings(m.GuildID)
	if err != nil {
		return
	}

	if settings.WelcomeChannel != nil && settings.WelcomeMessage != nil {
		msg := *settings.WelcomeMessage
		// Replace placeholders
		msg = replacePlaceholders(msg, m.User, m.GuildID)
		s.ChannelMessageSend(*settings.WelcomeChannel, msg)
	}
}

func (b *Bot) checkAFKMentions(s *discordgo.Session, m *discordgo.MessageCreate) {
	for _, mention := range m.Mentions {
		afk, err := b.DB.GetAFK(mention.ID)
		if err != nil || afk == nil {
			continue
		}

		msg := mention.Username + " is AFK"
		if afk.Message != nil {
			msg += ": " + *afk.Message
		}
		msg += " (since <t:" + formatUnixTime(afk.SetAt) + ":R>)"

		s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference())
	}
}

func (b *Bot) checkAFKReturn(s *discordgo.Session, m *discordgo.MessageCreate) {
	afk, err := b.DB.GetAFK(m.Author.ID)
	if err != nil || afk == nil {
		return
	}

	b.DB.RemoveAFK(m.Author.ID)
	s.ChannelMessageSendReply(m.ChannelID, "Welcome back! I've removed your AFK status.", m.Reference())
}

func (b *Bot) checkKeywordNotifications(s *discordgo.Session, m *discordgo.MessageCreate) {
	notifications, err := b.DB.GetAllKeywordNotifications()
	if err != nil {
		return
	}

	for _, n := range notifications {
		// Don't notify user of their own messages
		if n.UserID == m.Author.ID {
			continue
		}

		// Check guild filter
		if n.GuildID != nil && *n.GuildID != m.GuildID {
			continue
		}

		// Check if keyword is in message
		if containsWord(m.Content, n.Keyword) {
			// Send DM notification
			channel, err := s.UserChannelCreate(n.UserID)
			if err != nil {
				continue
			}

			embed := &discordgo.MessageEmbed{
				Title:       "Keyword Alert: " + n.Keyword,
				Description: m.Content,
				Color:       0x5865F2,
				Fields: []*discordgo.MessageEmbedField{
					{Name: "Author", Value: m.Author.Username, Inline: true},
					{Name: "Channel", Value: "<#" + m.ChannelID + ">", Inline: true},
				},
				Timestamp: m.Timestamp.Format(time.RFC3339),
			}

			s.ChannelMessageSendEmbed(channel.ID, embed)
		}
	}
}

func (b *Bot) runScheduledTasks() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	cleanupTicker := time.NewTicker(1 * time.Hour)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-b.stopChan:
			return
		case <-ticker.C:
			b.processScheduledMessages()
			b.processReminders()
		case <-cleanupTicker.C:
			// Clean up old deleted messages (older than 24 hours)
			b.DB.CleanOldDeletedMessages(24 * time.Hour)
		}
	}
}

func (b *Bot) processScheduledMessages() {
	messages, err := b.DB.GetPendingScheduledMessages()
	if err != nil {
		return
	}

	for _, msg := range messages {
		b.Session.ChannelMessageSend(msg.ChannelID, msg.Message)
		b.DB.MarkScheduledMessageExecuted(msg.ID)
	}
}

func (b *Bot) processReminders() {
	reminders, err := b.DB.GetPendingReminders()
	if err != nil {
		return
	}

	for _, r := range reminders {
		b.Session.ChannelMessageSend(r.ChannelID, "<@"+r.UserID+"> Reminder: "+r.Message)
		b.DB.MarkReminderCompleted(r.ID)
	}
}
