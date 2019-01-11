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
	Japanese []struct {
		Word    string `json:"word"`
		Reading string `json:"reading"`
	} `json:"japanese"`
	Senses []struct {
		EnglishDefinitions []string `json:"english_definitions"`
		PartsOfSpeech      []string `json:"parts_of_speech"`
		//Links              []interface{} `json:"links"`
		//Tags               []interface{} `json:"tags"`
		//Restrictions       []interface{} `json:"restrictions"`
		//SeeAlso            []interface{} `json:"see_also"`
		//Antonyms           []interface{} `json:"antonyms"`
		//Source             []interface{} `json:"source"`
		//Info               []interface{} `json:"info"`
	} `json:"senses"`
	Attribution struct {
		Jmdict   bool        `json:"jmdict"`
		Jmnedict bool        `json:"jmnedict"`
		Dbpedia  interface{} `json:"dbpedia"`
	} `json:"attribution"`
}

const SEP_JP = "、"
const SEP_EN = "; "
const PLEFT = "【"
const PRIGHT = "】"
const NBSP = "\u202f"

func JishoAPI(query string) ([]ResultEntry, error) {
	// build request
	qs := url.Values{}
	qs.Set("keyword", query)
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
			if ja[idx].Word == "" {
				entry.Header += ja[idx].Reading
				idx++
			} else if ja[idx].Reading == "" {
				entry.Header += ja[idx].Word
				idx++
			} else {
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
					entry.Header = writings[0] + PLEFT
					entry.Header += strings.Join(readings, SEP_JP) + PRIGHT
				}
				if len(readings) == 1 && len(writings) > 1 {
					entry.Header = strings.Join(writings, SEP_JP)
					entry.Header += PLEFT + readings[0] + PRIGHT
				}
			}
		}
		if data.IsCommon {
			entry.Header += "(P)"
		}
		// Senses
		for i, s := range data.Senses {
			def := fmt.Sprintf("%d.%s", i+1, NBSP)
			def += strings.Join(s.EnglishDefinitions, SEP_EN)
			if data.Attribution.Dbpedia != false {
				def += " (WikiDB)"
			}
			entry.Content = append(entry.Content, def)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}
