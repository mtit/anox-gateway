package main

import (
	"errors"
	"net/url"
	"sort"
	"sync"
)

var ErrNoAvailableInstance = errors.New("no available instance")

type RegistryState struct {
	mu       sync.RWMutex
	services map[string]map[string]*ServiceInstance
}

func NewRegistryState() *RegistryState {
	return &RegistryState{services: make(map[string]map[string]*ServiceInstance)}
}

func (r *RegistryState) ReplaceAll(services []*ServiceSummary) {
	r.mu.Lock()
	defer r.mu.Unlock()

	next := make(map[string]map[string]*ServiceInstance, len(services))
	for _, service := range services {
		instances := make(map[string]*ServiceInstance, len(service.Instances))
		for _, instance := range service.Instances {
			if instance == nil || !instanceReachable(instance) {
				continue
			}
			instances[instance.ID] = cloneInstance(instance)
		}
		if len(instances) > 0 {
			next[service.Name] = instances
		}
	}
	r.services = next
}

func (r *RegistryState) ApplyEvent(event serviceEventMessage) {
	if event.Instance == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	svc := event.Service
	if svc == "" {
		svc = event.Instance.ServiceName
	}

	switch event.Event {
	case "remove":
		if instances, ok := r.services[svc]; ok {
			delete(instances, event.Instance.ID)
			if len(instances) == 0 {
				delete(r.services, svc)
			}
		}
	default:
		if !instanceReachable(event.Instance) {
			return
		}
		if _, ok := r.services[svc]; !ok {
			r.services[svc] = make(map[string]*ServiceInstance)
		}
		r.services[svc][event.Instance.ID] = cloneInstance(event.Instance)
	}
}

func (r *RegistryState) Pick(service string) (*ServiceInstance, error) {
	r.mu.RLock()
	instances := r.services[service]
	list := make([]*ServiceInstance, 0, len(instances))
	for _, instance := range instances {
		if instanceReachable(instance) {
			list = append(list, cloneInstance(instance))
		}
	}
	r.mu.RUnlock()

	if len(list) == 0 {
		return nil, ErrNoAvailableInstance
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].CPUPercent != list[j].CPUPercent {
			return list[i].CPUPercent < list[j].CPUPercent
		}
		if list[i].MemoryAvailMB != list[j].MemoryAvailMB {
			return list[i].MemoryAvailMB > list[j].MemoryAvailMB
		}
		return list[i].ID < list[j].ID
	})
	return list[0], nil
}

func cloneInstance(instance *ServiceInstance) *ServiceInstance {
	if instance == nil {
		return nil
	}
	copy := *instance
	return &copy
}

func instanceReachable(instance *ServiceInstance) bool {
	if instance == nil || instance.HttpHost == "" || instance.HttpPort == "" {
		return false
	}
	_, err := url.Parse("http://" + netJoinHostPort(instance.HttpHost, instance.HttpPort))
	return err == nil
}