package resources

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetPods retrieves pods from the specified namespace
func GetPods(clientset *kubernetes.Clientset, namespace string) ([]PodInfo, error) {
	var pods []PodInfo

	// Get pod list from K8s API
	podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error fetching pods: %v", err)
	}

	// Process each pod
	for _, pod := range podList.Items {
		// Calculate pod age
		age := time.Since(pod.CreationTimestamp.Time).Round(time.Second)
		ageStr := FormatDuration(age)

		// Process container information
		containers := make([]ContainerInfo, 0, len(pod.Spec.Containers))
		for _, container := range pod.Spec.Containers {
			// Get container status
			var ready bool
			var state string
			var restartCount int32

			for _, status := range pod.Status.ContainerStatuses {
				if status.Name == container.Name {
					ready = status.Ready
					restartCount = status.RestartCount

					if status.State.Running != nil {
						state = string(ContainerRunning)
					} else if status.State.Waiting != nil {
						state = string(ContainerWaiting)
					} else if status.State.Terminated != nil {
						state = string(ContainerTerminated)
					}

					break
				}
			}

			// Process resource requests and limits
			cpuRequest := ""
			memRequest := ""
			cpuLimit := ""
			memLimit := ""

			if container.Resources.Requests != nil {
				if cpu, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
					cpuRequest = cpu.String()
				}
				if mem, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
					memRequest = mem.String()
				}
			}

			if container.Resources.Limits != nil {
				if cpu, ok := container.Resources.Limits[corev1.ResourceCPU]; ok {
					cpuLimit = cpu.String()
				}
				if mem, ok := container.Resources.Limits[corev1.ResourceMemory]; ok {
					memLimit = mem.String()
				}
			}

			// Process environment variables
			envVars := make(map[string]string)
			for _, env := range container.Env {
				if env.Value != "" {
					envVars[env.Name] = env.Value
				} else if env.ValueFrom != nil {
					envVars[env.Name] = "[from source]"
				}
			}

			// Create container info
			containers = append(containers, ContainerInfo{
				Name:            container.Name,
				Image:           container.Image,
				Ready:           ready,
				RestartCount:    int(restartCount),
				State:           state,
				CPURequest:      cpuRequest,
				MemoryRequest:   memRequest,
				CPULimit:        cpuLimit,
				MemoryLimit:     memLimit,
				EnvironmentVars: envVars,
			})
		}

		// Create pod info
		podInfo := PodInfo{
			Name:       pod.Name,
			Namespace:  pod.Namespace,
			Status:     string(pod.Status.Phase),
			Age:        ageStr,
			IP:         pod.Status.PodIP,
			Node:       pod.Spec.NodeName,
			Created:    pod.CreationTimestamp.Time,
			Labels:     pod.Labels,
			Containers: containers,
		}

		pods = append(pods, podInfo)
	}

	return pods, nil
}

