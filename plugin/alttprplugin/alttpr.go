package alttprplugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/itokatsu/nanogo/botutils"
)

const ItemURL = "https://alttpr.com/seed"
const EntranceURL = "https://alttpr.com/entrance/seed"
const SeedDownloadURL = "https://alttpr.com/en/h/"
const ManualGenerationURL = "https://alttpr.com/en/randomizer"

type alttprPlugin struct {
	Settings SeedOptions
	Cookie   string
}

type Config struct {
	Cookie string `json:"cookie"`
}

func New(cfg Config) (*alttprPlugin, error) {
	var p alttprPlugin
	p.Cookie = cfg.Cookie
	p.LoadSettings(DefaultSettingsJSON)
	return &p, nil
}

var Emotes = []string{
	"<:fairy:290795058284855307>",
	"<:bird:290795749254496257>",
	"<:ganon:290792210692046848>",
	"<:ded:290790229454094337>",
}

var DefaultSettingsJSON = []byte(`{
	"logic":"NoGlitches",
	"difficulty":"normal",
	"variation":"none",
	"mode":"open",
	"goal":"ganon",
	"weapons":"randomized",
	"tournament":false,
	"spoilers":false,
	"enemizer":false,
	"lang":"en" 
}`)

var fields = map[string][]string{
	"Logic":      {"NoGlitches", "OverworldGlitches", "MajorGlitches"},
	"Mode":       {"standard", "open", "inverted"},
	"Difficulty": {"easy", "normal", "hard", "expert", "insane", "crowdControl"},
	"Swords":     {"uncle", "randomized", "swordless"},
	"Variation":  {"none", "key-sanity", "retro", "timed-race", "timed-ohko", "ohko"},
	"Goal":       {"ganon", "dungeons", "pedestal", "triforce-hunt"},
	"Shuffle":    {"NoShuffle", "simple", "restricted", "full", "crossed", "insanity"},
}

var fieldOrder = []string{"Logic", "Mode", "Difficulty", "Swords", "Variation", "Goal", "Shuffle"}

type SeedOptions struct {
	Logic      string      `json:"logic"`
	Difficulty string      `json:"difficulty"`
	Variation  string      `json:"variation"`
	Mode       string      `json:"mode"`
	Goal       string      `json:"goal"`
	Shuffle    string      `json:"shuffle,omitempty"`
	Swords     string      `json:"weapons,omitempty"`
	Tournament bool        `json:"tournament"`
	Spoilers   bool        `json:"spoilers"`
	Enemizer   interface{} `json:"enemizer"` // False or EnemizerOptions
	Lang       string      `json:"en"`
}

type EnemizerOptions struct {
	Boss     string  `json:"bosses"`
	EnemyDmg string  `json:"enemy_damage"`
	EnemyHP  float64 `json:"enemy_health"` //NOT in Standard Mode
	Pot      bool    `json:"pot_shuffle"`
	Palette  bool    `json:"palette_shuffle"`
	Enemy    bool    `json:"enemy"` //NOT in standard Mode
}

func (s SeedOptions) Field(fieldName string) string {
	svalues := reflect.ValueOf(s)
	val := svalues.FieldByName(fieldName)
	if fieldName == "Shuffle" && val.String() == "" {
		return "NoShuffle"
	}
	return val.String()
}

func (s SeedOptions) SetField(fieldName string, value string) {
	if value == "NoShuffle" {
		s.Shuffle = ""
		return
	}
	reflect.ValueOf(&s).
		Elem().FieldByName(fieldName).SetString(value)
}

func SearchFields(term string) (key, value string) {
	for key, values := range fields {
		for _, v := range values {
			if strings.ToLower(v) == term {
				return key, v
			}
		}
	}
	return "", ""
}

// Load settings from JSON
func (p *alttprPlugin) LoadSettings(data []byte) error {
	return json.Unmarshal(data, &(p.Settings))
}

