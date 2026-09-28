package runner

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/shimshimney/pkg/config"
)

func waitForFile(t *testing.T, path, want string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if data, err := os.ReadFile(path); err == nil && strings.TrimSpace(string(data)) == want {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	data, _ := os.ReadFile(path)
	t.Fatalf("%s: want %q, got %q", path, want, data)
}

func TestRestartReplacesRunningProcess(t *testing.T) {
	dir := t.TempDir()
	version := filepath.Join(dir, "version")
	out := filepath.Join(dir, "out")
	if err := os.WriteFile(version, []byte("v1"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Build: "cp " + version + " " + filepath.Join(dir, "built"),
		Run:   "sh -c 'cat " + filepath.Join(dir, "built") + " > " + out + "; sleep 60'",
	}
	r := New(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	defer r.Stop()

	if err := r.Rebuild(); err != nil {
		t.Fatal(err)
	}
	if err := r.Start(); err != nil {
		t.Fatal(err)
	}
	waitForFile(t, out, "v1")
	first := r.current.Process.Pid

	if err := os.WriteFile(version, []byte("v2"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := r.Restart(); err != nil {
		t.Fatal(err)
	}
	waitForFile(t, out, "v2")
	if r.current.Process.Pid == first {
		t.Fatal("process was not replaced")
	}
}

func TestRestartKeepsOldProcessWhenBuildFails(t *testing.T) {
	r := New(config.Config{Build: "true", Run: "sleep 60"}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	defer r.Stop()
	if err := r.Start(); err != nil {
		t.Fatal(err)
	}
	pid := r.current.Process.Pid
	r.BuildCmd = "exit 1"
	if err := r.Restart(); err == nil {
		t.Fatal("expected build failure")
	}
	if r.current == nil || r.current.Process.Pid != pid {
		t.Fatal("running process should be kept after failed build")
	}
}
