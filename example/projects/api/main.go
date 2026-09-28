package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/shimshimney/pkg/logger"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	endpoints := make([]string, 8)
	for i := range endpoints {
		endpoints[i] = fmt.Sprintf("http://backend-%d.example.svc.cluster.local/hello", i+1)
	}
	logger.New("api").Info("starting api", "port", port)
	if err := http.ListenAndServe(":"+port, newHandler(endpoints, &http.Client{Timeout: 5 * time.Second})); err != nil {
		log.Fatal(err)
	}
}

func newHandler(endpoints []string, client *http.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		responses := make([]string, len(endpoints))
		errors := make([]error, len(endpoints))
		var wg sync.WaitGroup
		for i, endpoint := range endpoints {
			wg.Add(1)
			go func(i int, endpoint string) {
				defer wg.Done()
				req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, endpoint, nil)
				if err != nil {
					errors[i] = err
					return
				}
				resp, err := client.Do(req)
				if err != nil {
					errors[i] = err
					return
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					errors[i] = fmt.Errorf("%s returned %s", endpoint, resp.Status)
					return
				}
				body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
				if err != nil {
					errors[i] = err
					return
				}
				responses[i] = strings.TrimSpace(string(body))
			}(i, endpoint)
		}
		wg.Wait()
		for i, err := range errors {
			if err != nil {
				http.Error(w, fmt.Sprintf("backend %d failed: %v", i+1, err), http.StatusBadGateway)
				return
			}
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintln(w, strings.Join(responses, "\n"))
	})
}
