package catplugin

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/itokatsu/nanogo/botutils"
)

var baseUrl = "http://thecatapi.com/api/images/get?format=xml"

type catPlugin struct {
	name string
}

func New() (*catPlugin, error) {
	var pInstance catPlugin
	return &pInstance, nil
}

func (p *catPlugin) Name() string {
	return "cat"
}

func (p *catPlugin) HasData() bool {
	return false
}

type Response struct {
	Images []CatImg `xml:"data>images>image"`
}

type CatImg struct {
	Url        string `xml:"url"`
	Id         string `xml:"id"`
	Source_url string `xml:"source_url"`
}

func (p *catPlugin) HandleMsg(cmd *botutils.Cmd, s *discordgo.Session) {
	switch strings.ToLower(cmd.Name) {
	case "cat":
		reqUrl := fmt.Sprintf("%s&results_per_page=5", baseUrl)
		res := Response{}
		botutils.FetchXML(reqUrl, &res)
		for _, img := range res.Images {
			_, err := botutils.Client.Get(img.Url)
			if err != nil {
				continue
			}
			s.ChannelMessageSend(cmd.ChannelID, img.Url)
			return
		}
	}
}

func (p *catPlugin) Help() string {
	return "Get a random cat picture"
}
