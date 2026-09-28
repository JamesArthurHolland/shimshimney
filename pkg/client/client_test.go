package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shimshimney/pkg/api"
)

func TestRebuildIncludesRequiredNamespace(t *testing.T) {
	var got api.RebuildRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/rebuild" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true,"results":["backend-1 rebuilt"]}`))
	}))
	defer server.Close()
	client := New(server.URL)
	if _, err := client.Rebuild(""); err == nil {
		t.Fatal("missing namespace was accepted")
	}
	results, err := client.Rebuild("example")
	if err != nil {
		t.Fatal(err)
	}
	if got.Namespace != "example" || len(results) != 1 || results[0] != "backend-1 rebuilt" {
		t.Fatalf("unexpected namespace/results: %q, %v", got.Namespace, results)
	}
}
