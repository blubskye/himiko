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
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (ch *CommandHandler) registerLookupCommands() {
	// Weather
	ch.Register(&Command{
		Name:        "weather",
		Description: "Get current weather for a city",
		Category:    "Lookup",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "city",
				Description: "City name",
				Required:    true,
			},
		},
		Handler: ch.weatherHandler,
	})

	// Urban Dictionary
	ch.Register(&Command{
		Name:        "urban",
		Description: "Look up a term on Urban Dictionary",
		Category:    "Lookup",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "term",
				Description: "Term to look up",
				Required:    true,
			},
		},
		Handler: ch.urbanHandler,
	})

	// Wikipedia
	ch.Register(&Command{
		Name:        "wiki",
		Description: "Search Wikipedia",
		Category:    "Lookup",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "query",
				Description: "Search query",
				Required:    true,
			},
		},
		Handler: ch.wikiHandler,
	})

	// IP lookup
	ch.Register(&Command{
		Name:        "iplookup",
		Description: "Look up information about an IP address",
		Category:    "Lookup",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "ip",
				Description: "IP address to look up",
				Required:    true,
			},
		},
		Handler: ch.ipLookupHandler,
	})

	// Crypto price
	ch.Register(&Command{
		Name:        "crypto",
		Description: "Get cryptocurrency price",
		Category:    "Lookup",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "coin",
				Description: "Cryptocurrency (e.g., bitcoin, ethereum)",
				Required:    true,
			},
		},
		Handler: ch.cryptoHandler,
	})

	// Minecraft server
	ch.Register(&Command{
		Name:        "mcserver",
		Description: "Look up a Minecraft server",
		Category:    "Lookup",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "address",
				Description: "Server address",
				Required:    true,
			},
		},
		Handler: ch.mcServerHandler,
	})

	// GitHub user
	ch.Register(&Command{
		Name:        "github",
		Description: "Look up a GitHub user",
		Category:    "Lookup",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "username",
				Description: "GitHub username",
				Required:    true,
			},
		},
		Handler: ch.githubHandler,
	})

	// Npm package
	ch.Register(&Command{
		Name:        "npm",
		Description: "Look up an npm package",
		Category:    "Lookup",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "package",
				Description: "Package name",
				Required:    true,
			},
		},
		Handler: ch.npmHandler,
	})

	// Color
	ch.Register(&Command{
		Name:        "color",
		Description: "Get information about a color",
		Category:    "Lookup",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "hex",
				Description: "Hex color code (e.g., #FF0000)",
				Required:    true,
			},
		},
		Handler: ch.colorHandler,
	})
}

