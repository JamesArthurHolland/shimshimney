package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/shimshimney/pkg/api"
	pclient "github.com/shimshimney/pkg/client"
	shlogger "github.com/shimshimney/pkg/logger"
)

func TestColdModeOnlyBuildsWithoutContactingOperator(t *testing.T) {
	var requests atomic.Int32
	operator := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer operator.Close()

	output := filepath.Join(t.TempDir(), "commands")
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte("build_mode: cold\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SHIMNEY_CONFIG_PATH", configPath)
	t.Setenv("SHIMNEY_BUILD_MODE", "cold")
	t.Setenv("SHIMNEY_OPERATOR_URL", operator.URL)
	t.Setenv("SHIMNEY_BUILD", "printf 'build\\n' >> "+output)
	t.Setenv("SHIMNEY_RUN", "printf 'run\\n' >> "+output)

	if err := run(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "build\n" {
		t.Fatalf("unexpected commands: %q", data)
	}
	if got := requests.Load(); got != 0 {
		t.Fatalf("cold mode made %d operator requests", got)
	}
}

func TestColdModeDoesNotRunAfterBuildFailure(t *testing.T) {
	output := filepath.Join(t.TempDir(), "run")
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte("build_mode: cold\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SHIMNEY_CONFIG_PATH", configPath)
	t.Setenv("SHIMNEY_BUILD_MODE", "cold")
	t.Setenv("SHIMNEY_BUILD", "exit 1")
	t.Setenv("SHIMNEY_RUN", "touch "+output)

	if err := run(); err == nil || !strings.Contains(err.Error(), "build failed") {
		t.Fatalf("expected build failure, got %v", err)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("run command executed after failed build: %v", err)
	}
}

func TestHeartbeatReregistersWhenOperatorForgetsPod(t *testing.T) {
	var registered atomic.Bool
	var registers atomic.Int32
	operator := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/register":
			registers.Add(1)
			registered.Store(true)
			w.WriteHeader(http.StatusNoContent)
		case "/heartbeat":
			if !registered.Load() {
				http.Error(w, "pod is not registered", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer operator.Close()

	client := pclient.New(operator.URL)
	logger := shlogger.New("test")
	reg := api.RegisterRequest{Namespace: "ns", PodID: "pod", Name: "app", Port: 9090}
	hb := api.HeartbeatRequest{Namespace: "ns", PodID: "pod", Name: "app", Port: 9090}

	if err := heartbeat(client, reg, hb, logger); err != nil {
		t.Fatal(err)
	}
	if err := heartbeat(client, reg, hb, logger); err != nil {
		t.Fatal(err)
	}
	if got := registers.Load(); got != 1 {
		t.Fatalf("expected 1 re-registration, got %d", got)
	}
}

func TestHeartbeatDoesNotReregisterOnOtherErrors(t *testing.T) {
	var registers atomic.Int32
	operator := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/register" {
			registers.Add(1)
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer operator.Close()

	err := heartbeat(pclient.New(operator.URL), api.RegisterRequest{}, api.HeartbeatRequest{}, shlogger.New("test"))
	if err == nil {
		t.Fatal("expected heartbeat error")
	}
	if got := registers.Load(); got != 0 {
		t.Fatalf("expected no re-registration, got %d", got)
	}
}
