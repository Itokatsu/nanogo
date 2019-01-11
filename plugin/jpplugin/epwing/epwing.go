/* Nanogo Project

 * Load and access EPWING dictionaries
 * Only Works with dictionaries formated by zero-epwing
 */
package epwing

import (
	"encoding/json"
	//"errors"
	"fmt"
	"io"
	//"io/ioutil"
	//"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/itokatsu/nanogo/plugin/jpplugin/jp"
)

var gaijiRe *regexp.Regexp = regexp.MustCompile("[nw][0-9]{5}")

type Dict struct {
	Name    string
	Entries []Entry
	Fonts   []FontCouple
}

/* Dict Entry */
type Entry struct {
	Heading string
	Reading string
	Def     []interface{}
}

func (e Entry) String() string {
	return e.Heading
}

func (e Entry) Details() string {
	return e.Def[0].(string)
}

func (ep Dict) Lookup(query string) (results []jp.DictEntry) {
	for _, e := range ep.Entries {
		if e.Heading == query {
			results = append(results, e)
			continue
		}
		if e.Reading == query {
			results = append(results, e)
		}
	}
	return results
}

func (ep Dict) LookupRe(expr string) (results []jp.DictEntry, err error) {
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	for _, e := range ep.Entries {
		if re.MatchString(e.Heading) {
			results = append(results, e)
			continue
		}
		if re.MatchString(e.Reading) {
			results = append(results, e)
		}
	}
	return results, nil
}

func (ep *Dict) FindGaiji(code string) (Glyph, Font) {
	fmt.Println("searching for", code)
	for _, fonts := range ep.Fonts {
		fmt.Println("searching...")
		var font Font
		if code[0] == 'n' {
			font = fonts.Narrow
		} else {
			font = fonts.Wide
		}
		for _, g := range font.Glyphs {
			num, _ := strconv.Atoi(code[1:])
			if g.Code == num {
				return g, font
			}
		}
	}
	return Glyph{}, Font{}
}

func (ep *Dict) GetGaijiBMP(code string) string {
	g, font := ep.FindGaiji(code)
	bmp := g.Bitmap
	fmt.Printf("%v", bmp)
	res := ""
	for y := 0; y < font.Height; y++ {
		bits := fmt.Sprintf("%b", bmp[y])
		for i := 0; i < font.Width-len(bits); i++ {
			res += " "
		}
		for i := 0; i < len(bits); i++ {
			if rune(bits[i]) == '1' {
				res += "■" //black square
			} else {
				res += " " //space
			}
		}
		res += "\n"
	}
	return res
}

// Load from directory
func Load(dir string) (*Dict, error) {
	d := &Dict{}
	d.Name = filepath.Base(dir)

	/*files, err := filepath.Glob(dir + "/term_bank_*.json")
	fonts := dir + "/fonts.json"
	if err != nil {
		return nil, err
	}
	if len(files) < 1 {
		return nil, errors.New("No file matched")
	}
	for _, f := range files {
		r, _ := os.Open(f)
		err = LoadEntries(r, d)
		if err != nil {
			return nil, err
		}
	}*/

	d.Fonts, _ = LoadFont(dir)
	return d, nil
}

func TypedEntry(i []interface{}) Entry {
	e := Entry{}
	e.Heading = i[0].(string)
	e.Reading = i[1].(string)
	e.Def = i[5].([]interface{})
	return e
}

func LoadEntries(r io.Reader, d *Dict) error {
	var garbage [][]interface{}
	dec := json.NewDecoder(r)
	if err := dec.Decode(&garbage); err != nil {
		fmt.Printf("%v", err)
		return err
	}
	for _, e := range garbage {
		d.Entries = append(d.Entries, TypedEntry(e))
	}
	return nil
}
