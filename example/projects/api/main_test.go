package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestAggregatesBackendsConcurrentlyInOrder(t *testing.T) {
	const count = 8
	var started atomic.Int32
	release := make(chan struct{})
	endpoints := make([]string, count)
	for i := range endpoints {
		i := i
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/hello" {
				http.NotFound(w, r)
				return
			}
			if started.Add(1) == count {
				close(release)
			}
			select {
			case <-release:
			case <-r.Context().Done():
				return
			}
			fmt.Fprintf(w, "backend-%d\n", i+1)
		}))
		t.Cleanup(server.Close)
		endpoints[i] = server.URL + "/hello"
	}
	recorder := httptest.NewRecorder()
	done := make(chan struct{})
	go func() {
		newHandler(endpoints, &http.Client{Timeout: time.Second}).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("backends were not requested concurrently")
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status %d: %s", recorder.Code, recorder.Body.String())
	}
	lines := strings.Split(strings.TrimSpace(recorder.Body.String()), "\n")
	if len(lines) != count {
		t.Fatalf("expected %d responses, got %d", count, len(lines))
	}
	for i, line := range lines {
		if want := fmt.Sprintf("backend-%d", i+1); line != want {
			t.Fatalf("response line %d: got %q, want %q", i, line, want)
		}
	}
}

func TestBackendFailureReturnsBadGateway(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "failed", http.StatusInternalServerError)
	}))
	defer server.Close()
	recorder := httptest.NewRecorder()
	newHandler([]string{server.URL}, server.Client()).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("expected bad gateway, got %d", recorder.Code)
	}
}
