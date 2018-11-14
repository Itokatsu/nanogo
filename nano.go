package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/itokatsu/nanogo/botutils"
)

// plugins
import (
	"github.com/itokatsu/nanogo/plugin"

	//	"github.com/itokatsu/nanogo/testplugin"
	"github.com/itokatsu/nanogo/plugin/catplugin"
	"github.com/itokatsu/nanogo/plugin/diceplugin"
	"github.com/itokatsu/nanogo/plugin/googleplugin"
	"github.com/itokatsu/nanogo/plugin/infoplugin"
	"github.com/itokatsu/nanogo/plugin/jpplugin"
	"github.com/itokatsu/nanogo/plugin/tagplugin"
	"github.com/itokatsu/nanogo/plugin/wolframplugin"
	"github.com/itokatsu/nanogo/plugin/youtubeplugin"
)

// various Auth Tokens and API Keys
type ConfigKeys struct {
	Bot     BotConfig     `json:"bot"`
	Plugins PluginsConfig `json:"plugins"`
}
type BotConfig struct {
	Token       string   `json:"token"`
	Prefixes    []string `json:"prefixes,omitempty"`
	SaveFolder  string   `json:"saveFolder,omitempty"`
	TestingOnly bool     `json:"testing,omitempty"`
	TestChannel string   `json:"testChannel,omitempty"`
}
type PluginsConfig struct {
	Google  googleplugin.Config  `json:"google,omitempty"`
	Youtube youtubeplugin.Config `json:"youtube,omitempty"`
	Wolfram wolframplugin.Config `json:"wolfram,omitempty"`
	Jp      jpplugin.Config      `json:"japanese,omitempty"`
}

// Global variables
var DefaultConfig = BotConfig{
	// No bot token => useless
	Prefixes:    []string{"!"},
	SaveFolder:  ".saves",
	TestingOnly: false,
}

var (
	ph        *plugin.PluginHandler
	Cfg       ConfigKeys
	StartTime time.Time
	CmdPrefix string
)

func loadConfig() {
	file, err := ioutil.ReadFile("./confignew.json")
	if err != nil {
		fmt.Println("Error: Config file not found")
		os.Exit(1)
		/*
			fmt.Println("Loading default Config, some plugins will NOT start")
			Cfg.Bot = DefaultConfig
			return
		*/
	}
	err = json.Unmarshal(file, &Cfg)
	if err != nil {
		fmt.Println("Error: Couldn't unmarshal config file")
		os.Exit(1)
		/*
			fmt.Println("Loading default Config, some plugins will NOT start")
			Cfg.Bot = DefaultConfig
		*/
	}
}

func init() {
	// Load config file
	loadConfig()
	// Create plugin savestates folder
	Cfg.Bot.SaveFolder = path.Join(".", Cfg.Bot.SaveFolder)
	plugin.CreateSaveDir(Cfg.Bot.SaveFolder)
	// Generate pseudo random seed
	StartTime = time.Now()
	rand.Seed(time.Now().UnixNano())
}

func main() {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Cfg.Bot.Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Load Plugins
	ph = plugin.CreateHandler()
	go ph.Start(infoplugin.New(StartTime))
	go ph.Start(diceplugin.New())
	go ph.Start(tagplugin.New())
	go ph.Start(catplugin.New())

	go ph.Start(wolframplugin.New(Cfg.Plugins.Wolfram))
	go ph.Start(googleplugin.New(Cfg.Plugins.Google))
	go ph.Start(youtubeplugin.New(Cfg.Plugins.Youtube))
	go ph.Start(jpplugin.New(Cfg.Plugins.Jp))
	defer ph.SaveAll()
	defer ph.CleanupAll()

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	defer dg.Close()

	// Block until KILL
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Test channel only
	if Cfg.Bot.TestingOnly && m.ChannelID != Cfg.Bot.TestChannel {
		return
	}
	// Ignore messages from bots (and self)
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}
	cmd := botutils.ParseCmd(m.Content, Cfg.Bot.Prefixes...)
	if cmd.Name != "" {
		for _, p := range ph.GetPlugins() {
			go p.HandleMsg(&cmd, s, m)
		}
	}
}