func (ch *CommandHandler) weatherHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	city := getStringOption(i, "city")

	respondDeferred(s, i)

	// Using wttr.in for free weather data
	resp, err := http.Get(fmt.Sprintf("https://wttr.in/%s?format=j1", url.QueryEscape(city)))
	if err != nil {
		followUp(s, i, "Failed to fetch weather data.")
		return
	}
	defer resp.Body.Close()

	var weather struct {
		CurrentCondition []struct {
			TempC        string `json:"temp_C"`
			TempF        string `json:"temp_F"`
			FeelsLikeC   string `json:"FeelsLikeC"`
			Humidity     string `json:"humidity"`
			WeatherDesc  []struct{ Value string } `json:"weatherDesc"`
			WindspeedKmph string `json:"windspeedKmph"`
		} `json:"current_condition"`
		NearestArea []struct {
			AreaName []struct{ Value string } `json:"areaName"`
			Country  []struct{ Value string } `json:"country"`
		} `json:"nearest_area"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil || len(weather.CurrentCondition) == 0 {
		followUp(s, i, "Could not find weather for that location.")
		return
	}

	current := weather.CurrentCondition[0]
	location := city
	if len(weather.NearestArea) > 0 && len(weather.NearestArea[0].AreaName) > 0 {
		location = weather.NearestArea[0].AreaName[0].Value
		if len(weather.NearestArea[0].Country) > 0 {
			location += ", " + weather.NearestArea[0].Country[0].Value
		}
	}

	desc := "Clear"
	if len(current.WeatherDesc) > 0 {
		desc = current.WeatherDesc[0].Value
	}

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("Weather in %s", location),
		Color: 0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Condition", Value: desc, Inline: true},
			{Name: "Temperature", Value: fmt.Sprintf("%s°C / %s°F", current.TempC, current.TempF), Inline: true},
			{Name: "Feels Like", Value: fmt.Sprintf("%s°C", current.FeelsLikeC), Inline: true},
			{Name: "Humidity", Value: current.Humidity + "%", Inline: true},
			{Name: "Wind", Value: current.WindspeedKmph + " km/h", Inline: true},
		},
	}

	followUpEmbed(s, i, embed)
}

func (ch *CommandHandler) urbanHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	term := getStringOption(i, "term")

	respondDeferred(s, i)

	resp, err := http.Get(fmt.Sprintf("https://api.urbandictionary.com/v0/define?term=%s", url.QueryEscape(term)))
	if err != nil {
		followUp(s, i, "Failed to fetch definition.")
		return
	}
	defer resp.Body.Close()

	var data struct {
		List []struct {
			Word       string `json:"word"`
			Definition string `json:"definition"`
			Example    string `json:"example"`
			ThumbsUp   int    `json:"thumbs_up"`
			ThumbsDown int    `json:"thumbs_down"`
			Author     string `json:"author"`
		} `json:"list"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil || len(data.List) == 0 {
		followUp(s, i, "No definition found.")
		return
	}

	def := data.List[0]

	// Clean up definition (remove brackets)
	definition := strings.ReplaceAll(def.Definition, "[", "")
	definition = strings.ReplaceAll(definition, "]", "")
	if len(definition) > 1024 {
		definition = definition[:1021] + "..."
	}

	example := strings.ReplaceAll(def.Example, "[", "")
	example = strings.ReplaceAll(example, "]", "")
	if len(example) > 1024 {
		example = example[:1021] + "..."
	}

	embed := &discordgo.MessageEmbed{
		Title:       def.Word,
		Description: definition,
		Color:       0xEFFF00,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Example", Value: example, Inline: false},
			{Name: "Rating", Value: fmt.Sprintf("👍 %d | 👎 %d", def.ThumbsUp, def.ThumbsDown), Inline: true},
			{Name: "Author", Value: def.Author, Inline: true},
		},
	}

	followUpEmbed(s, i, embed)
}

func (ch *CommandHandler) wikiHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	query := getStringOption(i, "query")

	respondDeferred(s, i)

	resp, err := http.Get(fmt.Sprintf("https://en.wikipedia.org/api/rest_v1/page/summary/%s", url.QueryEscape(query)))
	if err != nil {
		followUp(s, i, "Failed to search Wikipedia.")
		return
	}
	defer resp.Body.Close()

	var data struct {
		Title       string `json:"title"`
		Extract     string `json:"extract"`
		ContentURLs struct {
			Desktop struct {
				Page string `json:"page"`
			} `json:"desktop"`
		} `json:"content_urls"`
		Thumbnail struct {
			Source string `json:"source"`
		} `json:"thumbnail"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil || data.Title == "" {
		followUp(s, i, "No Wikipedia article found.")
		return
	}

	extract := data.Extract
	if len(extract) > 2048 {
		extract = extract[:2045] + "..."
	}

	embed := &discordgo.MessageEmbed{
		Title:       data.Title,
		URL:         data.ContentURLs.Desktop.Page,
		Description: extract,
		Color:       0xFFFFFF,
	}

	if data.Thumbnail.Source != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: data.Thumbnail.Source}
	}

	followUpEmbed(s, i, embed)
}

func (ch *CommandHandler) ipLookupHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ip := getStringOption(i, "ip")

	respondDeferred(s, i)

	resp, err := http.Get(fmt.Sprintf("http://ip-api.com/json/%s", ip))
	if err != nil {
		followUp(s, i, "Failed to look up IP.")
		return
	}
	defer resp.Body.Close()

	var data struct {
		Status      string  `json:"status"`
		Country     string  `json:"country"`
		CountryCode string  `json:"countryCode"`
		Region      string  `json:"regionName"`
		City        string  `json:"city"`
		Zip         string  `json:"zip"`
		Lat         float64 `json:"lat"`
		Lon         float64 `json:"lon"`
		Timezone    string  `json:"timezone"`
		ISP         string  `json:"isp"`
		Org         string  `json:"org"`
		AS          string  `json:"as"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil || data.Status != "success" {
		followUp(s, i, "Could not find information for that IP.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("IP Lookup: %s", ip),
		Color: 0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Location", Value: fmt.Sprintf("%s, %s, %s", data.City, data.Region, data.Country), Inline: false},
			{Name: "Coordinates", Value: fmt.Sprintf("%.4f, %.4f", data.Lat, data.Lon), Inline: true},
			{Name: "Timezone", Value: data.Timezone, Inline: true},
			{Name: "ISP", Value: data.ISP, Inline: false},
			{Name: "Organization", Value: data.Org, Inline: true},
			{Name: "AS", Value: data.AS, Inline: true},
		},
	}

	followUpEmbed(s, i, embed)
}

