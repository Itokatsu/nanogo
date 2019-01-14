package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	"github.com/itokatsu/nanogo/plugin/alttprplugin"
	"github.com/itokatsu/nanogo/plugin/catplugin"
	"github.com/itokatsu/nanogo/plugin/diceplugin"
	"github.com/itokatsu/nanogo/plugin/googleplugin"
	"github.com/itokatsu/nanogo/plugin/infoplugin"
	"github.com/itokatsu/nanogo/plugin/jpplugin"
	"github.com/itokatsu/nanogo/plugin/tagplugin"
	"github.com/itokatsu/nanogo/plugin/wolframplugin"
	"github.com/itokatsu/nanogo/plugin/youtubeplugin"
)

// Config file
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
	Alttpr  alttprplugin.Config  `json:"alttpr,omitempty"`
}

// Global variables
var DefaultConfig = BotConfig{
	Prefixes:    []string{"!"},
	SaveFolder:  ".saves",
	TestingOnly: false,
}

var (
	ph        *plugin.PluginHandler
	Cfg       ConfigKeys
	StartTime time.Time
)

func loadConfig() {
	file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		//fmt.Println("Error: Config file not found")
		panic("Fatal: Config file not found")
	}
	err = json.Unmarshal(file, &Cfg)
	if err != nil {
		//fmt.Println("Error: Couldn't unmarshal config file")
		panic("Fatal: Couldn't unmarshal config file")
	}
}

func init() {
	StartTime = time.Now()
	// Load config file info Cfg
	loadConfig()
	// Create plugin savestates folder
	Cfg.Bot.SaveFolder = path.Join(".", Cfg.Bot.SaveFolder)
	plugin.CreateSaveDir(Cfg.Bot.SaveFolder)
	// Start plugins
	ph = plugin.CreateHandler()
}

func StartPlugins(ph *plugin.PluginHandler) {
	go ph.Start(infoplugin.New(StartTime))
	go ph.Start(diceplugin.New())
	go ph.Start(tagplugin.New())
	go ph.Start(catplugin.New())
	// using config file
	go ph.Start(alttprplugin.New(Cfg.Plugins.Alttpr))
	go ph.Start(wolframplugin.New(Cfg.Plugins.Wolfram))
	go ph.Start(googleplugin.New(Cfg.Plugins.Google))
	go ph.Start(youtubeplugin.New(Cfg.Plugins.Youtube))
	go ph.Start(jpplugin.New(Cfg.Plugins.Jp))
}

func main() {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Cfg.Bot.Token)
	if err != nil {
		panic("Fatal: Couldn't create Discord session")
		//fmt.Println("error creating Discord session,", err)
	}

	StartPlugins(ph)
	defer ph.SaveAll()
	defer ph.CleanupAll()
	dg.AddHandler(messageCreate) // Listen to new messages
	dg.AddHandler(messageUpdate) // Listen to message edits

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

// Handling messages
func handle(s *discordgo.Session, m *discordgo.Message) {
	// Ignore messages from bots (and self)
	if m.Author == nil || m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}
	// Test channel only
	if Cfg.Bot.TestingOnly && m.ChannelID != Cfg.Bot.TestChannel {
		return
	}

	//Message is worth parsing
	cmd := botutils.ParseCmd(m, Cfg.Bot.Prefixes...)
	if cmd.Name != "" {
		go ph.HandleMsg(&cmd, s)
	}

	// easy tests here
	switch cmd.Name {
	case "emo":
		sentMsg, _ := s.ChannelMessageSend(m.ChannelID, "reactions")
		for i := 0; i <= 10; i++ {
			s.MessageReactionAdd(m.ChannelID, sentMsg.ID, fmt.Sprintf("%d\u20E3", i))
		}

	// reboot the bot for config changes
	case "reboot", "reload":
		if !botutils.AuthorIsAdmin(cmd.Message, s) {
			return
		}
		ph.SaveAll()
		ph.CleanupAll()
		loadConfig()
		StartPlugins(ph)
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	handle(s, m.Message)
}

func messageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	handle(s, m.Message)
}
