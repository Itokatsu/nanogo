package jpplugin

import (
	"errors"
	"strings"
	//"time"

	"github.com/bwmarrin/discordgo"
	"github.com/itokatsu/nanogo/botutils"
	//"github.com/itokatsu/nanogo/plugin/jpplugin/epwing"
	//"github.com/itokatsu/nanogo/plugin/jpplugin/jmdict"
	"github.com/itokatsu/nanogo/plugin/jpplugin/jp"
	"github.com/itokatsu/nanogo/plugin/jpplugin/web"
)

type jpPlugin struct {
	dicts map[string]jp.Dict
}

type Config struct {
	Local     bool   `json:"local_dict"`
	JMdict    string `json:"jmdict"`
	Daijirin  string `json:"daijirin"`
	Kenkyusha string `json:"kenkyusha"`
}

func New(cfg Config) (*jpPlugin, error) {
	var p jpPlugin

	if cfg.Local {
		p.dicts = make(map[string]jp.Dict)
		//jmdict
		//if dict, err := jmdict.Load(cfg.JMdict); err == nil {
		//	p.dicts["j"] = dict
		//}
		//daijirin
		//if dict, err := epwing.Load(cfg.Daijirin); err == nil {
		//	p.dicts["dj"] = dict
		//}
		//kenkyusha
		//if dict, err := epwing.Load(cfg.Kenkyusha); err == nil {
		//	p.dicts["waei"] = dict
		//}

		if len(p.dicts) < 1 {
			return nil, errors.New("no dictionary loaded")
		}
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
	//now := time.Now()
	if len(cmd.Args) < 1 {
		return
	}

	switch strings.ToLower(cmd.Name) {
	case "j":

		var msg string
		var display int
		j, _ := web.JishoAPI(cmd.Args[0])
		if len(j) == 0 {
			return
		}
		if len(j) > 3 {
			display = 3
		} else {
			display = len(j)
		}
		msg += "```diff"
		// first three results
		for _, e := range j[:display] {
			msg += "\n+ " + e.Print()
		}
		msg += "```"
		s.ChannelMessageSend(m.ChannelID, msg)

	case "dj":
		var msg string
		var display int
		j, _ := web.Weblio(cmd.Args[0])
		if len(j) == 0 {
			return
		}
		if len(j) > 3 {
			display = 3
		} else {
			display = len(j)
		}
		msg += "```diff"
		// first three results
		for _, e := range j[:display] {
			msg += "\n+ " + e.Print()
		}
		msg += "```"
		s.ChannelMessageSend(m.ChannelID, msg)
	}

	/*
		var results []jp.DictEntry
		for key, dict := range p.dicts {
			//Exact Lookup
			if key == cmd.Name {
				results = dict.Lookup(cmd.Args[0])
				break
			}
			//Regexp Lookup
			if key+"r" == cmd.Name {
				results, _ = dict.LookupRe(cmd.Args[0])
				break
			}
			// Print Gaiji
			if key+"g" == cmd.Name {
				dictEp, ok := dict.(*epwing.Dict)
				if ok {
					gaiji := dictEp.GetGaijiBMP(cmd.Args[0])
					if gaiji != "" {
						gaiji = "```" + gaiji + "```"
					}
					s.ChannelMessageSend(m.ChannelID, gaiji)
				} else {
					s.ChannelMessageSend(m.ChannelID,
						fmt.Sprintf("isn't an EPWING dictionary is %T", dict))
				}
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

		s.ChannelMessageSend(m.ChannelID,
			fmt.Sprintf("%d Results (in %v) : ", n, time.Since(now)))

		c, err := botutils.NewMenu(s, results, " | ", m.ChannelID)
		if err != nil {
			fmt.Printf(err.Error())
		}
		go func() {
			for e := range c {
				entry := e.(jp.DictEntry)
				botutils.Send(s, m.ChannelID, entry.Details())
			}
		}()*/
}

func (p *jpPlugin) Help() string {
	return `
	Tools around Japanese language
	!j - Lookup japanese word
	!k - Lookup japanese ideogram`
}