func (ch *CommandHandler) cryptoHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	coin := getStringOption(i, "coin")

	respondDeferred(s, i)

	resp, err := http.Get(fmt.Sprintf("https://api.coingecko.com/api/v3/coins/%s", strings.ToLower(coin)))
	if err != nil {
		followUp(s, i, "Failed to fetch crypto data.")
		return
	}
	defer resp.Body.Close()

	var data struct {
		Name   string `json:"name"`
		Symbol string `json:"symbol"`
		Image  struct {
			Large string `json:"large"`
		} `json:"image"`
		MarketData struct {
			CurrentPrice struct {
				USD float64 `json:"usd"`
				EUR float64 `json:"eur"`
				GBP float64 `json:"gbp"`
			} `json:"current_price"`
			PriceChangePercentage24h float64 `json:"price_change_percentage_24h"`
			MarketCap                struct {
				USD float64 `json:"usd"`
			} `json:"market_cap"`
			TotalVolume struct {
				USD float64 `json:"usd"`
			} `json:"total_volume"`
			High24h struct {
				USD float64 `json:"usd"`
			} `json:"high_24h"`
			Low24h struct {
				USD float64 `json:"usd"`
			} `json:"low_24h"`
		} `json:"market_data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil || data.Name == "" {
		followUp(s, i, "Could not find that cryptocurrency.")
		return
	}

	changeEmoji := "📈"
	if data.MarketData.PriceChangePercentage24h < 0 {
		changeEmoji = "📉"
	}

	embed := &discordgo.MessageEmbed{
		Title:     fmt.Sprintf("%s (%s)", data.Name, strings.ToUpper(data.Symbol)),
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: data.Image.Large},
		Color:     0xF7931A,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Price (USD)", Value: fmt.Sprintf("$%.2f", data.MarketData.CurrentPrice.USD), Inline: true},
			{Name: "Price (EUR)", Value: fmt.Sprintf("€%.2f", data.MarketData.CurrentPrice.EUR), Inline: true},
			{Name: "24h Change", Value: fmt.Sprintf("%s %.2f%%", changeEmoji, data.MarketData.PriceChangePercentage24h), Inline: true},
			{Name: "24h High", Value: fmt.Sprintf("$%.2f", data.MarketData.High24h.USD), Inline: true},
			{Name: "24h Low", Value: fmt.Sprintf("$%.2f", data.MarketData.Low24h.USD), Inline: true},
			{Name: "Market Cap", Value: formatLargeNumber(data.MarketData.MarketCap.USD), Inline: true},
		},
	}

	followUpEmbed(s, i, embed)
}

func (ch *CommandHandler) mcServerHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	address := getStringOption(i, "address")

	respondDeferred(s, i)

	resp, err := http.Get(fmt.Sprintf("https://api.mcsrvstat.us/3/%s", url.QueryEscape(address)))
	if err != nil {
		followUp(s, i, "Failed to fetch server data.")
		return
	}
	defer resp.Body.Close()

	var data struct {
		Online  bool   `json:"online"`
		IP      string `json:"ip"`
		Port    int    `json:"port"`
		Version string `json:"version"`
		Players struct {
			Online int `json:"online"`
			Max    int `json:"max"`
		} `json:"players"`
		Motd struct {
			Clean []string `json:"clean"`
		} `json:"motd"`
		Icon string `json:"icon"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		followUp(s, i, "Could not parse server data.")
		return
	}

	status := "🔴 Offline"
	color := 0xED4245
	if data.Online {
		status = "🟢 Online"
		color = 0x57F287
	}

	embed := &discordgo.MessageEmbed{
		Title: address,
		Color: color,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Status", Value: status, Inline: true},
		},
	}

	if data.Online {
		motd := "N/A"
		if len(data.Motd.Clean) > 0 {
			motd = strings.Join(data.Motd.Clean, "\n")
		}

		embed.Fields = append(embed.Fields,
			&discordgo.MessageEmbedField{Name: "Players", Value: fmt.Sprintf("%d/%d", data.Players.Online, data.Players.Max), Inline: true},
			&discordgo.MessageEmbedField{Name: "Version", Value: data.Version, Inline: true},
			&discordgo.MessageEmbedField{Name: "IP", Value: fmt.Sprintf("%s:%d", data.IP, data.Port), Inline: true},
			&discordgo.MessageEmbedField{Name: "MOTD", Value: motd, Inline: false},
		)
	}

	followUpEmbed(s, i, embed)
}

