package api

import (
	"encoding/json"
	"testing"
)

func TestRegisterRequestContainsOnlyPodDetails(t *testing.T) {
	data, err := json.Marshal(RegisterRequest{
		PodID: "pod-1",
		Name:  "backend-1",
		Host:  "10.0.0.1",
		Port:  9090,
	})
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	if _, ok := fields["build_mode"]; ok {
		t.Fatalf("registration must not send build mode: %s", data)
	}
	if fields["pod_id"] != "pod-1" || fields["port"] != float64(9090) {
		t.Fatalf("registration lost pod details: %s", data)
	}
}
