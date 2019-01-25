package botutils

import (
	"github.com/bwmarrin/discordgo"

	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"
)

type PagerItem fmt.Stringer

type Pager struct {
	items        []PagerItem
	itemsPerPage int
	runesPerPage int
	currentPage  int
	pages        []int
}

const IPP_DEFAULT = 10
const RPP_DEFAULT = 1000

// Reflect and type generic items to PagerItem
func ReflectPagerItems(items interface{}) ([]PagerItem, error) {
	if reflect.TypeOf(items).Kind() != reflect.Slice {
		e := fmt.Errorf("elements passed to pager is not a slice")
		return nil, e
	}
	slice := reflect.ValueOf(items)
	itemSlice := make([]PagerItem, slice.Len())
	for i := 0; i < slice.Len(); i++ {
		// convert to Menu Item
		var ok bool
		itemSlice[i], ok = (slice.Index(i).Interface()).(PagerItem)
		if !ok {
			e := fmt.Errorf("elements passed to pager do not satisfy PagerItem")
			return nil, e
		}
	}
	return itemSlice, nil
}

// Creates a new Pager from item slice
func NewPager(items interface{}, limits ...int) (*Pager, error) {
	var p Pager
	reflectedItems, err := ReflectPagerItems(items)
	if err != nil {
		return nil, err
	}
	p.items = reflectedItems
	p.currentPage = 0
	switch {
	case len(limits) > 1:
		p.runesPerPage = limits[1]
		fallthrough
	case len(limits) > 0:
		p.itemsPerPage = limits[0]
	default:
		p.runesPerPage = RPP_DEFAULT
		p.itemsPerPage = IPP_DEFAULT
	}
	p.pages = []int{0}
	p.Pagination()
	return &p, nil
}

// Append items to an existing Pager
func (p *Pager) AddItems(items interface{}) (*Pager, error) {
	reflectedItems, err := ReflectPagerItems(items)
	if err != nil {
		return nil, err
	}
	p.items = append(p.items, reflectedItems...)
	p.Pagination()
	return p, nil
}

// Create page index starting from last page
func (p *Pager) Pagination() {
	lastPage := p.pages[len(p.pages)-1]
	var runes, items int
	for i, item := range p.items[lastPage:] {
		runes += utf8.RuneCountInString(item.String())
		items++
		if runes > p.runesPerPage || items > p.itemsPerPage {
			// One item per page even if limit was reached
			if items == 1 {
				p.pages = append(p.pages, lastPage+i+1)
			} else {
				p.pages = append(p.pages, lastPage+i)
			}
			runes = 0
			items = 0
		}
	}
}

func (p *Pager) GoToPage(i int) {
	if i >= 0 && i < len(p.pages) {
		p.currentPage = i
	}
}

// Returns items on current page
func (p *Pager) Page() []PagerItem {
	start := p.pages[p.currentPage]
	end := len(p.items)
	if p.currentPage < len(p.pages)-1 {
		end = p.pages[p.currentPage+1]
	}
	return p.items[start:end]
}

// Returns items on current page as a single string
func (p *Pager) PageString(sep string) string {
	var result string
	for _, item := range p.Page() {
		result += item.String() + sep
	}
	return strings.TrimSuffix(result, sep)
}

// Returns current page as an embed
func (p *Pager) PageEmbed() *Embed {
	e := NewEmbed()
	// Description
	e.SetDescription(p.PageString("\n"))
	e.TruncateDescription()
	// Title
	title := fmt.Sprintf("%d result", len(p.items))
	if len(p.items) > 1 {
		title += "s"
	}
	e.SetTitle(title)
	// Footer
	if len(p.pages) > 1 {
		footer := fmt.Sprintf("Page %d/%d", p.currentPage+1, len(p.pages))
		e.SetFooter(footer)
	}
	return e
}

func (p *Pager) UpdateMsg(s *discordgo.Session, msg *discordgo.Message) {
	s.ChannelMessageEditEmbed(msg.ChannelID, msg.ID, p.PageEmbed().MessageEmbed)
}

// Emojis
const REWIND = "\u23EA"
const LEFT_ARROW = "\u25C0"
const RIGHT_ARROW = "\u25B6"

func LinkPager(s *discordgo.Session, msg *discordgo.Message, p *Pager) {
	if len(p.pages) <= 1 {
		return
	}
	//Add Reactions
	if len(p.pages) > 2 {
		AddReactionButton(s, msg, REWIND, func() {
			p.GoToPage(0)
			p.UpdateMsg(s, msg)
		})
	}
	AddReactionButton(s, msg, LEFT_ARROW, func() {
		p.GoToPage(p.currentPage - 1)
		p.UpdateMsg(s, msg)
	})
	AddReactionButton(s, msg, RIGHT_ARROW, func() {
		p.GoToPage(p.currentPage + 1)
		p.UpdateMsg(s, msg)
	})
}
