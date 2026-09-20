package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	sharedapi "github.com/shimshimney/pkg/api"
	"github.com/shimshimney/operator/internal/k8s"
	"github.com/shimshimney/operator/internal/registry"
)

type Server struct {
	registry      *registry.Registry
	services      *k8s.Manager
	logger        *slog.Logger
	clientTimeout time.Duration
}

func NewServer(reg *registry.Registry, services *k8s.Manager, logger *slog.Logger) *Server {
	return &Server{registry: reg, services: services, logger: logger, clientTimeout: 5 * time.Second}
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
	if !s.registry.Heartbeat(req.PodID, req) {
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
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	results := s.rebuildAllPods()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "results": results})
}

func (s *Server) rebuildAllPods() []string {
	results := make([]string, 0)
	for _, pod := range s.registry.All() {
		endpoint := fmt.Sprintf("http://%s:%d/rebuild", pod.Host, pod.Port)
		if pod.Host == "" {
			endpoint = fmt.Sprintf("http://localhost:%d/rebuild", pod.Port)
		}
		client := &http.Client{Timeout: s.clientTimeout}
		resp, err := client.Post(endpoint, "application/json", nil)
		if err != nil {
			s.logger.Warn("rebuild request failed", "pod", pod.PodID, "endpoint", endpoint, "error", err)
			results = append(results, fmt.Sprintf("%s failed: %v", pod.PodID, err))
			continue
		}
		defer resp.Body.Close()
		results = append(results, fmt.Sprintf("%s rebuilt", pod.PodID))
	}
	return results
}
