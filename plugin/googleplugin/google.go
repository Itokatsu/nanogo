// Nanobot Project
//
// Plugin for google searches

/*
@TODO: googleimg ~ collage
@TODO: cmd with more results
*/

package googleplugin

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/itokatsu/nanogo/botutils"
)

type Config struct {
	ApiKey string `json:"apikey"`
}

type googlePlugin struct {
	name     string
	apiKey   string
	lastReqs map[string]*url.URL
}

func New(cfg Config) (*googlePlugin, error) {
	var pInstance googlePlugin
	if cfg.ApiKey == "" {
		e := errors.New("no Apikey found for google")
		return nil, e
	}
	pInstance.apiKey = cfg.ApiKey
	pInstance.lastReqs = make(map[string]*url.URL)
	return &pInstance, nil
}

func (p *googlePlugin) Name() string {
	return "google"
}

func (p *googlePlugin) HasData() bool {
	return false
}

type SearchResults struct {
	Items []Result
}
type Result struct {
	Title           string
	Link            string
	Snippet         string
	ThumbnailLink   string
	ThumbnailHeight int
	ThumbnailWidth  int
}

func (r Result) String() string {
	return fmt.Sprintf("%s\n%s", r.Title, r.Link)
}

func (p *googlePlugin) buildRequestURL(query string, n int) url.URL {
	numStr := strconv.Itoa(n)
	qs := url.Values{}
	qs.Set("key", p.apiKey)
	qs.Set("cx", "004895194701224026743:zdbrbrrm0bw")
	qs.Set("client", "google-csbe")
	qs.Set("num", numStr)
	qs.Set("ie", "utf8")
	qs.Set("oe", "utf8")
	qs.Set("fields", "items(title,link,snippet)")
	query = strings.Replace(query, "/", "", -1)
	query = strings.Replace(query, "&", "", -1)
	qs.Set("q", query)

	var reqUrl = url.URL{
		Scheme:   "https",
		Host:     "www.googleapis.com",
		Path:     "customsearch/v1",
		RawQuery: qs.Encode(),
	}
	return reqUrl
}

func (p *googlePlugin) HandleMsg(cmd *botutils.Cmd, s *discordgo.Session, m *discordgo.MessageCreate) {
	switch strings.ToLower(cmd.Name) {

	//first google result
	case "g", "google":
		if len(cmd.Args) == 0 {
			return
		}
		query := strings.Join(cmd.Args, " ")
		url := p.buildRequestURL(query, 1)

		p.lastReqs[m.ChannelID] = &url
		result := SearchResults{}
		err := botutils.FetchJSON(url.String(), &result)
		if err != nil {
			return
		}

		if len(result.Items) > 0 {
			s.ChannelMessageSend(m.ChannelID, result.Items[0].Link)
		} else {
			msg := fmt.Sprintf("No result found for %v", strings.Join(cmd.Args, " "))
			s.ChannelMessageSend(m.ChannelID, msg)
		}

	//google img
	case "gis":
		if len(cmd.Args) == 0 {
			return
		}
		query := strings.Join(cmd.Args, " ")
		qs := url.Values{}
		qs.Set("q", query)
		qs.Set("tbm", "isch")
		var reqUrl = url.URL{
			Scheme:   "https",
			Host:     "www.google.com",
			Path:     "search",
			RawQuery: qs.Encode(),
		}
		s.ChannelMessageSend(m.ChannelID, reqUrl.String())
		/*		url := p.buildRequestURL(query, 1)
				url.Query().Set("searchType", "image")

				p.lastReqs[m.ChannelID] = &url
				results := []SearchResults{}
				err := botutils.FetchJSON(url.String(), &results)
				if err != nil {
					return
				}*/

	//more results
	case "gm":
		// get last request
		url, ok := p.lastReqs[m.ChannelID]
		if !ok {
			return
		}

		// build new query url
		qs := url.Query()
		start, err := strconv.Atoi(qs.Get("start"))
		if err != nil {
			start = 1
		}
		offset, err := strconv.Atoi(qs.Get("num"))
		if err != nil {
			return
		}
		qs.Set("start", strconv.Itoa(start+offset))
		qs.Set("num", "5")
		url.RawQuery = qs.Encode()

		// Get results
		results := SearchResults{}
		errJson := botutils.FetchJSON(url.String(), &results)
		if errJson != nil {
			fmt.Printf("json error")
			return
		}
		p.lastReqs[m.ChannelID] = url

		if len(results.Items) > 0 {
			_, err := botutils.NewMenu(s, results.Items, "\n", m.ChannelID)
			if err != nil {
				fmt.Printf(err.Error())
				return
			}
		} else {
			msg := fmt.Sprintf("No more result for %v", strings.Join(cmd.Args, " "))
			s.ChannelMessageSend(m.ChannelID, msg)
		}
	}
}

func (p *googlePlugin) Help() string {
	return `
	!g <term> - Return first result from a Google search
	!gm - Return 5 more results
	
	!gis <term> - Return first result from a Google image search
	`
}
