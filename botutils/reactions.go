package botutils

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

const maxMessages = 2

type EmojiButton struct {
	emojiName string
	f         func()
	once      bool
}

var (
	buttons   = make(map[string][]*EmojiButton, maxMessages)
	messages  = make([]*discordgo.Message, 0) //needed to get correct order
	listening = false
)

func AddButton(s *discordgo.Session, m *discordgo.Message, b *EmojiButton) {
	s.MessageReactionAdd(m.ChannelID, m.ID, b.emojiName)
	_, exist := buttons[m.ID]
	if exist {
		buttons[m.ID] = append(buttons[m.ID], b)
	} else {
		buttons[m.ID] = append(make([]*EmojiButton, 0), b)
		messages = append(messages, m)
		// remove oldest if needed
		if len(messages) > maxMessages {
			s.MessageReactionsRemoveAll(messages[0].ChannelID, messages[0].ID)
			messages = messages[1:]
		}
	}

	//Bind to session
	if !listening {
		s.AddHandler(onReactionAdd)
		s.AddHandler(onReactionRm)
		listening = true
	}
}

func RemoveButtons(s *discordgo.Session, msgID string) {
	msgIdx := -1
	for i := range messages {
		if messages[i].ID == msgID {
			msgIdx = i
			break
		}
	}
	// not found ???
	if msgIdx == -1 {
		return
	}
	msg := messages[msgIdx]
	s.MessageReactionsRemoveAll(msg.ChannelID, msg.ID)
	delete(buttons, msg.ID)
	messages = append(messages[:msgIdx], messages[msgIdx+1:]...)
}

func AddReactionButton(s *discordgo.Session, m *discordgo.Message, e string, f func()) {
	btn := &EmojiButton{
		emojiName: e,
		f:         f,
		once:      false,
	}
	AddButton(s, m, btn)
}

func AddReactionButtonOnce(s *discordgo.Session, m *discordgo.Message, e string, f func()) {
	btn := &EmojiButton{
		emojiName: e,
		f:         f,
		once:      true,
	}
	AddButton(s, m, btn)
}

// discordgo event handling
func onReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	onReaction(s, r.MessageReaction)
}

func onReactionRm(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	onReaction(s, r.MessageReaction)
}

func onReaction(s *discordgo.Session, r *discordgo.MessageReaction) {
	if r.UserID == s.State.User.ID {
		return
	}
	// Try to match the Emoji
	btns, exist := buttons[r.MessageID]
	if !exist {
		fmt.Println("not found")
		return
	}
	for _, btn := range btns {
		if btn.emojiName == r.Emoji.Name {
			btn.f()
			if btn.once {
				RemoveButtons(s, r.MessageID)
				return
			}
		}
	}
}
