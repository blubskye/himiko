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
	"regexp"
	"strconv"
	"strings"

	"github.com/blubskye/himiko/internal/database"
	"github.com/bwmarrin/discordgo"
)

func (ch *CommandHandler) registerRanksCommands() {
	// Add rank
	ch.Register(&Command{
		Name:        "addrank",
		Description: "Add a level rank reward",
		Category:    "Ranks",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionRole,
				Name:        "role",
				Description: "Role to assign at level",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "level",
				Description: "Level required for this role",
				Required:    true,
				MinValue:    floatPtr(1),
				MaxValue:    1000,
			},
		},
		Handler: ch.addRankHandler,
	})

	// Remove rank
	ch.Register(&Command{
		Name:        "removerank",
		Description: "Remove a level rank reward",
		Category:    "Ranks",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionRole,
				Name:        "role",
				Description: "Role to remove from ranks",
				Required:    true,
			},
		},
		Handler: ch.removeRankHandler,
	})

	// List ranks
	ch.Register(&Command{
		Name:        "listranks",
		Description: "List all level rank rewards",
		Category:    "Ranks",
		Handler:     ch.listRanksHandler,
	})

	// Sync ranks from role names
	ch.Register(&Command{
		Name:        "syncranks",
		Description: "Auto-detect ranks from role names (e.g., 'Member (Lvl 5+)')",
		Category:    "Ranks",
		Handler:     ch.syncRanksHandler,
	})

	// Apply ranks to a user
	ch.Register(&Command{
		Name:        "applyranks",
		Description: "Apply appropriate rank roles to a user based on their level",
		Category:    "Ranks",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "User to apply ranks to (leave empty for all users)",
				Required:    false,
			},
		},
		Handler: ch.applyRanksHandler,
	})
}

func (ch *CommandHandler) addRankHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to manage ranks.")
		return
	}

	role := getRoleOption(i, "role")
	level := int(getIntOption(i, "level"))

	if role == nil {
		respondEphemeral(s, i, "Please specify a role.")
		return
	}

	err := ch.bot.DB.AddLevelRank(i.GuildID, role.ID, level)
	if err != nil {
		respondEphemeral(s, i, "Failed to add rank.")
		return
	}

	embed := successEmbed("Rank Added",
		fmt.Sprintf("%s will be given at level **%d**", role.Mention(), level))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) removeRankHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to manage ranks.")
		return
	}

	role := getRoleOption(i, "role")
	if role == nil {
		respondEphemeral(s, i, "Please specify a role.")
		return
	}

	err := ch.bot.DB.RemoveLevelRank(i.GuildID, role.ID)
	if err != nil {
		respondEphemeral(s, i, "Failed to remove rank.")
		return
	}

	embed := successEmbed("Rank Removed",
		fmt.Sprintf("%s has been removed from rank rewards", role.Mention()))
	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) listRanksHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ranks, err := ch.bot.DB.GetLevelRanks(i.GuildID)
	if err != nil {
		respondEphemeral(s, i, "Failed to get ranks.")
		return
	}

	if len(ranks) == 0 {
		respondEphemeral(s, i, "No rank rewards configured. Use `/addrank` to add some!")
		return
	}

	var description strings.Builder
	for _, r := range ranks {
		description.WriteString(fmt.Sprintf("**Level %d** → <@&%s>\n", r.Level, r.RoleID))
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Level Rank Rewards (%d)", len(ranks)),
		Description: description.String(),
		Color:       0x5865F2,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Users will automatically receive roles when reaching these levels",
		},
	}

	respondEmbed(s, i, embed)
}

func (ch *CommandHandler) syncRanksHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to sync ranks.")
		return
	}

	respondDeferred(s, i)

	// Get all roles
	roles, err := s.GuildRoles(i.GuildID)
	if err != nil {
		followUp(s, i, "Failed to get server roles.")
		return
	}

	// Pattern to match "(Lvl X+)" or "(Level X+)" in role names
	pattern := regexp.MustCompile(`\((?:Lvl|Level)\s*(\d+)\+?\)`)

	synced := 0
	for _, role := range roles {
		matches := pattern.FindStringSubmatch(role.Name)
		if len(matches) >= 2 {
			level, err := strconv.Atoi(matches[1])
			if err == nil && level > 0 {
				err = ch.bot.DB.AddLevelRank(i.GuildID, role.ID, level)
				if err == nil {
					synced++
				}
			}
		}
	}

	if synced == 0 {
		followUp(s, i, "No roles found with level patterns (e.g., 'Member (Lvl 5+)')")
		return
	}

	embed := successEmbed("Ranks Synced",
		fmt.Sprintf("Synced **%d** rank roles from role names", synced))
	followUpEmbed(s, i, embed)
}

func (ch *CommandHandler) applyRanksHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		respondEphemeral(s, i, "You need administrator permission to apply ranks.")
		return
	}

	user := getUserOption(i, "user")

	respondDeferred(s, i)

	ranks, err := ch.bot.DB.GetLevelRanks(i.GuildID)
	if err != nil || len(ranks) == 0 {
		followUp(s, i, "No rank rewards configured.")
		return
	}

	if user != nil {
		// Apply to single user
		applied, err := ch.applyRanksToUser(s, i.GuildID, user.ID, ranks)
		if err != nil {
			followUp(s, i, fmt.Sprintf("Failed to apply ranks: %v", err))
			return
		}
		embed := successEmbed("Ranks Applied",
			fmt.Sprintf("Applied **%d** rank roles to %s", applied, user.Mention()))
		followUpEmbed(s, i, embed)
	} else {
		// Apply to all users with XP
		members, err := s.GuildMembers(i.GuildID, "", 1000)
		if err != nil {
			followUp(s, i, "Failed to get server members.")
			return
		}

		totalApplied := 0
		usersUpdated := 0
		for _, member := range members {
			applied, _ := ch.applyRanksToUser(s, i.GuildID, member.User.ID, ranks)
			if applied > 0 {
				totalApplied += applied
				usersUpdated++
			}
		}

		embed := successEmbed("Ranks Applied",
			fmt.Sprintf("Applied **%d** rank roles to **%d** users", totalApplied, usersUpdated))
		followUpEmbed(s, i, embed)
	}
}

func (ch *CommandHandler) applyRanksToUser(s *discordgo.Session, guildID, userID string, ranks []database.LevelRank) (int, error) {
	xpData, err := ch.bot.DB.GetUserXP(guildID, userID)
	if err != nil {
		return 0, err
	}

	applied := 0
	for _, rank := range ranks {
		if xpData.Level >= rank.Level {
			err := s.GuildMemberRoleAdd(guildID, userID, rank.RoleID)
			if err == nil {
				applied++
			}
		}
	}

	return applied, nil
}
