package stats

import (
	"fmt"
	"log"
	"net/http"
)

func MustStartServerBackground(snapshotHistory *SnapshotHistory, port uint32) {
	go func() {
		if err := StartServer(snapshotHistory, port); err != nil {
			log.Fatal(err)
		}
	}()
}

func StartServer(snapshotHistory *SnapshotHistory, port uint32) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", Index)
	AddPprof(mux)
	AddMetrics(mux)
	AddSnapshots(mux, snapshotHistory)
	return http.ListenAndServe(fmt.Sprintf(":%v", port), mux)
}
