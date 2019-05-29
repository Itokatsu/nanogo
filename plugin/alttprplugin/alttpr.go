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

const ItemURL = "https://alttpr.com/seed"                      // Item Randomizer URL
const EntranceURL = "https://alttpr.com/entrance/seed"         // Entrance Randomizer URL
const SeedDownloadURL = "https://alttpr.com/en/h/"             // Download Page
const ManualGenerationURL = "https://alttpr.com/en/randomizer" // Generation Site URL
const DailySeedURL = "https://alttpr.com/en/daily"             // Daily Run

type alttprPlugin struct {
	Settings *SeedOptions
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

var fieldsValues = map[string][]string{
	"Logic":      {"NoGlitches", "OverworldGlitches", "MajorGlitches"},
	"Mode":       {"standard", "open", "inverted"},
	"Difficulty": {"easy", "normal", "hard", "expert", "insane"},
	"Swords":     {"uncle", "randomized", "swordless"},
	"Variation":  {"none", "key-sanity", "retro", "timed-race", "timed-ohko", "ohko"},
	"Goal":       {"ganon", "dungeons", "pedestal", "triforce-hunt"},
	"Shuffle":    {"NoShuffle", "simple", "restricted", "full", "crossed", "insanity"},
}

var enemizerFields = map[string][]string{
	"Boss Shuffle":    {"Off", "Simple", "Full", "Chaos"},
	"Enemy Damage":    {"Default", "Shuffled", "Chaos"},
	"Enemy Health":    {"Default", "Easy", "Normal", "Hard", "Brickwall"},
	"Pot Shuffle":     {"Off", "On"},
	"Enemy Shuffle":   {"Off", "On"},
	"Palette Shuffle": {"Off", "On"},
}

var fieldsOrder = []string{"Logic", "Mode", "Difficulty", "Swords", "Variation", "Goal", "Shuffle"}

//Search field from *lowercased* value
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

// Get a Field value from variable name
func (op *SeedOptions) Field(fieldName string) string {
	svalues := reflect.ValueOf(op).Elem()
	val := svalues.FieldByName(fieldName)
	if fieldName == "Shuffle" && val.String() == "" {
		return "NoShuffle"
	}
	return val.String()
}

// Set a field to value
func (op *SeedOptions) SetField(fieldName string, value string) {
	if value == "NoShuffle" {
		op.Shuffle = ""
		return
	}
	field := reflect.ValueOf(op).Elem().FieldByName(fieldName)
	field.SetString(value)
}

// Emotes used in embed
var Emotes = []string{
	"<:fairy:290795058284855307>",
	"<:bird:290795749254496257>",
	"<:ganon:290792210692046848>",
	"<:ded:290790229454094337>",
}

// Generate Embed from Settings
func (op *SeedOptions) Embed() *botutils.Embed {
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
			if op.Field(key) == v {
				text += fmt.Sprintf("__**%s**__ ", strings.Title(v))
			} else {
				text += strings.Title(v) + " "
			}
		}
	}
	text = strings.Replace(text, "Ohko", "OHKO", -1)
	return embed.SetDescription(text)
}

func (p *alttprPlugin) HandleMsg(cmd *botutils.Cmd, s *discordgo.Session) {
	switch strings.ToLower(cmd.Name) {
	case "alttp", "lttp", "zeldo", "z":
		generate := false
		randomized := false
		display := false
		options := p.Settings
        randWeights := myWeights
        fixedOptions := make(map[string]string)

		// Parse command arguments
		if len(cmd.Args) < 1 {
			generate = true
		} else {
			if strings.ToLower(cmd.Args[0]) == "daily" {
				s.ChannelMessageSend(cmd.ChannelID, DailySeedURL)
				return
			}
			if strings.HasPrefix(strings.ToLower(cmd.Args[0]), "settings") {
				display = true
			} else {
				// search for arguments in possible fields
				for _, a := range cmd.Args {
					a = strings.ToLower(a)
					key, value := SearchFields(a)
					if value != "" {
                        fixedOptions[key] = value
                        if value != options.Field(key) {
                            // change value
    						options.SetField(key, value)
						    display = true
                        }
                    } else {
						switch a {
                        // full randomize
                        case "fullrand", "??":
							randWeights = fullRandomWeights
							display = true
							randomized = true
                        // classic weighted randomize
						case "rand", "rando", "?":
							display = true
							randomized = true
						case "generate", "gen", "go":
							generate = true
                        // reset settings
						case "default", "def", "reset":
							p.LoadSettings(DefaultSettingsJSON)
							display = true
						}
					}
				}
			}
		}
        if randomized {
            options = GenerateRandOptions(randWeights, fixedOptions)
        }
		if !randomized && !generate {
            // Save changes
			p.Settings = options
		}
		// Seed Generation
		if generate {
			msgComplex := &discordgo.MessageSend{
				Content: "Generating seed…",
				Embed:   options.Embed().MessageEmbed,
			}
			msg, _ := s.ChannelMessageSendComplex(cmd.ChannelID, msgComplex)
			go p.GenerateSeedEditMsg(s, msg, options)
			return
		}
		// Display Settings
		if display {
			msg, _ := s.ChannelMessageSendEmbed(cmd.ChannelID, options.Embed().MessageEmbed)
			// Randomize Button
			botutils.AddReactionButton(s, msg, "🔁", func() {
				edit := discordgo.NewMessageEdit(msg.ChannelID, msg.ID)
				options = GenerateRandOptions(randWeights, fixedOptions)
				s.ChannelMessageEditComplex(edit.SetEmbed(options.Embed().MessageEmbed))
			})

			// Generate Button
			botutils.AddReactionButtonOnce(s, msg, "\u25B6", func() {
				edit := discordgo.NewMessageEdit(msg.ChannelID, msg.ID)
				s.ChannelMessageEditComplex(edit.SetContent("Generating seed…"))
				go p.GenerateSeedEditMsg(s, msg, options)
				botutils.RemoveButtonAll(s, msg.ID)
			})
		}
	}
}

