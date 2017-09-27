package tagplugin

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/itokatsu/nanogo/botutils"
)

type tagPlugin struct {
	name string
	Tags map[string]map[string]Tag
}

type Tag struct {
	AuthorID string
	Message  string
}

func New() *tagPlugin {
	var pInstance tagPlugin
	pInstance.Tags = make(map[string]map[string]Tag)
	pInstance.name = "tag"
	return &pInstance
}

func (p *tagPlugin) Name() string {
	return p.name
}

func (p *tagPlugin) HasSaves() bool {
	return true
}

func (p *tagPlugin) HandleMsg(cmd *botutils.Cmd, s *discordgo.Session, m *discordgo.MessageCreate) {

	if key := botutils.ParseCmd(cmd.Name, "!"); key.Name != "" {
		channel, err := s.State.Channel(m.ChannelID)
		if err != nil {
			return
		}
		guildID := channel.GuildID
		if tag, exist := p.Tags[guildID][key.Name]; exist {
			s.ChannelMessageSend(m.ChannelID, tag.Message)
		}
		return
	}

	switch strings.ToLower(cmd.Name) {
	case "label", "tag":
		if len(cmd.Args) < 1 {
			return
		}
		channel, err := s.State.Channel(m.ChannelID)
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
			case "add", "del", "list":
				msg := fmt.Sprintf("Error : %v is a reserved keyword for the Tag plugin.", key)
				s.ChannelMessageSend(m.ChannelID, msg)
				return
			}
			// Tag already exist, only admins or author can modify it
			if tag, exist := p.Tags[guildID][key]; exist {
				if m.Author.ID != tag.AuthorID && !botutils.AuthorIsAdmin(s, m) {
					msg := fmt.Sprintf(":nano: The tag `%s` belongs to %s.", key, m.Author.Username)
					s.ChannelMessageSend(m.ChannelID, msg)
					return
				}
			}
			p.Tags[guildID][key] = Tag{
				AuthorID: m.Author.ID,
				Message:  strings.Join(cmd.Args[2:], " "),
			}
			s.ChannelMessageSend(m.ChannelID, p.Tags[guildID][key].Message)
			return

		case "del":
			if len(cmd.Args) < 2 {
				return
			}
			key := strings.ToLower(cmd.Args[1])
			tag, exist := p.Tags[guildID][key]
			if !exist {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Tag `%v` not found", key))
				return
			}
			if m.Author.ID != tag.AuthorID && !botutils.AuthorIsAdmin(s, m) {
				msg := fmt.Sprintf(":nano: Tag `%s` belongs to %s.", key, m.Author.Username)
				s.ChannelMessageSend(m.ChannelID, msg)
				return
			}
			delete(p.Tags[guildID], key)
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Command `%v` deleted", key))
			return

		case "list":

		default:
			if len(cmd.Args) < 1 {
				return
			}
			key := strings.ToLower(cmd.Args[0])
			tag, exist := p.Tags[guildID][key]
			if !exist {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Tag `%v` not found", key))
				return
			}
			s.ChannelMessageSend(m.ChannelID, tag.Message)
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

func (p *tagPlugin) Save() []byte {
	buf, err := json.Marshal(p)
	if err != nil {
		fmt.Println("error marshaling")
		return nil
	}
	return buf
}

func (p *tagPlugin) Load(data []byte) error {
	if data == nil {
		fmt.Println("No data to load")
		return fmt.Errorf("No data")
	}
	err := json.Unmarshal(data, &p)
	if err != nil {
		fmt.Println("Error loading data", err)
		return err
	}
	return nil
}

func (p *tagPlugin) Cleanup() {

}
