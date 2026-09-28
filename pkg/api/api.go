package api

import "time"

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type RegisterRequest struct {
	Namespace string `json:"namespace"`
	PodID     string `json:"pod_id"`
	Name      string `json:"name"`
	Host      string `json:"host,omitempty"`
	Port      int    `json:"port"`
}

type HeartbeatRequest struct {
	Namespace string `json:"namespace"`
	PodID     string `json:"pod_id"`
	Name      string `json:"name"`
	Host      string `json:"host,omitempty"`
	Port      int    `json:"port"`
	State     string `json:"state,omitempty"`
}

type PodStatus struct {
	Namespace     string    `json:"namespace"`
	PodID         string    `json:"pod_id"`
	Name          string    `json:"name"`
	Host          string    `json:"host"`
	Port          int       `json:"port"`
	Status        string    `json:"status"`
	RegisteredAt  time.Time `json:"registered_at"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
}

type RebuildRequest struct {
	Namespace string `json:"namespace"`
}

type MessageResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}
