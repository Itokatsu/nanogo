package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/itokatsu/nanogo/parser"
	"github.com/itokatsu/nanogo/pingplugin"
)

// Global variables
var (
	DiscordToken string
	CmdPrefix    string
	ph           pluginHandler
)

func init() {
	CmdPrefix = "!"
	flag.StringVar(&DiscordToken, "t", "", "Bot Token")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())
}

func main() {

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + DiscordToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Load Plugins
	ph.plugins = make(map[string]Plugin)
	ph.Load(pingplugin.New())
	//	ph.Load(rollplugin.New())

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

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}

	if cmd := parser.Cmd(m.Content, CmdPrefix); cmd.Name == "" {
		return
	} else {
		//@TODO: use goroutines
		for _, p := range ph.plugins {
			p.HandleMsg(&cmd, s, m)
		}
	}
	/*		if cmd == "roll" {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			s.ChannelMessageSend(m.ChannelID, strconv.Itoa(r.Intn(20)+1))
		}*/

}
