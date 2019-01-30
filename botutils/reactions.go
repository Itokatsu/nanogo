package botutils

import (
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
	err := s.MessageReactionAdd(m.ChannelID, m.ID, b.emojiName)
	if err != nil {
		return
	}
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

func FindMsgFromID(msgID string) (*discordgo.Message, int) {
	for i := range messages {
		if messages[i].ID == msgID {
			return messages[i], i
		}
	}
	return nil, -1
}

func RemoveButton(s *discordgo.Session, msgID string, emoji string) {
	msg, msgIdx := FindMsgFromID(msgID)
	if msg == nil {
		return
	}

	buttonIdx := -1
	for i := range buttons[msgID] {
		if buttons[msgID][i].emojiName == emoji {
			buttonIdx = i
			break
		}
	}
	if buttonIdx == -1 {
		return
	}

	//remove reactions
	users, err := s.MessageReactions(msg.ChannelID, msg.ID, emoji, 100)
	if err != nil || len(users) == 0 {
		return
	}
	for _, user := range users {
		s.MessageReactionRemove(msg.ChannelID, msg.ID, emoji, user.ID)
	}

	//remove button
	buttons[msg.ID] = append(buttons[msg.ID][:buttonIdx], buttons[msg.ID][buttonIdx+1:]...)
	//remove msg if no more button
	if len(buttons[msg.ID]) == 0 {
		delete(buttons, msg.ID)
		messages = append(messages[:msgIdx], messages[msgIdx+1:]...)
	}
}

func RemoveButtonAll(s *discordgo.Session, msgID string) {
	msg, msgIdx := FindMsgFromID(msgID)
	if msg == nil {
		return
	}
	s.MessageReactionsRemoveAll(msg.ChannelID, msg.ID)
	delete(buttons, msg.ID)
	messages = append(messages[:msgIdx], messages[msgIdx+1:]...)
}

func CleanReactions(s *discordgo.Session) {
	for _, msg := range messages {
		RemoveButtonAll(s, msg.ID)
	}
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
		return
	}
	for _, btn := range btns {
		if btn.emojiName == r.Emoji.Name {
			btn.f()
			if btn.once {
				RemoveButton(s, r.MessageID, btn.emojiName)
				return
			}
		}
	}
}
