package discord

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (guild *GuildState) handleGameEndMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	// unmute all players
	guild.handleTrackedMembers(s, false, false)
	// clear any existing game state message
	guild.Room = ""
	guild.Region = ""
	guild.clearGameTracking(s)
}

func (guild *GuildState) handleGameStartMessage(s *discordgo.Session, m *discordgo.MessageCreate, room string, region string) {
	guild.Room = room
	guild.Region = region

	guild.clearGameTracking(s)

	guild.GameStateMessage = sendMessage(s, m.ChannelID, gameStateResponse(guild))
	log.Println("Added self game state message")

	for _, e := range guild.StatusEmojis[true] {
		addReaction(s, guild.GameStateMessage.ChannelID, guild.GameStateMessage.ID, e.FormatForReaction())
	}
}

func (guild *GuildState) handleGameStateMessage(s *discordgo.Session) {
	if guild.GameStateMessage == nil {
		//log.Println("Game State Message is scuffed, try .au start again!")
		return
	}
	editMessage(s, guild.GameStateMessage.ChannelID, guild.GameStateMessage.ID, gameStateResponse(guild))
}

// TODO this probably deals with too much direct state-changing;
//probably want to bubble it up to some higher authority?
func (guild *GuildState) handleReactionGameStartAdd(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	if guild.GameStateMessage != nil {

		//verify that the user is reacting to the state/status message
		if IsUserReactionToStateMsg(m, guild.GameStateMessage) {
			for color, e := range guild.StatusEmojis[true] {
				if e.ID == m.Emoji.ID {
					log.Printf("Player %s reacted with color %s", m.UserID, GetColorStringForInt(color))

					//pair up the discord user with the relevant in-game data, matching by the color
					_, matched := guild.matchByColor(m.UserID, GetColorStringForInt(color), guild.AmongUsData)

					//then remove the player's reaction if we matched, or if we didn't
					err := s.MessageReactionRemove(m.ChannelID, m.MessageID, e.FormatForReaction(), m.UserID)
					if err != nil {
						log.Println(err)
					}

					if matched {
						guild.handleGameStateMessage(s)
					}
					break
				}
			}

		}
	}
}

func (guild *GuildState) handlePlayerAddMessage(s *discordgo.Session, m *discordgo.MessageCreate, name string, color string) bool {
	// we need to determine if it is a valid color
	// then we need to matchByColor
	if strings.HasPrefix(name, "<@!") && strings.HasSuffix(name, ">") && IsColorString(color) {
		//strip the special characters off front and end
		idLookup := name[3 : len(name)-1]
		g, err := s.State.Guild(guild.ID)
		if err != nil {
			log.Println(err)
		}
		for _, member := range g.Members {
			if idLookup == member.User.Username {
				_, matched := guild.matchByColor(member.User.ID, color, guild.AmongUsData)
				return matched
			}
		}
	}

	return false
}

// sendMessage provides a single interface to send a message to a channel via discord
func sendMessage(s *discordgo.Session, channelID string, message string) *discordgo.Message {
	msg, err := s.ChannelMessageSend(channelID, message)
	if err != nil {
		log.Println(err)
	}
	return msg
}

// editMessage provides a single interface to edit a message in a channel via discord
func editMessage(s *discordgo.Session, channelID string, messageID string, message string) *discordgo.Message {
	msg, err := s.ChannelMessageEdit(channelID, messageID, message)
	if err != nil {
		log.Println(err)
	}
	return msg
}

func deleteMessage(s *discordgo.Session, channelID string, messageID string) {
	err := s.ChannelMessageDelete(channelID, messageID)
	if err != nil {
		log.Println(err)
	}
}

func addReaction(s *discordgo.Session, channelID, messageID, emojiID string) {
	err := s.MessageReactionAdd(channelID, messageID, emojiID)
	if err != nil {
		log.Println(err)
	}
}

func removeAllReactions(s *discordgo.Session, channelID, messageID string) {
	err := s.MessageReactionsRemoveAll(channelID, messageID)
	if err != nil {
		log.Println(err)
	}
}
