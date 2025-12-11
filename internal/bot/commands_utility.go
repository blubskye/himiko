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
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func (ch *CommandHandler) registerUtilityCommands() {
	// Ping
	ch.Register(&Command{
		Name:        "ping",
		Description: "Check bot latency",
		Category:    "Utility",
		Handler:     ch.pingHandler,
	})

	// Snipe
	ch.Register(&Command{
		Name:        "snipe",
		Description: "Retrieve recently deleted messages",
		Category:    "Utility",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "amount",
				Description: "Number of messages to retrieve (1-15)",
				Required:    false,
				MinValue:    floatPtr(1),
				MaxValue:    15,
			},
		},
		Handler: ch.snipeHandler,
	})

	// AFK
	ch.Register(&Command{
		Name:        "afk",
		Description: "Set your AFK status",
		Category:    "Utility",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "message",
				Description: "Your AFK message",
				Required:    false,
			},
		},
		Handler: ch.afkHandler,
	})

	// Remind
	ch.Register(&Command{
		Name:        "remind",
		Description: "Set a reminder",
		Category:    "Utility",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "time",
				Description: "When to remind (e.g., 1h30m, 2d)",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "message",
				Description: "What to remind you about",
				Required:    true,
			},
		},
		Handler: ch.remindHandler,
	})

	// Schedule
	ch.Register(&Command{
		Name:        "schedule",
		Description: "Schedule a message",
		Category:    "Utility",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "time",
				Description: "When to send (e.g., 1h30m, 2d)",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "message",
				Description: "Message to send",
				Required:    true,
			},
		},
		Handler: ch.scheduleHandler,
	})

	// Poll
	ch.Register(&Command{
		Name:        "poll",
		Description: "Create a poll",
		Category:    "Utility",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "question",
				Description: "The poll question",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "options",
				Description: "Options separated by | (max 10)",
				Required:    false,
			},
		},
		Handler: ch.pollHandler,
	})

	// Embed builder
	ch.Register(&Command{
		Name:        "embed",
		Description: "Create a custom embed",
		Category:    "Utility",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "title",
				Description: "Embed title",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "description",
				Description: "Embed description",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "color",
				Description: "Hex color (e.g., #FF0000)",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "image",
				Description: "Image URL",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "thumbnail",
				Description: "Thumbnail URL",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "footer",
				Description: "Footer text",
				Required:    false,
			},
		},
		Handler: ch.embedHandler,
	})

	// Clean (delete your messages)
	ch.Register(&Command{
		Name:        "clean",
		Description: "Delete your own messages",
		Category:    "Utility",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "amount",
				Description: "Number of messages to delete (1-100)",
				Required:    true,
				MinValue:    floatPtr(1),
				MaxValue:    100,
			},
		},
		Handler: ch.cleanHandler,
	})

	// First message
	ch.Register(&Command{
		Name:        "firstmessage",
		Description: "Get the first message in the channel",
		Category:    "Utility",
		Handler:     ch.firstMessageHandler,
	})

	// Uptime
	ch.Register(&Command{
		Name:        "uptime",
		Description: "Check bot uptime",
		Category:    "Utility",
		Handler:     ch.uptimeHandler,
	})

	// Say
	ch.Register(&Command{
		Name:        "say",
		Description: "Make the bot say something",
		Category:    "Utility",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "message",
				Description: "Message to say",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "Channel to send to (default: current)",
				Required:    false,
			},
		},
		Handler: ch.sayHandler,
	})

	// Steal emoji
	ch.Register(&Command{
		Name:        "stealemoji",
		Description: "Add an emoji to this server",
		Category:    "Utility",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "emoji",
				Description: "Emoji to steal (paste the emoji)",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "name",
				Description: "Name for the emoji",
				Required:    false,
			},
		},
		Handler: ch.stealEmojiHandler,
	})

	// Math
	ch.Register(&Command{
		Name:        "math",
		Description: "Simple math evaluation",
		Category:    "Utility",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "expression",
				Description: "Math expression (e.g., 2+2, 10*5)",
				Required:    true,
			},
		},
		Handler: ch.mathHandler,
	})
}

var botStartTime = time.Now()

