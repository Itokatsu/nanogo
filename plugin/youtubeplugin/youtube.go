package youtubeplugin

import (
	"bufio"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dpup/gohubbub"
	"github.com/itokatsu/nanogo/botutils"
	"github.com/itokatsu/nanogo/plugin"
)

type Config struct {
	ApiKey string `json:"apikey"`
	Ip     string `json:"ip"`
	Port   int    `json:"port"`
}

type youtubePlugin struct {
	name   string
	apiKey string
	client *gohubbub.Client
	port   int
	Subs   map[string]*Subscription `json:"subs,omitempty"`
}

func New(cfg Config) (*youtubePlugin, error) {
	if cfg.ApiKey == "" || cfg.Ip == "" || cfg.Port == 0 {
		e := errors.New("Lacking config for plugin youtube")
		return nil, e
	}
	var pInstance youtubePlugin
	pInstance.name = "youtube"
	pInstance.apiKey = cfg.ApiKey
	pInstance.port = cfg.Port

	// HTTP Client for subscription notifications
	addr := fmt.Sprintf("%v:%v", cfg.Ip, cfg.Port)
	pInstance.client = gohubbub.NewClient(addr, "Nanogo")

	pInstance.Subs = make(map[string]*Subscription)
	go pInstance.client.StartAndServe("", cfg.Port)
	return &pInstance, nil
}

// filters value
const whitelist = 1
const blacklist = -1
const off = 0

func (p *youtubePlugin) Name() string {
	return "youtube"
}

func (p *youtubePlugin) Help() string {
	return "oupas"
}

func (p *youtubePlugin) HasData() bool {
	return true
}

type Subscription struct {
	Channel         YtChannel
	FeedUrl         string
	AddedAt         time.Time
	AddedBy         *discordgo.User
	NotifChannelIDs []string
	LastNotif       string
	Filter          int
}

func (p *youtubePlugin) GetSubscriptions(str string) []*Subscription {
	var matches []*Subscription
	for id, sub := range p.Subs {
		title := strings.ToLower(sub.Channel.Snippet.Title)
		if id == str || strings.Contains(title, strings.ToLower(str)) {
			matches = append(matches, sub)
		}
	}
	return matches
}

// sorter wrapper
type Subscriptions []*Subscription

func (s Subscriptions) Len() int {
	return len(s)
}
func (s Subscriptions) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type ByName struct{ Subscriptions }
type ByDate struct{ Subscriptions }
type ByUser struct{ Subscriptions }
type ByVids struct{ Subscriptions }
type BySubs struct{ Subscriptions }

func (s ByName) Less(i, j int) bool {
	return s.Subscriptions[i].Channel.Snippet.Title < s.Subscriptions[j].Channel.Snippet.Title
}
func (s ByDate) Less(i, j int) bool {
	return s.Subscriptions[i].AddedAt.Before(s.Subscriptions[j].AddedAt)
}
func (s ByUser) Less(i, j int) bool {
	return s.Subscriptions[i].AddedBy.Username < s.Subscriptions[j].AddedBy.Username
}
func (s ByVids) Less(i, j int) bool {
	return s.Subscriptions[i].Channel.Stats.VidCount < s.Subscriptions[j].Channel.Stats.VidCount
}
func (s BySubs) Less(i, j int) bool {
	return s.Subscriptions[i].Channel.Stats.SubCount < s.Subscriptions[j].Channel.Stats.SubCount
}

/* Youtube API */
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
		SubHidden bool   `json:"hiddenSubscriberCount"`
	} `json:"statistics"`
}

func (c *YtChannel) Details() string {
	if c.Stats.SubHidden {
		return fmt.Sprintf("Now subscribed to %v \n %v Videos | %v Views",
			c.Snippet.Title, c.Stats.VidCount, c.Stats.ViewCount)
	}
	return fmt.Sprintf("Now subscribed to %v \n %v Subscribers | %v Videos | %v Views",
		c.Snippet.Title, c.Stats.SubCount, c.Stats.VidCount, c.Stats.ViewCount)
}

type VideoListResponse struct {
	Items []YtVideo `json:"items"`
}

type YtVideo struct {
	Kind    string `json:"kind"`
	Etag    string `json:"etag"`
	Id      string `json:"id"`
	Snippet struct {
		PublishedAt time.Time `json:"publishedAt"`
		ChannelID   string    `json:"channelId"`
		Title       string    `json:"title"`
		Desc        string    `json:"description"`
	} `json:"snippet"`
	ChannelName string `json:"channelTitle"`
}

