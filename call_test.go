package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestClient(t *testing.T) {

	tt := []struct {
		name    string
		method  string
		path    string
		payload string
	}{
		{name: "update", method: "PUT", path: "/rest/api/2/user?username=foo%40bar.com", payload: `{"emailAddress":"foo@bat.com"}`},
		{name: "add", method: "POST", path: "/rest/servicedeskapi/organization/7/user", payload: `{"usernames":["alice"]}`},
		{name: "remove", method: "DELETE", path: "/rest/servicedeskapi/organization/8/user", payload: `{"usernames":["bob"]}`},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			setEnv()

			testSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				if r.Method != tc.method {
					t.Errorf("wrong HTTP method %v", r.Method)
				}

				ct := r.Header.Get("Content-Type")
				if ct != "application/json" {
					t.Errorf("wrong content type: %v", ct)
				}

				ex := r.Header.Get("X-ExperimentalApi")
				if ex != "opt-in" {
					t.Errorf("wrong experimental api header: %v", ex)
				}

				sa := r.Header.Get("Authorization")
				if sa != "Basic Zm9vOmJhcg==" {
					t.Errorf("wrong auth header: %v", sa)
				}

				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Errorf("could not read request body: %v", sa)
				}

				if string(body) != tc.payload {
					t.Errorf("expected %v, got %v", tc.payload, string(body))
				}
			}))

			u, _ := url.Parse(testSrv.URL)
			c := &Client{
				BaseURL:    u,
				httpClient: &http.Client{Timeout: 10 * time.Second},
			}

			req, err := c.newRequest(tc.method, tc.path, []byte(tc.payload))
			if err != nil {
				t.Fatalf("could not make request: %q", err)
			}

			if req.URL.String() != (u.String() + tc.path) {
				t.Errorf("wrong target url: %v", req.URL.String())
			}

			resp, err := c.do(req)
			if err != nil {
				t.Fatalf("call failed: %v", err)
			}
			defer resp.Body.Close()

		})
	}
}
