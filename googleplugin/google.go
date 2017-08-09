package googleplugin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/itokatsu/nanogo/parser"
)

var client = &http.Client{Timeout: 10 * time.Second}

type googlePlugin struct {
	name   string
	apiKey string
}

func (p *googlePlugin) Name() string {
	return "google"
}

type SearchResults struct {
	Items []struct {
		Title   string
		Link    string
		Snipper string
	}
}

func (p *googlePlugin) buildRequestURL(query string) url.URL {
	qs := url.Values{}
	qs.Add("key", p.apiKey)
	qs.Add("cx", "004895194701224026743:zdbrbrrm0bw")
	qs.Add("client", "google-csbe")
	qs.Add("num", "1")
	qs.Add("ie", "utf8")
	qs.Add("oe", "utf8")
	qs.Add("fields", "items(title,link,snippet)")
	query = strings.Replace(query, "/", "", -1)
	query = strings.Replace(query, "&", "", -1)
	qs.Add("q", query)

	var reqUrl = url.URL{
		Scheme:   "https",
		Host:     "www.googleapis.com",
		Path:     "customsearch/v1",
		RawQuery: qs.Encode(),
	}
	return reqUrl
}

func getJSON(url url.URL, target interface{}) error {
	r, err := client.Get(url.String())
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func (p *googlePlugin) HandleMsg(cmd *parser.ParsedCmd, s *discordgo.Session, m *discordgo.MessageCreate) {

	switch strings.ToLower(cmd.Name) {
	case "g":
		if len(cmd.Args) == 0 {
			return
		}
		query := strings.Join(cmd.Args, " ")
		url := p.buildRequestURL(query)
		result := SearchResults{}
		getJSON(url, &result)

		if len(result.Items) > 0 {
			s.ChannelMessageSend(m.ChannelID, result.Items[0].Link)
		} else {
			msg := fmt.Sprintf("No result found for %v", strings.Join(cmd.Args, " "))
			s.ChannelMessageSend(m.ChannelID, msg)
		}
	}
}

func (p *googlePlugin) Help() string {
	return "return first result of google search"
}

func New(apiKey string) *googlePlugin {
	var pInstance googlePlugin
	pInstance.apiKey = apiKey
	return &pInstance
}