/* Notification Feed */
type Link struct {
	Rel string `xml:"rel,attr"`
	Url string `xml:"href,attr"`
}
type Feed struct {
	Title   string    `xml:title"`
	Links   []Link    `xml:"link"`
	Entry   Entry     `xml:"entry"`
	Updated time.Time `xml:"updated"`
}

type Entry struct {
	VideoId string `xml:"id"`
	Title   string `xml:"title"`
	Author  struct {
		Name string `xml:"name"`
		Url  string `xml:"url"`
	} `xml:"author"`
	Link      Link      `xml:"link"`
	Published time.Time `xml:"published"`
	Updated   time.Time `xml:"updated"`
}

// GetChannel Info
func (p *youtubePlugin) GetChannelBy(fieldName string, searchTerm string) (YtChannel, error) {
	url := "https://www.googleapis.com/youtube/v3/"
	url += fmt.Sprintf("channels?part=snippet,statistics&%v=%v&key=%v", fieldName, searchTerm, p.apiKey)
	result := ChannelListResponse{}
	jsonerr := botutils.FetchJSON(url, &result)
	if jsonerr != nil {
		return YtChannel{}, fmt.Errorf("Failed to parse Youtube API response for %v", searchTerm)
	}
	if len(result.Items) < 1 {
		return YtChannel{}, fmt.Errorf("No channel found for %v", searchTerm)
	}
	return result.Items[0], nil
}

// Get Video info from Youtube API
func (p *youtubePlugin) GetVideo(id string) (YtVideo, error) {
	url := "https://www.googleapis.com/youtube/v3/"
	url += fmt.Sprintf("videos?part=snippet&id=%v&key=%v", id, p.apiKey)
	result := VideoListResponse{}
	jsonerr := botutils.FetchJSON(url, &result)
	if jsonerr != nil {
		return YtVideo{}, fmt.Errorf("Failed to parse Youtube API response for videos/%v", id)
	}
	if len(result.Items) < 1 {
		return YtVideo{}, fmt.Errorf("No Video found with id %v", id)
	}
	return result.Items[0], nil
}

func matchFilter(filepath string, str string) bool {
	file, err := os.Open(filepath)
	if err != nil {
		return false
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(str, scanner.Text()) {
			return true
		}
	}
	return false
}

func (p *youtubePlugin) AddSubscription(sub *Subscription, s *discordgo.Session) error {
	err := p.client.DiscoverAndSubscribe(sub.FeedUrl, func(contentType string, body []byte) {
		var feed Feed
		xmlError := xml.Unmarshal(body, &feed)
		if xmlError != nil {
			log.Printf("XML Parse Error %v", xmlError)
			return
		}
		// check for mismatch
		for _, link := range feed.Links {
			if link.Rel == "self" && link.Url != sub.FeedUrl {
				return
			}
		}

		entry := feed.Entry
		//check for valid id
		if len(entry.VideoId) <= len("yt:video:") {
			return
		}

		//try to filter updates out
		vidID := entry.VideoId[len("yt:video:"):]
		if sub.LastNotif == vidID {
			return
		}
		if entry.Updated.Sub(entry.Published) >= 2*time.Hour {
			return
		}

		//apply filter
		if sub.Filter != off {
			path := fmt.Sprintf("./youtube/%v.filter", sub.Channel.Id)
			r := matchFilter(path, entry.Title)
			if r && sub.Filter == blacklist {
				return
			}
			if !r && sub.Filter == whitelist {
				return
			}
		}

		sub.LastNotif = vidID
		msg := fmt.Sprintf("New video from **%s** \n `%s` \n http://www.youtube.com/watch?v=%s",
			entry.Author.Name, entry.Title, entry.VideoId[len("yt:video:"):])
		for _, c := range sub.NotifChannelIDs {
			s.ChannelMessageSend(c, msg)
		}
	}) // end of callback func
	if err != nil {
		return err
	} else {
		p.Subs[sub.Channel.Id] = sub
		return nil
	}
}

