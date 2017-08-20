package youtubeplugin

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dpup/gohubbub"
	"github.com/itokatsu/nanogo/parser"
)

var client = &http.Client{Timeout: 10 * time.Second}

type youtubePlugin struct {
	name          string
	apiKey        string
	client        *gohubbub.Client
	port          int
	subscriptions map[string]*Subscription
}

func New(args ...string) *youtubePlugin {
	if len(args) < 3 {
		log.Fatal("Youtubeplugin requires 2 args")
	}
	var pInstance youtubePlugin
	pInstance.apiKey = args[0]
	pInstance.port, _ = strconv.Atoi(args[2])
	addr := fmt.Sprintf("%v:%v", args[1], args[2])
	pInstance.client = gohubbub.NewClient(addr, "Nanogo")
	pInstance.subscriptions = make(map[string]*Subscription)
	go pInstance.client.StartAndServe("", pInstance.port)
	return &pInstance
}

func (p *youtubePlugin) Name() string {
	return "youtube"
}

type Subscription struct {
	Channel        YtChannel
	Url            string
	AddedAt        discordgo.Timestamp
	AddedBy        *discordgo.User
	NotifChannelID string
}

type ChannelListResponse struct {
	Items []YtChannel `json:"items"`
}

type YtChannel struct {
	Kind    string `json:"kind"`
	Etag    string `json:"etag"`
	Id      string `json:"id"`
	Snippet struct {
		Title string `json:"title"`
		Desc  string `json:"description"`
	} `json:"snippet"`
	Stats struct {
		ViewCount string `json:"viewCount"`
		SubCount  string `json:"subscriberCount"`
		VidCount  string `json:"videoCount"`
	} `json:"statistics"`
}

type Feed struct {
	Title string `xml:title"`
	Video Entry  `xml:"entry"`
}

type Entry struct {
	VideoId string `xml:"id"`
	Title   string `xml:"title"`
	Author  struct {
		Name string `xml:"name"`
		Url  string `xml:"url"`
	} `xml:"author"`
}

func getJSON(url string, target interface{}) error {
	r, err := client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func (p *youtubePlugin) AddSubscription(sub *Subscription, s *discordgo.Session) error {
	err := p.client.DiscoverAndSubscribe(sub.Url, func(contentType string, body []byte) {
		var feed Feed
		xmlError := xml.Unmarshal(body, &feed)
		if xmlError != nil {
			log.Printf("XML Parse Error %v", xmlError)
		} else {
			vid := feed.Video
			msg := fmt.Sprintf("New video from **%s** \n «%s» \n http://www.youtube.com/watch?v=%s",
				vid.Author.Name, vid.Title, vid.VideoId[len("yt:video:"):])
			s.ChannelMessageSend(sub.NotifChannelID, msg)
		}
	})
	if err != nil {
		return err
	} else {
		p.subscriptions[sub.Channel.Snippet.Title] = sub
		return nil
	}

}

func (p *youtubePlugin) GetChannelBy(fieldName string, searchTerm string) (YtChannel, error) {
	url := "https://www.googleapis.com/youtube/v3/"
	url += fmt.Sprintf("channels?part=snippet,statistics&%v=%v&key=%v", fieldName, searchTerm, p.apiKey)
	result := ChannelListResponse{}
	fmt.Printf("Trying... %s ", url)
	jsonerr := getJSON(url, &result)
	if jsonerr != nil {
		return YtChannel{}, fmt.Errorf("Failed to parse Youtube API response for %v", searchTerm)
	}
	if len(result.Items) < 1 {
		return YtChannel{}, fmt.Errorf("No channel found for %v", searchTerm)
	}
	return result.Items[0], nil
}

func (p *youtubePlugin) HandleMsg(cmd *parser.ParsedCmd, s *discordgo.Session, m *discordgo.MessageCreate) {
	switch strings.ToLower(cmd.Name) {
	case "yt", "youtube":
		if len(cmd.Args) < 2 {
			return
		}

		switch cmd.Args[0] {
		case "sub":
			theChannel := YtChannel{}
			theChannel, err := p.GetChannelBy("forUsername", cmd.Args[1])
			if err != nil {
				theChannel, err = p.GetChannelBy("id", cmd.Args[1])
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, err.Error())
					return
				}
			}

			//Build Subscription
			var sub Subscription
			sub.Channel = theChannel
			sub.AddedAt = m.Timestamp
			sub.AddedBy = m.Author
			if len(cmd.Args) > 2 {
				//@TODO handle channel mention
				sub.NotifChannelID = cmd.Args[2]
			} else {
				sub.NotifChannelID = m.ChannelID
			}
			sub.Url = "https://www.youtube.com/xml/feeds/"
			sub.Url += fmt.Sprintf("videos.xml?channel_id=%v", sub.Channel.Id)

			err = p.AddSubscription(&sub, s)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}
			msg := fmt.Sprintf("Now subscribed to %v \n %v Videos | %v Subscribers | %v Views",
				sub.Channel.Snippet.Title, sub.Channel.Stats.VidCount,
				sub.Channel.Stats.SubCount, sub.Channel.Stats.ViewCount)
			s.ChannelMessageSend(m.ChannelID, msg)

		case "unsub":
			/*if v, ok := p.subscriptions[strings.ToLower(ytArgs)]; ok {
				p.client.Unsubscribe(v)
			}*/

		case "list":

		}
	}
}

func (p *youtubePlugin) Help() string {
	return "oupas"
}

func (p *youtubePlugin) Cleanup() {
	for _, s := range p.subscriptions {
		p.client.Unsubscribe(s.Url)
	}
	time.Sleep(time.Second * 5)
}
