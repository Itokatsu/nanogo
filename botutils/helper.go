package botutils

import (
	"github.com/bwmarrin/discordgo"
	"unicode/utf8"
	"strings"
)

func AuthorIsAdmin(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	perm, _ := s.UserChannelPermissions(m.Author.ID, m.ChannelID)
	return (perm&discordgo.PermissionAdministrator == discordgo.PermissionAdministrator)
}

var maxRunes = 2000
func Send(s *discordgo.Session, chanID string, msg string) {
	nRunes := utf8.RuneCountInString(msg)
	if nRunes < maxRunes {
		s.ChannelMessageSend(chanID, msg)
	} else {
		msgRune := []rune(msg)
		msgTrunc := string(msgRune[:2000])
		lineBreak := strings.LastIndexByte(msgTrunc, '\n')
		if lineBreak == -1 {
			lineBreak = len(msgTrunc)
			return
		}
		Send(s, chanID, msg[:lineBreak])
		Send(s, chanID, msg[lineBreak:])
	}
}