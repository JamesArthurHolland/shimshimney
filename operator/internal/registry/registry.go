package registry

import (
	"sync"
	"time"

	sharedapi "github.com/shimshimney/pkg/api"
)

type Pod struct {
	Namespace     string
	PodID         string
	Name          string
	Host          string
	Port          int
	Status        string
	RegisteredAt  time.Time
	LastHeartbeat time.Time
}

type Registry struct {
	mu   sync.RWMutex
	pods map[podKey]*Pod
}

type podKey struct {
	namespace string
	podID     string
}

func New() *Registry {
	return &Registry{pods: map[podKey]*Pod{}}
}

func (r *Registry) Upsert(req sharedapi.RegisterRequest) *Pod {
	r.mu.Lock()
	defer r.mu.Unlock()
	podID := req.PodID
	if podID == "" {
		podID = req.Name
	}
	key := podKey{req.Namespace, podID}
	pod := r.pods[key]
	if pod == nil {
		pod = &Pod{Namespace: req.Namespace, PodID: podID, RegisteredAt: time.Now()}
		r.pods[key] = pod
	}
	pod.Name = req.Name
	pod.Host = req.Host
	pod.Port = req.Port
	pod.Status = "registered"
	pod.LastHeartbeat = time.Now()
	if pod.RegisteredAt.IsZero() {
		pod.RegisteredAt = time.Now()
	}
	return pod
}

func (r *Registry) Heartbeat(req sharedapi.HeartbeatRequest) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	pod, ok := r.pods[podKey{req.Namespace, req.PodID}]
	if !ok {
		return false
	}
	pod.LastHeartbeat = time.Now()
	pod.Status = "healthy"
	if req.Name != "" {
		pod.Name = req.Name
	}
	if req.Host != "" {
		pod.Host = req.Host
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
			Namespace:     pod.Namespace,
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

func (r *Registry) InNamespace(namespace string) []Pod {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Pod, 0)
	for _, pod := range r.pods {
		if pod.Namespace == namespace {
			result = append(result, *pod)
		}
	}
	return result
}
