package timeplugin

import (
    "time" 
    "strings"

	"github.com/bwmarrin/discordgo"
	"github.com/itokatsu/nanogo/botutils"
)

type timePlugin struct {
	name string
}

func New() (*timePlugin, error) {
	var pInstance timePlugin
	return &pInstance, nil
}

func (p *timePlugin) Name() string {
	return "time"
}

func (p *timePlugin) HasData() bool {
	return false
}

func (p *timePlugin) HandleMsg(cmd *botutils.Cmd, s *discordgo.Session) {
	switch strings.ToLower(cmd.Name) {
	case "t", "time":
        now := time.Now()
        if len(cmd.Args) < 1 {
            s.ChannelMessageSend(cmd.ChannelID, now.String())
        }
        return
	
    case "remindme", "rm":
        if len(cmd.Args) > 1 {
            return
        }
        t, err := time.ParseDuration(cmd.Args[0])
        if err != nil {
            errMsg := "Could not parse message, valid units are 'h', 'm', 's'"
            s.ChannelMessageSend(cmd.ChannelID, errMsg)
        }
        time.AfterFunc(t, func() {
            userm := cmd.Message.Author.Mention()
            s.ChannelMessageSend(cmd.ChannelID, userm + " !!" )
        })
    }
}

func (p *timePlugin) Help() string {
	return "Time plugin. timers and timezone stuff"
}
