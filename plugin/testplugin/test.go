package googleplugin

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/itokatsu/nanogo/botutils"
)

type googlePlugin struct {
	name   string
	apiKey string
}

func New(apiKey string) *googlePlugin, error {
	var pInstance googlePlugin
	pInstance.apiKey = apiKey
	return &pInstance
}
func (p *testPlugin) HasData() bool {
	return false
}
func (p *testPlugin) Name() string {
	return "google"
}

type SearchResults struct {
	Items []Result
}
type Result struct {
	Title   string
	Link    string
	Snipper string
}

func (r Result) String() string {
	if len(r.Title) > 40 {
		return fmt.Sprintf("%v…", r.Title[:40])
	}
	return r.Title
}

func (p *googlePlugin) buildRequestURL(query string) url.URL {
	qs := url.Values{}
	qs.Add("key", p.apiKey)
	qs.Add("cx", "004895194701224026743:zdbrbrrm0bw")
	qs.Add("client", "google-csbe")
	qs.Add("num", "5")
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

func (p *googlePlugin) HandleMsg(cmd *botutils.Cmd, s *discordgo.Session) {
	switch strings.ToLower(cmd.Name) {
	case "g":
		if len(cmd.Args) == 0 {
			return
		}
		query := strings.Join(cmd.Args, " ")
		url := p.buildRequestURL(query)
		result := SearchResults{}
		err := botutils.FetchJSON(url.String(), &result)
		if err != nil {
			return
		}

		if len(result.Items) == 1 {
			s.ChannelMessageSend(cmd.ChannelID, result.Items[0].Link)
			return
		}
		if len(result.Items) < 1 {
			msg := fmt.Sprintf("No result found for %v", strings.Join(cmd.Args, " "))
			s.ChannelMessageSend(cmd.ChannelID, msg)
			return
		}

		var strgers []fmt.Stringer
		for _, r := range result.Items {
			strgers = append(strgers, r)
		}
		botutils.NewMenu(s, strgers, "\n", cmd.ChannelID, func(stger fmt.Stringer) {
			res := stger.(Result)
			s.ChannelMessageSend(cmd.ChannelID, res.Link)
		})

	}
}

func (p *googlePlugin) Help() string {
	return "return first result of google search"
}

func (p *googlePlugin) Save() []byte {
	buf, err := json.Marshal(p)
	if err != nil {
		fmt.Errorf("Failed to convert plugin state to json")
	}
	return buf
}

func (p *googlePlugin) Load(data []byte) error {
	if data == nil {
		return nil
	}
	err := json.Unmarshal(data, p)
	if err != nil {
		fmt.Println("Error loading data", err)
		return err
	}
	return nil
}

func (p *googlePlugin) Cleanup() {
}
