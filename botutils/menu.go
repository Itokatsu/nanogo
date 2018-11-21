package botutils

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"reflect"
	"strconv"
	"strings"
	"time"
)

/*
@TODO pagination ?
 fix interfaces
*/

var duration = 30 * time.Minute
var activeMenus = make(map[string]*Menu)

type Item fmt.Stringer

type Menu struct {
	items     []Item      // itemSlice
	separator string      // separator used to print menu
	channelID string      // discord channel ID
	emitter   chan Item   // send trigger back to caller
	handler   func()      // handler used to stop listening
	timer     *time.Timer // timeout timer
}

func (menu Menu) String() string {
	msg := ""
	for i, e := range menu.items {
		msg += fmt.Sprintf("%d. %v%v", i+1, e, menu.separator)
	}
	msg = strings.TrimSuffix(msg, menu.separator)
	return msg
}

// TypeError
func NewMenu(s *discordgo.Session, items interface{}, separator string, chID string) (chan Item, error) {
	if oldMenu, ok := activeMenus[chID]; ok {
		//desactivate old menu
		oldMenu.desactivate(s)
	}

	// check if items is a slice
	if reflect.TypeOf(items).Kind() != reflect.Slice {
		e := errors.New("elements passed to menu is not a slice")
		return nil, e
	}
	slice := reflect.ValueOf(items)
	itemSlice := make([]Item, slice.Len())
	for i := 0; i < slice.Len(); i++ {
		// convert to Item type
		var ok bool
		itemSlice[i], ok = (slice.Index(i).Interface()).(Item)
		if !ok {
			e := errors.New("elements passed to menu do not satisfy MenuItem interface")
			return nil, e
		}
	}
	// create Menu
	menu := Menu{
		items:     itemSlice,
		separator: separator,
		channelID: chID,
		emitter:   make(chan Item),
	}
	menu.activate(s)

	// print menu on discord
	s.ChannelMessageSend(menu.channelID, menu.String())
	return menu.emitter, nil
}

func (menu *Menu) activate(s *discordgo.Session) {
	// discordgo start listening for responses
	menu.handler = s.AddHandler(menu.onMessage)
	// disable it after duration
	menu.timer = time.AfterFunc(duration, func() {
		menu.desactivate(s)
	})
	// insert menu in map
	activeMenus[menu.channelID] = menu
}

func (menu *Menu) desactivate(s *discordgo.Session) {
	// discordgo stop listening
	menu.handler()
	// stop timer
	menu.timer.Stop()
	// close channel
	close(menu.emitter)
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
	if i <= 0 || i > len(menu.items) {
		return
	}
	menu.emitter <- menu.items[i-1]
}
