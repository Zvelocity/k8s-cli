package resources

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetServices retrieves services from the specified namespace
func GetServices(clientset *kubernetes.Clientset, namespace string) ([]ServiceInfo, error) {
	var services []ServiceInfo

	// Get service list from K8s API
	serviceList, err := clientset.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error fetching services: %v", err)
	}

	// Process each service
	for _, svc := range serviceList.Items {
		// Calculate service age
		age := time.Since(svc.CreationTimestamp.Time).Round(time.Second)
		ageStr := FormatDuration(age)

		// Process ports
		var ports []ServicePort
		for _, port := range svc.Spec.Ports {
			svcPort := ServicePort{
				Name:       port.Name,
				Protocol:   string(port.Protocol),
				Port:       port.Port,
				TargetPort: port.TargetPort.IntVal,
				NodePort:   port.NodePort,
			}
			ports = append(ports, svcPort)
		}

		// Format external IP
		externalIP := "<none>"
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			if ip := svc.Status.LoadBalancer.Ingress[0].IP; ip != "" {
				externalIP = ip
			} else if hostname := svc.Status.LoadBalancer.Ingress[0].Hostname; hostname != "" {
				externalIP = hostname
			}
		} else if svc.Spec.Type == corev1.ServiceTypeNodePort || svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
			externalIP = "<pending>"
		}

		// Create service info
		serviceInfo := ServiceInfo{
			Name:       svc.Name,
			Namespace:  svc.Namespace,
			Type:       string(svc.Spec.Type),
			ClusterIP:  svc.Spec.ClusterIP,
			ExternalIP: externalIP,
			Ports:      FormatPortsForDisplay(ports),
			Age:        ageStr,
			Selector:   svc.Spec.Selector,
		}

		services = append(services, serviceInfo)
	}

	return services, nil
}

// GetServiceDetail returns detailed information about a specific service
func GetServiceDetail(clientset *kubernetes.Clientset, namespace, serviceName string) (string, error) {
	// Get the service from the API
	svc, err := clientset.CoreV1().Services(namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("error fetching service details: %v", err)
	}

	// Format external IP
	externalIP := "<none>"
	if len(svc.Status.LoadBalancer.Ingress) > 0 {
		if ip := svc.Status.LoadBalancer.Ingress[0].IP; ip != "" {
			externalIP = ip
		} else if hostname := svc.Status.LoadBalancer.Ingress[0].Hostname; hostname != "" {
			externalIP = hostname
		}
	} else if svc.Spec.Type == corev1.ServiceTypeNodePort || svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
		externalIP = "<pending>"
	}

	// Process ports
	var ports []ServicePort
	for _, port := range svc.Spec.Ports {
		svcPort := ServicePort{
			Name:       port.Name,
			Protocol:   string(port.Protocol),
			Port:       port.Port,
			TargetPort: port.TargetPort.IntVal,
			NodePort:   port.NodePort,
		}
		ports = append(ports, svcPort)
	}

	// Build the detail string
	detail := fmt.Sprintf("Service: %s\n", svc.Name)
	detail += fmt.Sprintf("Namespace: %s\n", svc.Namespace)
	detail += fmt.Sprintf("Type: %s\n", svc.Spec.Type)
	detail += fmt.Sprintf("Cluster IP: %s\n", svc.Spec.ClusterIP)
	detail += fmt.Sprintf("External IP: %s\n", externalIP)

	// Format ports
	detail += "\nPorts:\n"
	if len(svc.Spec.Ports) == 0 {
		detail += "  No ports defined\n"
	} else {
		for _, port := range svc.Spec.Ports {
			if port.NodePort > 0 {
				detail += fmt.Sprintf("  - %d:%d/%s", port.Port, port.NodePort, port.Protocol)
			} else {
				detail += fmt.Sprintf("  - %d/%s", port.Port, port.Protocol)
			}

			if port.Name != "" {
				detail += fmt.Sprintf(" (name: %s)", port.Name)
			}

			detail += "\n"
		}
	}

	// Selectors
	detail += "\nSelector:\n"
	if len(svc.Spec.Selector) == 0 {
		detail += "  No selector defined\n"
	} else {
		for key, value := range svc.Spec.Selector {
			detail += fmt.Sprintf("  %s: %s\n", key, value)
		}
	}

	// Session affinity
	detail += fmt.Sprintf("\nSession Affinity: %s\n", svc.Spec.SessionAffinity)

	// Labels
	if len(svc.Labels) > 0 {
		detail += "\nLabels:\n"
		for key, value := range svc.Labels {
			detail += fmt.Sprintf("  %s: %s\n", key, value)
		}
	}

	// Annotations
	if len(svc.Annotations) > 0 {
		detail += "\nAnnotations:\n"
		for key, value := range svc.Annotations {
			detail += fmt.Sprintf("  %s: %s\n", key, value)
		}
	}

	// Creation timestamp
	detail += fmt.Sprintf("\nCreated: %s\n", svc.CreationTimestamp.Format(time.RFC3339))

	return detail, nil
}
