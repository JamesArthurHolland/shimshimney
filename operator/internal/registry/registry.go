package registry

import (
	"sync"
	"time"

	sharedapi "github.com/shimshimney/pkg/api"
)

type Pod struct {
	PodID         string
	Name          string
	Host          string
	Port          int
	Status        string
	BuildMode     string
	RegisteredAt  time.Time
	LastHeartbeat time.Time
}

type Registry struct {
	mu   sync.RWMutex
	pods map[string]*Pod
}

func New() *Registry {
	return &Registry{pods: map[string]*Pod{}}
}

func (r *Registry) Upsert(req sharedapi.RegisterRequest) *Pod {
	r.mu.Lock()
	defer r.mu.Unlock()
	podID := req.PodID
	if podID == "" {
		podID = req.Name
	}
	pod := r.pods[podID]
	if pod == nil {
		pod = &Pod{PodID: podID, RegisteredAt: time.Now()}
		r.pods[podID] = pod
	}
	pod.Name = req.Name
	pod.Host = req.Host
	pod.Port = req.Port
	pod.BuildMode = req.BuildMode
	pod.Status = "registered"
	pod.LastHeartbeat = time.Now()
	if pod.RegisteredAt.IsZero() {
		pod.RegisteredAt = time.Now()
	}
	return pod
}

func (r *Registry) Heartbeat(podID string, req sharedapi.HeartbeatRequest) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	pod, ok := r.pods[podID]
	if !ok {
		return false
	}
	pod.LastHeartbeat = time.Now()
	pod.Status = "healthy"
	if req.Name != "" {
		pod.Name = req.Name
	}
	if req.Port != 0 {
		pod.Port = req.Port
	}
	return true
}

func (r *Registry) List() []sharedapi.PodStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]sharedapi.PodStatus, 0, len(r.pods))
	for _, pod := range r.pods {
		list = append(list, sharedapi.PodStatus{
			PodID:         pod.PodID,
			Name:          pod.Name,
			Host:          pod.Host,
			Port:          pod.Port,
			Status:        pod.Status,
			RegisteredAt:  pod.RegisteredAt,
			LastHeartbeat: pod.LastHeartbeat,
		})
	}
	return list
}

func (r *Registry) All() []*Pod {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*Pod, 0, len(r.pods))
	for _, pod := range r.pods {
		result = append(result, pod)
	}
	return result
}
