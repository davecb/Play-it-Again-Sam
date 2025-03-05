package main

// dummy -- This is a web service that just returns the parameter it was called with
// 	Based on "hello world", https://www.atlantic.net/dedicated-server-hosting/deploying-a-go-web-application-using-nginx-on-ubuntu-22-04/

import (
	"fmt"
	"log"
	"net/http"
)

const addr = ":9990"

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//grab the next part of the url as a parameter
		param := r.URL.Path[len("/"):]
		if param == "" {
			param = "nothing at all"
		}
		fmt.Fprintf(w, "Got %q\n", param)
	})

	http.ListenAndServe(addr, nil)
	log.Fatal(http.ListenAndServe(addr, nil))
}
