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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

func (ch *CommandHandler) registerAICommands() {
	// AI Ask
	ch.Register(&Command{
		Name:        "ask",
		Description: "Ask AI a question",
		Category:    "AI",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "question",
				Description: "Your question",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "system",
				Description: "Custom system prompt",
				Required:    false,
			},
		},
		Handler: ch.askAIHandler,
	})
}

func (ch *CommandHandler) askAIHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	question := getStringOption(i, "question")
	systemPrompt := getStringOption(i, "system")

	// Check if AI is configured
	if ch.bot.Config.APIs.OpenAIKey == "" {
		respondEphemeral(s, i, "AI is not configured. Please set an OpenAI API key in the config.")
		return
	}

	respondDeferred(s, i)

	if systemPrompt == "" {
		systemPrompt = "You are a helpful assistant. Keep responses concise and under 2000 characters."
	}

	// Prepare the request
	requestBody := map[string]interface{}{
		"model": ch.bot.Config.APIs.OpenAIModel,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": question},
		},
		"max_tokens":  1000,
		"temperature": 0.7,
	}

	jsonBody, _ := json.Marshal(requestBody)

	req, err := http.NewRequest("POST", ch.bot.Config.APIs.OpenAIBaseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		followUp(s, i, "Failed to create request.")
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ch.bot.Config.APIs.OpenAIKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		followUp(s, i, "Failed to contact AI service.")
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		followUp(s, i, "Failed to parse AI response.")
		return
	}

	if response.Error.Message != "" {
		followUp(s, i, "AI Error: "+response.Error.Message)
		return
	}

	if len(response.Choices) == 0 {
		followUp(s, i, "No response from AI.")
		return
	}

	answer := response.Choices[0].Message.Content
	if len(answer) > 2000 {
		answer = answer[:1997] + "..."
	}

	embed := &discordgo.MessageEmbed{
		Title:       "AI Response",
		Description: answer,
		Color:       0x10A37F,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Model: %s", ch.bot.Config.APIs.OpenAIModel),
		},
	}

	followUpEmbed(s, i, embed)
}
