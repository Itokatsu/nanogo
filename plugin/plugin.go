package plugin

import "fmt"
import "github.com/bwmarrin/discordgo"
import "github.com/itokatsu/nanogo/parser"

type Plugin struct {
	BasePlugin
	isActive bool
}

type BasePlugin interface {
	New(args ...string) *Plugin
	Name() string
	Help() string
	HandleMsg(cmd *parser.ParsedCmd,
		s *discordgo.Session,
		m *discordgo.MessageCreate)
}

func (p *Plugin) Activate() {
	p.isActive = true
}
func (p *Plugin) Desactivate() {
	p.isActive = false
}

// plugin handler
type Handler struct {
	Plugins map[string]Plugin
}

func (ph *Handler) GetPlugin(pname string) (*Plugin, error) {
	if p, ok := ph.Plugins[pname]; ok {
		return nil, fmt.Errorf("plugin %q not found", pname)
	} else {
		return &p, nil
	}
}

func (ph *Handler) Load(p Plugin) *Plugin {
	ph.Plugins[p.Name()] = p
	p.Activate()
	fmt.Printf("+ Plugin '%v' loaded \n", p.Name())
	return &p
}

func (ph *Handler) Unload(p Plugin) {
	p.Desactivate()
}
