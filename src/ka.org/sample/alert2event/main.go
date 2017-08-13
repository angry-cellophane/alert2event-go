package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

const prometheusEndpointResponse = `# HELP alerts_number The total number of alerts
# TYPE alerts_number counter
alerts_number %d`

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

func alert2eventHandler(w http.ResponseWriter, r *http.Request) {
	alertsNumber++
	var alert Alert
	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	eventInBytes, err := json.Marshal(Event{
		Name:    "Event",
		Summary: alert.Summary,
		Type:    alert.Severity,
	})
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := http.Post(eventApiUrl, "application/json", bytes.NewBuffer(eventInBytes))
	if err != nil || resp.StatusCode != 200 {
		switch {
		case err != nil:
			log.Println("Cannot send event to EventAPI: " + err.Error())
		case resp.StatusCode != 200:
			log.Println("POST " + eventApiUrl + " returned " + resp.Status)
		}
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf(prometheusEndpointResponse, alertsNumber)))
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
	eventApiUrl = os.Getenv("EVENTAPI_URL")
	if len(eventApiUrl) == 0 {
		fmt.Println("Env var EVENTAPI_URL is not defined. Define the var and rerun the app")
		os.Exit(1)
	}

	handler("/alert", "POST", alert2eventHandler)
	handler("/prometheus", "GET", metricsHandler)

	app := http.Server{Addr: ":8080"}
	if err := app.ListenAndServe(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
