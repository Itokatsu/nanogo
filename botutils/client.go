package botutils

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"time"
)

var Client = &http.Client{Timeout: 10 * time.Second}

func FetchJSON(url string, target interface{}) error {
	r, err := Client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func FetchXML(url string, target interface{}) error {
	r, err := Client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return xml.NewDecoder(r.Body).Decode(target)
}
