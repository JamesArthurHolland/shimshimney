package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/shimshimney/operator/internal/k8s"
	"github.com/shimshimney/operator/internal/registry"
	sharedapi "github.com/shimshimney/pkg/api"
)

type Server struct {
	registry      *registry.Registry
	services      *k8s.Manager
	logger        *slog.Logger
	clientTimeout time.Duration
}

func NewServer(reg *registry.Registry, services *k8s.Manager, logger *slog.Logger) *Server {
	return &Server{registry: reg, services: services, logger: logger, clientTimeout: 2 * time.Minute}
}

func (s *Server) Run(addr string) error {
	http.HandleFunc("/health", s.handleHealth)
	http.HandleFunc("/register", s.handleRegister)
	http.HandleFunc("/heartbeat", s.handleHeartbeat)
	http.HandleFunc("/rebuild", s.handleRebuild)
	http.HandleFunc("/pods", s.handlePods)
	return http.ListenAndServe(addr, nil)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sharedapi.HealthResponse{Status: "ok", Timestamp: time.Now()})
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req sharedapi.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Namespace) == "" || (req.PodID == "" && req.Name == "") {
		http.Error(w, "namespace and pod_id or name are required", http.StatusBadRequest)
		return
	}
	pod := s.registry.Upsert(req)
	s.services.EnsurePodService(pod)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sharedapi.MessageResponse{OK: true, Message: fmt.Sprintf("registered %s", pod.PodID)})
}

func (s *Server) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req sharedapi.HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Namespace) == "" || req.PodID == "" {
		http.Error(w, "namespace and pod_id are required", http.StatusBadRequest)
		return
	}
	if !s.registry.Heartbeat(req) {
		http.Error(w, "pod is not registered", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handlePods(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.registry.List())
}

func (s *Server) handleRebuild(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req sharedapi.RebuildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Namespace) == "" {
		http.Error(w, "namespace is required", http.StatusBadRequest)
		return
	}
	results := s.rebuildPods(req.Namespace)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "results": results})
}

func (s *Server) rebuildPods(namespace string) []string {
	pods := s.registry.InNamespace(namespace)
	results := make([]string, len(pods))
	client := &http.Client{Timeout: s.clientTimeout}
	var wg sync.WaitGroup
	for i, pod := range pods {
		wg.Add(1)
		go func(i int, pod registry.Pod) {
			defer wg.Done()
			results[i] = s.rebuildPod(client, pod)
		}(i, pod)
	}
	wg.Wait()
	return results
}

func (s *Server) rebuildPod(client *http.Client, pod registry.Pod) string {
	endpoint := fmt.Sprintf("http://%s:%d/rebuild", pod.Host, pod.Port)
	if pod.Host == "" {
		endpoint = fmt.Sprintf("http://localhost:%d/rebuild", pod.Port)
	}
	resp, err := client.Post(endpoint, "application/json", nil)
	if err != nil {
		s.logger.Warn("rebuild request failed", "pod", pod.PodID, "endpoint", endpoint, "error", err)
		return fmt.Sprintf("%s failed: %v", pod.PodID, err)
	}
	resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Sprintf("%s failed: %s", pod.PodID, resp.Status)
	}
	return fmt.Sprintf("%s rebuilt", pod.PodID)
}
