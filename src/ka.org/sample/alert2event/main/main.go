package main

import (
	"ka.org/sample/alert2event/app"
	"os"
)

func main() {
	eventApi := os.Getenv("EVENTAPI_URL")
	if len(eventApi) == 0 {
		panic("Env var EVENTAPI_URL is not defined. Define the var and rerun the app")
	}

	server := app.App{Port: 8080, EventApiUrl: eventApi}
	server.Start()
}
