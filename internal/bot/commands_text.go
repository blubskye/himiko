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
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"unicode"

	"github.com/bwmarrin/discordgo"
)

func (ch *CommandHandler) registerTextCommands() {
	// ASCII art
	ch.Register(&Command{
		Name:        "ascii",
		Description: "Convert text to ASCII art",
		Category:    "Text",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "text",
				Description: "Text to convert",
				Required:    true,
			},
		},
		Handler: ch.asciiHandler,
	})

	// Zalgo text
	ch.Register(&Command{
		Name:        "zalgo",
		Description: "Convert text to zalgo style",
		Category:    "Text",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "text",
				Description: "Text to convert",
				Required:    true,
			},
		},
		Handler: ch.zalgoHandler,
	})

	// Reverse text
	ch.Register(&Command{
		Name:        "reverse",
		Description: "Reverse text",
		Category:    "Text",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "text",
				Description: "Text to reverse",
				Required:    true,
			},
		},
		Handler:       ch.reverseHandler,
		PrefixHandler: ch.reversePrefixHandler,
	})

	// Upside down
	ch.Register(&Command{
		Name:        "upsidedown",
		Description: "Flip text upside down",
		Category:    "Text",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "text",
				Description: "Text to flip",
				Required:    true,
			},
		},
		Handler: ch.upsidedownHandler,
	})

	// Morse code
	ch.Register(&Command{
		Name:        "morse",
		Description: "Convert text to morse code",
		Category:    "Text",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "text",
				Description: "Text to convert",
				Required:    true,
			},
		},
		Handler: ch.morseHandler,
	})

	// Vaporwave
	ch.Register(&Command{
		Name:        "vaporwave",
		Description: "Convert text to vaporwave style",
		Category:    "Text",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "text",
				Description: "Text to convert",
				Required:    true,
			},
		},
		Handler:       ch.vaporwaveHandler,
		PrefixHandler: ch.vaporwavePrefixHandler,
	})

	// OwO
	ch.Register(&Command{
		Name:        "owo",
		Description: "OwOify your text",
		Category:    "Text",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "text",
				Description: "Text to OwOify",
				Required:    true,
			},
		},
		Handler:       ch.owoHandler,
		PrefixHandler: ch.owoPrefixHandler,
	})

	// Smart/Mock text
	ch.Register(&Command{
		Name:        "mock",
		Description: "CoNvErT tExT tO mOcK sTyLe",
		Category:    "Text",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "text",
				Description: "Text to mock",
				Required:    true,
			},
		},
		Handler:       ch.mockHandler,
		PrefixHandler: ch.mockPrefixHandler,
	})

	// 1337 speak
	ch.Register(&Command{
		Name:        "leet",
		Description: "Convert text to 1337 speak",
		Category:    "Text",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "text",
				Description: "Text to convert",
				Required:    true,
			},
		},
		Handler:       ch.leetHandler,
		PrefixHandler: ch.leetPrefixHandler,
	})

	// Regional indicators (emoji letters)
	ch.Register(&Command{
		Name:        "regional",
		Description: "Convert text to regional indicator emojis",
		Category:    "Text",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "text",
				Description: "Text to convert",
				Required:    true,
			},
		},
		Handler: ch.regionalHandler,
	})

	// Spoiler each character
	ch.Register(&Command{
		Name:        "spoiler",
		Description: "Spoiler each character",
		Category:    "Text",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "text",
				Description: "Text to spoilerify",
				Required:    true,
			},
		},
		Handler: ch.spoilerHandler,
	})

	// Spaced text
	ch.Register(&Command{
		Name:        "space",
		Description: "Add spaces between characters",
		Category:    "Text",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "text",
				Description: "Text to space out",
				Required:    true,
			},
		},
		Handler:       ch.spaceHandler,
		PrefixHandler: ch.spacePrefixHandler,
	})

	// Italic/fancy text
	ch.Register(&Command{
		Name:        "fancy",
		Description: "Convert to fancy italic text",
		Category:    "Text",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "text",
				Description: "Text to convert",
				Required:    true,
			},
		},
		Handler: ch.fancyHandler,
	})

	// Encode/Decode
	ch.Register(&Command{
		Name:        "encode",
		Description: "Encode text to various formats",
		Category:    "Text",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "format",
				Description: "Encoding format",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "Base64", Value: "base64"},
					{Name: "Hex", Value: "hex"},
					{Name: "Binary", Value: "binary"},
					{Name: "ROT13", Value: "rot13"},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "text",
				Description: "Text to encode",
				Required:    true,
			},
		},
		Handler: ch.encodeHandler,
	})

	// Decode
	ch.Register(&Command{
		Name:        "decode",
		Description: "Decode text from various formats",
		Category:    "Text",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "format",
				Description: "Decoding format",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "Base64", Value: "base64"},
					{Name: "Hex", Value: "hex"},
					{Name: "Binary", Value: "binary"},
					{Name: "ROT13", Value: "rot13"},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "text",
				Description: "Text to decode",
				Required:    true,
			},
		},
		Handler: ch.decodeHandler,
	})

	// Codeblock
	ch.Register(&Command{
		Name:        "codeblock",
		Description: "Wrap text in a codeblock",
		Category:    "Text",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "text",
				Description: "Text to wrap",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "language",
				Description: "Code language for syntax highlighting",
				Required:    false,
			},
		},
		Handler:       ch.codeblockHandler,
		PrefixHandler: ch.codeblockPrefixHandler,
	})

	// Hyperlink
	ch.Register(&Command{
		Name:        "hyperlink",
		Description: "Create a hyperlink",
		Category:    "Text",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "text",
				Description: "Display text",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "url",
				Description: "URL to link to",
				Required:    true,
			},
		},
		Handler: ch.hyperlinkHandler,
	})
}

