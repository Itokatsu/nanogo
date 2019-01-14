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

func (p *infoPlugin) HandleMsg(cmd *botutils.Cmd, s *discordgo.Session) {
	switch strings.ToLower(cmd.Name) {
	case "uptime":
		uptime := time.Since(p.startTime)
		uptime -= uptime % time.Second
		s.ChannelMessageSend(cmd.ChannelID, uptime.String())
		return

	case "src", "source":
		s.ChannelMessageSend(cmd.ChannelID, "https://github.com/Itokatsu/nanogo")
		return

	case "ping":
		s.ChannelMessageSend(cmd.ChannelID, "Pong!")
		return

	case "pong":
		s.ChannelMessageSend(cmd.ChannelID, "Ping!")
		return

	case "tenhou":
		s.ChannelMessageSend(cmd.ChannelID, "http://tenhou.net/0/?L7133")
		return

	case "chapu":
		s.ChannelMessageSend(cmd.ChannelID, "https://www.youtube.com/watch?v=FTfBe9NPzyU")
		return
	}
}

func (p *infoPlugin) Help() string {
	return ``
}
