package botutils

import (
	"github.com/bwmarrin/discordgo"

	"errors"
	"fmt"
)

const ErrorEmote = "<:awateru:362402110979047425>"

func SendErrorMsg(cmd *Cmd, s *discordgo.Session, e error) {
	s.ChannelMessageSend(cmd.ChannelID, e.Error())
}

func ErrorEmbed(e error) *discordgo.MessageEmbed {
	errorEmbed := NewEmbed().
		SetDescription(e.Error()).
		SetColor(0xFF0000)
	return errorEmbed.MessageEmbed
}

func AddErrorReaction(cmd *Cmd, s *discordgo.Session) {
	s.MessageReactionAdd(cmd.ChannelID, cmd.ID, ErrorEmote)
}

// some Error Creating functions
func HttpStatusCodeError(code int, src string) error {
	msg := fmt.Sprintf("%d received from %s", code, src)
	return errors.New(msg)
}

// Permission related
func NoPermissionError() error {
	msg := fmt.Sprintf("You don't have enough permission for this")
	return errors.New(msg)
}

func AuthorIsAdmin(m *discordgo.Message, s *discordgo.Session) bool {
	perm, _ := s.UserChannelPermissions(m.Author.ID, m.ChannelID)
	return (perm&discordgo.PermissionAdministrator == discordgo.PermissionAdministrator)
}

func CheckArgs(cmd *Cmd, num int) (e error) {
	if len(cmd.Args) < num {
		return NotEnoughArgsError(cmd, num)
	}
	return nil
}

func NotEnoughArgsError(cmd *Cmd, needed int) error {
	msg := fmt.Sprintf("%s need at least %d arguments (got %d)", cmd.Name, needed, len(cmd.Args))
	return errors.New(msg)
}

func NotFoundError(cmd *Cmd, what string) error {
	msg := fmt.Sprintf("Could not find %s", what)
	return errors.New(msg)
}