func (ch *CommandHandler) asciiHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	text := getStringOption(i, "text")

	if len(text) > 10 {
		respondEphemeral(s, i, "Text too long. Maximum 10 characters.")
		return
	}

	// Simple ASCII art font
	font := map[rune][]string{
		'A': {"  █  ", " █ █ ", "█████", "█   █", "█   █"},
		'B': {"████ ", "█   █", "████ ", "█   █", "████ "},
		'C': {" ████", "█    ", "█    ", "█    ", " ████"},
		'D': {"████ ", "█   █", "█   █", "█   █", "████ "},
		'E': {"█████", "█    ", "████ ", "█    ", "█████"},
		'F': {"█████", "█    ", "████ ", "█    ", "█    "},
		'G': {" ████", "█    ", "█  ██", "█   █", " ████"},
		'H': {"█   █", "█   █", "█████", "█   █", "█   █"},
		'I': {"█████", "  █  ", "  █  ", "  █  ", "█████"},
		'J': {"█████", "   █ ", "   █ ", "█  █ ", " ██  "},
		'K': {"█   █", "█  █ ", "███  ", "█  █ ", "█   █"},
		'L': {"█    ", "█    ", "█    ", "█    ", "█████"},
		'M': {"█   █", "██ ██", "█ █ █", "█   █", "█   █"},
		'N': {"█   █", "██  █", "█ █ █", "█  ██", "█   █"},
		'O': {" ███ ", "█   █", "█   █", "█   █", " ███ "},
		'P': {"████ ", "█   █", "████ ", "█    ", "█    "},
		'Q': {" ███ ", "█   █", "█ █ █", "█  █ ", " ██ █"},
		'R': {"████ ", "█   █", "████ ", "█  █ ", "█   █"},
		'S': {" ████", "█    ", " ███ ", "    █", "████ "},
		'T': {"█████", "  █  ", "  █  ", "  █  ", "  █  "},
		'U': {"█   █", "█   █", "█   █", "█   █", " ███ "},
		'V': {"█   █", "█   █", "█   █", " █ █ ", "  █  "},
		'W': {"█   █", "█   █", "█ █ █", "██ ██", "█   █"},
		'X': {"█   █", " █ █ ", "  █  ", " █ █ ", "█   █"},
		'Y': {"█   █", " █ █ ", "  █  ", "  █  ", "  █  "},
		'Z': {"█████", "   █ ", "  █  ", " █   ", "█████"},
		' ': {"     ", "     ", "     ", "     ", "     "},
	}

	var lines [5]string
	for _, char := range strings.ToUpper(text) {
		if art, ok := font[char]; ok {
			for j := 0; j < 5; j++ {
				lines[j] += art[j] + " "
			}
		}
	}

	result := "```\n"
	for _, line := range lines {
		result += line + "\n"
	}
	result += "```"

	if len(result) > 2000 {
		respondEphemeral(s, i, "Result too long to display.")
		return
	}

	respond(s, i, result)
}

