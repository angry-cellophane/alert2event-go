package main

import (
	"testing"
	"time"
	"net/http"
	"encoding/json"
	"ka.org/sample/alert2event/app"
	"os"
	"bytes"
	"sync"
)

type Events struct {
	events []app.Event
	mux    sync.Mutex
}

func (e *Events) add(event app.Event) {
	e.mux.Lock()
	defer e.mux.Unlock()

	e.events = append(e.events, event)
}

func (e *Events) find(summary string) *app.Event {
	e.mux.Lock()
	defer e.mux.Unlock()

	for _, event := range e.events {
		if event.Summary == summary {
			return &event
		}
	}

	return nil
}

var events = Events{}

func TestMain(m *testing.M) {
	runEventApiStub()
	os.Setenv("EVENTAPI_URL", "http://localhost:8080/event")
	go func() {
		main()
	}()

	m.Run()
}

func TestServerIsAvailable(t *testing.T) {
	for i := 0; i < 5; i++ {
		resp, err := http.Get("http://localhost:8080/prometheus")

		switch {
		case err != nil && i != 4:
			continue
		case err != nil && i == 4:
			t.Fatal("Server hasn't started up in 500ms")
		case err != nil:
			time.Sleep(100 * time.Millisecond)
		case err == nil && resp.StatusCode == 200:
			break
		case err == nil && resp.StatusCode != 200:
			t.Fatal("Server is up but /prometheus returned unknown answer " + resp.Status)
		default:
			t.FailNow()
		}
	}
}

func TestEventApiReceivesEvents(t *testing.T) {
	alert := app.Alert{
		Summary: "test TestEventApiReceivesEvents summary",
		Severity: "WARNING",
	}
	alertBytes, _ := json.Marshal(alert)

	resp, err := http.Post("http://localhost:8080/alert", "application/json", bytes.NewBuffer(alertBytes))
	if err != nil {
		t.Fatal("POST /alert returned an exception " + err.Error())
	}

	if resp.StatusCode != 200 {
		t.Fatal("POST /alert status != 200 -> " + resp.Status)
	}

	foundEvent := events.find(alert.Summary)

	if foundEvent == nil {
		t.Fatal("EventAPI stub has not received the expected test event")
	}
}

func runEventApiStub() {
	http.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var event app.Event
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		events.add(event)
		w.WriteHeader(http.StatusOK)
	})
}
