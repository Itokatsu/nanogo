package botutils

import (
	"github.com/bwmarrin/discordgo"
	"strings"
)

type Cmd struct {
	*discordgo.Message
	Name string
	Args []string
}

func ParseCmd(msg *discordgo.Message, prefixes ...string) (c Cmd) {
	text := strings.TrimSpace(msg.Content)
	for _, p := range prefixes {
		matched := strings.HasPrefix(strings.ToLower(text),
			strings.ToLower(p))
		if matched {
			c.Message = msg
			// Build and return Cmd
			text = text[len(p):]
			f := strings.Fields(text)
			length := len(text)
			if length > 0 {
				c.Name = f[0]
			}
			if length > 1 {
				c.Args = f[1:]
			}
			return c
		}
	}
	return
}
