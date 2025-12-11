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
	"net/http"

	"github.com/bwmarrin/discordgo"
)

func (ch *CommandHandler) registerImageCommands() {
	// Animal images
	ch.Register(&Command{
		Name:        "animal",
		Description: "Get a random animal image",
		Category:    "Images",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "type",
				Description: "Type of animal",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "Cat", Value: "cat"},
					{Name: "Dog", Value: "dog"},
					{Name: "Fox", Value: "fox"},
					{Name: "Bird", Value: "bird"},
					{Name: "Bunny", Value: "bunny"},
					{Name: "Duck", Value: "duck"},
					{Name: "Koala", Value: "koala"},
					{Name: "Panda", Value: "panda"},
					{Name: "Red Panda", Value: "red_panda"},
				},
			},
		},
		Handler: ch.animalHandler,
	})

	// User avatar
	ch.Register(&Command{
		Name:        "avatar",
		Description: "Get a user's avatar",
		Category:    "Images",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "User to get avatar of",
				Required:    false,
			},
		},
		Handler: ch.avatarHandler,
	})

	// User banner
	ch.Register(&Command{
		Name:        "banner",
		Description: "Get a user's banner",
		Category:    "Images",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "User to get banner of",
				Required:    false,
			},
		},
		Handler: ch.bannerHandler,
	})

	// Server icon
	ch.Register(&Command{
		Name:        "servericon",
		Description: "Get the server icon",
		Category:    "Images",
		Handler:     ch.serverIconHandler,
	})

	// Cat fact with image
	ch.Register(&Command{
		Name:        "catfact",
		Description: "Get a random cat fact",
		Category:    "Images",
		Handler:     ch.catFactHandler,
	})

	// Dog fact with image
	ch.Register(&Command{
		Name:        "dogfact",
		Description: "Get a random dog fact",
		Category:    "Images",
		Handler:     ch.dogFactHandler,
	})

	// Random meme
	ch.Register(&Command{
		Name:        "meme",
		Description: "Get a random meme from Reddit",
		Category:    "Images",
		Handler:     ch.memeHandler,
	})
}

func (ch *CommandHandler) animalHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	animalType := getStringOption(i, "type")

	respondDeferred(s, i)

	var imageURL string
	var err error

	switch animalType {
	case "cat":
		imageURL, err = fetchCatImage()
	case "dog":
		imageURL, err = fetchDogImage()
	case "fox":
		imageURL, err = fetchFoxImage()
	case "bird":
		imageURL, err = fetchSomeRandomAPI("bird")
	case "bunny":
		imageURL, err = fetchSomeRandomAPI("bunny") // Note: may not work
	case "duck":
		imageURL, err = fetchDuckImage()
	case "koala":
		imageURL, err = fetchSomeRandomAPI("koala")
	case "panda":
		imageURL, err = fetchSomeRandomAPI("panda")
	case "red_panda":
		imageURL, err = fetchSomeRandomAPI("red_panda")
	default:
		followUp(s, i, "Unknown animal type.")
		return
	}

	if err != nil {
		followUp(s, i, "Failed to fetch image: "+err.Error())
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("Random %s", animalType),
		Image: &discordgo.MessageEmbedImage{URL: imageURL},
		Color: 0x5865F2,
	}

	followUpEmbed(s, i, embed)
}