func (p *youtubePlugin) HandleMsg(cmd *botutils.Cmd, s *discordgo.Session) {
	switch strings.ToLower(cmd.Name) {
	case "yt", "youtube":
		if len(cmd.Args) < 1 {
			return
		}
		switch cmd.Args[0] {
		case "sub":
			if len(cmd.Args) < 2 {
				return
			}
			YTChan := YtChannel{}
			YTChan, err := p.GetChannelBy("forUsername", cmd.Args[1])
			if err != nil {
				YTChan, err = p.GetChannelBy("id", cmd.Args[1])
				if err != nil {
					s.ChannelMessageSend(cmd.ChannelID, err.Error())
					return
				}
			}

			notif := cmd.ChannelID
			if len(cmd.Args) > 2 {
				//@TODO handle channel mention
				notif = cmd.Args[2]
			}

			if sub, ok := p.Subs[YTChan.Id]; ok {
				for _, id := range sub.NotifChannelIDs {
					if notif == id {
						s.ChannelMessageSend(cmd.ChannelID, "Already subscribed")
						return
					}
				}
				sub.NotifChannelIDs = append(sub.NotifChannelIDs, notif)
				s.ChannelMessageSend(cmd.ChannelID, sub.Channel.Details())
				plugin.Save(p)
				return
			}
			//Build Subscription
			var sub Subscription
			sub.Channel = YTChan
			sub.AddedAt, _ = cmd.Timestamp.Parse()
			sub.AddedBy = cmd.Author
			sub.Filter = off
			sub.FeedUrl = "https://www.youtube.com/xml/feeds/"
			sub.FeedUrl += fmt.Sprintf("videos.xml?channel_id=%v", sub.Channel.Id)
			sub.NotifChannelIDs = make([]string, 0)
			sub.NotifChannelIDs = append(sub.NotifChannelIDs, notif)

			err = p.AddSubscription(&sub, s)
			if err != nil {
				s.ChannelMessageSend(cmd.ChannelID, err.Error())
				return
			}
			plugin.Save(p)
			s.ChannelMessageSend(cmd.ChannelID, sub.Channel.Details())

		case "unsub":
			if len(cmd.Args) < 2 {
				return
			}
			subs := p.GetSubscriptions(cmd.Args[1])
			if len(subs) == 0 {
				return
			}

			if cmd.Author.ID != subs[0].AddedBy.ID && !botutils.AuthorIsAdmin(s, cmd.Message) {
				s.ChannelMessageSend(cmd.ChannelID, "You cannot do that :nano:")
				return
			}
			p.client.Unsubscribe(subs[0].FeedUrl)
			delete(p.Subs, subs[0].Channel.Id)
			plugin.Save(p)
			msg := fmt.Sprintf("Unsubscribed from %v", subs[0].Channel.Snippet.Title)
			s.ChannelMessageSend(cmd.ChannelID, msg)
			return

		case "filter":
			if len(cmd.Args) < 2 {
				return
			}
			msg := ""
			subs := p.GetSubscriptions(cmd.Args[1])
			if len(subs) == 0 {
				return
			}
			if len(cmd.Args) < 3 {
				switch subs[0].Filter {
				case off:
					msg = "Filter : off"
				case blacklist:
					msg = "Filter : blacklist"
				case whitelist:
					msg = "Filter : whitelist"
				}
			} else {
				switch strings.ToLower(cmd.Args[2]) {
				case "whitelist", "white", "wl", "+":
					subs[0].Filter = whitelist
				case "blacklist", "black", "bl", "-":
					subs[0].Filter = blacklist
				case "=", "off":
					subs[0].Filter = off
				}
			}
			s.ChannelMessageSend(cmd.ChannelID, msg)
			plugin.Save(p)

		case "list":
			if len(p.Subs) < 1 {
				s.ChannelMessageSend(cmd.ChannelID, "No subscription found.")
				return
			}
			var subs Subscriptions
			for _, sub := range p.Subs {
				subs = append(subs, sub)
			}
			sort.Sort(ByName{subs})
			/* page, format and print this shit*/
			msg := fmt.Sprintf("```%3s %-30s %-15s %s\n", "#", "Chaîne", "Ajouté par", "Notifs")
			for idx, sub := range subs {
				msg += fmt.Sprintf("%3d %-30s %-15s %d\n",
					idx+1, sub.Channel.Snippet.Title,
					sub.AddedBy.Username, len(sub.NotifChannelIDs))
			}
			msg += fmt.Sprintf("\nSubscribed to %d channels.```", len(subs))
			s.ChannelMessageSend(cmd.ChannelID, msg)
		}
	}
}

func (p *youtubePlugin) Cleanup() {
	for _, sub := range p.Subs {
		p.client.Unsubscribe(sub.FeedUrl)
	}
	time.Sleep(time.Second * 5)
}
