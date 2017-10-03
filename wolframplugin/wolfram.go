// Nanobot Project
//
// Plugin for WolframAlpha integration

package wolframplugin

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/itokatsu/nanogo/botutils"
)

type wolframPlugin struct {
	name   string
	apiKey string
}

func New(apiKey string) *wolframPlugin {
	var pInstance wolframPlugin
	pInstance.apiKey = apiKey
	return &pInstance
}

func (p *wolframPlugin) Name() string {
	return "wolfram"
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

func (p *wolframPlugin) buildRequestURL(query string) url.URL {
	qs := url.Values{}
	qs.Set("input", query)
	qs.Set("appid", p.apiKey)
	qs.Set("location", "Paris")
	qs.Set("output", "json")
	//qs.Set("format", "image,plaintext")
	var reqUrl = url.URL{
		Scheme:   "https",
		Host:     "api.wolframalpha.com",
		Path:     "v2/query",
		RawQuery: qs.Encode(),
	}
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
		result := resp.Res
		strers := make([]fmt.Stringer, len(result.Pods))
		for i, pod := range result.Pods {
			strers[i] = pod
		}
		botutils.NewMenu(s, strers, " | ", m.ChannelID, func(strer fmt.Stringer) {
			pod := strer.(Pod)
			pod.Display(s, m.ChannelID)
		})

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
		result := resp.Res
		if err != nil {
			return
		}
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