func (ch *CommandHandler) zalgoHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	text := getStringOption(i, "text")

	// Zalgo combining characters (using hex codes for reliability)
	zalgoUp := []rune{0x030D, 0x030E, 0x0304, 0x0305, 0x033F, 0x0311, 0x0306, 0x0310, 0x0352, 0x0357, 0x0351, 0x0307, 0x0308, 0x030A, 0x0342, 0x0343, 0x0344, 0x034A, 0x034B, 0x034C, 0x0303, 0x0302, 0x030C, 0x0350, 0x0300, 0x0301, 0x030B, 0x030F, 0x0312, 0x0313, 0x0314, 0x033D, 0x033E, 0x035B, 0x0346, 0x031A}
	zalgoDown := []rune{0x0316, 0x0317, 0x0318, 0x0319, 0x031C, 0x031D, 0x031E, 0x031F, 0x0320, 0x0324, 0x0325, 0x0326, 0x0329, 0x032A, 0x032B, 0x032C, 0x032D, 0x032E, 0x032F, 0x0330, 0x0331, 0x0332, 0x0333, 0x0339, 0x033A, 0x033B, 0x033C, 0x0345, 0x0347, 0x0348, 0x0349, 0x034D, 0x034E, 0x0353, 0x0354, 0x0355, 0x0356, 0x0359, 0x035A, 0x0323}
	zalgoMid := []rune{0x0315, 0x031B, 0x0300, 0x0301, 0x0358, 0x0321, 0x0322, 0x0327, 0x0328, 0x0334, 0x0335, 0x0336, 0x035C, 0x035D, 0x035E, 0x035F, 0x0360, 0x0362, 0x0338, 0x0337, 0x0361}

	var result strings.Builder
	for _, char := range text {
		result.WriteRune(char)
		for j := 0; j < rand.Intn(3)+1; j++ {
			result.WriteRune(zalgoUp[rand.Intn(len(zalgoUp))])
		}
		for j := 0; j < rand.Intn(3)+1; j++ {
			result.WriteRune(zalgoMid[rand.Intn(len(zalgoMid))])
		}
		for j := 0; j < rand.Intn(3)+1; j++ {
			result.WriteRune(zalgoDown[rand.Intn(len(zalgoDown))])
		}
	}

	respond(s, i, result.String())
}

func (ch *CommandHandler) reverseHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	text := getStringOption(i, "text")

	runes := []rune(text)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	respond(s, i, string(runes))
}

func (ch *CommandHandler) upsidedownHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	text := getStringOption(i, "text")

	flipMap := map[rune]rune{
		'a': 'ɐ', 'b': 'q', 'c': 'ɔ', 'd': 'p', 'e': 'ǝ', 'f': 'ɟ', 'g': 'ƃ',
		'h': 'ɥ', 'i': 'ᴉ', 'j': 'ɾ', 'k': 'ʞ', 'l': 'l', 'm': 'ɯ', 'n': 'u',
		'o': 'o', 'p': 'd', 'q': 'b', 'r': 'ɹ', 's': 's', 't': 'ʇ', 'u': 'n',
		'v': 'ʌ', 'w': 'ʍ', 'x': 'x', 'y': 'ʎ', 'z': 'z',
		'A': '∀', 'B': 'q', 'C': 'Ɔ', 'D': 'p', 'E': 'Ǝ', 'F': 'Ⅎ', 'G': 'פ',
		'H': 'H', 'I': 'I', 'J': 'ſ', 'K': 'ʞ', 'L': '˥', 'M': 'W', 'N': 'N',
		'O': 'O', 'P': 'Ԁ', 'Q': 'Q', 'R': 'ɹ', 'S': 'S', 'T': '┴', 'U': '∩',
		'V': 'Λ', 'W': 'M', 'X': 'X', 'Y': '⅄', 'Z': 'Z',
		'1': 'Ɩ', '2': 'ᄅ', '3': 'Ɛ', '4': 'ㄣ', '5': 'ϛ', '6': '9', '7': 'ㄥ',
		'8': '8', '9': '6', '0': '0',
		'.': '˙', ',': 0x0027, '?': '¿', '!': '¡', 0x0027: ',', '"': '„',
		'(': ')', ')': '(', '[': ']', ']': '[', '{': '}', '}': '{',
		'<': '>', '>': '<', '&': '⅋', '_': '‾',
	}

	runes := []rune(text)
	// Reverse and flip
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	var result strings.Builder
	for _, r := range runes {
		if flipped, ok := flipMap[r]; ok {
			result.WriteRune(flipped)
		} else {
			result.WriteRune(r)
		}
	}

	respond(s, i, result.String())
}

