package main

import "time"

type ServiceInstance struct {
	ID            string    `json:"id"`
	ServiceName   string    `json:"service_name"`
	RegisteredAt  time.Time `json:"registered_at"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	CPUCores      int       `json:"cpu_cores"`
	CPUPercent    float64   `json:"cpu_percent"`
	MemoryTotalMB int64     `json:"memory_total_mb"`
	MemoryAvailMB int64     `json:"memory_avail_mb"`
	GlobalVersion int64     `json:"global_version"`
	ServiceVersion int64    `json:"service_version"`
	HttpHost      string    `json:"http_host"`
	HttpPort      string    `json:"http_port"`
}

type ServiceSummary struct {
	Name          string             `json:"name"`
	InstanceCount int                `json:"instance_count"`
	Instances     []*ServiceInstance `json:"instances"`
}

type servicesSnapshotMessage struct {
	Type     string            `json:"type"`
	Services []*ServiceSummary `json:"services"`
}

type serviceEventMessage struct {
	Type      string           `json:"type"`
	Event     string           `json:"event"`
	Service   string           `json:"service"`
	Instance  *ServiceInstance `json:"instance"`
	Instances int              `json:"instances"`
}