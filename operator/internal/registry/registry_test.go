package registry

import (
	"testing"

	"github.com/shimshimney/pkg/api"
)

func TestHeartbeatUpdatesPodHostAfterReplacement(t *testing.T) {
	reg := New()
	reg.Upsert(api.RegisterRequest{Namespace: "example", PodID: "backend-1", Host: "10.42.0.1", Port: 9090})
	if !reg.Heartbeat(api.HeartbeatRequest{Namespace: "example", PodID: "backend-1", Host: "10.42.0.2"}) {
		t.Fatal("registered pod heartbeat rejected")
	}
	if got := reg.All()[0].Host; got != "10.42.0.2" {
		t.Fatalf("expected replacement pod IP, got %q", got)
	}
}

func TestRegistryIsolatesNamespaces(t *testing.T) {
	reg := New()
	reg.Upsert(api.RegisterRequest{Namespace: "example", PodID: "same", Host: "10.0.0.1"})
	reg.Upsert(api.RegisterRequest{Namespace: "other", PodID: "same", Host: "10.0.0.2"})
	if reg.Heartbeat(api.HeartbeatRequest{Namespace: "missing", PodID: "same"}) {
		t.Fatal("heartbeat accepted for a different namespace")
	}
	if got := reg.InNamespace("example"); len(got) != 1 || got[0].Host != "10.0.0.1" {
		t.Fatalf("example pods: %+v", got)
	}
	if got := reg.InNamespace("other"); len(got) != 1 || got[0].Host != "10.0.0.2" {
		t.Fatalf("other pods: %+v", got)
	}
}
