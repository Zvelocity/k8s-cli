package resources

import (
	"fmt"
	"time"
)

// ViewType represents different UI views
type ViewType string

const (
	// PodView is the view that shows pods
	PodView ViewType = "pods"

	// ServiceView is the view that shows services
	ServiceView ViewType = "services"

	// DetailView is the view that shows detailed information
	DetailView ViewType = "detail"

	// NamespaceView is the view for selecting namespaces
	NamespaceView ViewType = "namespaces"
)

// PodInfo contains essential pod information
type PodInfo struct {
	Name       string
	Namespace  string
	Status     string
	Age        string
	IP         string
	Node       string
	Created    time.Time
	Labels     map[string]string
	Containers []ContainerInfo
}

// ContainerInfo contains container details
type ContainerInfo struct {
	Name            string
	Image           string
	Ready           bool
	RestartCount    int
	State           string
	CPURequest      string
	MemoryRequest   string
	CPULimit        string
	MemoryLimit     string
	EnvironmentVars map[string]string
}

// ServiceInfo contains essential service information
type ServiceInfo struct {
	Name       string
	Namespace  string
	Type       string
	ClusterIP  string
	ExternalIP string
	Ports      string
	Age        string
	Selector   map[string]string
}

// ResourceData contains all resource information
type ResourceData struct {
	Pods     []PodInfo
	Services []ServiceInfo
}

// FormatDuration converts a duration to a human-readable string like "5d12h"
func FormatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd%dh", days, hours)
	} else if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// FormatPortsForDisplay formats service ports for display
func FormatPortsForDisplay(ports []ServicePort) string {
	if len(ports) == 0 {
		return ""
	}

	var result string
	for i, port := range ports {
		if i > 0 {
			result += ", "
		}

		if port.NodePort > 0 {
			result += fmt.Sprintf("%d:%d/%s", port.Port, port.NodePort, port.Protocol)
		} else {
			result += fmt.Sprintf("%d/%s", port.Port, port.Protocol)
		}
	}

	return result
}

// ServicePort represents a port mapping in a service
type ServicePort struct {
	Name       string
	Protocol   string
	Port       int32
	TargetPort int32
	NodePort   int32
}

// ContainerState represents the state of a container
type ContainerState string

const (
	// ContainerRunning means the container is currently running
	ContainerRunning ContainerState = "Running"

	// ContainerWaiting means the container is waiting to start
	ContainerWaiting ContainerState = "Waiting"

	// ContainerTerminated means the container has terminated
	ContainerTerminated ContainerState = "Terminated"
)
