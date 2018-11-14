package jpplugin

import (
	"strings"
	//"os"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/itokatsu/nanogo/botutils"
	//"github.com/itokatsu/nanogo/plugin/jpplugin/jmdict"
	"github.com/itokatsu/nanogo/plugin/jpplugin/epwing"
)

type jpPlugin struct {
	//jmdict *jmdict.JMdict
	daijirin *epwing.Dict
}

type Config struct {
	JMdict    string `json:"jmdict"`
	Daijirin  string `json:"daijirin"`
	Kenkyusha string `json:"kenkyusha"`
}

func New(cfg Config) (*jpPlugin, error) {
	var pInstance jpPlugin

	/*tests := []string{"おくじょう", "臆病", "かみ"}
	testsre := []string{".ぬ", "(.)\\1しい"}*/

	//pInstance.jmdict = jmdict.LoadFromFile(cfg.JMdict)
	pInstance.daijirin, _ = epwing.LoadDir(cfg.Daijirin)

	/*
		var results []epwing.Entry
		for _, t := range tests {
			//results = pInstance.jmdict.Lookup(t)
			results = pInstance.daijirin.Lookup(t)
			fmt.Printf("%v", results)
		}
		for _, tr := range testsre {
			//results = pInstance.jmdict.Lookup(t)
			results, err := pInstance.daijirin.LookupRe(tr)
			if err != nil { continue }
			fmt.Printf("%v", results)
		}*/
	return &pInstance, nil
}

func (p *jpPlugin) Name() string {
	return "japanese"
}

func (p *jpPlugin) HasData() bool {
	return false
}

// Hiragana : U+3040 ~ U+309F
// Katakana : U+30A0 ~ U+30FF

func (p *jpPlugin) HandleMsg(cmd *botutils.Cmd, s *discordgo.Session, m *discordgo.MessageCreate) {
	now := time.Now()
	if len(cmd.Args) < 1 {
		return
	}
	query := cmd.Args[0]
	var results []epwing.Entry
	var err error
	switch strings.ToLower(cmd.Name) {
	case "dj":
		results = p.daijirin.Lookup(query)

	case "djr":
		results, err = p.daijirin.LookupRe(query)
		if err != nil {
			return
		}
	}
	n := len(results)
	if n < 1 {
		return
	}
	if n == 1 {
		s.ChannelMessageSend(m.ChannelID, results[0].Details())
	}
	if n > 25 {
		results = results[:25]
	}

	resultsStr := make([]fmt.Stringer, len(results))
	for i, v := range results {
		resultsStr[i] = fmt.Stringer(v)
	}
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%d Results (in %v) : ", n, time.Since(now)))
	botutils.NewMenu(s, resultsStr, " | ", m.ChannelID, func(strger fmt.Stringer) {
		e := epwing.Entry{}
		e = strger.(epwing.Entry)
		s.ChannelMessageSend(m.ChannelID, e.Details())
	})
}

/*func (p *jpPlugin) searchJMDict(query string) (results []jmdict.Entry) {
	now := time.Now()
	for _, entry := range p.jmdict.Entries {
		for _, kanji := range entry.KanjiElements {
			if query == kanji.Phrase {
				results = append(results, entry)
				continue
			}
		}
		for _, reading := range entry.ReadingElements {
			if query == reading.Phrase ||
				query == reading.PhraseNoKanji {
					results = append(results, entry)
				}
		}
	}
	fmt.Printf("%v", time.Since(now))
}*/

func (p *jpPlugin) Help() string {
	return `
	Tools around Japanese language
	!j - Lookup japanese word
	!k - Lookup japanese ideogram`
}