type Seed struct {
	Hash string `json:"hash"`
}


func (p *alttprPlugin) GenerateSeedLink(options *SeedOptions) (string, error) {
	if options == nil {
		options = p.Settings
	}
	// @TODO Validate Settings
	data, err := json.Marshal(options)
	if err != nil {
		return "", err
	}
	// Build Request
	reqUrl := ItemURL
	if options.Shuffle != "" {
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
    fmt.Println(body)
	var seed Seed
	err = json.Unmarshal(body, &seed)
	if err != nil {
		return "", err
	}
	link := SeedDownloadURL + seed.Hash
	return link, nil
}

func (p *alttprPlugin) GenerateSeedEditMsg(s *discordgo.Session, msg *discordgo.Message, options *SeedOptions) {
	edit := discordgo.NewMessageEdit(msg.ChannelID, msg.ID)
	link, err := p.GenerateSeedLink(options)
	if err != nil {
		s.MessageReactionAdd(msg.ChannelID, msg.ID, botutils.ErrorEmote)
		edit.SetContent("Something went terribly wrong…").
			SetEmbed(
				botutils.NewEmbed().
					SetDescription(err.Error()).
					SetColor(0xFF0000).
					SetTitle("ALLTP: Randomizer").
					SetURL(ManualGenerationURL).
					SetFooter("Click the URL and try to generate it yourself").
					MessageEmbed)
	} else {
		edit.SetContent(link).
			SetEmbed(
				options.Embed().
					SetTitle("Your seed is ready !").
					SetColor(0x99FF99).
					SetURL(link).MessageEmbed)
	}
	s.ChannelMessageEditComplex(edit)
}

// Randomize Options
var myWeights = [][]int{
	{80, 20, 0},          // {"NoGlitches", "OverworldGlitches", "MajorGlitches"},
	{5, 100, 10},         // {"standard", "open", "inverted"},
	{5, 100, 25, 5, 0},   // {"easy", "normal", "hard", "expert", "insane"},
	{20, 80, 0},          // {"uncle", "randomized", "swordless"},
	{100, 5, 0, 1, 1, 0}, // {"none", "key-sanity", "retro", "timed-race", "timed-ohko", "ohko"},
	{50, 20, 20, 5},      // {"ganon", "dungeons", "pedestal", "triforce-hunt"},
	{100, 2, 2, 2, 2, 2}, // {"NoShuffle", "simple", "restricted", "full", "crossed", "insanity"},
}
var fullRandomWeights = [][]int{
	{1, 1, 1},
	{1, 1, 1},
	{1, 1, 1, 1, 1},
	{1, 1, 1},
	{1, 1, 1, 1, 1, 1},
	{1, 1, 1, 1},
	{1, 1, 1, 1, 1, 1},
}

func GenerateRandOptions(weights [][]int, fixOptions map[string]string) *SeedOptions {
	var options SeedOptions
	for idx, field := range fieldsOrder {
        if fixValue, ok := fixOptions[field]; ok {
            options.SetField(field, fixValue)
        } else {
            valueIdx := botutils.RandWeights(weights[idx])
            options.SetField(field, fieldsValues[field][valueIdx])
        }
	}
	options.Enemizer = false
	return &options
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

// Abbreviations for fields
var fieldsAbbrs = map[string][]string{
	"ng":        {"Logic", "NoGlitches", "NoShuffle"},
	"ow":        {"Logic", "OverworldGlitches"},
	"major":     {"Logic", "MajorGlitches"},
	"nl":        {"Logic", "NoLogic"},
	"std":       {"Mode", "standard"},
	"inv":       {"Mode", "inverted"},
	"key":       {"Variation", "key-sanity"},
	"ks":        {"Variation", "key-sanity"},
	"trace":     {"Variation", "timed-race"},
	"t-race":    {"Variation", "timed-race"},
	"timedrace": {"Variation", "timed-race"},
	"tohko":     {"Variation", "timed-ohko"},
	"t-ohko":    {"Variation", "timed-ohko"},
	"timedohko": {"Variation", "timed-ohko"},
    "dun":  {"Goal", "dungeons"},
    "ped":  {"Goal", "pedestal"},
	"triforce":  {"Goal", "triforce-hunt"},
	"tri":       {"Goal", "triforce-hunt"},
	"hunt":      {"Goal", "triforce-hunt"},
	"ns":        {"Shuffle", "NoShuffle"},
}
