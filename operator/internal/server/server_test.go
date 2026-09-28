package server

import (
	"bytes"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/shimshimney/operator/internal/k8s"
	"github.com/shimshimney/operator/internal/registry"
	"github.com/shimshimney/pkg/api"
)

func TestRebuildOnlySpecifiedNamespace(t *testing.T) {
	var exampleHits, otherHits atomic.Int32
	shim := func(hits *atomic.Int32) string {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hits.Add(1)
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(server.Close)
		return strings.TrimPrefix(server.URL, "http://")
	}
	exampleAddr := shim(&exampleHits)
	otherAddr := shim(&otherHits)
	reg := registry.New()
	for _, entry := range []struct{ namespace, address string }{{"example", exampleAddr}, {"other", otherAddr}} {
		// The same pod ID in different namespaces must remain independent.
		host, portString, err := net.SplitHostPort(entry.address)
		if err != nil {
			t.Fatal(err)
		}
		port, err := strconv.Atoi(portString)
		if err != nil {
			t.Fatal(err)
		}
		reg.Upsert(api.RegisterRequest{Namespace: entry.namespace, PodID: "backend-1", Host: host, Port: port})
	}
	srv := NewServer(reg, k8s.NewManager(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	for _, tc := range []struct {
		body string
		want int
	}{{`{"namespace":"example"}`, http.StatusOK}, {`{}`, http.StatusBadRequest}} {
		recorder := httptest.NewRecorder()
		srv.handleRebuild(recorder, httptest.NewRequest(http.MethodPost, "/rebuild", bytes.NewBufferString(tc.body)))
		if recorder.Code != tc.want {
			t.Fatalf("body %s: got %d, want %d", tc.body, recorder.Code, tc.want)
		}
	}
	if exampleHits.Load() != 1 || otherHits.Load() != 0 {
		t.Fatalf("unexpected rebuilds: example=%d other=%d", exampleHits.Load(), otherHits.Load())
	}
}
