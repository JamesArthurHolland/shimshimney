package common

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func Serve(name string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprintf(w, "hello world from %s\n", name); err != nil {
			log.Printf("write response failed: %v", err)
		}
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return http.ListenAndServe(":"+port, mux)
}
