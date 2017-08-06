package infoplugin

import "github.com/bwmarrin/discordgo"
import "github.com/itokatsu/nanogo/parser"
import "github.com/itokatsu/nanogo/plugin"
import "strings"

type infoPlugin struct {
	plugin.Plugin
}

func (p *infoPlugin) Name() string {
	return "info"
}

func (p *infoPlugin) HandleMsg(cmd *parser.ParsedCmd, s *discordgo.Session, m *discordgo.MessageCreate) {

	switch strings.ToLower(cmd.Name) {
	case "info":
		s.ChannelMessageSend(m.ChannelID, "https://github.com/Itokatsu/nanogo")
	}

}

func (p *infoPlugin) Help() string {
	return "on s'en tape"
}

func New() *infoPlugin {
	var pInstance infoPlugin
	return &pInstance
}
