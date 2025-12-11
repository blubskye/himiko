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
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

func (ch *CommandHandler) registerRandomCommands() {
	// Random advice
	ch.Register(&Command{
		Name:        "advice",
		Description: "Get random advice",
		Category:    "Random",
		Handler:     ch.adviceHandler,
	})

	// Random quote
	ch.Register(&Command{
		Name:        "quote",
		Description: "Get a random inspirational quote",
		Category:    "Random",
		Handler:     ch.quoteHandler,
	})

	// Random fact
	ch.Register(&Command{
		Name:        "fact",
		Description: "Get a random fact",
		Category:    "Random",
		Handler:     ch.factHandler,
	})

	// Random trivia
	ch.Register(&Command{
		Name:        "trivia",
		Description: "Get a random trivia question",
		Category:    "Random",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "category",
				Description: "Trivia category",
				Required:    false,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "General Knowledge", Value: "9"},
					{Name: "Science & Nature", Value: "17"},
					{Name: "Computers", Value: "18"},
					{Name: "Mathematics", Value: "19"},
					{Name: "Sports", Value: "21"},
					{Name: "Geography", Value: "22"},
					{Name: "History", Value: "23"},
					{Name: "Art", Value: "25"},
					{Name: "Animals", Value: "27"},
					{Name: "Vehicles", Value: "28"},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "difficulty",
				Description: "Difficulty level",
				Required:    false,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "Easy", Value: "easy"},
					{Name: "Medium", Value: "medium"},
					{Name: "Hard", Value: "hard"},
				},
			},
		},
		Handler: ch.triviaHandler,
	})

	// Would you rather
	ch.Register(&Command{
		Name:        "wyr",
		Description: "Would you rather question",
		Category:    "Random",
		Handler:     ch.wyrHandler,
	})

	// Truth or Dare
	ch.Register(&Command{
		Name:        "truthordare",
		Description: "Truth or Dare",
		Category:    "Random",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "type",
				Description: "Truth or Dare",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "Truth", Value: "truth"},
					{Name: "Dare", Value: "dare"},
				},
			},
		},
		Handler: ch.truthOrDareHandler,
	})

	// Never have I ever
	ch.Register(&Command{
		Name:        "nhie",
		Description: "Never have I ever",
		Category:    "Random",
		Handler:     ch.nhieHandler,
	})

	// Dad joke
	ch.Register(&Command{
		Name:        "dadjoke",
		Description: "Get a dad joke",
		Category:    "Random",
		Handler:     ch.dadJokeHandler,
	})

	// Generate password
	ch.Register(&Command{
		Name:        "password",
		Description: "Generate a random password",
		Category:    "Random",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "length",
				Description: "Password length (8-64)",
				Required:    false,
				MinValue:    floatPtr(8),
				MaxValue:    64,
			},
		},
		Handler: ch.passwordHandler,
	})
}

func (ch *CommandHandler) adviceHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	respondDeferred(s, i)

	resp, err := http.Get("https://api.adviceslip.com/advice")
	if err != nil {
		followUp(s, i, "Failed to fetch advice.")
		return
	}
	defer resp.Body.Close()

	var data struct {
		Slip struct {
			Advice string `json:"advice"`
		} `json:"slip"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		followUp(s, i, "Failed to parse advice.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Random Advice",
		Description: data.Slip.Advice,
		Color:       0x5865F2,
	}

	followUpEmbed(s, i, embed)
}

func (ch *CommandHandler) quoteHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	respondDeferred(s, i)

	resp, err := http.Get("https://api.quotable.io/random")
	if err != nil {
		followUp(s, i, "Failed to fetch quote.")
		return
	}
	defer resp.Body.Close()

	var data struct {
		Content string `json:"content"`
		Author  string `json:"author"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		followUp(s, i, "Failed to parse quote.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Description: fmt.Sprintf("*\"%s\"*", data.Content),
		Color:       0x5865F2,
		Footer:      &discordgo.MessageEmbedFooter{Text: "— " + data.Author},
	}

	followUpEmbed(s, i, embed)
}

