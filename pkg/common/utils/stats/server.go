package stats

import (
	"fmt"
	"log"
	"net/http"
)

func MustStartServerBackground(port uint32) {
	go func() {
		if err := StartServer(port); err != nil {
			log.Fatal(err)
		}
	}()
}

func StartServer(port uint32) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", Index)
	AddPprof(mux)
	AddMetrics(mux)
	return http.ListenAndServe(fmt.Sprintf(":%v", port), mux)
}
