// Nanobot Project
//
// Plugin for WolframAlpha integration

package wolframplugin

import (
	"errors"
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/itokatsu/nanogo/botutils"
)

type Config struct {
	ApiKey string `json:"apikey"`
}

type wolframPlugin struct {
	name   string
	apiKey string
}

func New(cfg Config) (*wolframPlugin, error) {
	var pInstance wolframPlugin
	if cfg.ApiKey == "" {
		e := errors.New("no Apikey found for wolfram")
		return &pInstance, e
	}
	pInstance.apiKey = cfg.ApiKey
	return &pInstance, nil
}

func (p *wolframPlugin) Name() string {
	return "wolfram"
}

func (p *wolframPlugin) HasData() bool {
	return false
}

type SearchResults struct {
	Res struct {
		Success bool    `json:"success"`
		Numpods float64 `json:"numpods"`
		Pods    []Pod   `json:"pods"`
	} `json:"queryresult"`
}

type Pod struct {
	Title      string   `json:"title"`
	Numsubpods float64  `json:"numsubpods"`
	SubPods    []SubPod `json:"subpods"`
	/*	States []struct {
		Name  string `json:"name"`
		Input string `json:"input"`
	} `json:"states"`*/
}

type SubPod struct {
	Text string `json:"plaintext"`
	Img  struct {
		Url string `json:"src"`
	} `json:"img"`
}

func (pod Pod) String() string {
	return pod.Title
}

func (pod Pod) Display(s *discordgo.Session, channelID string) {
	spod := pod.SubPods[0]
	if spod.Text != "" {
		s.ChannelMessageSend(channelID, spod.Text)
		return
	}
	resp, err := botutils.Client.Get(spod.Img.Url)
	if err != nil {
		return
	}
	s.ChannelFileSend(channelID, "wolfram.gif", resp.Body)
}

var WolframEndpoint = "https://api.wolframalpha.com/v2/query"

func (p *wolframPlugin) buildRequestURL(query string) *url.URL {
	qs := url.Values{}
	qs.Set("input", query)
	qs.Set("appid", p.apiKey)
	qs.Set("location", "Paris")
	qs.Set("output", "json")
	//qs.Set("format", "image,plaintext")
	reqUrl, _ := url.Parse(WolframEndpoint)
	reqUrl.RawQuery = qs.Encode()
	return reqUrl
}

func (p *wolframPlugin) HandleMsg(cmd *botutils.Cmd, s *discordgo.Session, m *discordgo.MessageCreate) {
	switch strings.ToLower(cmd.Name) {
	case "wa":
		if len(cmd.Args) == 0 {
			return
		}
		query := strings.Join(cmd.Args, " ")
		url := p.buildRequestURL(query)

		resp := SearchResults{}
		err := botutils.FetchJSON(url.String(), &resp)
		if err != nil {
			return
		}
		results := resp.Res.Pods
		if len(results) == 1 {
			results[0].Display(s, m.ChannelID)
		} else {
			c, err := botutils.NewMenu(s, results, " | ", m.ChannelID)
			if err != nil {
				return
			}

			go func() {
				for resp := range c {
					pod := resp.(Pod)
					pod.Display(s, m.ChannelID)
				}
			}()
		}

	//result pod only
	case "war":
		if len(cmd.Args) == 0 {
			return
		}
		query := strings.Join(cmd.Args, " ")
		url := p.buildRequestURL(query)
		qs := url.Query()
		qs.Set("includepodid", "Result")
		url.RawQuery = qs.Encode()
		resp := SearchResults{}
		err := botutils.FetchJSON(url.String(), &resp)
		if err != nil {
			return
		}
		result := resp.Res
		if !result.Success {
			return
		}
		if result.Numpods < 1 {
			c := botutils.Cmd{
				Name: "wa",
				Args: cmd.Args,
			}
			p.HandleMsg(&c, s, m)
			return
		}
		result.Pods[0].Display(s, m.ChannelID)
	}
}

func (p *wolframPlugin) Help() string {
	return `
	`
}
