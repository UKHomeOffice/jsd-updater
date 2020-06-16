package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/tidwall/gjson"
)

func (m *Message) parseUpdateReq(input string) error {

	rpf := gjson.Get(input, os.Getenv("REPORTER_FIELD"))
	aef := gjson.Get(input, os.Getenv("ALT_EMAIL_FIELD"))
	if !rpf.Exists() || !aef.Exists() {
		return errors.New("missing environment variable")
	}

	var n []string
	m.Usernames = append(n, rpf.Str)
	m.EmailAddr = aef.Str
	log.Printf("updating email address for %v with (%v) ...", m.Usernames[0], m.EmailAddr)

	m.Method = "PUT"

	var u url.URL
	u.Path += "/rest/api/2/user"
	parameters := url.Values{}
	parameters.Add("username", m.Usernames[0])
	u.RawQuery = parameters.Encode()
	m.URI = u.String()

	j := Message{
		EmailAddr: m.EmailAddr,
	}
	p, err := json.Marshal(j)
	if err != nil {
		return err
	}
	m.Payload = p

	return nil
}

func (m *Message) updateHandler(w http.ResponseWriter, r *http.Request) {

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	input := buf.String()

	err = m.parseUpdateReq(input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

}
