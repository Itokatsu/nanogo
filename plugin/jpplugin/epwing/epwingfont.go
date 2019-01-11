/* Nanogo Project

 * Load and access EPWING dictionaries
 * Gaiji bitmaps from font files provided by zero-epwing
 */

package epwing

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type FontFile struct {
	CharCode string `json:"charCode"`
	DiscCode string `json:"discCode"`
	Subbooks []struct {
		Title     string       `json:"title"`
		Copyright string       `json:"copyright"`
		Fonts     []FontCouple `json:"fonts"`
	} `json:"subbooks"`
}

type FontCouple struct {
	Narrow Font `json:"narrow"`
	Wide   Font `json:"wide"`
}

type Font struct {
	Glyphs []Glyph `json:"glyphs"`
	Width  int     `json:"width"`
	Height int     `json:"height"`
}

type Glyph struct {
	Bitmap []int `json:"bitmap"`
	Code   int   `json:"code"`
}

func LoadFont(dir string) ([]FontCouple, error) {
	fonts := dir + "/fonts.json"
	content, _ := ioutil.ReadFile(fonts)

	var ff FontFile
	err := json.Unmarshal(content, &ff)
	if err != nil {
		fmt.Println("Error loading data", err)
		return nil, err
	}
	return ff.Subbooks[0].Fonts, nil
}
