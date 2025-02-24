package main

// dummy -- This is a web service that waits a specific period
// 	and then returns success. Based on "hello world" (:-))
//	See https://www.atlantic.net/dedicated-server-hosting/deploying-a-go-web-application-using-nginx-on-ubuntu-22-04/

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Dave, you want http://localhost:9990/greet/ with an optional name to greet\n")
	})

	http.HandleFunc("/greet/", func(w http.ResponseWriter, r *http.Request) {
		//grab the next part of the url as a parameter
		name := r.URL.Path[len("/greet/"):]
		if name == "" {
			name = "zaphod beeblebrox"
		}
		fmt.Fprintf(w, "Hello %s\n", name)
		fmt.Fprintf(w, "waiting %v seconds\n", 1)
		time.Sleep(1 * time.Second)
		fmt.Fprintf(w, "done waiting\n")
	})

	http.ListenAndServe(":9990", nil)
}
