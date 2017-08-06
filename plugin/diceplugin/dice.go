package diceplugin

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/itokatsu/nanogo/parser"
	"github.com/itokatsu/nanogo/plugin"
)

type dicePlugin struct {
	plugin.Plugin
}

func (p *dicePlugin) Name() string {
	return "dice"
}

func (p *dicePlugin) HandleMsg(cmd *parser.ParsedCmd, s *discordgo.Session, m *discordgo.MessageCreate) {

	switch strings.ToLower(cmd.Name) {
	case "roll":
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v (max : %v)", r.Intn(20)+1, 20))
	}

}

func (p *dicePlugin) Help() string {
	return "roll a die"
}

func New() *dicePlugin {
	var pInstance dicePlugin
	return &pInstance
}
