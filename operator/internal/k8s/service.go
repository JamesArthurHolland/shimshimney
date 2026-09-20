package k8s

import (
	"fmt"
	"sync"

	"github.com/shimshimney/operator/internal/registry"
)

type Service struct {
	Name       string
	Port       int
	TargetPort int
	Selector   map[string]string
}

type Manager struct {
	mu       sync.Mutex
	services map[string]Service
}

func NewManager() *Manager {
	return &Manager{services: map[string]Service{}}
}

func (m *Manager) EnsurePodService(pod *registry.Pod) Service {
	m.mu.Lock()
	defer m.mu.Unlock()
	name := fmt.Sprintf("svc-%s", pod.PodID)
	service := Service{
		Name:       name,
		Port:       pod.Port,
		TargetPort: pod.Port,
		Selector: map[string]string{
			"pod_id": pod.PodID,
		},
	}
	m.services[name] = service
	return service
}

func (m *Manager) List() []Service {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]Service, 0, len(m.services))
	for _, service := range m.services {
		result = append(result, service)
	}
	return result
}
