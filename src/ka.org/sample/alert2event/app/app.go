package app

import (
	"net/http"
	"log"
	"encoding/json"
	"bytes"
	"strconv"
	"fmt"
)


var alertsNumber int = 0
var eventApiUrl string

type Alert struct {
	Summary  string `json: summary`
	Severity string `json: severity`
}

type Event struct {
	Name    string `json: name`
	Type    string `json: type`
	Summary string `json: summary`
}

func alert2event(a Alert) Event {
	return Event{
		Name: "Event",
		Summary: a.Summary,
		Type: a.Severity,
	}
}

func alert2eventHandler(w http.ResponseWriter, r *http.Request) {
	alertsNumber++
	var alert Alert
	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event := alert2event(alert)
	eventBytes, err := json.Marshal(event)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, err := http.Post(eventApiUrl, "application/json", bytes.NewBuffer(eventBytes))

	if err != nil || resp.StatusCode != 200 {
		switch  {
		case err != nil:
			log.Println("Cannot send event to EventAPI: " + err.Error())
		case resp.StatusCode != 200:
			log.Println("POST " + eventApiUrl + " returned " + resp.Status)
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
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
	EventApiUrl string
	Port int
	server *http.Server
}

func (app *App) Start() {
	app.server = &http.Server{Addr: ":" + strconv.Itoa(app.Port)}
	eventApiUrl = app.EventApiUrl

	handler("/alert", "POST", alert2eventHandler)
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
