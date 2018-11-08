package botutils

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strconv"
	"strings"
	"time"
)

/*
@TODO swag editing embed pagination
*/

var duration = 30 * time.Minute
var activeMenus = make(map[string]*Menu)

type Menu struct {
	elems     []fmt.Stringer
	separator string
	channelID string
	callback  func(fmt.Stringer)
	handler   func()
	timer     *time.Timer
}

func (menu Menu) String() string {
	msg := ""
	for i, e := range menu.elems {
		msg += fmt.Sprintf("%d. %v%v", i+1, e, menu.separator)
	}
	msg = strings.TrimSuffix(msg, menu.separator)
	return msg
}

func NewMenu(s *discordgo.Session, elements []fmt.Stringer, separator string,
	chID string, f func(fmt.Stringer)) *Menu {
	if activem, ok := activeMenus[chID]; ok {
		//manually desactivate old menu
		activem.handler()
		activem.timer.Stop()
	}
	menu := Menu{
		elems:     elements,
		separator: separator,
		channelID: chID,
		callback:  f,
	}
	menu.activate(s)

	//print menu
	s.ChannelMessageSend(menu.channelID, menu.String())
	return &menu
}

func (menu *Menu) activate(s *discordgo.Session) {
	menu.handler = s.AddHandler(menu.onMessage)
	menu.timer = time.AfterFunc(duration, menu.handler)
	activeMenus[menu.channelID] = menu
}

func (menu *Menu) onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	//ignore other channels
	if m.ChannelID != menu.channelID {
		return
	}
	//ignore bots and self
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}
	//ignore non-numbers
	i, err := strconv.Atoi(m.Content)
	if err != nil {
		return
	}
	//ignore out of bounds
	if i <= 0 || i > len(menu.elems) {
		return
	}
	go menu.callback(menu.elems[i-1])
}
