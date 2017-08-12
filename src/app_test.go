package main_test

import (
	"testing"
	"time"
	"net/http"
	"."
)

func TestServerAvailable(t *testing.T) {
	app := main.App{Port: 8080}
	defer app.Stop()

	go func() {
		app.Start()
	}()

	for i:=0; i<5; i++ {
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
