package main

// This is a limiter from https://blog.logrocket.com/rate-limiting-go-application/
// It is a blanket limit, system-wide.
// It is modified to add a delay, and set the limit to 30 TPS

import (
	"encoding/json"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	"time"
)

const (
	port  = ":9990"
	limit = 30
	burst = 5
	delay = 0.1
)

type Message struct {
	Status string `json:"status"`
	Body   string `json:"body"`
}

func main() {
	http.Handle("/", rateLimiter(endpointHandler))
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Printf("There was an error listening on port %s: %v\n", port, err)
	}
}

func rateLimiter(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	limiter := rate.NewLimiter(limit, burst)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			message := Message{
				Status: "Request Failed",
				Body:   "The API is at capacity, try again later.",
			}

			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(&message)
			return
		} else {
			next(w, r)
		}
	})
}

func endpointHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	message := Message{
		Status: "Successful",
		Body:   "Hi! You've reached the API. How may I help you?",
	}
	time.Sleep(1 * time.Minute)
	err := json.NewEncoder(writer).Encode(&message)
	if err != nil {
		return
	}
}
