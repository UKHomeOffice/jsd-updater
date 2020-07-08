package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/tidwall/gjson"
)

// setEnv sets some test envars
func setEnv() {
	os.Setenv("ADMIN_USER", "foo")
	os.Setenv("ADMIN_PASS", "bar")
	os.Setenv("REPORTER_FIELD", "issue.fields.reporter.name")
	os.Setenv("ALT_EMAIL_FIELD", "issue.fields.customfield_10147")
	os.Setenv("NEW_ORG_FIELD", "issue.fields.customfield_10002.0.id")
	os.Setenv("NEW_USER_FIELD", "issue.fields.customfield_10600.name")
	os.Setenv("OLD_ORG_FIELD", "issue.fields.customfield_10002.0.id")
	os.Setenv("OLD_USER_FIELD", "issue.fields.customfield_11100.name")
}

//  getMsg gets some test input
func getMsg(p int) (string, error) {

	body, err := ioutil.ReadFile("payloads.json")
	if err != nil {
		return "", err
	}

	path := fmt.Sprintf("cases.%v", p)
	res := gjson.GetManyBytes(body, path)

	return res[0].Raw, nil
}

func TestHandlers(t *testing.T) {

	var names []string

	tt := []struct {
		name      string
		handler   string
		input     int
		usernames []string
		email     string
		org       string
		method    string
		path      string
		err       string
	}{
		{name: "missingUpdate", handler: "update", input: 1, err: "missing environment variable"},
		{name: "missingAdd", handler: "add", input: 2, err: "missing environment variable"},
		{name: "missingRemove", handler: "remove", input: 3, err: "missing environment variable"},
		{name: "updateEmail", handler: "update", input: 4, path: "/rest/api/2/user?username=foo%40bar.com", method: "PUT", usernames: append(names, "foo@bar.com"), email: "foo@bat.com"},
		{name: "addUser", handler: "add", input: 5, path: "/rest/servicedeskapi/organization/7/user", method: "POST", usernames: append(names, "alice"), org: "7"},
		{name: "removeUser", handler: "remove", input: 6, path: "/rest/servicedeskapi/organization/8/user", method: "DELETE", usernames: append(names, "bob"), org: "8"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			// init env & object
			setEnv()
			var a Message

			// create inbound payload
			m, err := getMsg(tc.input)
			if err != nil {
				t.Fatalf("could not get message: %v", err)
			}
			rawM := json.RawMessage(m)
			p, err := json.Marshal(rawM)
			if err != nil {
				t.Fatalf("could not make incoming payload: %v", err)
			}
			pld := bytes.NewReader(p)

			// create inbound request
			r, err := http.NewRequest("POST", "/", pld)
			if err != nil {
				t.Fatalf("could not make incoming request: %v", err)
			}

			// create response recorder
			rr := httptest.NewRecorder()

			// select relevant handler
			switch {
			case tc.handler == "update":
				handler := http.HandlerFunc(a.updateHandler)
				handler.ServeHTTP(rr, r)
			case tc.handler == "add":
				handler := http.HandlerFunc(a.addHandler)
				handler.ServeHTTP(rr, r)
			case tc.handler == "remove":
				handler := http.HandlerFunc(a.removeHandler)
				handler.ServeHTTP(rr, r)
			}

			res := rr.Result()
			defer res.Body.Close()

			b, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("could not read response: %v", err)
			}

			if tc.err == "" {

				if res.StatusCode != http.StatusOK {
					t.Errorf("expected status OK, got %v", res.Status)
				}

				if a.Usernames[0] != tc.usernames[0] {
					t.Errorf("expected %v, got %v", tc.usernames[0], a.Usernames[0])
				}

				if a.EmailAddr != tc.email {
					t.Errorf("expected %v, got %v", tc.email, a.EmailAddr)
				}

				if a.Org != tc.org {
					t.Errorf("expected %v, got %v", tc.org, a.Org)
				}

				if a.Method != tc.method {
					t.Errorf("expected %v, got %v", tc.method, a.Method)
				}

				if a.URI != tc.path {
					t.Errorf("expected %v, got %v", tc.path, a.URI)
				}
			}

			if msg := string(bytes.TrimSpace(b)); msg != tc.err {
				t.Errorf("expected error %q, got: %q", tc.err, msg)
			}

		})
	}
}
