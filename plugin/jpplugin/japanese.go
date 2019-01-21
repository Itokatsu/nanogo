package jpplugin

import (
	"fmt"
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
		/*
			p.dicts = make(map[string]jp.Dict)
			//jmdict
			if dict, err := jmdict.Load(cfg.JMdict); err == nil {
				p.dicts["j"] = dict
			//}
			//daijirin
			if dict, err := epwing.Load(cfg.Daijirin); err == nil {
				p.dicts["dj"] = dict
			//}
			//kenkyusha
			if dict, err := epwing.Load(cfg.Kenkyusha); err == nil {
				p.dicts["waei"] = dict
			//}

			if len(p.dicts) < 1 {
				return nil, errors.New("no dictionary loaded")
			}*/
	}
	return &p, nil
}

func (p *jpPlugin) Name() string {
	return "japanese"
}

func (p *jpPlugin) HasData() bool {
	return false
}

const PLUS_EMOJI = "\u2795"

func (p *jpPlugin) HandleMsg(cmd *botutils.Cmd, s *discordgo.Session) {
	//now := time.Now()
	if len(cmd.Args) < 1 {
		return
	}

	switch strings.ToLower(cmd.Name) {
	case "jtag":
		desc, ok := web.JishoTags[cmd.Args[0]]
		if ok {
			s.ChannelMessageSend(cmd.ChannelID, desc)
		}

	case "j":
		var search = 1
		results, err := web.JishoAPI(cmd.Args[0], search)
		// Request error ?
		if err != nil {
			botutils.SendErrorMsg(cmd, s, err)
		}
		if len(results) < 1 {
			fmt.Println("no results")
			return
		}

		var pager *botutils.Pager
		pager, err = botutils.NewPager(results)
		if err != nil {
			fmt.Println(err)
			return
		}

		if len(results) == 20 {
			pager.AddCustomEmoji(PLUS_EMOJI, func() {
				search++
				results, err := web.JishoAPI(cmd.Args[0], search)
				if err != nil {
					fmt.Println(err)
					return
				}
				pager.AddItems(results)
			})
		}

		msg, _ := s.ChannelMessageSendEmbed(cmd.ChannelID, pager.PageEmbed().MessageEmbed)
		botutils.LinkPager(s, msg, pager)
	}
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
				s.ChannelMessageSend(cmd.ChannelID, gaiji)
			} else {
				s.ChannelMessageSend(cmd.ChannelID,
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
		botutils.Send(s, cmd.ChannelID, results[0].Details())
		return
	}
	if n > 25 {
		results = results[:25]
	}

	s.ChannelMessageSend(cmd.ChannelID,
		fmt.Sprintf("%d Results (in %v) : ", n, time.Since(now)))

	c, err := botutils.NewMenu(s, results, " | ", cmd.ChannelID)
	if err != nil {
		fmt.Printf(err.Error())
	}
	go func() {
		for e := range c {
			entry := e.(jp.DictEntry)
			botutils.Send(s, cmd.ChannelID, entry.Details())
		}
	}()*/

func (p *jpPlugin) Help() string {
	return `
	Tools around Japanese language
	!j - Lookup japanese word
	!k - Lookup japanese ideogram`
}
