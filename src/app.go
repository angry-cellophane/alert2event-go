package main

import (
	"net/http"
	"log"
	"encoding/json"
	"bytes"
	"strconv"
	"fmt"
)


var alertsNumber int = 0


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

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	response := `# HELP alerts_number The total number of alerts
# TYPE alerts_number counter
alerts_number %d`
	w.Write([]byte(fmt.Sprintf(response, alertsNumber)))
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


type App struct {
	Port int
	server *http.Server
}

func (app *App) Start() {
	app.server = &http.Server{Addr: ":" + strconv.Itoa(app.Port)}

	handler("/alert", "POST", alert2eventHandler)
	handler("/event", "POST", dummyEventApi)
	handler("/prometheus", "GET", metricsHandler)

	if err := app.server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func (app *App) Stop() {
	if err:= app.server.Shutdown(nil); err != nil {
		panic(err)
	}
}
func main() {
	app := App{Port: 8080}
	app.Start()
}