func (ch *CommandHandler) pingHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	start := time.Now()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Pinging...",
		},
	})

	latency := time.Since(start).Milliseconds()
	wsLatency := s.HeartbeatLatency().Milliseconds()

	embed := &discordgo.MessageEmbed{
		Title: "Pong!",
		Color: 0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "API Latency", Value: fmt.Sprintf("%dms", latency), Inline: true},
			{Name: "WebSocket", Value: fmt.Sprintf("%dms", wsLatency), Inline: true},
		},
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: strPtr(""),
		Embeds:  &[]*discordgo.MessageEmbed{embed},
	})
}

func (ch *CommandHandler) snipeHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	amount := getIntOption(i, "amount")
	if amount == 0 {
		amount = 1
	}

	messages, err := ch.bot.DB.GetDeletedMessages(i.ChannelID, int(amount))
	if err != nil || len(messages) == 0 {
		respondEphemeral(s, i, "No deleted messages found in this channel.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: "Sniped Messages",
		Color: 0x5865F2,
	}

	for _, msg := range messages {
		user, _ := s.User(msg.UserID)
		username := msg.UserID
		if user != nil {
			username = user.Username
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("%s - <t:%s:R>", username, formatUnixTime(msg.DeletedAt)),
			Value: truncate(msg.Content, 1024),
		})
	}

	respondEmbedEphemeral(s, i, embed)
}

func (ch *CommandHandler) afkHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	message := getStringOption(i, "message")

	if message == "" {
		message = "AFK"
	}

	err := ch.bot.DB.SetAFK(i.Member.User.ID, message)
	if err != nil {
		respondEphemeral(s, i, "Failed to set AFK status.")
		return
	}

	respond(s, i, fmt.Sprintf("You are now AFK: %s", message))
}

func (ch *CommandHandler) remindHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	timeStr := getStringOption(i, "time")
	message := getStringOption(i, "message")

	duration, err := parseDuration(timeStr)
	if err != nil || duration <= 0 {
		respondEphemeral(s, i, "Invalid time format. Use format like: 1h30m, 2d, 30m")
		return
	}

	remindAt := time.Now().Add(duration)

	err = ch.bot.DB.AddReminder(i.Member.User.ID, i.ChannelID, message, remindAt)
	if err != nil {
		respondEphemeral(s, i, "Failed to set reminder.")
		return
	}

	embed := successEmbed("Reminder Set",
		fmt.Sprintf("I'll remind you <t:%d:R>\n**Message:** %s", remindAt.Unix(), message))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) scheduleHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	timeStr := getStringOption(i, "time")
	message := getStringOption(i, "message")

	duration, err := parseDuration(timeStr)
	if err != nil || duration <= 0 {
		respondEphemeral(s, i, "Invalid time format. Use format like: 1h30m, 2d, 30m")
		return
	}

	scheduledFor := time.Now().Add(duration)

	err = ch.bot.DB.ScheduleMessage(i.GuildID, i.ChannelID, i.Member.User.ID, message, scheduledFor)
	if err != nil {
		respondEphemeral(s, i, "Failed to schedule message.")
		return
	}

	embed := successEmbed("Message Scheduled",
		fmt.Sprintf("Message will be sent <t:%d:R>", scheduledFor.Unix()))
	respondEmbedEphemeral(s, i, embed)
}

func (ch *CommandHandler) pollHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	question := getStringOption(i, "question")
	optionsStr := getStringOption(i, "options")

	var options []string
	if optionsStr != "" {
		options = strings.Split(optionsStr, "|")
		for idx := range options {
			options[idx] = strings.TrimSpace(options[idx])
		}
	}

	embed := &discordgo.MessageEmbed{
		Title: question,
		Color: 0x5865F2,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Poll by %s", i.Member.User.Username),
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})

	// Get the message to add reactions
	msg, err := s.InteractionResponse(i.Interaction)
	if err != nil {
		return
	}

	if len(options) == 0 {
		// Yes/No poll
		s.MessageReactionAdd(i.ChannelID, msg.ID, "✅")
		s.MessageReactionAdd(i.ChannelID, msg.ID, "❌")
	} else {
		// Multiple choice
		emojis := []string{"1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣", "6️⃣", "7️⃣", "8️⃣", "9️⃣", "🔟"}
		var desc string
		for idx, opt := range options {
			if idx >= 10 {
				break
			}
			desc += fmt.Sprintf("%s %s\n", emojis[idx], opt)
			s.MessageReactionAdd(i.ChannelID, msg.ID, emojis[idx])
		}

		embed.Description = desc
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{embed},
		})
	}
}