func (ch *CommandHandler) morseHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	text := getStringOption(i, "text")

	morseMap := map[rune]string{
		'a': ".-", 'b': "-...", 'c': "-.-.", 'd': "-..", 'e': ".", 'f': "..-.",
		'g': "--.", 'h': "....", 'i': "..", 'j': ".---", 'k': "-.-", 'l': ".-..",
		'm': "--", 'n': "-.", 'o': "---", 'p': ".--.", 'q': "--.-", 'r': ".-.",
		's': "...", 't': "-", 'u': "..-", 'v': "...-", 'w': ".--", 'x': "-..-",
		'y': "-.--", 'z': "--..",
		'0': "-----", '1': ".----", '2': "..---", '3': "...--", '4': "....-",
		'5': ".....", '6': "-....", '7': "--...", '8': "---..", '9': "----.",
		' ': "/",
	}

	var result strings.Builder
	for _, char := range strings.ToLower(text) {
		if morse, ok := morseMap[char]; ok {
			result.WriteString(morse)
			result.WriteString(" ")
		}
	}

	respond(s, i, result.String())
}

func (ch *CommandHandler) vaporwaveHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	text := getStringOption(i, "text")

	var result strings.Builder
	for _, char := range text {
		if char >= '!' && char <= '~' {
			result.WriteRune(rune(int(char) + 0xFEE0))
		} else if char == ' ' {
			result.WriteRune('　') // Full-width space
		} else {
			result.WriteRune(char)
		}
	}

	respond(s, i, result.String())
}

func (ch *CommandHandler) owoHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	text := getStringOption(i, "text")

	faces := []string{" owo", " UwU", " >w<", " ^w^", " OwO", " :3", " x3"}

	result := strings.ToLower(text)
	result = strings.ReplaceAll(result, "r", "w")
	result = strings.ReplaceAll(result, "l", "w")
	result = strings.ReplaceAll(result, "R", "W")
	result = strings.ReplaceAll(result, "L", "W")
	result = strings.ReplaceAll(result, "no", "nyo")
	result = strings.ReplaceAll(result, "No", "Nyo")
	result = strings.ReplaceAll(result, "NO", "NYO")
	result = strings.ReplaceAll(result, "na", "nya")
	result = strings.ReplaceAll(result, "Na", "Nya")
	result = strings.ReplaceAll(result, "NA", "NYA")

	// Add random face at the end
	result += faces[rand.Intn(len(faces))]

	respond(s, i, result)
}

func (ch *CommandHandler) mockHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	text := getStringOption(i, "text")

	var result strings.Builder
	upper := false
	for _, char := range text {
		if unicode.IsLetter(char) {
			if upper {
				result.WriteRune(unicode.ToUpper(char))
			} else {
				result.WriteRune(unicode.ToLower(char))
			}
			upper = !upper
		} else {
			result.WriteRune(char)
		}
	}

	respond(s, i, result.String())
}

func (ch *CommandHandler) leetHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	text := getStringOption(i, "text")

	leetMap := map[rune]string{
		'a': "4", 'A': "4",
		'b': "8", 'B': "8",
		'e': "3", 'E': "3",
		'g': "9", 'G': "9",
		'i': "1", 'I': "1",
		'l': "1", 'L': "1",
		'o': "0", 'O': "0",
		's': "5", 'S': "5",
		't': "7", 'T': "7",
	}

	var result strings.Builder
	for _, char := range text {
		if leet, ok := leetMap[char]; ok {
			result.WriteString(leet)
		} else {
			result.WriteRune(char)
		}
	}

	respond(s, i, result.String())
}

