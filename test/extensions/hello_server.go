package extensions

import (
	"fmt"
	"net/http"
)

func RunHelloServer(helloMsg string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(helloMsg))
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})
	return http.ListenAndServe(fmt.Sprintf(":%v", HelloServerPort), mux)
}