func (ch *CommandHandler) embedHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	title := getStringOption(i, "title")
	description := getStringOption(i, "description")
	colorStr := getStringOption(i, "color")
	image := getStringOption(i, "image")
	thumbnail := getStringOption(i, "thumbnail")
	footer := getStringOption(i, "footer")

	if title == "" && description == "" {
		respondEphemeral(s, i, "Please provide at least a title or description.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
	}

	// Parse color
	if colorStr != "" {
		colorStr = strings.TrimPrefix(colorStr, "#")
		var color int64
		fmt.Sscanf(colorStr, "%x", &color)
		embed.Color = int(color)
	} else {
		embed.Color = 0x5865F2
	}

	if image != "" {
		embed.Image = &discordgo.MessageEmbedImage{URL: image}
	}
	if thumbnail != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: thumbnail}
	}
	if footer != "" {
		embed.Footer = &discordgo.MessageEmbedFooter{Text: footer}
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) cleanHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	amount := int(getIntOption(i, "amount"))

	respondDeferredEphemeral(s, i)

	messages, err := s.ChannelMessages(i.ChannelID, 100, "", "", "")
	if err != nil {
		followUp(s, i, "Failed to fetch messages.")
		return
	}

	var toDelete []string
	for _, msg := range messages {
		if msg.Author.ID == i.Member.User.ID {
			msgTime, _ := discordgo.SnowflakeTimestamp(msg.ID)
			if time.Since(msgTime) < 14*24*time.Hour {
				toDelete = append(toDelete, msg.ID)
				if len(toDelete) >= amount {
					break
				}
			}
		}
	}

	if len(toDelete) == 0 {
		followUp(s, i, "No messages found to delete.")
		return
	}

	if len(toDelete) == 1 {
		s.ChannelMessageDelete(i.ChannelID, toDelete[0])
	} else {
		s.ChannelMessagesBulkDelete(i.ChannelID, toDelete)
	}

	followUp(s, i, fmt.Sprintf("Deleted %d of your messages.", len(toDelete)))
}

func (ch *CommandHandler) firstMessageHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	respondDeferred(s, i)

	messages, err := s.ChannelMessages(i.ChannelID, 1, "", "", "")
	if err != nil || len(messages) == 0 {
		followUp(s, i, "Failed to fetch messages.")
		return
	}

	// We need to get the actual first message by going backwards
	// This is a simplified version - Discord doesn't have a direct "first message" API
	// We'll get the channel creation and find messages around that time

	channel, err := s.Channel(i.ChannelID)
	if err != nil {
		followUp(s, i, "Failed to fetch channel info.")
		return
	}

	messages, err = s.ChannelMessages(i.ChannelID, 1, "", channel.ID, "")
	if err != nil || len(messages) == 0 {
		followUp(s, i, "Could not find the first message.")
		return
	}

	msg := messages[0]
	msgTime, _ := discordgo.SnowflakeTimestamp(msg.ID)

	embed := &discordgo.MessageEmbed{
		Title:       "First Message",
		Description: msg.Content,
		Color:       0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Author", Value: msg.Author.Username, Inline: true},
			{Name: "Sent", Value: fmt.Sprintf("<t:%d:F>", msgTime.Unix()), Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Message ID: %s", msg.ID),
		},
	}

	followUpEmbed(s, i, embed)
}

func (ch *CommandHandler) uptimeHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	uptime := time.Since(botStartTime)

	days := int(uptime.Hours()) / 24
	hours := int(uptime.Hours()) % 24
	minutes := int(uptime.Minutes()) % 60
	seconds := int(uptime.Seconds()) % 60

	var uptimeStr string
	if days > 0 {
		uptimeStr = fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	} else if hours > 0 {
		uptimeStr = fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		uptimeStr = fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		uptimeStr = fmt.Sprintf("%ds", seconds)
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Bot Uptime",
		Description: uptimeStr,
		Color:       0x57F287,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Started", Value: fmt.Sprintf("<t:%d:F>", botStartTime.Unix()), Inline: true},
		},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) sayHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	message := getStringOption(i, "message")
	channel := getChannelOption(i, "channel")

	channelID := i.ChannelID
	if channel != nil {
		channelID = channel.ID
	}

	_, err := s.ChannelMessageSend(channelID, message)
	if err != nil {
		respondEphemeral(s, i, "Failed to send message: "+err.Error())
		return
	}

	respondEphemeral(s, i, "Message sent!")
}

