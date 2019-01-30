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
const DailySeedURL = "https://alttpr.com/en/daily"

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
	EnemyHP  float64 `json:"enemy_health,omitempty"` //NOT in Standard Mode
	Pot      bool    `json:"pot_shuffle"`
	Palette  bool    `json:"palette_shuffle"`
	Enemy    bool    `json:"enemy,omitempty"` //NOT in standard Mode
}

// Load options from JSON
func (p *alttprPlugin) LoadSettings(data []byte) error {
	return json.Unmarshal(data, &(p.Settings))
}

func (s *SeedOptions) Validate() error {
	if s.Enemizer != false {
		/*if s.Mode == "standard" {
			if s.Enemizer.EnemyHP != "" {
				return fmt.Errorf("Standard Mode incompatible with Enemizer.EnemyHP")
			}
			if s.Enemizer.Enemy != "" {
				return fmt.Errorf("Standard Mode incompatible with Enemizer.Enemy")
			}
		}*/
		if s.Variation == "timer-ohko" || s.Variation == "ohko" {
			return fmt.Errorf("OHKO may not be completable with Enemizer enabled")
		}
	}
	if s.Shuffle != "" {
		if s.Logic != "NoGlitches" {
			return fmt.Errorf("Entrance Shuffle only compatible with No Glitches Logic")
		}
		if s.Mode == "standard" {
			return fmt.Errorf("Entrance Shuffle incompatible with  Standard Mode")
		}
	}

	return nil
}

var fieldsValues = map[string][]string{
	"Logic":      {"NoGlitches", "OverworldGlitches", "MajorGlitches"},
	"Mode":       {"standard", "open", "inverted"},
	"Difficulty": {"easy", "normal", "hard", "expert", "insane"},
	"Swords":     {"uncle", "randomized", "swordless"},
	"Variation":  {"none", "key-sanity", "retro", "timed-race", "timed-ohko", "ohko"},
	"Goal":       {"ganon", "dungeons", "pedestal", "triforce-hunt"},
	"Shuffle":    {"NoShuffle", "simple", "restricted", "full", "crossed", "insanity"},
}

// Enemizer fields
var enemizerFields = map[string][]string{
	"Boss Shuffle":    {"Off", "Simple", "Full", "Chaos"},
	"Enemy Damage":    {"Default", "Shuffled", "Chaos"},
	"Enemy Health":    {"Default", "Easy", "Normal", "Hard", "Brickwall"},
	"Pot Shuffle":     {"Off", "On"},
	"Enemy Shuffle":   {"Off", "On"},
	"Palette Shuffle": {"Off", "On"},
}

var fieldsOrder = []string{"Logic", "Mode", "Difficulty", "Swords", "Variation", "Goal", "Shuffle"}

var fieldsAbbrs = map[string][]string{
	"ng":        {"Logic", "NoGlitches", "NoShuffle"},
	"ow":        {"Logic", "OverworldGlitches"},
	"major":     {"Logic", "MajorGlitches"},
	"std":       {"Mode", "standard"},
	"inv":       {"Mode", "inverted"},
	"key":       {"Variation", "key-sanity"},
	"trace":     {"Variation", "timed-race"},
	"t-race":    {"Variation", "timed-race"},
	"timedrace": {"Variation", "timed-race"},
	"tohko":     {"Variation", "timed-ohko"},
	"t-ohko":    {"Variation", "timed-ohko"},
	"timedohko": {"Variation", "timed-ohko"},
	"triforce":  {"Goal", "triforce-hunt"},
	"ns":        {"Shuffle", "NoShuffle"},
}

func (p *alttprPlugin) Field(fieldName string) string {
	svalues := reflect.ValueOf(p.Settings)
	val := svalues.FieldByName(fieldName)
	if fieldName == "Shuffle" && val.String() == "" {
		return "NoShuffle"
	}
	return val.String()
}

func SearchFields(term string) (k string, v string) {
	res, ok := fieldsAbbrs[term]
	if ok {
		return res[0], res[1]
	}
	for key, values := range fieldsValues {
		for _, v := range values {
			if strings.ToLower(v) == term {
				return key, v
			}
		}
	}
	return "", ""
}

func (p *alttprPlugin) SetField(fieldName string, value string) {
	if value == "NoShuffle" {
		p.Settings.Shuffle = ""
		return
	}
	field := reflect.ValueOf(&(p.Settings)).Elem().FieldByName(fieldName)
	field.SetString(value)
}

func (p *alttprPlugin) GetSettingsEmbed() *discordgo.MessageEmbed {
	embed := botutils.NewEmbed().
		SetTitle("ALTTP:Randomizer - Options").
		SetURL("https://alttpr.com/en/options").
		SetColor(0x006600)

	var text string
	emote := Emotes[botutils.RandInt(len(Emotes))]
	for _, key := range fieldsOrder {
		text += fmt.Sprintf("\n%s`%s:` ", emote, key)
		values := fieldsValues[key]
		for _, v := range values {
			if p.Field(key) == v {
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
		display := false
		// Parse command arguments
		if len(cmd.Args) < 1 {
			generate = true
		} else {
			if strings.ToLower(cmd.Args[0]) == "daily" {
				s.ChannelMessageSend(cmd.ChannelID, DailySeedURL)
				return
			}
			if strings.HasPrefix(strings.ToLower(cmd.Args[0]), "set") {
				display = true
			} else {
				// search for arguments in possible fields
				for _, a := range cmd.Args {
					a = strings.ToLower(a)
					key, value := SearchFields(a)
					if value != "" && value != p.Field(key) {
						p.SetField(key, value)
						display = true
					} else {
						switch a {
						case "generate", "gen", "go":
							generate = true
						case "default", "def", "reset":
							p.LoadSettings(DefaultSettingsJSON)
							display = true
						}
					}
				}
			}
		}

		if generate {
			msgComplex := &discordgo.MessageSend{
				Content: "Generating seed…",
				Embed:   p.GetSettingsEmbed(),
			}
			msg, _ := s.ChannelMessageSendComplex(cmd.ChannelID, msgComplex)
			go p.GenerateSeedEditMsg(s, msg)
			return
		}
		if display {
			msg, _ := s.ChannelMessageSendEmbed(cmd.ChannelID, p.GetSettingsEmbed())
			botutils.AddReactionButtonOnce(s, msg, "\u25B6", func() {
				edit := discordgo.NewMessageEdit(msg.ChannelID, msg.ID)
				s.ChannelMessageEditComplex(edit.SetContent("Generating seed…"))
				go p.GenerateSeedEditMsg(s, msg)
			})
		}
	}
}

type Seed struct {
	Hash string `json:"hash"`
}

func (p *alttprPlugin) GenerateSeedLink() (string, error) {
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

func (p *alttprPlugin) GenerateSeedEditMsg(s *discordgo.Session, msg *discordgo.Message) {
	edit := discordgo.NewMessageEdit(msg.ChannelID, msg.ID)
	link, err := p.GenerateSeedLink()
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
	} else {
		edit.SetContent(link).
			SetEmbed(
				botutils.GetEmbed(msg).
					SetTitle("Your seed is ready !").
					SetColor(0x99FF99).
					SetURL(link).MessageEmbed)
	}
	s.ChannelMessageEditComplex(edit)
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
