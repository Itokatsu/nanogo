package jpplugin

import (
	"fmt"
	"time"
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/itokatsu/nanogo/botutils"
	"github.com/itokatsu/nanogo/plugin/jpplugin/jmdict"
	"github.com/itokatsu/nanogo/plugin/jpplugin/epwing"
)

type jpPlugin struct {
	dicts map[string]Dict
}

type Config struct {
	JMdict    string `json:"jmdict"`
	Daijirin  string `json:"daijirin"`
	Kenkyusha string `json:"kenkyusha"`
}

func New(cfg Config) (*jpPlugin, error) {
	var p jpPlugin

	p.dicts = make(map[string]Dict)
	//jmdict
	if dict, err := jmdict.Load(cfg.JMdict);
		err == nil {
		p.dicts["j"] = dict
	}
	//daijirin
	if dict, err := epwing.Load(cfg.Daijirin); 
		err == nil {
		p.dicts["dj"] = dict
	}
	//kenkyusha
	if dict, err := epwing.Load(cfg.Kenkyusha);
		err == nil {
		p.dicts["kks"] = dict
	}

	if len(p.dicts) < 1 {
		return nil, errors.New("no dictionary loaded")
	}
	return &p, nil
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

	/*switch strings.ToLower(cmd.Name) {
	}*/

	// Lookups
	var results []DictEntry
	for key, dict := range p.dicts {
		if key == cmd.Name {
			results = dict.Lookup(cmd.Args[0])
			break
		}
		if key + "r" == cmd.Name {
			results, _ = dict.LookupRe(cmd.Args[0])
			break
		}
	}

	n := len(results)
	if n < 1 {
		return
	}
	if n == 1 {
		botutils.Send(s, m.ChannelID, results[0].Details())
		return
	}
	if n > 25 {
		results = results[:25]
	}

	resultsStr := make([]fmt.Stringer, len(results))
	for i, v := range results {
		resultsStr[i] = fmt.Stringer(v)
	}
	s.ChannelMessageSend(m.ChannelID, 
		fmt.Sprintf("%d Results (in %v) : ", n, time.Since(now)))
	botutils.NewMenu(s, resultsStr, " | ", m.ChannelID, func(strger fmt.Stringer) {
		/*
		type assert
		*/
		e := epwing.Entry{}
		e = strger.(epwing.Entry)
		botutils.Send(s, m.ChannelID, e.Details())
	})
}

func (p *jpPlugin) Help() string {
	return `
	Tools around Japanese language
	!j - Lookup japanese word
	!k - Lookup japanese ideogram`
}