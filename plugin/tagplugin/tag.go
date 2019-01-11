// Nanobot Project
//
// custom commands plugin

package tagplugin

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/itokatsu/nanogo/botutils"
	"github.com/itokatsu/nanogo/plugin"
)

type tagPlugin struct {
	name string
	Tags map[string]map[string]Tag `json:"tags,omitempty"`
}

type Tag struct {
	AuthorID string
	Message  string
	LastEdit time.Time
}

func New() (*tagPlugin, error) {
	var pInstance tagPlugin
	pInstance.Tags = make(map[string]map[string]Tag)
	pInstance.name = "tag"
	return &pInstance, nil
}

func (p *tagPlugin) Name() string {
	return p.name
}

func (p *tagPlugin) HasData() bool {
	return true
}

const shortcut = "!"

func (p *tagPlugin) HandleMsg(cmd *botutils.Cmd, s *discordgo.Session) {

	// shortcut with prefix
	if strings.HasPrefix(cmd.Name, shortcut) {
		key := cmd.Name[len(shortcut):]
		channel, err := s.State.Channel(cmd.ChannelID)
		if err != nil {
			return
		}
		guildID := channel.GuildID
		if tag, exist := p.Tags[guildID][key]; exist {
			s.ChannelMessageSend(cmd.ChannelID, tag.Message)
		}
		return
	}

	switch strings.ToLower(cmd.Name) {
	case "label", "tag":
		if len(cmd.Args) < 1 {
			return
		}
		channel, err := s.State.Channel(cmd.ChannelID)
		if err != nil {
			return
		}
		guildID := channel.GuildID
		if p.Tags[guildID] == nil {
			p.Tags[guildID] = make(map[string]Tag)
		}

		switch strings.ToLower(cmd.Args[0]) {
		case "add":
			if len(cmd.Args) < 3 {
				return
			}
			key := strings.ToLower(cmd.Args[1])
			switch key {
			// Illegal keywords
			case "add", "del", "list":
				msg := fmt.Sprintf("Error : %v is a reserved keyword for the Tag plugin.", key)
				s.ChannelMessageSend(cmd.ChannelID, msg)
				return
			}
			// Tag already exist, only admins or author can modify it
			if tag, exist := p.Tags[guildID][key]; exist {
				if cmd.Author.ID != tag.AuthorID && !botutils.AuthorIsAdmin(s, cmd.Message) {
					msg := fmt.Sprintf("The tag `%s` belongs to %s.", key, cmd.Author.Username)
					s.ChannelMessageSend(cmd.ChannelID, msg)
					return
				}
			}
			// Create Tag
			p.Tags[guildID][key] = Tag{
				AuthorID: cmd.Author.ID,
				Message:  strings.Join(cmd.Args[2:], " "),
				LastEdit: time.Now()}
			s.ChannelMessageSend(cmd.ChannelID, p.Tags[guildID][key].Message)
			plugin.Save(p)
			return

		case "del":
			if len(cmd.Args) < 2 {
				return
			}
			key := strings.ToLower(cmd.Args[1])
			tag, exist := p.Tags[guildID][key]
			if !exist {
				s.ChannelMessageSend(cmd.ChannelID, fmt.Sprintf("Tag `%v` not found", key))
				return
			}
			if cmd.Author.ID != tag.AuthorID && !botutils.AuthorIsAdmin(s, cmd.Message) {
				msg := fmt.Sprintf("Tag `%s` belongs to %s.", key, cmd.Author.Username)
				s.ChannelMessageSend(cmd.ChannelID, msg)
				return
			}
			delete(p.Tags[guildID], key)
			s.ChannelMessageSend(cmd.ChannelID, fmt.Sprintf("Command `%v` deleted", key))
			plugin.Save(p)
			return

		case "list":

		default:
			if len(cmd.Args) < 1 {
				return
			}
			key := strings.ToLower(cmd.Args[0])
			tag, exist := p.Tags[guildID][key]
			if !exist {
				s.ChannelMessageSend(cmd.ChannelID, fmt.Sprintf("Tag `%v` not found", key))
				return
			}
			s.ChannelMessageSend(cmd.ChannelID, tag.Message)
			return
		}
	}
}

func (p *tagPlugin) Help() string {
	return `
	!label add <key> <message> : create/edit a custom command
	!label del <key> : delete a custom command
	!label <key> : print mesasge corresponding to <key>
	!!<key> : print message corresponding to <key>
	`
}