// GetPodDetail returns detailed information about a specific pod
func GetPodDetail(clientset *kubernetes.Clientset, namespace, podName string) (string, error) {
	// Get the pod from the API
	pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("error fetching pod details: %v", err)
	}

	// Build the detail string
	var sb strings.Builder

	// Basic pod information
	sb.WriteString(fmt.Sprintf("Pod: %s\n", pod.Name))
	sb.WriteString(fmt.Sprintf("Namespace: %s\n", pod.Namespace))
	sb.WriteString(fmt.Sprintf("Status: %s\n", pod.Status.Phase))
	sb.WriteString(fmt.Sprintf("IP: %s\n", pod.Status.PodIP))
	sb.WriteString(fmt.Sprintf("Node: %s\n", pod.Spec.NodeName))
	sb.WriteString(fmt.Sprintf("Created: %s\n", pod.CreationTimestamp.Format(time.RFC3339)))

	// Labels
	if len(pod.Labels) > 0 {
		sb.WriteString("\nLabels:\n")
		for key, value := range pod.Labels {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	// Container details
	sb.WriteString("\nContainers:\n")
	for _, container := range pod.Spec.Containers {
		sb.WriteString(fmt.Sprintf("  - %s (Image: %s)\n", container.Name, container.Image))

		// Resource requests and limits
		if container.Resources.Requests != nil || container.Resources.Limits != nil {
			sb.WriteString("    Resources:\n")
			if cpu, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
				sb.WriteString(fmt.Sprintf("      CPU Request: %s\n", cpu.String()))
			}
			if mem, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
				sb.WriteString(fmt.Sprintf("      Memory Request: %s\n", mem.String()))
			}
			if cpu, ok := container.Resources.Limits[corev1.ResourceCPU]; ok {
				sb.WriteString(fmt.Sprintf("      CPU Limit: %s\n", cpu.String()))
			}
			if mem, ok := container.Resources.Limits[corev1.ResourceMemory]; ok {
				sb.WriteString(fmt.Sprintf("      Memory Limit: %s\n", mem.String()))
			}
		}

		// Container status
		for _, status := range pod.Status.ContainerStatuses {
			if status.Name == container.Name {
				sb.WriteString(fmt.Sprintf("    Status:\n"))
				sb.WriteString(fmt.Sprintf("      Ready: %v\n", status.Ready))
				sb.WriteString(fmt.Sprintf("      Restart Count: %d\n", status.RestartCount))

				if status.State.Running != nil {
					sb.WriteString(fmt.Sprintf("      State: Running (started at %s)\n",
						status.State.Running.StartedAt.Format(time.RFC3339)))
				} else if status.State.Waiting != nil {
					sb.WriteString(fmt.Sprintf("      State: Waiting (reason: %s)\n",
						status.State.Waiting.Reason))
					if status.State.Waiting.Message != "" {
						sb.WriteString(fmt.Sprintf("      Message: %s\n", status.State.Waiting.Message))
					}
				} else if status.State.Terminated != nil {
					sb.WriteString(fmt.Sprintf("      State: Terminated (reason: %s)\n",
						status.State.Terminated.Reason))
					if status.State.Terminated.Message != "" {
						sb.WriteString(fmt.Sprintf("      Message: %s\n", status.State.Terminated.Message))
					}
				}

				break
			}
		}
	}

	// Environment variables
	sb.WriteString("\nEnvironment Variables:\n")
	for _, container := range pod.Spec.Containers {
		sb.WriteString(fmt.Sprintf("  %s:\n", container.Name))
		if len(container.Env) == 0 {
			sb.WriteString("    No environment variables defined\n")
		} else {
			for _, env := range container.Env {
				if env.Value != "" {
					sb.WriteString(fmt.Sprintf("    - %s: %s\n", env.Name, env.Value))
				} else if env.ValueFrom != nil {
					var source string
					if env.ValueFrom.ConfigMapKeyRef != nil {
						source = fmt.Sprintf("ConfigMap %s (key: %s)",
							env.ValueFrom.ConfigMapKeyRef.Name, env.ValueFrom.ConfigMapKeyRef.Key)
					} else if env.ValueFrom.SecretKeyRef != nil {
						source = fmt.Sprintf("Secret %s (key: %s)",
							env.ValueFrom.SecretKeyRef.Name, env.ValueFrom.SecretKeyRef.Key)
					} else if env.ValueFrom.FieldRef != nil {
						source = fmt.Sprintf("Field %s", env.ValueFrom.FieldRef.FieldPath)
					} else {
						source = "Unknown source"
					}
					sb.WriteString(fmt.Sprintf("    - %s: [from %s]\n", env.Name, source))
				}
			}
		}
	}

	// Volumes
	if len(pod.Spec.Volumes) > 0 {
		sb.WriteString("\nVolumes:\n")
		for _, volume := range pod.Spec.Volumes {
			sb.WriteString(fmt.Sprintf("  - %s:\n", volume.Name))

			if volume.PersistentVolumeClaim != nil {
				sb.WriteString(fmt.Sprintf("    Type: PersistentVolumeClaim\n"))
				sb.WriteString(fmt.Sprintf("    Claim Name: %s\n", volume.PersistentVolumeClaim.ClaimName))
			} else if volume.ConfigMap != nil {
				sb.WriteString(fmt.Sprintf("    Type: ConfigMap\n"))
				sb.WriteString(fmt.Sprintf("    Name: %s\n", volume.ConfigMap.Name))
			} else if volume.Secret != nil {
				sb.WriteString(fmt.Sprintf("    Type: Secret\n"))
				sb.WriteString(fmt.Sprintf("    Secret Name: %s\n", volume.Secret.SecretName))
			} else if volume.EmptyDir != nil {
				sb.WriteString(fmt.Sprintf("    Type: EmptyDir\n"))
			} else if volume.HostPath != nil {
				sb.WriteString(fmt.Sprintf("    Type: HostPath\n"))
				sb.WriteString(fmt.Sprintf("    Path: %s\n", volume.HostPath.Path))
			} else {
				sb.WriteString(fmt.Sprintf("    Type: Other\n"))
			}
		}
	}

	// Events (would require additional API calls, simplified version here)
	sb.WriteString("\nUse 'kubectl describe pod' for events and additional information\n")

	return sb.String(), nil
}
