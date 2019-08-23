package v2scar

import (
	"encoding/json"
	"net/http"
)

func getJson(c *http.Client, url string, target interface{}) error {
	r, err := c.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}
