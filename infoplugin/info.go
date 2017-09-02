package infoplugin

import (
	"github.com/bwmarrin/discordgo"
	"github.com/itokatsu/nanogo/parser"
	"os"
	"strings"
	"time"
)

type infoPlugin struct {
	name      string
	startTime time.Time
}

func New(t time.Time) *infoPlugin {
	var pInstance infoPlugin
	pInstance.startTime = t
	return &pInstance
}

func (p *infoPlugin) Name() string {
	return "info"
}

func (p *infoPlugin) HandleMsg(cmd *parser.ParsedCmd, s *discordgo.Session, m *discordgo.MessageCreate) {
	switch strings.ToLower(cmd.Name) {
	case "uptime":
		uptime := time.Since(p.startTime)
		uptime -= uptime % time.Second
		s.ChannelMessageSend(m.ChannelID, uptime.String())
		return
	case "source":
		s.ChannelMessageSend(m.ChannelID, "https://github.com/Itokatsu/nanogo")
		return
	case "tenhou":
		s.ChannelMessageSend(m.ChannelID, "http://tenhou.net/0/?L7133")
		return
	case "essences":
		imgReader, err := os.Open("./media/img/essences.png")
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, ":nano: Cannot open file")
			return
		}
		s.ChannelFileSend(m.ChannelID, "essences.png", imgReader)
		return
	case "gemtd":
		imgReader, err := os.Open("./media/img/gemtd.jpg")
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, ":nano: Cannot open file")
			return
		}
		s.ChannelFileSend(m.ChannelID, "gemtd.jpg", imgReader)
		return
	}
}

func (p *infoPlugin) Help() string {
	return "on s'en tape"
}

func (p *infoPlugin) Cleanup() {
}
