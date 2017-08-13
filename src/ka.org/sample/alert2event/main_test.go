package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"testing"
)

/*
Start server first, then run other tests
*/
func TestMain(m *testing.M) {
	runEventApiStub()
	os.Setenv("EVENTAPI_URL", "http://localhost:8080/event")
	go func() {
		main()
	}()

	m.Run()
}

func TestServerIsAvailable(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/prometheus")
	switch {
	case err != nil:
		t.Fatal("GET /prometheus returned an error " + err.Error())
	case resp.StatusCode != 200:
		t.Fatal("Server is up but /prometheus returned unknown answer " + resp.Status)
	}
}

func TestEventApiReceivesEvents(t *testing.T) {
	alert := Alert{
		Summary:  "test TestEventApiReceivesEvents summary",
		Severity: "WARNING",
	}
	alertBytes, _ := json.Marshal(alert)

	resp, err := http.Post("http://localhost:8080/alert", "application/json", bytes.NewBuffer(alertBytes))
	switch {
	case err != nil:
		t.Fatal("POST /alert returned an exception " + err.Error())
	case resp.StatusCode != 200:
		t.Fatal("POST /alert status != 200 -> " + resp.Status)
	}

	if _, ok := events.find(func(e Event) bool { return e.Summary == alert.Summary }); !ok {
		t.Fatal("EventAPI stub has not received the expected test event")
	}
}

func runEventApiStub() {
	http.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var event Event
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		events.add(event)
		w.WriteHeader(http.StatusOK)
	})
}

type Events struct {
	events []Event
	mux    sync.Mutex
}

func (e *Events) add(event Event) {
	e.mux.Lock()
	defer e.mux.Unlock()

	e.events = append(e.events, event)
}

func (e *Events) find(predicate func(Event) bool) (*Event, bool) {
	e.mux.Lock()
	defer e.mux.Unlock()

	for _, event := range e.events {
		if predicate(event) {
			return &event, true
		}
	}

	return nil, false
}

var events = Events{}
