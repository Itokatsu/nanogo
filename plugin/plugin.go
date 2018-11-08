package plugin

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

import (
	"github.com/bwmarrin/discordgo"
	"github.com/itokatsu/nanogo/botutils"
)

type Plugin interface {
	HandleMsg(cmd *botutils.Cmd, s *discordgo.Session, m *discordgo.MessageCreate)
	Help() string
	Name() string
}

type HasSave interface {
	Load([]byte) error
	Save() []byte
}

type HasCleanup interface {
	Cleanup()
}

// Global variable
var SaveDir string

// plugin Handler
type PluginHandler struct {
	plugins 	map[string]Plugin
}

func CreateHandler() *PluginHandler {
	var handler PluginHandler
	handler.plugins = make(map[string]Plugin)
	return &handler
}

func CreateSaveDir(saveDir string) {
	SaveDir = saveDir
	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		err := os.Mkdir(saveDir, 0777)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func getSavefile(p Plugin) string {
	return path.Join(SaveDir, p.Name())
}

func (ph *PluginHandler) GetPlugins() map[string]Plugin {
	return ph.plugins
}

func (ph *PluginHandler) Load(p Plugin, err error) Plugin {
	// Handle plugins that failed to initialize
	if err != nil {
		fmt.Printf("! Plugin '%v' failed \n", p.Name())
		return nil
	}

	// Start plugins with no savefile
	psl, ok := p.(HasSave)
	if !ok {
		return ph.Start(p)
	}
	// Read savefile
	filename := getSavefile(p)
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return ph.Start(p)
	}
	// load savefile
	if err := psl.Load(content); err != nil {
		return ph.Start(p)
	}
	fmt.Printf("+ Plugin '%v' loaded \n", p.Name())
	return ph.Start(p)
}

func (ph *PluginHandler) Start(p Plugin) Plugin {
	ph.plugins[p.Name()] = p
	fmt.Printf("+ Plugin '%v' started \n", p.Name())
	return p
}

func Save(p Plugin) {
	// test if plugin implements HasSave interface
	psl, ok := p.(HasSave)
	if !ok {
		return
	}
	// retrieve data
	data := psl.Save()
	if data == nil {
		return
	}
	filename := getSavefile(p)
	if err := ioutil.WriteFile(filename, data, 0777); err != nil {
		fmt.Printf("! Plugin '%v' Failed to save \n", p.Name())
		return
	}
	fmt.Printf("* Plugin '%v' saved \n", p.Name())
}

func (ph *PluginHandler) SaveAll() {
	for _, p := range ph.plugins {
		Save(p)
	}
}

func (ph *PluginHandler) CleanupAll() {
	// test if plugin implements HasCleanup interface
	for _, p := range ph.plugins {
		pclnup, ok := p.(HasCleanup)
		if !ok {
			continue
		}
		pclnup.Cleanup()
	}
}
