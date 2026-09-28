package common

import (
	"fmt"
	"net/http"
	"os"

	"github.com/shimshimney/pkg/logger"
)

func Serve(name string) error {
	log := logger.New(name)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Info("starting backend", "name", name, "port", port)
	return http.ListenAndServe(":"+port, newHandler(name))
}

func newHandler(name string) http.Handler {
	log := logger.New(name)
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if _, err := fmt.Fprintf(w, "%s %s\n", name, RandomString); err != nil {
			log.Error("write response failed", "error", err)
			return
		}
		log.Info("handled request", "path", r.URL.Path, "remote", r.RemoteAddr)
	})
	return mux
}