func (ch *CommandHandler) factHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	respondDeferred(s, i)

	resp, err := http.Get("https://uselessfacts.jsph.pl/api/v2/facts/random")
	if err != nil {
		followUp(s, i, "Failed to fetch fact.")
		return
	}
	defer resp.Body.Close()

	var data struct {
		Text string `json:"text"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		followUp(s, i, "Failed to parse fact.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Random Fact",
		Description: data.Text,
		Color:       0x5865F2,
	}

	followUpEmbed(s, i, embed)
}

func (ch *CommandHandler) triviaHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	category := getStringOption(i, "category")
	difficulty := getStringOption(i, "difficulty")

	respondDeferred(s, i)

	url := "https://opentdb.com/api.php?amount=1&type=multiple"
	if category != "" {
		url += "&category=" + category
	}
	if difficulty != "" {
		url += "&difficulty=" + difficulty
	}

	resp, err := http.Get(url)
	if err != nil {
		followUp(s, i, "Failed to fetch trivia.")
		return
	}
	defer resp.Body.Close()

	var data struct {
		Results []struct {
			Category         string   `json:"category"`
			Difficulty       string   `json:"difficulty"`
			Question         string   `json:"question"`
			CorrectAnswer    string   `json:"correct_answer"`
			IncorrectAnswers []string `json:"incorrect_answers"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil || len(data.Results) == 0 {
		followUp(s, i, "Failed to parse trivia.")
		return
	}

	trivia := data.Results[0]

	// Unescape HTML entities
	question := unescapeHTML(trivia.Question)
	correct := unescapeHTML(trivia.CorrectAnswer)

	// Mix answers
	answers := append(trivia.IncorrectAnswers, trivia.CorrectAnswer)
	rand.Shuffle(len(answers), func(i, j int) {
		answers[i], answers[j] = answers[j], answers[i]
	})

	var answerText string
	letters := []string{"A", "B", "C", "D"}
	for idx, ans := range answers {
		answerText += fmt.Sprintf("**%s.** %s\n", letters[idx], unescapeHTML(ans))
	}

	embed := &discordgo.MessageEmbed{
		Title:       question,
		Description: answerText,
		Color:       0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Category", Value: trivia.Category, Inline: true},
			{Name: "Difficulty", Value: trivia.Difficulty, Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Answer: %s", correct),
		},
	}

	followUpEmbed(s, i, embed)
}

func (ch *CommandHandler) wyrHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	questions := []struct {
		Option1 string
		Option2 string
	}{
		{"Be able to fly", "Be able to read minds"},
		{"Never use the internet again", "Never watch TV again"},
		{"Be rich but alone", "Be poor but surrounded by loved ones"},
		{"Have unlimited money", "Have unlimited time"},
		{"Know how you'll die", "Know when you'll die"},
		{"Live without music", "Live without movies"},
		{"Be famous but broke", "Be unknown but wealthy"},
		{"Always be cold", "Always be hot"},
		{"Speak all languages", "Talk to animals"},
		{"Be invisible", "Be able to teleport"},
		{"Have no taste", "Have no smell"},
		{"Live in the past", "Live in the future"},
		{"Only eat pizza forever", "Never eat pizza again"},
		{"Be a genius with no friends", "Be average with many friends"},
		{"Have super strength", "Have super speed"},
	}

	q := questions[rand.Intn(len(questions))]

	embed := &discordgo.MessageEmbed{
		Title:       "Would You Rather",
		Description: fmt.Sprintf("🔴 %s\n\n**OR**\n\n🔵 %s", q.Option1, q.Option2),
		Color:       0x5865F2,
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})

	// Add reactions
	msg, _ := s.InteractionResponse(i.Interaction)
	if msg != nil {
		s.MessageReactionAdd(i.ChannelID, msg.ID, "🔴")
		s.MessageReactionAdd(i.ChannelID, msg.ID, "🔵")
	}
}

