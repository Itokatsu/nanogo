package plugin

import (
	"encoding/json"
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
	HasData() bool
}

type HasCleanup interface {
	Cleanup()
}

// Global variable
var SaveDir string

func CreateSaveDir(saveDir string) {
	// into global var
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

// Load a plugin state
func Load(p Plugin) error {
	if !p.HasData() {
		return nil
	}
	filename := getSavefile(p)
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Couldn't open savefile")
		return err
	}
	if content == nil {
		fmt.Println("No data to load")
		return fmt.Errorf("No data")
	}
	err2 := json.Unmarshal(content, &p)
	if err2 != nil {
		fmt.Println("Error loading data", err)
		return err
	}
	return nil
}

// Save a plugin state
func Save(p Plugin) {
	if !p.HasData() {
		return
	}
	//get Data
	data, err := json.Marshal(p)
	if err != nil {
		fmt.Println("error marshaling")
		return
	}
	// retrieve data
	filename := getSavefile(p)
	if err := ioutil.WriteFile(filename, data, 0777); err != nil {
		fmt.Printf("! Plugin '%v' Failed to save \n", p.Name())
		return
	}
	fmt.Printf("* Plugin '%v' saved \n", p.Name())
}

/*
Plugin Handler
*/
type PluginHandler struct {
	plugins map[string]Plugin
}

func CreateHandler() *PluginHandler {
	var handler PluginHandler
	handler.plugins = make(map[string]Plugin)
	return &handler
}

// Start a plugin and tries to load it
func (ph *PluginHandler) Start(p Plugin, err error) {
	// Handle plugins that failed to initialize
	if err != nil {
		fmt.Printf("! Plugin '%v' failed\n", p.Name())
		return
	}

	Load(p)
	ph.plugins[p.Name()] = p
	fmt.Printf("+ Plugin '%v' started \n", p.Name())
}

// Get Plugins
func (ph *PluginHandler) GetPlugins() map[string]Plugin {
	return ph.plugins
}

// Save all plugins
func (ph *PluginHandler) SaveAll() {
	for _, p := range ph.plugins {
		Save(p)
	}
}

// Cleanup all plugins
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
