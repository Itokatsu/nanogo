package pingplugin

import "github.com/bwmarrin/discordgo"
import "github.com/itokatsu/nanogo/parser"
import "strings"

type pingPlugin struct {
	name string
}

func (p *pingPlugin) Name() string {
	return "ping"
}

func New() *pingPlugin {
	var pInstance pingPlugin
	return &pInstance
}

func (p *pingPlugin) HandleMsg(cmd *parser.ParsedCmd, s *discordgo.Session, m *discordgo.MessageCreate) {
	switch strings.ToLower(cmd.Name) {
	case "ping":
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	case "pong":
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}
}

func (p *pingPlugin) Help() string {
	return "on s'en tape"
}

func (p *pingPlugin) Cleanup() {
}
