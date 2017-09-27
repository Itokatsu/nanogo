package botutils

import (
	"encoding/json"
	"net/http"
	"time"
)

var client = &http.Client{Timeout: 10 * time.Second}

func FetchJSON(url string, target interface{}) error {
	r, err := client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}
