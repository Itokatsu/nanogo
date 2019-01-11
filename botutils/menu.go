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

//var duration2 = 10 * time.Hour
//var activeMenus2 = make([]*Menu2)

type Item fmt.Stringer // MenuItem
type ItemEmbed interface {
	Embed() *discordgo.MessageEmbed
}

type Menu struct {
	items     []Item      // itemSlice
	separator string      // separator used to print menu
	channelID string      // discord channel ID
	emitter   chan Item   // send trigger back to caller
	handler   func()      // handler used to stop listening
	timer     *time.Timer // timeout timer
}

type Menu2 struct {
	items map[string]ItemEmbed // [emojiID]ItemEmbed
	msg   *discordgo.Message
}

func (menu Menu) String() string {
	msg := ""
	for i, e := range menu.items {
		msg += fmt.Sprintf("%d. %v%v", i+1, e, menu.separator)
	}
	msg = strings.TrimSuffix(msg, menu.separator)
	return msg
}

// Creates a menu
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
		// convert to Menu Item
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

func GetDefaultEmojis(num int) []string {
	emojis := make([]string, num)
	for i := 0; i < num; i++ {
		emojis[i] = fmt.Sprintf("%d\u20E3", i+1)
	}
	return emojis
}

// Creates a embed menu
/*
func NewMenu2(s *discordgo.Session, items interface{}, emojis []string, chID string) error {
	// >> PAS DES ITEMEMBED MAIS DES CHANNELS D'ITEMEMBED pour pas attendre.
	// check if items is a slice
	if reflect.TypeOf(items).Kind() != reflect.Slice {
		return errors.New("elements passed to menu is not a slice")
	}
	slice := reflect.ValueOf(items)
	length := slice.Len()
	if length > 8 {
		length = 8
	}
	if emojis.Len() >= length {
		itemMap := make(map[string]ItemEmbed, length)
	} else {
		itemMap := make(map[string]ItemEmbed, length)
		emojis = GetDefaultEmojis(length)
	}
	for i := 0; i < length; i++ {
		// convert to Menu Item2
		var ok bool
		key = emojis[i]
		itemSlice[key], ok = (slice.Index(i).Interface()).(ItemEmbed)
		if !ok {
			return errors.New("elements passed to menu do not satisfy MenuItem interface")
		}
	}
	// print menu on discord
	sentMsg, err := s.ChannelMessageSendEmbed(chID, items[0].Embed)
	if err != nil {
		return err
	}
	// create Menu
	menu := Menu2{
		items: itemMap,
		msg:   sentMsg,
	}

	if emojis.Len() > 0 {
		for _, emoji := range emojis {
			s.MessageReactionAdd(chID, sentMsg.ID, emoji)
		}
	} else { //default to number emojis
		for i := 0; i < itemSlice.Len(); i++ {
			s.MessageReactionAdd(chID, sentMsg.ID, fmt.Sprintf("%d\u20E3", i+1))
		}
	}
	//menu.activate(s)
	// func (s *Session) MessageReactionsRemoveAll(channelID, messageID string) error
}*/

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
	menu.handler()      // discordgo stop listening
	menu.timer.Stop()   // stop timer
	close(menu.emitter) // close channel
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

/*
func (menu *Menu2) onReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	menu.EditEmbed(s, r.ChannelID, r.Emoji)
}

func (menu *Menu2) onReactionRm(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	menu.EditEmbed(s, r.ChannelID, r.Emoji)
}

func (menu *Menu2) EditEmbed(s *discordgo.Session, ChannelID string, e *discordgo.Emoji) {
	if menu.msg.ChannelID != channelID {
		return
	}
	embed, ok := menu.items[e.ID]
	if !ok {
		return
	}
	s.ChannelMessageEditEmbed(menu.msg.ChannelID, menu.msg.ID, embed)
}
*/
