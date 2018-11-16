/* Nanogo Project

 * Load and access EPWING dictionaries
 * Only Works with dictionaries formated by zero-epwing
 */
package epwing

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"fmt"
	"github.com/itokatsu/nanogo/botutils"
)

type Dict struct {
	Name    string
	Entries []Entry
}

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

func (ep *Dict) Lookup(query string) (results []botutils.DictEntry) {
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

func (ep *Dict) LookupRe(expr string) (results []botutils.DictEntry, err error) {
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

// Load from directory
func Load(dir string) (*Dict, error) {
	d := &Dict{}
	d.Name = filepath.Base(dir)

	files, err := filepath.Glob(dir + "/term_bank_*.json")
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
	}
	return d, nil
}

func TypeEntry(i []interface{}) Entry {
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
		d.Entries = append(d.Entries, TypeEntry(e))
	}
	return nil
}
