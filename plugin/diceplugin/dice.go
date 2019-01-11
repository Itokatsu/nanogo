package diceplugin

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/itokatsu/nanogo/botutils"
)

var maxRolls = 10
var maxSize = 100

type dicePlugin struct {
	name string
}

func New() (*dicePlugin, error) {
	var pInstance dicePlugin
	return &pInstance, nil
}

func (p *dicePlugin) Name() string {
	return "dice"
}

func (p *dicePlugin) HasData() bool {
	return false
}

var validArg = `^(\d*?)d?(\d+)$`
var re = regexp.MustCompile(validArg)

func (p *dicePlugin) HandleMsg(cmd *botutils.Cmd, s *discordgo.Session) {
	switch strings.ToLower(cmd.Name) {
	case "roll":
		var nRolls, dieSize int
		if len(cmd.Args) > 0 && re.MatchString(cmd.Args[0]) {
			subm := re.FindStringSubmatch(cmd.Args[0])
			nRolls, _ = strconv.Atoi(subm[1])
			dieSize, _ = strconv.Atoi(subm[2])
			if nRolls == 0 {
				nRolls = 1
			}
			if nRolls > maxRolls {
				nRolls = maxRolls
			}
			if dieSize < 2 || dieSize > maxSize {
				dieSize = maxSize
			}
		} else {
			nRolls = 1
			dieSize = 20
		}
		if nRolls == 1 {
			s.ChannelMessageSend(cmd.ChannelID,
				fmt.Sprintf("%d (max : %d)", botutils.RandInt(dieSize)+1, dieSize))
			return
		}

		var results []string
		sum := 0
		for i := nRolls; i > 0; i-- {
			v := botutils.RandInt(dieSize) + 1
			results = append(results, strconv.Itoa(v))
			sum += v
		}
		msg := fmt.Sprintf("%v ➜ %v (avg : %.2f ; max : %v)",
			strings.Join(results, " | "), sum, float64(sum)/float64(nRolls), dieSize)
		s.ChannelMessageSend(cmd.ChannelID, msg)
		return
	}
}

func (p *dicePlugin) Help() string {
	return "roll a die"
}
