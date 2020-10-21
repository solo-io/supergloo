package extensions

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func RunEchoSerer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := func() error {
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return err
			}
			for k := range r.Header {
				w.Header().Set(k, r.Header.Get(k))
			}
			_, err = w.Write(b)
			return err
		}(); err != nil {
			http.Error(w, err.Error(), 500)
		}
	})
	return http.ListenAndServe(fmt.Sprintf(":%v", EchoServerPort), mux)
}