func (ch *CommandHandler) truthOrDareHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	choice := getStringOption(i, "type")

	truths := []string{
		"What's your biggest fear?",
		"What's the most embarrassing thing you've done?",
		"What's a secret you've never told anyone?",
		"What's your biggest regret?",
		"Who was your first crush?",
		"What's the worst lie you've ever told?",
		"What's your guilty pleasure?",
		"What's the most childish thing you still do?",
		"What's the worst thing you've ever said to someone?",
		"What's something you're glad your mom doesn't know about you?",
		"What's the biggest misconception about you?",
		"What's the most trouble you've been in?",
		"What's your most embarrassing childhood memory?",
		"What's something you've done that you still feel guilty about?",
		"What's the worst date you've ever been on?",
	}

	dares := []string{
		"Send a message to your crush right now",
		"Post an embarrassing photo on social media",
		"Let someone post something on your social media",
		"Do your best impression of someone in the call",
		"Speak in an accent for the next 5 minutes",
		"Show everyone your screen time",
		"Let someone send a text from your phone",
		"Share your most recently deleted photo",
		"Do 20 pushups",
		"Sing a song of the group's choice",
		"Call a random contact and sing happy birthday",
		"Let someone change your profile picture for 24 hours",
		"Share your most recent search history",
		"Let someone write your status for 24 hours",
		"Send a screenshot of your DMs to the group",
	}

	var prompt string
	var title string
	if choice == "truth" {
		prompt = truths[rand.Intn(len(truths))]
		title = "Truth"
	} else {
		prompt = dares[rand.Intn(len(dares))]
		title = "Dare"
	}

	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: prompt,
		Color:       0x5865F2,
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) nhieHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	statements := []string{
		"Never have I ever lied to get out of work/school",
		"Never have I ever pretended to be sick",
		"Never have I ever stalked someone on social media",
		"Never have I ever sent a text to the wrong person",
		"Never have I ever had a crush on a friend's partner",
		"Never have I ever cried at a movie",
		"Never have I ever eaten food off the floor",
		"Never have I ever broken something and blamed it on someone else",
		"Never have I ever regretted a haircut",
		"Never have I ever fallen asleep at work/school",
		"Never have I ever ghosted someone",
		"Never have I ever been kicked out of somewhere",
		"Never have I ever forgotten someone's name mid-conversation",
		"Never have I ever pretended to laugh at a joke I didn't understand",
		"Never have I ever walked into a glass door",
	}

	statement := statements[rand.Intn(len(statements))]

	embed := &discordgo.MessageEmbed{
		Title:       "Never Have I Ever",
		Description: statement,
		Color:       0x5865F2,
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) dadJokeHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	respondDeferred(s, i)

	req, _ := http.NewRequest("GET", "https://icanhazdadjoke.com/", nil)
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		followUp(s, i, "Failed to fetch dad joke.")
		return
	}
	defer resp.Body.Close()

	var data struct {
		Joke string `json:"joke"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		followUp(s, i, "Failed to parse dad joke.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Dad Joke",
		Description: data.Joke,
		Color:       0xFEE75C,
	}

	followUpEmbed(s, i, embed)
}

func (ch *CommandHandler) passwordHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	length := getIntOption(i, "length")
	if length == 0 {
		length = 16
	}

	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;:,.<>?"

	password := make([]byte, length)
	for i := range password {
		password[i] = chars[rand.Intn(len(chars))]
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Generated Password",
		Description: fmt.Sprintf("```%s```", string(password)),
		Color:       0x57F287,
		Footer:      &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("Length: %d characters", length)},
	}

	respondEmbedEphemeral(s, i, embed)
}

func unescapeHTML(s string) string {
	replacements := map[string]string{
		"&amp;":  "&",
		"&lt;":   "<",
		"&gt;":   ">",
		"&quot;": "\"",
		"&#039;": "'",
		"&apos;": "'",
	}

	for old, new := range replacements {
		s = replaceAll(s, old, new)
	}
	return s
}

func replaceAll(s, old, new string) string {
	for {
		idx := indexOf(s, old)
		if idx == -1 {
			break
		}
		s = s[:idx] + new + s[idx+len(old):]
	}
	return s
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
