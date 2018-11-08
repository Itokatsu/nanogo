package botutils

import (
	"github.com/bwmarrin/discordgo"
)

func AuthorIsAdmin(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	perm, _ := s.UserChannelPermissions(m.Author.ID, m.ChannelID)
	return (perm&discordgo.PermissionAdministrator == discordgo.PermissionAdministrator)
}