func (ch *CommandHandler) stealEmojiHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !hasPermission(s, i.GuildID, i.Member.User.ID, discordgo.PermissionManageEmojis) {
		respondEphemeral(s, i, "You don't have permission to manage emojis.")
		return
	}

	emojiStr := getStringOption(i, "emoji")
	name := getStringOption(i, "name")

	// Parse emoji ID from string like <:name:id> or <a:name:id>
	var emojiID string
	var animated bool

	if strings.HasPrefix(emojiStr, "<a:") {
		animated = true
		parts := strings.Split(strings.Trim(emojiStr, "<>"), ":")
		if len(parts) >= 3 {
			emojiID = parts[2]
			if name == "" {
				name = parts[1]
			}
		}
	} else if strings.HasPrefix(emojiStr, "<:") {
		parts := strings.Split(strings.Trim(emojiStr, "<>"), ":")
		if len(parts) >= 3 {
			emojiID = parts[2]
			if name == "" {
				name = parts[1]
			}
		}
	} else {
		respondEphemeral(s, i, "Please provide a custom emoji (not a default one).")
		return
	}

	if emojiID == "" {
		respondEphemeral(s, i, "Could not parse emoji.")
		return
	}

	// Construct URL
	ext := "png"
	if animated {
		ext = "gif"
	}
	url := fmt.Sprintf("https://cdn.discordapp.com/emojis/%s.%s", emojiID, ext)

	respondDeferred(s, i)

	// Download the image
	resp, err := httpClient.Get(url)
	if err != nil {
		followUp(s, i, "Failed to download emoji.")
		return
	}
	defer resp.Body.Close()

	// Create the emoji
	emoji, err := s.GuildEmojiCreate(i.GuildID, &discordgo.EmojiParams{
		Name:  name,
		Image: fmt.Sprintf("data:image/%s;base64,", ext),
	})

	if err != nil {
		followUp(s, i, "Failed to create emoji: "+err.Error())
		return
	}

	followUp(s, i, fmt.Sprintf("Successfully added emoji: %s", emoji.MessageFormat()))
}

func (ch *CommandHandler) mathHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	expression := getStringOption(i, "expression")

	// Simple and safe math evaluation
	result, err := evaluateMath(expression)
	if err != nil {
		respondEphemeral(s, i, "Invalid expression: "+err.Error())
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: "Math Result",
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Expression", Value: fmt.Sprintf("`%s`", expression), Inline: true},
			{Name: "Result", Value: fmt.Sprintf("`%s`", result), Inline: true},
		},
		Color: 0x5865F2,
	}

	respondEmbed(s, i, embed)
}

func strPtr(s string) *string {
	return &s
}

// Simple math evaluator (basic operations only for safety)
func evaluateMath(expr string) (string, error) {
	// Remove spaces
	expr = strings.ReplaceAll(expr, " ", "")

	// Very basic evaluation - only handles simple expressions
	// For production, you'd want a proper expression parser
	var result float64
	_, err := fmt.Sscanf(expr, "%f", &result)
	if err == nil {
		return fmt.Sprintf("%.2f", result), nil
	}

	// Try basic operations
	var a, b float64
	var op string

	for _, operator := range []string{"+", "-", "*", "/", "^", "%"} {
		if strings.Contains(expr, operator) {
			parts := strings.SplitN(expr, operator, 2)
			if len(parts) == 2 {
				_, err1 := fmt.Sscanf(parts[0], "%f", &a)
				_, err2 := fmt.Sscanf(parts[1], "%f", &b)
				if err1 == nil && err2 == nil {
					op = operator
					break
				}
			}
		}
	}

	if op == "" {
		return "", fmt.Errorf("could not parse expression")
	}

	switch op {
	case "+":
		result = a + b
	case "-":
		result = a - b
	case "*":
		result = a * b
	case "/":
		if b == 0 {
			return "", fmt.Errorf("division by zero")
		}
		result = a / b
	case "^":
		result = 1
		for i := 0; i < int(b); i++ {
			result *= a
		}
	case "%":
		result = float64(int(a) % int(b))
	}

	// Format nicely
	if result == float64(int(result)) {
		return fmt.Sprintf("%.0f", result), nil
	}
	return fmt.Sprintf("%.4f", result), nil
}

var httpClient = &http.Client{Timeout: 10 * time.Second}
