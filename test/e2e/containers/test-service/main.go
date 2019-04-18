package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

func main() {
	port := flag.Uint64("p", 8080, "port")
	failEvenRequests := flag.Bool("fail-half", false, "fail every other request. used to test retries")
	flag.Parse()
	log.Fatal(run(*port, *failEvenRequests))
}

func logReq(r *http.Request) {
	log.Printf("request count: %v", requestCount)
	b, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Printf("error dumping req: %v", err)
	} else {
		log.Printf("request dump:\n%s\n", b)
	}
}

var requestCount int

func run(port uint64, failEvenRequests bool) error {
	m := http.NewServeMux()
	m.HandleFunc("/retry-this-route", func(w http.ResponseWriter, r *http.Request) {
		logReq(r)
		requestCount++

		if failEvenRequests && requestCount%2 == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
	return http.ListenAndServe(fmt.Sprintf(":%v", port), m)
}