func (ch *CommandHandler) avatarHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := getUserOption(i, "user")
	if user == nil {
		user = i.Member.User
	}

	url := avatarURL(user)

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s's Avatar", user.Username),
		Image: &discordgo.MessageEmbedImage{URL: url},
		Color: 0x5865F2,
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) bannerHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	targetUser := getUserOption(i, "user")
	if targetUser == nil {
		targetUser = i.Member.User
	}

	// Fetch full user data to get banner
	fullUser, err := s.User(targetUser.ID)
	if err != nil {
		respondEphemeral(s, i, "Failed to fetch user data.")
		return
	}

	url := bannerURL(fullUser)
	if url == "" {
		respondEphemeral(s, i, "This user doesn't have a banner.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s's Banner", fullUser.Username),
		Image: &discordgo.MessageEmbedImage{URL: url},
		Color: 0x5865F2,
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) serverIconHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guild, err := s.Guild(i.GuildID)
	if err != nil {
		respondEphemeral(s, i, "Failed to fetch server data.")
		return
	}

	url := guildIconURL(guild)
	if url == "" {
		respondEphemeral(s, i, "This server doesn't have an icon.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s's Icon", guild.Name),
		Image: &discordgo.MessageEmbedImage{URL: url},
		Color: 0x5865F2,
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) catFactHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	respondDeferred(s, i)

	// Fetch cat fact
	resp, err := http.Get("https://catfact.ninja/fact")
	if err != nil {
		followUp(s, i, "Failed to fetch cat fact.")
		return
	}
	defer resp.Body.Close()

	var factData struct {
		Fact string `json:"fact"`
	}
	json.NewDecoder(resp.Body).Decode(&factData)

	// Fetch cat image
	imageURL, _ := fetchCatImage()

	embed := &discordgo.MessageEmbed{
		Title:       "Cat Fact",
		Description: factData.Fact,
		Color:       0xFFB6C1,
	}
	if imageURL != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: imageURL}
	}

	followUpEmbed(s, i, embed)
}

func (ch *CommandHandler) dogFactHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	respondDeferred(s, i)

	// Fetch dog fact
	resp, err := http.Get("https://dog-api.kinduff.com/api/facts")
	if err != nil {
		followUp(s, i, "Failed to fetch dog fact.")
		return
	}
	defer resp.Body.Close()

	var factData struct {
		Facts []string `json:"facts"`
	}
	json.NewDecoder(resp.Body).Decode(&factData)

	fact := ""
	if len(factData.Facts) > 0 {
		fact = factData.Facts[0]
	}

	// Fetch dog image
	imageURL, _ := fetchDogImage()

	embed := &discordgo.MessageEmbed{
		Title:       "Dog Fact",
		Description: fact,
		Color:       0x8B4513,
	}
	if imageURL != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: imageURL}
	}

	followUpEmbed(s, i, embed)
}

func (ch *CommandHandler) memeHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	respondDeferred(s, i)

	resp, err := http.Get("https://meme-api.com/gimme")
	if err != nil {
		followUp(s, i, "Failed to fetch meme.")
		return
	}
	defer resp.Body.Close()

	var meme struct {
		Title     string `json:"title"`
		URL       string `json:"url"`
		PostLink  string `json:"postLink"`
		Subreddit string `json:"subreddit"`
		Author    string `json:"author"`
		Ups       int    `json:"ups"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&meme); err != nil {
		followUp(s, i, "Failed to parse meme data.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: meme.Title,
		URL:   meme.PostLink,
		Image: &discordgo.MessageEmbedImage{URL: meme.URL},
		Color: 0xFF4500,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("r/%s | u/%s | %d upvotes", meme.Subreddit, meme.Author, meme.Ups),
		},
	}

	followUpEmbed(s, i, embed)
}

// API helper functions
func fetchCatImage() (string, error) {
	resp, err := http.Get("https://api.thecatapi.com/v1/images/search")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data []struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	if len(data) > 0 {
		return data[0].URL, nil
	}
	return "", fmt.Errorf("no image found")
}

func fetchDogImage() (string, error) {
	resp, err := http.Get("https://dog.ceo/api/breeds/image/random")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	return data.Message, nil
}

func fetchFoxImage() (string, error) {
	resp, err := http.Get("https://randomfox.ca/floof/")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data struct {
		Image string `json:"image"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	return data.Image, nil
}

func fetchDuckImage() (string, error) {
	resp, err := http.Get("https://random-d.uk/api/random")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	return data.URL, nil
}

func fetchSomeRandomAPI(animal string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://some-random-api.com/animal/%s", animal))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data struct {
		Image string `json:"image"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	return data.Image, nil
}
