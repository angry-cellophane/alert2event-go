package main

import (
	"net/http"
	"log"
	"encoding/json"
	"bytes"
)

type Alert struct {
	Summary  string `json: summary`
	Severity string `json: severity`
}

type Event struct {
	Name    string `json: name`
	Type    string `json: type`
	Summary string `json: summary`
}

func dummyEventApi(w http.ResponseWriter, r *http.Request) {
	var event Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println(event)
	w.WriteHeader(http.StatusOK)
}

func alert2event(a Alert) Event {
	return Event{
		Name: "Event",
		Summary: a.Summary,
		Type: a.Severity,
	}
}

func sendEvent(e *Event, cb func(*http.Response, error)) {
	event, err := json.Marshal(e)
	if err != nil {
		cb(nil, err)
		return
	}

	resp, err := http.Post("http://localhost:8080/event", "application/json", bytes.NewBuffer(event))
	cb(resp, err)
}

func alert2eventHandler(w http.ResponseWriter, r *http.Request) {
	var alert Alert
	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		log.Println(err)
	}

	event := alert2event(alert)
	go sendEvent(&event, func(resp *http.Response, err error) {
		if err != nil {
			log.Println("sendEvent cb: " + err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}

func handler(pattern, method string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		handler(w, r)
	})
}

func main() {
	handler("/alert", "POST", alert2eventHandler)
	handler("/event", "POST", dummyEventApi)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