func (ch *CommandHandler) regionalHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	text := getStringOption(i, "text")

	var result strings.Builder
	for _, char := range strings.ToLower(text) {
		if char >= 'a' && char <= 'z' {
			result.WriteString(fmt.Sprintf(":regional_indicator_%c: ", char))
		} else if char == ' ' {
			result.WriteString("   ")
		} else if char >= '0' && char <= '9' {
			numbers := []string{":zero:", ":one:", ":two:", ":three:", ":four:", ":five:", ":six:", ":seven:", ":eight:", ":nine:"}
			result.WriteString(numbers[char-'0'] + " ")
		} else {
			result.WriteRune(char)
		}
	}

	if len(result.String()) > 2000 {
		respondEphemeral(s, i, "Result too long to display.")
		return
	}

	respond(s, i, result.String())
}

func (ch *CommandHandler) spoilerHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	text := getStringOption(i, "text")

	var result strings.Builder
	for _, char := range text {
		if char == ' ' {
			result.WriteString(" ")
		} else {
			result.WriteString("||")
			result.WriteRune(char)
			result.WriteString("||")
		}
	}

	if len(result.String()) > 2000 {
		respondEphemeral(s, i, "Result too long to display.")
		return
	}

	respond(s, i, result.String())
}

func (ch *CommandHandler) spaceHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	text := getStringOption(i, "text")

	var result strings.Builder
	runes := []rune(text)
	for idx, char := range runes {
		result.WriteRune(char)
		if idx < len(runes)-1 {
			result.WriteString(" ")
		}
	}

	respond(s, i, result.String())
}

func (ch *CommandHandler) fancyHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	text := getStringOption(i, "text")

	// Mathematical italic characters
	var result strings.Builder
	for _, char := range text {
		if char >= 'a' && char <= 'z' {
			result.WriteRune(rune(0x1D44E + int(char-'a')))
		} else if char >= 'A' && char <= 'Z' {
			result.WriteRune(rune(0x1D434 + int(char-'A')))
		} else {
			result.WriteRune(char)
		}
	}

	respond(s, i, result.String())
}

func (ch *CommandHandler) encodeHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	format := getStringOption(i, "format")
	text := getStringOption(i, "text")

	var result string
	switch format {
	case "base64":
		result = base64.StdEncoding.EncodeToString([]byte(text))
	case "hex":
		result = hex.EncodeToString([]byte(text))
	case "binary":
		var binary strings.Builder
		for _, char := range text {
			binary.WriteString(fmt.Sprintf("%08b ", char))
		}
		result = strings.TrimSpace(binary.String())
	case "rot13":
		result = rot13(text)
	}

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("Encoded (%s)", format),
		Description: fmt.Sprintf("```\n%s\n```", result),
		Color: 0x5865F2,
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) decodeHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	format := getStringOption(i, "format")
	text := getStringOption(i, "text")

	var result string
	var err error
	switch format {
	case "base64":
		decoded, e := base64.StdEncoding.DecodeString(text)
		if e != nil {
			err = e
		} else {
			result = string(decoded)
		}
	case "hex":
		decoded, e := hex.DecodeString(text)
		if e != nil {
			err = e
		} else {
			result = string(decoded)
		}
	case "binary":
		parts := strings.Fields(text)
		var bytes []byte
		for _, part := range parts {
			var b byte
			fmt.Sscanf(part, "%08b", &b)
			bytes = append(bytes, b)
		}
		result = string(bytes)
	case "rot13":
		result = rot13(text)
	}

	if err != nil {
		respondEphemeral(s, i, "Failed to decode: "+err.Error())
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("Decoded (%s)", format),
		Description: fmt.Sprintf("```\n%s\n```", result),
		Color: 0x5865F2,
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) codeblockHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	text := getStringOption(i, "text")
	lang := getStringOption(i, "language")

	respond(s, i, fmt.Sprintf("```%s\n%s\n```", lang, text))
}

func (ch *CommandHandler) hyperlinkHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	text := getStringOption(i, "text")
	url := getStringOption(i, "url")

	respond(s, i, fmt.Sprintf("[%s](%s)", text, url))
}

func rot13(text string) string {
	var result strings.Builder
	for _, char := range text {
		if char >= 'a' && char <= 'z' {
			result.WriteRune('a' + (char-'a'+13)%26)
		} else if char >= 'A' && char <= 'Z' {
			result.WriteRune('A' + (char-'A'+13)%26)
		} else {
			result.WriteRune(char)
		}
	}
	return result.String()
}

// Prefix handlers for text commands

