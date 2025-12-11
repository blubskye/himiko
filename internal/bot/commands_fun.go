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
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func (ch *CommandHandler) registerFunCommands() {
	// 8ball
	ch.Register(&Command{
		Name:        "8ball",
		Description: "Ask the magic 8-ball a question",
		Category:    "Fun",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "question",
				Description: "Your question",
				Required:    true,
			},
		},
		Handler: ch.eightBallHandler,
	})

	// Dice roll
	ch.Register(&Command{
		Name:        "dice",
		Description: "Roll a dice",
		Category:    "Fun",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "sides",
				Description: "Number of sides (default: 6)",
				Required:    false,
				MinValue:    floatPtr(2),
				MaxValue:    100,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "count",
				Description: "Number of dice to roll (default: 1)",
				Required:    false,
				MinValue:    floatPtr(1),
				MaxValue:    10,
			},
		},
		Handler: ch.diceHandler,
	})

	// Coinflip
	ch.Register(&Command{
		Name:        "coinflip",
		Description: "Flip a coin",
		Category:    "Fun",
		Handler:     ch.coinflipHandler,
	})

	// RPS
	ch.Register(&Command{
		Name:        "rps",
		Description: "Rock, Paper, Scissors",
		Category:    "Fun",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "choice",
				Description: "Your choice",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "Rock", Value: "rock"},
					{Name: "Paper", Value: "paper"},
					{Name: "Scissors", Value: "scissors"},
				},
			},
		},
		Handler: ch.rpsHandler,
	})

	// Random number
	ch.Register(&Command{
		Name:        "random",
		Description: "Generate a random number",
		Category:    "Fun",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "min",
				Description: "Minimum number",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "max",
				Description: "Maximum number",
				Required:    false,
			},
		},
		Handler: ch.randomHandler,
	})

	// Joke
	ch.Register(&Command{
		Name:        "joke",
		Description: "Get a random joke",
		Category:    "Fun",
		Handler:     ch.jokeHandler,
	})

	// Rate
	ch.Register(&Command{
		Name:        "rate",
		Description: "Rate something out of 10",
		Category:    "Fun",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "thing",
				Description: "What to rate",
				Required:    true,
			},
		},
		Handler: ch.rateHandler,
	})

	// Ship
	ch.Register(&Command{
		Name:        "ship",
		Description: "Calculate love compatibility",
		Category:    "Fun",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user1",
				Description: "First user",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user2",
				Description: "Second user",
				Required:    true,
			},
		},
		Handler: ch.shipHandler,
	})

	// IQ test
	ch.Register(&Command{
		Name:        "iq",
		Description: "Check someone's IQ",
		Category:    "Fun",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "User to test",
				Required:    false,
			},
		},
		Handler: ch.iqHandler,
	})

	// Gay test
	ch.Register(&Command{
		Name:        "gay",
		Description: "How gay is someone?",
		Category:    "Fun",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "User to check",
				Required:    false,
			},
		},
		Handler: ch.gayHandler,
	})

	// PP size
	ch.Register(&Command{
		Name:        "pp",
		Description: "Check PP size",
		Category:    "Fun",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "User to measure",
				Required:    false,
			},
		},
		Handler: ch.ppHandler,
	})

	// Hug
	ch.Register(&Command{
		Name:        "hug",
		Description: "Hug someone",
		Category:    "Fun",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "User to hug",
				Required:    true,
			},
		},
		Handler: ch.hugHandler,
	})

	// Slap
	ch.Register(&Command{
		Name:        "slap",
		Description: "Slap someone",
		Category:    "Fun",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "User to slap",
				Required:    true,
			},
		},
		Handler: ch.slapHandler,
	})

	// Pat
	ch.Register(&Command{
		Name:        "pat",
		Description: "Pat someone",
		Category:    "Fun",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "User to pat",
				Required:    true,
			},
		},
		Handler: ch.patHandler,
	})

	// Kiss
	ch.Register(&Command{
		Name:        "kiss",
		Description: "Kiss someone",
		Category:    "Fun",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "User to kiss",
				Required:    true,
			},
		},
		Handler: ch.kissHandler,
	})

	// Respect (F)
	ch.Register(&Command{
		Name:        "f",
		Description: "Pay respects",
		Category:    "Fun",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "thing",
				Description: "What to pay respects to",
				Required:    false,
			},
		},
		Handler: ch.respectHandler,
	})

	// Choose
	ch.Register(&Command{
		Name:        "choose",
		Description: "Choose between options",
		Category:    "Fun",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "options",
				Description: "Options separated by commas or 'or'",
				Required:    true,
			},
		},
		Handler: ch.chooseHandler,
	})
}

