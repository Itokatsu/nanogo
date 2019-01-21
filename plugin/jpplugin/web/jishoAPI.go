/* Getting results from jisho.org API */

package web

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/itokatsu/nanogo/botutils"
)

const JishoAPIEndpoint = "https://jisho.org/api/v1/search/words"

type JishoAPIResult struct {
	Meta struct {
		Status int `json:"status"`
	} `json:"meta"`
	Data []JishoEntry `json:"data"`
}
type JishoEntry struct {
	IsCommon bool `json:"is_common"`
	//Tags     []interface{} `json:"tags"`
	Japanese    []JishoEntryJapanese `json:"japanese"`
	Senses      []JishoEntrySense    `json:"senses"`
	Attribution struct {
		Jmdict   bool        `json:"jmdict"`
		Jmnedict bool        `json:"jmnedict"`
		Dbpedia  interface{} `json:"dbpedia"`
	} `json:"attribution"`
}

type JishoEntryJapanese struct {
	Word    string `json:"word"`
	Reading string `json:"reading"`
}

type JishoEntrySense struct {
	EnglishDefinitions []string `json:"english_definitions"`
	PartsOfSpeech      []string `json:"parts_of_speech"`
	//Links              []string `json:"links"`
	Tags         []string `json:"tags"`
	Restrictions []string `json:"restrictions"`
	//SeeAlso            []interface{} `json:"see_also"`
	//Antonyms           []interface{} `json:"antonyms"`
	//Source             []interface{} `json:"source"`
	Info []string `json:"info"`
}

const SEP_JP = "、"
const SEP_EN = "; "
const PLEFT = "【"
const PRIGHT = "】"
const NBSP = "\u202f"

func JishoAPI(query string, page int) ([]ResultEntry, error) {
	// build request
	qs := url.Values{}
	qs.Set("keyword", query)
	qs.Set("page", fmt.Sprintf("%d", page))
	reqUrl, _ := url.Parse(JishoAPIEndpoint)
	reqUrl.RawQuery = qs.Encode()
	// send request
	var res JishoAPIResult
	err := botutils.FetchJSON(reqUrl.String(), &res)
	if err != nil {
		return nil, err
	}

	// parse results
	var entries []ResultEntry
	for _, data := range res.Data {
		entry := ResultEntry{}
		ja := data.Japanese
		length := len(data.Japanese)

		// make clean header
		idx := 0
		for idx < length {
			//multiple writings
			var writings []string
			oldReading := ja[idx].Reading
			for idx < length && ja[idx].Reading == oldReading {
				writings = append(writings, ja[idx].Word)
				idx++
			}
			idx--
			//multiple reading
			var readings []string
			oldWord := ja[idx].Word
			for idx < length && ja[idx].Word == oldWord {
				readings = append(readings, ja[idx].Reading)
				idx++
			}

			if len(writings) == 1 {
				if writings[0] == "" {
					entry.Header = strings.Join(readings, SEP_JP)
				} else {
					entry.Header = writings[0]
					entry.Header += PLEFT + strings.Join(readings, SEP_JP) + PRIGHT
				}
			} else if len(writings) > 1 && len(readings) == 1 {
				entry.Header = strings.Join(writings, SEP_JP)
				entry.Header += PLEFT + readings[0] + PRIGHT
			}

		}
		entry.Header = entry.Header
		if data.IsCommon {
			entry.Header += "`(P)`"
		}

		// Content
		var prevTags string
		for i, s := range data.Senses {
			var dbpedia = (data.Attribution.Dbpedia != false)
			var def string
			// Tags
			if len(s.PartsOfSpeech)+len(s.Tags) > 0 {
				var tagsStr string
				for _, pos := range s.PartsOfSpeech {
					if a, ok := reverseTags[pos]; ok {
						tagsStr += fmt.Sprintf("**`%s`** ", a)
					}
				}
				for _, t := range s.Tags {
					if a, ok := reverseTags[t]; ok {
						tagsStr += fmt.Sprintf("**`%s`** ", a)
					}
				}
				if len(tagsStr) > 0 && tagsStr != prevTags {
					def += fmt.Sprintf("%s%s\n", offset, tagsStr)
					prevTags = tagsStr
				}
			}
			// Definitions
			def += fmt.Sprintf("%s%d.%s", offset, i+1, NBSP)
			def += strings.Join(s.EnglishDefinitions, SEP_EN)
			// More Info
			if len(s.Restrictions) > 0 {
				def += fmt.Sprintf(" (%s only)", strings.Join(s.Restrictions, ", "))
			}
			if dbpedia {
				def += " `WikiDB`"
			}
			if len(s.Info) > 0 {
				def += fmt.Sprintf("\n%s| %s", offset+offset, strings.Join(s.Info, SEP_EN))
			}
			entry.Content = append(entry.Content, def)
		}

		// Done
		entries = append(entries, entry)
	}
	return entries, nil
}

