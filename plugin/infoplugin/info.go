package infoplugin

import (
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/itokatsu/nanogo/botutils"
)

type infoPlugin struct {
	name      string
	startTime time.Time
}

func New(t time.Time) (*infoPlugin, error) {
	var pInstance infoPlugin
	pInstance.startTime = t
	return &pInstance, nil
}

func (p *infoPlugin) Name() string {
	return "info"
}

func (p *infoPlugin) HasData() bool {
	return false
}

func (p *infoPlugin) HandleMsg(cmd *botutils.Cmd, s *discordgo.Session, m *discordgo.MessageCreate) {
	switch strings.ToLower(cmd.Name) {
	case "uptime":
		uptime := time.Since(p.startTime)
		uptime -= uptime % time.Second
		s.ChannelMessageSend(m.ChannelID, uptime.String())
		return

	case "src", "source":
		s.ChannelMessageSend(m.ChannelID, "https://github.com/Itokatsu/nanogo")
		return

	case "ping":
		s.ChannelMessageSend(m.ChannelID, "Pong!")
		return

	case "pong":
		s.ChannelMessageSend(m.ChannelID, "Ping!")
		return

	case "tenhou":
		s.ChannelMessageSend(m.ChannelID, "http://tenhou.net/0/?L7133")
		return
	}
}

func (p *infoPlugin) Help() string {
	return ``
}
