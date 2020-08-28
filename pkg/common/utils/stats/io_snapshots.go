package stats

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type SnapshotHistory struct {
	lock         sync.RWMutex
	latestInput  json.Marshaler
	latestOutput json.Marshaler
}

func NewSnapshotHistory() *SnapshotHistory {
	return &SnapshotHistory{}
}

func (h *SnapshotHistory) SetInput(latestInput json.Marshaler) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.latestInput = latestInput
}

func (h *SnapshotHistory) GetInput() ([]byte, error) {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return h.latestInput.MarshalJSON()
}

func (h *SnapshotHistory) SetOutput(latestOutput json.Marshaler) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.latestOutput = latestOutput
}

func (h *SnapshotHistory) GetOutput() ([]byte, error) {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return h.latestOutput.MarshalJSON()
}

func AddSnapshots(mux *http.ServeMux, history *SnapshotHistory) {

	profileDescriptions["/snapshots/input"] = "Latest Input Snapshot"
	profileDescriptions["/snapshots/output"] = "Latest Output Snapshot"

	mux.HandleFunc("/snapshots/input", func(w http.ResponseWriter, r *http.Request) {
		b, err := history.GetInput()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = fmt.Fprint(w, string(b))
	})

	mux.HandleFunc("/snapshots/output", func(w http.ResponseWriter, r *http.Request) {
		b, err := history.GetOutput()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = fmt.Fprint(w, string(b))
	})
}