func (p *alttprPlugin) GetSettingsEmbed() *discordgo.MessageEmbed {
	embed := botutils.NewEmbed().
		SetTitle("ALTTP:Randomizer - Options").
		SetURL("https://alttpr.com/en/options").
		SetColor(0x006600)

	var text string
	emote := Emotes[botutils.RandInt(len(Emotes))]
	for _, key := range fieldOrder {
		text += fmt.Sprintf("\n%s`%s:` ", emote, key)
		values := fields[key]
		for _, v := range values {
			if p.Settings.Field(key) == v {
				text += fmt.Sprintf("__**%s**__ ", strings.Title(v))
			} else {
				text += strings.Title(v) + " "
			}
		}
	}
	text = strings.Replace(text, "Ohko", "OHKO", -1)
	return embed.SetDescription(text).MessageEmbed
}

func (p *alttprPlugin) HandleMsg(cmd *botutils.Cmd, s *discordgo.Session) {
	switch strings.ToLower(cmd.Name) {
	case "alttp", "lttp", "zeldo", "z":
		generate := false
		updated := false
		// Parse command arguments
		if len(cmd.Args) < 1 {
			generate = true
		} else {
			if strings.HasPrefix(strings.ToLower(cmd.Args[0]), "set") {
				s.ChannelMessageSendEmbed(cmd.ChannelID, p.GetSettingsEmbed())
				return
			}
			// search for arguments in possible fields
			for _, a := range cmd.Args {
				a = strings.ToLower(a)
				//Search fields
				key, value := SearchFields(a)
				if value != "" && value != p.Settings.Field(key) {
					p.Settings.SetField(key, value)
					updated = true
				}
				if a == "go" || a == "gen" {
					generate = true
				}
			}
		}

		if generate {
			msgComplex := &discordgo.MessageSend{
				Content: "Generating seed…",
				Embed:   p.GetSettingsEmbed(),
			}
			msg, _ := s.ChannelMessageSendComplex(cmd.ChannelID, msgComplex)

			go func(msg *discordgo.Message) {
				edit := discordgo.NewMessageEdit(msg.ChannelID, msg.ID)
				resp, err := p.GenerateSeed()
				if err != nil {
					s.MessageReactionAdd(msg.ChannelID, msg.ID, botutils.ErrorEmote)
					edit.SetContent("Something went terribly wrong…").
						SetEmbed(
							botutils.NewEmbed().
								SetDescription(err.Error()).
								SetColor(0xFF0000).
								SetURL(ManualGenerationURL).
								SetFooter("Click the URL and try to generate it yourself").
								MessageEmbed)
					return
				}

				edit.SetContent(resp).
					SetEmbed(
						botutils.GetEmbed(msg).
							SetTitle("Your seed is ready !").
							SetColor(0x99FF99).
							SetURL(resp).MessageEmbed)
				s.ChannelMessageEditComplex(edit)
			}(msg) //end of goroutine
			return
		}
		if updated {
			s.ChannelMessageSendEmbed(cmd.ChannelID, p.GetSettingsEmbed())
		}
	}
}

type Seed struct {
	Hash string `json:"hash"`
}

func (p *alttprPlugin) GenerateSeed() (string, error) {
	// @TODO Validate Settings
	data, err := json.Marshal(p.Settings)
	if err != nil {
		return "", err
	}
	// Build Request
	reqUrl := ItemURL
	if p.Settings.Shuffle != "" {
		reqUrl = EntranceURL
	}
	req, err := http.NewRequest("POST", reqUrl, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	req.Header.Set("Cookie", p.Cookie)
	// Send Request
	resp, err := botutils.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	// Read Response
	body, _ := ioutil.ReadAll(resp.Body)
	var seed Seed
	err = json.Unmarshal(body, &seed)
	if err != nil {
		return "", err
	}
	link := SeedDownloadURL + seed.Hash
	return link, nil
}

func (p *alttprPlugin) Name() string {
	return "alttpr"
}

func (p *alttprPlugin) HasData() bool {
	return true
}

func (p *alttprPlugin) Help() string {
	return ""
}
