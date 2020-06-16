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

func (m *Message) parseRemoveReq(input string) error {

	ouf := gjson.Get(input, os.Getenv("OLD_USER_FIELD"))
	oof := gjson.Get(input, os.Getenv("OLD_ORG_FIELD"))
	if !ouf.Exists() || !oof.Exists() {
		return errors.New("missing environment variable")
	}

	var n []string
	m.Usernames = append(n, ouf.Str)
	m.Org = oof.Str
	log.Printf("removing %v from org %v ...", m.Usernames[0], m.Org)

	m.Method = "DELETE"
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

func (m *Message) removeHandler(w http.ResponseWriter, r *http.Request) {

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	input := buf.String()

	err = m.parseRemoveReq(input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

}
