package main

import (
	"ka.org/sample/alert2event/app"
	"os"
	"fmt"
)

func main() {
	eventApi := os.Getenv("EVENTAPI_URL")
	if len(eventApi) == 0 {
		fmt.Println("Env var EVENTAPI_URL is not defined. Define the var and rerun the app")
		os.Exit(1)
	}

	server := app.App{Port: 8080, EventApiUrl: eventApi}
	server.Start()
}