func (ch *CommandHandler) eightBallHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	question := getStringOption(i, "question")

	responses := []string{
		"It is certain.", "It is decidedly so.", "Without a doubt.",
		"Yes definitely.", "You may rely on it.", "As I see it, yes.",
		"Most likely.", "Outlook good.", "Yes.", "Signs point to yes.",
		"Reply hazy, try again.", "Ask again later.", "Better not tell you now.",
		"Cannot predict now.", "Concentrate and ask again.",
		"Don't count on it.", "My reply is no.", "My sources say no.",
		"Outlook not so good.", "Very doubtful.",
	}

	answer := responses[rand.Intn(len(responses))]

	embed := &discordgo.MessageEmbed{
		Title: "Magic 8-Ball",
		Color: 0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Question", Value: question, Inline: false},
			{Name: "Answer", Value: answer, Inline: false},
		},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) diceHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	sides := getIntOption(i, "sides")
	count := getIntOption(i, "count")

	if sides == 0 {
		sides = 6
	}
	if count == 0 {
		count = 1
	}

	var results []string
	var total int64
	for j := int64(0); j < count; j++ {
		roll := rand.Int63n(sides) + 1
		total += roll
		results = append(results, fmt.Sprintf("%d", roll))
	}

	desc := strings.Join(results, ", ")
	if count > 1 {
		desc += fmt.Sprintf("\n**Total:** %d", total)
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Rolling %dd%d", count, sides),
		Description: desc,
		Color:       0x5865F2,
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) coinflipHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	result := "Heads"
	if rand.Intn(2) == 1 {
		result = "Tails"
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Coin Flip",
		Description: fmt.Sprintf("The coin landed on **%s**!", result),
		Color:       0xF1C40F,
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) rpsHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userChoice := getStringOption(i, "choice")

	choices := []string{"rock", "paper", "scissors"}
	botChoice := choices[rand.Intn(3)]

	var result string
	if userChoice == botChoice {
		result = "It's a tie!"
	} else if (userChoice == "rock" && botChoice == "scissors") ||
		(userChoice == "paper" && botChoice == "rock") ||
		(userChoice == "scissors" && botChoice == "paper") {
		result = "You win!"
	} else {
		result = "I win!"
	}

	emojis := map[string]string{
		"rock":     "rock",
		"paper":    "page_facing_up",
		"scissors": "scissors",
	}

	embed := &discordgo.MessageEmbed{
		Title: "Rock Paper Scissors",
		Color: 0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "You chose", Value: fmt.Sprintf(":%s: %s", emojis[userChoice], strings.Title(userChoice)), Inline: true},
			{Name: "I chose", Value: fmt.Sprintf(":%s: %s", emojis[botChoice], strings.Title(botChoice)), Inline: true},
			{Name: "Result", Value: result, Inline: false},
		},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) randomHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	min := getIntOption(i, "min")
	max := getIntOption(i, "max")

	if max == 0 {
		max = 100
	}
	if min >= max {
		respondEphemeral(s, i, "Min must be less than max.")
		return
	}

	result := rand.Int63n(max-min+1) + min

	embed := &discordgo.MessageEmbed{
		Title:       "Random Number",
		Description: fmt.Sprintf("**%d**\n(between %d and %d)", result, min, max),
		Color:       0x5865F2,
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) jokeHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	respondDeferred(s, i)

	resp, err := http.Get("https://official-joke-api.appspot.com/random_joke")
	if err != nil {
		followUp(s, i, "Failed to fetch joke.")
		return
	}
	defer resp.Body.Close()

	var joke struct {
		Setup     string `json:"setup"`
		Punchline string `json:"punchline"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&joke); err != nil {
		followUp(s, i, "Failed to parse joke.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       joke.Setup,
		Description: fmt.Sprintf("||%s||", joke.Punchline),
		Color:       0xFEE75C,
	}

	followUpEmbed(s, i, embed)
}

func (ch *CommandHandler) rateHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	thing := getStringOption(i, "thing")
	rating := rand.Intn(11)

	var emoji string
	switch {
	case rating <= 2:
		emoji = "terrible"
	case rating <= 4:
		emoji = "bad"
	case rating <= 6:
		emoji = "meh"
	case rating <= 8:
		emoji = "good"
	default:
		emoji = "excellent"
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Rating: %s", thing),
		Description: fmt.Sprintf("I rate **%s** a **%d/10** (%s)", thing, rating, emoji),
		Color:       0x5865F2,
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) shipHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user1 := getUserOption(i, "user1")
	user2 := getUserOption(i, "user2")

	if user1 == nil || user2 == nil {
		respondEphemeral(s, i, "Please specify two users.")
		return
	}

	// Generate deterministic percentage based on user IDs
	seed := int64(0)
	for _, c := range user1.ID + user2.ID {
		seed += int64(c)
	}
	rng := rand.New(rand.NewSource(seed))
	percentage := rng.Intn(101)

	var status string
	var emoji string
	switch {
	case percentage <= 20:
		status = "Not meant to be..."
		emoji = "broken_heart"
	case percentage <= 40:
		status = "Just friends"
		emoji = "blue_heart"
	case percentage <= 60:
		status = "There's potential!"
		emoji = "yellow_heart"
	case percentage <= 80:
		status = "A good match!"
		emoji = "orange_heart"
	default:
		status = "Perfect match!"
		emoji = "heart"
	}

	// Create ship name
	name1 := user1.Username
	name2 := user2.Username
	shipName := name1[:len(name1)/2] + name2[len(name2)/2:]

	// Progress bar
	filled := percentage / 10
	bar := strings.Repeat("█", filled) + strings.Repeat("░", 10-filled)

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf(":%s: %s", emoji, shipName),
		Description: fmt.Sprintf("%s + %s\n\n%s %d%%\n\n**%s**",
			user1.Mention(), user2.Mention(), bar, percentage, status),
		Color: 0xFF69B4,
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) iqHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := getUserOption(i, "user")
	if user == nil {
		user = i.Member.User
	}

	// Deterministic based on user ID
	seed := int64(0)
	for _, c := range user.ID {
		seed += int64(c)
	}
	rng := rand.New(rand.NewSource(seed))
	iq := rng.Intn(200) + 1

	var comment string
	switch {
	case iq < 70:
		comment = "Needs some help..."
	case iq < 90:
		comment = "Below average"
	case iq < 110:
		comment = "Average"
	case iq < 130:
		comment = "Above average"
	case iq < 150:
		comment = "Gifted!"
	default:
		comment = "Genius level!"
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("IQ Test: %s", user.Username),
		Description: fmt.Sprintf("**IQ:** %d\n**Assessment:** %s", iq, comment),
		Color:       0x5865F2,
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: avatarURL(user)},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) gayHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := getUserOption(i, "user")
	if user == nil {
		user = i.Member.User
	}

	// Deterministic based on user ID
	seed := int64(0)
	for _, c := range user.ID {
		seed += int64(c)
	}
	rng := rand.New(rand.NewSource(seed))
	percentage := rng.Intn(101)

	bar := strings.Repeat("█", percentage/10) + strings.Repeat("░", 10-percentage/10)

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Gay-o-meter: %s", user.Username),
		Description: fmt.Sprintf("**%s is %d%% gay**\n\n%s", user.Username, percentage, bar),
		Color:       0xFF69B4,
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: avatarURL(user)},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) ppHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := getUserOption(i, "user")
	if user == nil {
		user = i.Member.User
	}

	// Deterministic based on user ID
	seed := int64(0)
	for _, c := range user.ID {
		seed += int64(c)
	}
	rng := rand.New(rand.NewSource(seed))
	size := rng.Intn(15) + 1

	pp := "8" + strings.Repeat("=", size) + "D"

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("PP Size: %s", user.Username),
		Description: pp,
		Color:       0x5865F2,
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) hugHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := getUserOption(i, "user")
	if user == nil {
		respondEphemeral(s, i, "Please specify a user to hug.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Description: fmt.Sprintf("**%s** hugs **%s**!", i.Member.User.Username, user.Username),
		Color:       0xFF69B4,
		Image:       &discordgo.MessageEmbedImage{URL: getInteractionGif("hug")},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) slapHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := getUserOption(i, "user")
	if user == nil {
		respondEphemeral(s, i, "Please specify a user to slap.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Description: fmt.Sprintf("**%s** slaps **%s**!", i.Member.User.Username, user.Username),
		Color:       0xED4245,
		Image:       &discordgo.MessageEmbedImage{URL: getInteractionGif("slap")},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) patHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := getUserOption(i, "user")
	if user == nil {
		respondEphemeral(s, i, "Please specify a user to pat.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Description: fmt.Sprintf("**%s** pats **%s**!", i.Member.User.Username, user.Username),
		Color:       0x57F287,
		Image:       &discordgo.MessageEmbedImage{URL: getInteractionGif("pat")},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) kissHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := getUserOption(i, "user")
	if user == nil {
		respondEphemeral(s, i, "Please specify a user to kiss.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Description: fmt.Sprintf("**%s** kisses **%s**!", i.Member.User.Username, user.Username),
		Color:       0xFF69B4,
		Image:       &discordgo.MessageEmbedImage{URL: getInteractionGif("kiss")},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) respectHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	thing := getStringOption(i, "thing")

	msg := fmt.Sprintf("**%s** has paid their respects.", i.Member.User.Username)
	if thing != "" {
		msg = fmt.Sprintf("**%s** has paid their respects to **%s**.", i.Member.User.Username, thing)
	}

	embed := &discordgo.MessageEmbed{
		Description: msg + "\n\nPress F to pay respects.",
		Color:       0x5865F2,
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) chooseHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	input := getStringOption(i, "options")

	// Split by "or" or comma
	var options []string
	if strings.Contains(strings.ToLower(input), " or ") {
		options = strings.Split(strings.ToLower(input), " or ")
	} else {
		options = strings.Split(input, ",")
	}

	// Clean up options
	var cleaned []string
	for _, opt := range options {
		opt = strings.TrimSpace(opt)
		if opt != "" {
			cleaned = append(cleaned, opt)
		}
	}

	if len(cleaned) < 2 {
		respondEphemeral(s, i, "Please provide at least 2 options separated by commas or 'or'.")
		return
	}

	choice := cleaned[rand.Intn(len(cleaned))]

	embed := &discordgo.MessageEmbed{
		Title:       "I choose...",
		Description: fmt.Sprintf("**%s**", choice),
		Color:       0x5865F2,
		Footer:      &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("From %d options", len(cleaned))},
	}

	respondEmbed(s, i, embed)
}

// Helper function to get interaction GIFs
func getInteractionGif(action string) string {
	gifs := map[string][]string{
		"hug": {
			"https://media.tenor.com/ULZU0gBw2d0AAAAC/hug.gif",
			"https://media.tenor.com/E6fMkQRZBdgAAAAC/anime-hug.gif",
			"https://media.tenor.com/9e1aE_xBLCsAAAAC/anime-hug.gif",
		},
		"slap": {
			"https://media.tenor.com/Ws6Dm1ZW_vMAAAAC/anime-slap.gif",
			"https://media.tenor.com/mBmFvU_7qvYAAAAC/slap-anime.gif",
			"https://media.tenor.com/DKvymPDFk1EAAAAC/slap-anime.gif",
		},
		"pat": {
			"https://media.tenor.com/N41zKEDABuUAAAAC/pat-head.gif",
			"https://media.tenor.com/3lIdWk9HgtMAAAAC/head-pat.gif",
			"https://media.tenor.com/xpIgJcLsQakAAAAC/anime-head-pat.gif",
		},
		"kiss": {
			"https://media.tenor.com/9e-PWLSdW3EAAAAC/kiss-anime.gif",
			"https://media.tenor.com/0T1A9B0LoQkAAAAC/anime-kiss.gif",
			"https://media.tenor.com/oLZ0u-VnIHMAAAAC/kiss-anime.gif",
		},
	}

	if urls, ok := gifs[action]; ok {
		return urls[rand.Intn(len(urls))]
	}
	return ""
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
