package main

import (
	"log"
	"net/http"

	"github.com/apex/gateway"
)

// Message represents an update request
type Message struct {
	Usernames []string `json:"usernames,omitempty"`
	EmailAddr string   `json:"emailAddress,omitempty"`
	Org       string   `json:"org,omitempty"`
	Method    string   `json:"method,omitempty"`
	URI       string   `json:"uri,omitempty"`
	Payload   []byte   `json:"payload,omitempty"`
}

var m Message

// Caller is a middleware handler that makes outgoing requests
type Caller struct {
	handler http.Handler
}

// NewCaller constructs a new Caller middleware handler
func NewCaller(handlerToWrap http.Handler) *Caller {
	return &Caller{handlerToWrap}
}

// serveHTTP passes the request from main handler to middleware
func (c *Caller) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	c.handler.ServeHTTP(w, r)

	err := callJSD(&m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

// handler serves a wrapped mux
func handler() http.Handler {

	mux := http.NewServeMux()
	mux.HandleFunc("/add", m.addHandler)
	mux.HandleFunc("/remove", m.removeHandler)
	mux.HandleFunc("/update", m.updateHandler)
	return NewCaller(mux)
}

func main() {

	log.Fatal(gateway.ListenAndServe("", handler()))
}
