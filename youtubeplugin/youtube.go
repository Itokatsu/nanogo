package youtubeplugin

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dpup/gohubbub"
	"github.com/itokatsu/nanogo/parser"
)

//all purpose client
var client = &http.Client{Timeout: 10 * time.Second}

type youtubePlugin struct {
	name   string
	apiKey string
	client *gohubbub.Client
	port   int
	subs   map[string]*Subscription
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
	pInstance.subs = make(map[string]*Subscription)
	go pInstance.client.StartAndServe("", pInstance.port)
	return &pInstance
}

func (p *youtubePlugin) Name() string {
	return "youtube"
}

type Subscription struct {
	Channel        YtChannel
	FeedUrl        string
	AddedAt        time.Time
	AddedBy        *discordgo.User
	NotifChannelID string
	LastNotif      string
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
	} `json:"statistics"`
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

func getJSON(url string, target interface{}) error {
	r, err := client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func (p *youtubePlugin) GetChannelBy(fieldName string, searchTerm string) (YtChannel, error) {
	url := "https://www.googleapis.com/youtube/v3/"
	url += fmt.Sprintf("channels?part=snippet,statistics&%v=%v&key=%v", fieldName, searchTerm, p.apiKey)
	fmt.Println(url)
	result := ChannelListResponse{}
	jsonerr := getJSON(url, &result)
	if jsonerr != nil {
		return YtChannel{}, fmt.Errorf("Failed to parse Youtube API response for %v", searchTerm)
	}
	if len(result.Items) < 1 {
		return YtChannel{}, fmt.Errorf("No channel found for %v", searchTerm)
	}
	return result.Items[0], nil
}

func (p *youtubePlugin) GetVideo(id string) (YtVideo, error) {
	url := "https://www.googleapis.com/youtube/v3/"
	url += fmt.Sprintf("videos?part=snippet&id=%v&key=%v", id, p.apiKey)
	result := VideoListResponse{}
	jsonerr := getJSON(url, &result)
	if jsonerr != nil {
		return YtVideo{}, fmt.Errorf("Failed to parse Youtube API response for videos/%v", id)
	}
	if len(result.Items) < 1 {
		return YtVideo{}, fmt.Errorf("No Video found with id %v", id)
	}
	return result.Items[0], nil
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

		sub.LastNotif = vidID
		// timing info
		timeMsg := fmt.Sprintf("```Entry Published: %v \nEntry Update: %v \nFeed Update: %v```", entry.Published, entry.Updated, feed.Updated)
		msg := fmt.Sprintf("New video from **%s** \n `%s` \n http://www.youtube.com/watch?v=%s",
			entry.Author.Name, entry.Title, entry.VideoId[len("yt:video:"):])
		s.ChannelMessageSend(sub.NotifChannelID, msg)
		s.ChannelMessageSend(sub.NotifChannelID, timeMsg)
	})
	if err != nil {
		return err
	} else {
		p.subs[sub.Channel.Id] = sub
		return nil
	}
}

func (p *youtubePlugin) HandleMsg(cmd *parser.ParsedCmd, s *discordgo.Session, m *discordgo.MessageCreate) {
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
			theChannel := YtChannel{}
			theChannel, err := p.GetChannelBy("forUsername", cmd.Args[1])
			if err != nil {
				theChannel, err = p.GetChannelBy("id", cmd.Args[1])
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, err.Error())
					return
				}
			}
			if _, ok := p.subs[theChannel.Id]; ok {
				s.ChannelMessageSend(m.ChannelID, "Already subscribed")
				return
			}
			//Build Subscription
			var sub Subscription
			sub.Channel = theChannel
			sub.AddedAt, _ = m.Timestamp.Parse()
			sub.AddedBy = m.Author
			if len(cmd.Args) > 2 {
				//@TODO handle channel mention
				sub.NotifChannelID = cmd.Args[2]
			} else {
				sub.NotifChannelID = m.ChannelID
			}
			sub.FeedUrl = "https://www.youtube.com/xml/feeds/"
			sub.FeedUrl += fmt.Sprintf("videos.xml?channel_id=%v", sub.Channel.Id)

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
			if len(cmd.Args) < 2 {
				return
			}
			perm, _ := s.UserChannelPermissions(s.State.User.ID, m.ChannelID)
			admin := (perm&discordgo.PermissionAdministrator == discordgo.PermissionAdministrator)
			for id, sub := range p.subs {
				title := strings.ToLower(sub.Channel.Snippet.Title)
				if id == cmd.Args[1] || title == strings.ToLower(cmd.Args[1]) {
					if m.Author.ID != sub.AddedBy.ID && !admin {
						s.ChannelMessageSend(m.ChannelID, "You cannot do that :nano:")
						return
					}
					msg := fmt.Sprintf("Unsubscribed from %v", title)
					p.client.Unsubscribe(sub.FeedUrl)
					delete(p.subs, id)
					s.ChannelMessageSend(m.ChannelID, msg)
					return
				}
			}

		case "list":
			if len(p.subs) < 1 {
				s.ChannelMessageSend(m.ChannelID, "No subscription found.")
				return
			}
			var subs Subscriptions
			for _, sub := range p.subs {
				subs = append(subs, sub)
			}
			sort.Sort(ByName{subs})
			/* page, format and print this shit*/
			msg := fmt.Sprintf("```%3s %-30s %-15s %s\n", "#", "Chaîne", "Ajouté par", "Notif")
			for idx, sub := range subs {
				notifChannel, _ := s.Channel(sub.NotifChannelID)
				msg += fmt.Sprintf("%3d %-30s %-15s %s\n",
					idx+1, sub.Channel.Snippet.Title,
					sub.AddedBy.Username, fmt.Sprintf("#%s", notifChannel.Name))
			}
			msg += fmt.Sprintf("\nSubscribed to %d channels.```", len(subs))
			s.ChannelMessageSend(m.ChannelID, msg)
		}
	}
}

func (p *youtubePlugin) Help() string {
	return "oupas"
}

func (p *youtubePlugin) SaveTo(file string) {
	buf, err := json.Marshal(p.subs)
	if err != nil {
		fmt.Errorf("Failed to convert youtube/subscriptions struct to json")
		return
	}
	wErr := ioutil.WriteFile(file, buf, 0777)
	if wErr != nil {
		fmt.Errorf("Failed to write to file %s", file)
		return
	}
}

func (p *youtubePlugin) LoadFrom(file string) {

}

func (p *youtubePlugin) Cleanup() {
	p.SaveTo("youtubecfg.json")
	for _, s := range p.subs {
		p.client.Unsubscribe(s.FeedUrl)
	}
	time.Sleep(time.Second * 5)
}
