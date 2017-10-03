package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/itokatsu/nanogo/botutils"
)

// plugins
import (
	"github.com/itokatsu/nanogo/diceplugin"
	"github.com/itokatsu/nanogo/googleplugin"
	"github.com/itokatsu/nanogo/infoplugin"
	"github.com/itokatsu/nanogo/tagplugin"
	//	"github.com/itokatsu/nanogo/testplugin"
	"github.com/itokatsu/nanogo/catplugin"
	"github.com/itokatsu/nanogo/wolframplugin"
	"github.com/itokatsu/nanogo/youtubeplugin"
)

// various Auth Tokens and API Keys
type ConfigKeys struct {
	BotToken   string `json:"token"`
	NASAKey    string `json:"nasa"`
	GoogleKey  string `json:"google"`
	YoutubeKey string `json:"youtube"`
	WolframKey string `json:"wolfram"`
	IP         string `json:"ip"`
	Port       string `json:"port"`

	TestingOnly bool   `json:"testing"`
	TestChannel string `json:"testChannel"`
}

// Global variables
var (
	ph        pluginHandler
	CmdPrefix string
	Cfg       ConfigKeys
	StartTime time.Time
)

func loadConfig() {
	file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		fmt.Println("fatal")
		os.Exit(1)
	}
	json.Unmarshal(file, &Cfg)
}

func init() {
	CmdPrefix = "!"
	loadConfig()
	StartTime = time.Now()
	rand.Seed(time.Now().UnixNano())
}

func main() {

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Cfg.BotToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Load Plugins
	ph.plugins = make(map[string]Plugin)
	ph.Load(infoplugin.New(StartTime))
	ph.Load(diceplugin.New())
	ph.Load(tagplugin.New())
	ph.Load(catplugin.New())

	ph.Load(wolframplugin.New(Cfg.WolframKey))
	ph.Load(googleplugin.New(Cfg.GoogleKey))
	ph.Load(youtubeplugin.New(Cfg.YoutubeKey, Cfg.IP, Cfg.Port))
	defer ph.Save()

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	defer dg.Close()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore messages from bots
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}
	// Test channel only
	if Cfg.TestingOnly && m.ChannelID != Cfg.TestChannel {
		return
	}
	if cmd := botutils.ParseCmd(m.Content, CmdPrefix); cmd.Name != "" {
		//@TODO: use goroutines?
		for _, p := range ph.plugins {
			p.HandleMsg(&cmd, s, m)
		}
	}
}
