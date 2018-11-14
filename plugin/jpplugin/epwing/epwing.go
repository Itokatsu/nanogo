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
)

type Dict struct {
	Name    string
	Entries []Entry
}

type Entry struct {
	Heading string
	Reading string
	Unknown string
	POS     string
	Zero    float64
	Def     []interface{}
}

func (e Entry) String() string {
	return e.Heading
}

func (e Entry) Details() string {
	return e.Def[0].(string)
}

func (ep *Dict) Lookup(query string) (results []Entry) {
	for _, e := range ep.Entries {
		if e.Heading == query {
			results = append(results, e)
			continue
		}
		// reading
		if e.Reading == query {
			results = append(results, e)
		}
	}
	return results
}

func (ep *Dict) LookupRe(expr string) (results []Entry, err error) {
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	for _, e := range ep.Entries {
		// heading
		if re.MatchString(e.Heading) {
			results = append(results, e)
			continue
		}
		// reading
		if re.MatchString(e.Reading) {
			results = append(results, e)
		}
	}
	return results, nil
}

func LoadDir(dir string) (*Dict, error) {
	d := &Dict{}
	d.Name = filepath.Base(dir)
	fmt.Printf(dir)

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
	e.Unknown = i[2].(string)
	e.POS = i[3].(string)
	e.Zero = i[4].(float64)
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

/*
func Load(r io.Reader) (*Dict, error) {
	d := &Dict{}
	dec := xml.NewDecoder(r)
	dec.Entity = Entities
	if err := dec.Decode(d); err != nil {
		return nil, err
	}

	return d, nil
}
*/
