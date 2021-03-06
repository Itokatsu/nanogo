/* Scraping results from weblio.jp */
package web

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/itokatsu/nanogo/botutils"
)

const WeblioURL = "https://www.weblio.jp/content/"

const NumKanji = "一二三四五六七八九十"
const NumBlack = "❶❷❸❹❺❻❼❽❾❿⓫⓬⓭⓮⓯⓰⓱⓲⓳⓴"
const NumWhite = "①②③④⑤⑥⑦⑧⑨⑩⑪⑫⑬⑭⑮⑯⑰⑱⑲⑳"
const NumKana = "㋐㋑㋒㋓㋔"

func Weblio(query string) ([]ResultEntry, error) {
	// Get HTML page
	reqUrl, _ := url.Parse(WeblioURL + query)
	res, err := botutils.Client.Get(reqUrl.String())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, botutils.HttpStatusCodeError(res.StatusCode, reqUrl.String())
	}

	// Load HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	// parse document
	var entries []ResultEntry
	doc.Find(".kiji .NetDicHead").Each(func(i int, s *goquery.Selection) {
		var entry ResultEntry
		entry.Header = s.Text()
		var content string
		text := s.Next().Text()
		offset := ""
		for _, n := range NumKanji {
			from := fmt.Sprintf("  %s", n)
			to := fmt.Sprintf("\n%s%s", offset, n)
			text = strings.Replace(text, from, to, -1)
		}
		offset += "　"
		for _, n := range NumWhite {
			from := fmt.Sprintf("%s", n)
			to := fmt.Sprintf("\n%s%s", offset, n)
			text = strings.Replace(text, from, to, -1)
		}
		for _, n := range NumBlack {
			from := fmt.Sprintf("%s", n)
			to := fmt.Sprintf("\n%s%s", offset, n)
			text = strings.Replace(text, from, to, -1)
		}
		offset += "　"
		for _, n := range NumKana {
			from := fmt.Sprintf("%s", n)
			to := fmt.Sprintf("\n%s%s", offset, n)
			text = strings.Replace(text, from, to, -1)
		}
		content += text + "\n"

		//gros bordel !!
		entry.Content = append(entry.Content, content)
		entries = append(entries, entry)
	})
	return entries, nil
}
