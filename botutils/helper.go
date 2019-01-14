package botutils

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"unicode/utf8"
)

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
	}
}
