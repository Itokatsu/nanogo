package main

import "fmt"
import "github.com/bwmarrin/discordgo"
import "github.com/itokatsu/nanogo/parser"

type Plugin interface {
	Name() string
	HandleMsg(cmd *parser.ParsedCmd,
		s *discordgo.Session,
		m *discordgo.MessageCreate)
	Help() string
}

type pluginHandler struct {
	plugins map[string]Plugin
}

func (ph *pluginHandler) Load(p Plugin) *Plugin {
	ph.plugins[p.Name()] = p
	fmt.Printf("+ Plugin '%v' loadedi \n", p.Name())
	return &p
}
