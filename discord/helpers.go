package discord

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func guildMemberDeafenAndMute(s *discordgo.Session, guildID string, userID string, deaf bool, mute bool) (err error) {
	log.Printf("Issuing mute=%v deaf=%v request to discord\n", mute, deaf)
	data := struct {
		Deaf bool `json:"deaf"`
		Mute bool `json:"mute"`
	}{deaf, mute}

	_, err = s.RequestWithBucketID("PATCH", discordgo.EndpointGuildMember(guildID, userID), data, discordgo.EndpointGuildMember(guildID, ""))
	return
}

func guildMemberMute(session *discordgo.Session, guildID, userID string, mute bool) (err error) {
	log.Printf("Issuing mute=%v request to discord\n", mute)
	data := struct {
		Mute bool `json:"mute"`
	}{mute}

	_, err = session.RequestWithBucketID("PATCH", discordgo.EndpointGuildMember(guildID, userID), data, discordgo.EndpointGuildMember(guildID, ""))
	return
}

func isVoiceChannelTracked(channelID string, trackedChannels map[string]Tracking) bool {
	for _, v := range trackedChannels {
		if v.channelID == channelID {
			return true
		}
	}
	return false
}

func (guild *GuildState) matchByColor(userID, text string, allAuData map[string]*AmongUserData) (string, bool) {
	//guild.AmongUsDataLock.Lock()
	//defer guild.AmongUsDataLock.Unlock()

	for _, auData := range allAuData {
		if GetColorStringForInt(auData.Color) == strings.ToLower(text) {
			if user, ok := guild.UserData[userID]; ok {
				user.auData = auData //point to the single copy in memory
				//user.visualTrack = true
				guild.UserData[userID] = user
				log.Printf("Linked %s to %s", userID, user.auData.ToString())
				return fmt.Sprintf("Successfully linked player via Color!"), true
			}
			return fmt.Sprintf("No user found with userID %s", userID), false
		}
	}
	return fmt.Sprintf(":x: No in-game player data was found matching that color!\n"), false
}