func (ch *CommandHandler) githubHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	username := getStringOption(i, "username")

	respondDeferred(s, i)

	resp, err := http.Get(fmt.Sprintf("https://api.github.com/users/%s", url.QueryEscape(username)))
	if err != nil {
		followUp(s, i, "Failed to fetch GitHub data.")
		return
	}
	defer resp.Body.Close()

	var data struct {
		Login       string `json:"login"`
		Name        string `json:"name"`
		AvatarURL   string `json:"avatar_url"`
		HTMLURL     string `json:"html_url"`
		Bio         string `json:"bio"`
		PublicRepos int    `json:"public_repos"`
		Followers   int    `json:"followers"`
		Following   int    `json:"following"`
		CreatedAt   string `json:"created_at"`
		Company     string `json:"company"`
		Location    string `json:"location"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil || data.Login == "" {
		followUp(s, i, "Could not find that GitHub user.")
		return
	}

	name := data.Login
	if data.Name != "" {
		name = data.Name
	}

	embed := &discordgo.MessageEmbed{
		Title:       name,
		URL:         data.HTMLURL,
		Description: data.Bio,
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: data.AvatarURL},
		Color:       0x333333,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Repositories", Value: fmt.Sprintf("%d", data.PublicRepos), Inline: true},
			{Name: "Followers", Value: fmt.Sprintf("%d", data.Followers), Inline: true},
			{Name: "Following", Value: fmt.Sprintf("%d", data.Following), Inline: true},
		},
	}

	if data.Company != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name: "Company", Value: data.Company, Inline: true,
		})
	}
	if data.Location != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name: "Location", Value: data.Location, Inline: true,
		})
	}

	followUpEmbed(s, i, embed)
}

func (ch *CommandHandler) npmHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	pkg := getStringOption(i, "package")

	respondDeferred(s, i)

	resp, err := http.Get(fmt.Sprintf("https://registry.npmjs.org/%s", url.QueryEscape(pkg)))
	if err != nil {
		followUp(s, i, "Failed to fetch npm data.")
		return
	}
	defer resp.Body.Close()

	var data struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		DistTags    struct {
			Latest string `json:"latest"`
		} `json:"dist-tags"`
		Author struct {
			Name string `json:"name"`
		} `json:"author"`
		License    string `json:"license"`
		Homepage   string `json:"homepage"`
		Repository struct {
			URL string `json:"url"`
		} `json:"repository"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil || data.Name == "" {
		followUp(s, i, "Could not find that npm package.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       data.Name,
		URL:         fmt.Sprintf("https://www.npmjs.com/package/%s", data.Name),
		Description: data.Description,
		Color:       0xCB3837,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Version", Value: data.DistTags.Latest, Inline: true},
			{Name: "License", Value: data.License, Inline: true},
		},
	}

	if data.Author.Name != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name: "Author", Value: data.Author.Name, Inline: true,
		})
	}

	followUpEmbed(s, i, embed)
}

func (ch *CommandHandler) colorHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	hexStr := getStringOption(i, "hex")
	hexStr = strings.TrimPrefix(hexStr, "#")

	var color int64
	fmt.Sscanf(hexStr, "%x", &color)

	// Convert to RGB
	r := (color >> 16) & 0xFF
	g := (color >> 8) & 0xFF
	b := color & 0xFF

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("#%s", strings.ToUpper(hexStr)),
		Color: int(color),
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Hex", Value: fmt.Sprintf("#%s", strings.ToUpper(hexStr)), Inline: true},
			{Name: "RGB", Value: fmt.Sprintf("rgb(%d, %d, %d)", r, g, b), Inline: true},
			{Name: "Integer", Value: fmt.Sprintf("%d", color), Inline: true},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: fmt.Sprintf("https://singlecolorimage.com/get/%s/100x100", hexStr),
		},
	}

	respondEmbed(s, i, embed)
}

func formatLargeNumber(n float64) string {
	if n >= 1e12 {
		return fmt.Sprintf("$%.2fT", n/1e12)
	} else if n >= 1e9 {
		return fmt.Sprintf("$%.2fB", n/1e9)
	} else if n >= 1e6 {
		return fmt.Sprintf("$%.2fM", n/1e6)
	}
	return fmt.Sprintf("$%.2f", n)
}
