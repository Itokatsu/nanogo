package web

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"
	"github.com/itokatsu/nanogo/botutils"
)

const offset = "\u3000\u3000"

type ResultEntry struct {
	Header  string
	Content []string
}

func (e ResultEntry) Print() string {
	contentText := strings.Join(e.Content, "\n")
	return fmt.Sprintf(":small_orange_diamond: %s\n%s", e.Header, contentText)
}

type ResultEntryPager struct {
	Entries     []ResultEntry
	CurrentPage int
	Pages       []int //stores index of the first entry of the page
}

func NewResultPager(e []ResultEntry) *ResultEntryPager {
	var p ResultEntryPager
	p.Entries = e
	p.CurrentPage = 0
	p.Pagination(600, 10)
	return &p
}

func (p *ResultEntryPager) Pagination(charLimit int, itemLimit int) {
	p.Pages = []int{0}
	var runes int
	for i, entry := range p.Entries {
		runes += utf8.RuneCountInString(entry.Print())
		if runes > charLimit || i > p.Pages[len(p.Pages)-1]+itemLimit {
			// At least one item per page even if >limit
			if i == p.Pages[len(p.Pages)-1] {
				p.Pages = append(p.Pages, i+1)
			} else {
				p.Pages = append(p.Pages, i)
			}
			runes = 0
		}
	}
}

func (p *ResultEntryPager) Page(i int) {
	if i >= 0 && i < len(p.Pages) {
		p.CurrentPage = i
	}
}

func (p *ResultEntryPager) PopulateEmbed(e *botutils.Embed) *botutils.Embed {
	offset := p.Pages[p.CurrentPage]
	til := len(p.Entries)
	if p.CurrentPage < len(p.Pages)-1 {
		til = p.Pages[p.CurrentPage+1]
	}
	var text string
	for _, e := range p.Entries[offset:til] {
		text += e.Print() + "\n"
	}
	footer := fmt.Sprintf("Page %d/%d", p.CurrentPage+1, len(p.Pages))
	return e.SetDescription(text).TruncateDescription().SetFooter(footer)
}

type PagerHandler struct {
	pagers    map[string]*ResultEntryPager
	listening bool
}

var PagerH = PagerHandler{
	pagers:    make(map[string]*ResultEntryPager),
	listening: false,
}

const REWIND = "\u23EA"
const LEFT_ARROW = "\u25C0"
const RIGHT_ARROW = "\u25B6"

func (h *PagerHandler) Add(s *discordgo.Session, msg *discordgo.Message, p *ResultEntryPager) {
	if len(p.Pages) <= 1 {
		return
	}
	h.pagers[msg.ID] = p
	//Add Reactions
	s.MessageReactionAdd(msg.ChannelID, msg.ID, REWIND)
	s.MessageReactionAdd(msg.ChannelID, msg.ID, LEFT_ARROW)
	s.MessageReactionAdd(msg.ChannelID, msg.ID, RIGHT_ARROW)
	h.Listen(s)
}

func (h *PagerHandler) Listen(s *discordgo.Session) {
	if !h.listening {
		s.AddHandler(h.OnReactionAdd)
		s.AddHandler(h.OnReactionRm)
		h.listening = true
	}
}

func (h *PagerHandler) OnReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	h.onReaction(s, r.MessageReaction)
}

func (h *PagerHandler) OnReactionRm(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	h.onReaction(s, r.MessageReaction)
}

func (h *PagerHandler) onReaction(s *discordgo.Session, r *discordgo.MessageReaction) {
	if r.UserID == s.State.User.ID {
		return
	}
	var pager *ResultEntryPager
	pager, exist := h.pagers[r.MessageID]
	if !exist {
		return
	}
	old := pager.CurrentPage
	// Match the Emoji
	switch r.Emoji.Name {
	case REWIND:
		pager.Page(0)
	case LEFT_ARROW:
		pager.Page(pager.CurrentPage - 1)
	case RIGHT_ARROW:
		pager.Page(pager.CurrentPage + 1)
	default:
		return
	}
	if pager.CurrentPage == old {
		return
	}
	// Edit msg
	edit := discordgo.NewMessageEdit(r.ChannelID, r.MessageID)
	edit.SetEmbed(pager.PopulateEmbed(botutils.NewEmbed()).MessageEmbed)
	s.ChannelMessageEditComplex(edit)
}
