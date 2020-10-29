package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/tidwall/gjson"
)

func (m *Message) parseAddReq(input string) error {

	nuf := gjson.Get(input, os.Getenv("NEW_USER_FIELD"))
	nof := gjson.Get(input, os.Getenv("NEW_ORG_FIELD"))
	if !nuf.Exists() || !nof.Exists() {
		return errors.New("missing environment variable")
	}

	var n []string
	m.Usernames = append(n, nuf.Str)
	m.Org = nof.Str
	log.Printf("adding %v to org %v ...", m.Usernames[0], m.Org)

	m.Method = "POST"
	m.URI = fmt.Sprintf("/rest/servicedeskapi/organization/%s/user", m.Org)

	j := Message{
		Usernames: m.Usernames,
	}
	p, err := json.Marshal(j)
	if err != nil {
		return err
	}
	m.Payload = p

	return nil
}

func (m *Message) addHandler(w http.ResponseWriter, r *http.Request) {

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	input := buf.String()

	err = m.parseAddReq(input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

}