func (ch *CommandHandler) codeblockPrefixHandler(ctx *PrefixContext) {
	if len(ctx.Args) == 0 {
		ctx.Reply("Usage: `" + ctx.Prefix + "codeblock <text>` or `" + ctx.Prefix + "codeblock <lang> <text>`")
		return
	}

	var lang, text string
	if len(ctx.Args) >= 2 {
		// Check if first arg looks like a language
		possibleLang := ctx.Args[0]
		if len(possibleLang) <= 10 && !strings.Contains(possibleLang, " ") {
			lang = possibleLang
			text = ctx.GetArgRest(1)
		} else {
			text = ctx.GetArgRest(0)
		}
	} else {
		text = ctx.GetArgRest(0)
	}

	ctx.Reply(fmt.Sprintf("```%s\n%s\n```", lang, text))
}

func (ch *CommandHandler) reversePrefixHandler(ctx *PrefixContext) {
	if len(ctx.Args) == 0 {
		ctx.Reply("Usage: `" + ctx.Prefix + "reverse <text>`")
		return
	}
	text := ctx.GetArgRest(0)
	runes := []rune(text)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	ctx.Reply(string(runes))
}

func (ch *CommandHandler) mockPrefixHandler(ctx *PrefixContext) {
	if len(ctx.Args) == 0 {
		ctx.Reply("Usage: `" + ctx.Prefix + "mock <text>`")
		return
	}
	text := ctx.GetArgRest(0)
	var result strings.Builder
	upper := false
	for _, char := range text {
		if unicode.IsLetter(char) {
			if upper {
				result.WriteRune(unicode.ToUpper(char))
			} else {
				result.WriteRune(unicode.ToLower(char))
			}
			upper = !upper
		} else {
			result.WriteRune(char)
		}
	}
	ctx.Reply(result.String())
}

func (ch *CommandHandler) owoPrefixHandler(ctx *PrefixContext) {
	if len(ctx.Args) == 0 {
		ctx.Reply("Usage: `" + ctx.Prefix + "owo <text>`")
		return
	}
	text := ctx.GetArgRest(0)
	replacer := strings.NewReplacer(
		"r", "w", "R", "W",
		"l", "w", "L", "W",
		"ove", "uv",
		"OVE", "UV",
	)
	result := replacer.Replace(text)
	faces := []string{" OwO", " UwU", " >w<", " ^w^", " :3", " nyaa~"}
	result += faces[rand.Intn(len(faces))]
	ctx.Reply(result)
}

func (ch *CommandHandler) vaporwavePrefixHandler(ctx *PrefixContext) {
	if len(ctx.Args) == 0 {
		ctx.Reply("Usage: `" + ctx.Prefix + "vaporwave <text>`")
		return
	}
	text := ctx.GetArgRest(0)
	var result strings.Builder
	for _, char := range text {
		if char >= '!' && char <= '~' {
			result.WriteRune(char + 0xFEE0)
		} else if char == ' ' {
			result.WriteString("  ")
		} else {
			result.WriteRune(char)
		}
	}
	ctx.Reply(result.String())
}

func (ch *CommandHandler) leetPrefixHandler(ctx *PrefixContext) {
	if len(ctx.Args) == 0 {
		ctx.Reply("Usage: `" + ctx.Prefix + "leet <text>`")
		return
	}
	text := ctx.GetArgRest(0)
	leetMap := map[rune]string{
		'a': "4", 'A': "4", 'b': "8", 'B': "8",
		'e': "3", 'E': "3", 'g': "9", 'G': "9",
		'i': "1", 'I': "1", 'l': "1", 'L': "1",
		'o': "0", 'O': "0", 's': "5", 'S': "5",
		't': "7", 'T': "7",
	}
	var result strings.Builder
	for _, char := range text {
		if leet, ok := leetMap[char]; ok {
			result.WriteString(leet)
		} else {
			result.WriteRune(char)
		}
	}
	ctx.Reply(result.String())
}

func (ch *CommandHandler) spacePrefixHandler(ctx *PrefixContext) {
	if len(ctx.Args) == 0 {
		ctx.Reply("Usage: `" + ctx.Prefix + "space <text>`")
		return
	}
	text := ctx.GetArgRest(0)
	var result strings.Builder
	runes := []rune(text)
	for i, char := range runes {
		result.WriteRune(char)
		if i < len(runes)-1 {
			result.WriteString(" ")
		}
	}
	ctx.Reply(result.String())
}