var JishoTags = map[string]string{
	"verb":      "Verb of any type",
	"adjective": "Adjective of any type",
	"counter":   "Counter words",
	"abbr":      "Abbreviation",
	"adj":       "Adjective",
	"adj-f":     "Noun or verb acting prenominally",
	"adj-i":     "I-adjective",
	"adj-ix":    "I-adjective (yoi/ii class)",
	"adj-kari":  "Kari adjective (archaic)",
	"adj-ku":    "Ku-adjective (archaic)",
	"adj-na":    "Na-adjective",
	"adj-nari":  "Archaic/formal form of na-adjective",
	"adj-no":    "No-adjective",
	"adj-pn":    "Pre-noun adjectival",
	"adj-shiku": "Shiku adjective (archaic)",
	"adj-t":     "Taru-adjective",
	"adv":       "Adverb",
	"adv-to":    "Adverb taking the 'to' particle",
	"anat":      "Anatomical term",
	"ant":       "Antonym",
	"arch":      "Archaism",
	"archit":    "Architecture term",
	"astron":    "Astronomy, etc. term",
	"ateji":     "Ateji (phonetic) reading",
	"aux":       "Auxiliary",
	"aux-adj":   "Auxiliary adjective",
	"aux-v":     "Auxiliary verb",
	"baseb":     "Baseball term",
	"biol":      "Biology term",
	"bot":       "Botany term",
	"bus":       "Business term",
	"chem":      "Chemistry term",
	"chn":       "Children's language",
	"col":       "Colloquialism",
	"comp":      "Computer terminology",
	"conj":      "Conjunction",
	"cop-da":    "Copula",
	"ctr":       "Counter",
	"derog":     "Derogatory",
	"eK":        "Exclusively kanji",
	"econ":      "Economics term",
	"ek":        "Exclusively kana",
	"engr":      "Engineering term",
	"equ":       "Equivalent",
	"ex":        "Usage example",
	"exp":       "Expression",
	"expl":      "Explanatory",
	"fam":       "Familiar language",
	"fem":       "Female term or language",
	"fig":       "Figuratively",
	"finc":      "Finance term",
	"food":      "Food term",
	"geol":      "Geology, etc. term",
	"geom":      "Geometry term",
	"gikun":     "Gikun (meaning as reading) or jukujikun (special kanji reading)",
	"go":        "On reading",
	"hob":       "Hokkaido dialect",
	"hon":       "Honorific or respectful (sonkeigo)",
	"hum":       "Humble (kenjougo)",
	"iK":        "Irregular kanji usage",
	"id":        "Idiomatic expression",
	"ik":        "Irregular kana usage",
	"int":       "Interjection",
	"io":        "Irregular okurigana usage",
	"iv":        "Irregular verb",
	"jlpt-n1":   "JLPT N1",
	"jlpt-n2":   "JLPT N2",
	"jlpt-n3":   "JLPT N3",
	"jlpt-n4":   "JLPT N4",
	"jlpt-n5":   "JLPT N5",
	"joc":       "Jocular, humorous term",
	"jouyou":    "Approved reading for jouyou kanji",
	"kan":       "On reading",
	"kanyou":    "On reading",
	"ksb":       "Kansai dialect",
	"ktb":       "Kantou dialect",
	"kun":       "Kun reading",
	"kvar":      "Kanji variant",
	"kyb":       "Kyoto dialect",
	"kyu":       "Kyuushuu dialect",
	"law":       "Law, etc. term",
	"ling":      "linguistics terminology",
	"lit":       "Literaly",
	"m-sl":      "Manga slang",
	"mahj":      "Mahjong term",
	"male":      "Male term or language",
	"male-sl":   "Male slang",
	"math":      "Mathematics",
	"med":       "Medicine, etc. term",
	"mil":       "Military",
	"music":     "Music term",
	"n":         "Noun",
	"n-adv":     "Adverbial noun",
	"n-pr":      "Proper noun",
	"n-pref":    "Noun - used as a prefix",
	"n-suf":     "Noun - used as a suffix",
	"n-t":       "Temporal noun",
	"nab":       "Nagano dialect",
	"name":      "Name reading (nanori)",
	"num":       "Numeric",
	"oK":        "Out-dated kanji",
	"obs":       "Obsolete term",
	"obsc":      "Obscure term",
	"oik":       "Old or irregular kana form",
	"ok":        "Out-dated or obsolete kana usage",
	"on":        "On reading",
	"on-mim":    "Onomatopoeic or mimetic word",
	"osb":       "Osaka dialect",
	"physics":   "Physics terminology",
	"pn":        "Pronoun",
	"poet":      "Poetical term",
	"pol":       "Polite (teineigo)",
	"pref":      "Prefix",
	"proverb":   "Proverb",
	"prt":       "Particle",
	"rad":       "Reading used as name of radical",
	"rare":      "Rare",
	"rkb":       "Ryuukyuu dialect",
	"see":       "See also",
	"sens":      "Sensitive",
	"shogi":     "Shogi term",
	"sl":        "Slang",
	"sports":    "Sports term",
	"suf":       "Suffix",
	"sumo":      "Sumo term",
	"syn":       "Synonym",
	"thb":       "Touhoku dialect",
	"tou":       "On reading",
	"tsb":       "Tosa dialect",
	"tsug":      "Tsugaru dialect",
	"uK":        "Usually written using kanji alone",
	"uk":        "Usually written using kana alone",
	"unc":       "Unclassified",
	"v-unspec":  "Verb unspecified",
	"v1":        "Ichidan verb",
	"v1-s":      "Ichidan verb (kureru special class)",
	"v2a-s":     "Nidan verb with u ending (archaic)",
	"v2b-k":     "Nidan verb (upper class) with bu ending (archaic)",
	"v2b-s":     "Nidan verb (lower class) with bu ending (archaic)",
	"v2d-k":     "Nidan verb (upper class) with dzu ending (archaic)",
	"v2d-s":     "Nidan verb (lower class) with dzu ending (archaic)",
	"v2g-k":     "Nidan verb (upper class) with gu ending (archaic)",
	"v2g-s":     "Nidan verb (lower class) with gu ending (archaic)",
	"v2h-k":     "Nidan verb (upper class) with hu/fu ending (archaic)",
	"v2h-s":     "Nidan verb (lower class) with hu/fu ending (archaic)",
	"v2k-k":     "Nidan verb (upper class) with ku ending (archaic)",
	"v2k-s":     "Nidan verb (lower class) with ku ending (archaic)",
	"v2m-k":     "Nidan verb (upper class) with mu ending (archaic)",
	"v2m-s":     "Nidan verb (lower class) with mu ending (archaic)",
	"v2n-s":     "Nidan verb (lower class) with nu ending (archaic)",
	"v2r-k":     "Nidan verb (upper class) with ru ending (archaic)",
	"v2r-s":     "Nidan verb (lower class) with ru ending (archaic)",
	"v2s-s":     "Nidan verb (lower class) with su ending (archaic)",
	"v2t-k":     "Nidan verb (upper class) with tsu ending (archaic)",
	"v2t-s":     "Nidan verb (lower class) with tsu ending (archaic)",
	"v2w-s":     "Nidan verb (lower class) with u ending and we conjugation (archaic)",
	"v2y-k":     "Nidan verb (upper class) with yu ending (archaic)",
	"v2y-s":     "Nidan verb (lower class) with yu ending (archaic)",
	"v2z-s":     "Nidan verb (lower class) with zu ending (archaic)",
	"v4b":       "Yodan verb with bu ending (archaic)",
	"v4g":       "Yodan verb with gu ending (archaic)",
	"v4h":       "Yodan verb with hu/fu ending (archaic)",
	"v4k":       "Yodan verb with ku ending (archaic)",
	"v4m":       "Yodan verb with mu ending (archaic)",
	"v4n":       "Yodan verb with nu ending (archaic)",
	"v4r":       "Yodan verb with ru ending (archaic)",
	"v4s":       "Yodan verb with su ending (archaic)",
	"v4t":       "Yodan verb with tsu ending (archaic)",
	"v5aru":     "Godan verb - aru special class",
	"v5b":       "Godan verb with bu ending",
	"v5g":       "Godan verb with gu ending",
	"v5k":       "Godan verb with ku ending",
	"v5k-s":     "Godan verb - Iku/Yuku special class",
	"v5m":       "Godan verb with mu ending",
	"v5n":       "Godan verb with nu ending",
	"v5r":       "Godan verb with ru ending",
	"v5r-i":     "Godan verb with ru ending (irregular verb)",
	"v5s":       "Godan verb with su ending",
	"v5t":       "Godan verb with tsu ending",
	"v5u":       "Godan verb with u ending",
	"v5u-s":     "Godan verb with u ending (special class)",
	"v5uru":     "Godan verb - Uru old class verb (old form of Eru)",
	"v5z":       "Godan verb with zu ending",
	"vi":        "intransitive verb",
	"vk":        "Kuru verb - special class",
	"vn":        "Irregular nu verb",
	"vr":        "Irregular ru verb, plain form ends with -ri",
	"vs":        "Suru verb",
	"vs-c":      "Su verb - precursor to the modern suru",
	"vs-i":      "Suru verb - irregular",
	"vs-s":      "Suru verb - special class",
	"vt":        "Transitive verb",
	"vulg":      "Vulgar",
	"vz":        "Ichidan verb - zuru verb (alternative form of -jiru verbs)",
	"yoji":      "Yojijukugo (four character compound)",
	"zool":      "Zoology term",
	"Buddh":     "Buddhist term",
	"MA":        "Martial arts term",
	"Shinto":    "Shinto term",
	"X":         "Rude or X-rated term",
	"wasei":     "Wasei, word made in Japan",
}

var reverseTags = reverseMap(JishoTags)

func reverseMap(m map[string]string) map[string]string {
	reversed := make(map[string]string)
	for k, v := range m {
		reversed[v] = k
	}
	return reversed
}
