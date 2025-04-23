package client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/zvelocity/k8s-cli/internal/resources"
)

// K8sClient wraps kubernetes clientset with helper methods
type K8sClient struct {
	Clientset *kubernetes.Clientset
}

// New creates a new K8sClient
func New() (*K8sClient, error) {
	// Find kubeconfig
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		// Try default location if not specified
		homeDir, err := os.UserHomeDir()
		if err == nil {
			kubeconfig = filepath.Join(homeDir, ".kube", "config")
		}
	}

	// Build config from kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error building kubeconfig: %v", err)
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating Kubernetes client: %v", err)
	}

	return &K8sClient{
		Clientset: clientset,
	}, nil
}

// GetNamespaces returns all namespaces in the cluster
func (c *K8sClient) GetNamespaces() ([]string, error) {
	// Get namespace list from K8s API
	nsList, err := c.Clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error fetching namespaces: %v", err)
	}

	// Extract namespace names
	var namespaces []string
	for _, ns := range nsList.Items {
		namespaces = append(namespaces, ns.Name)
	}

	return namespaces, nil
}

// GetPods returns pods in the given namespace
func (c *K8sClient) GetPods(namespace string) ([]resources.PodInfo, error) {
	return resources.GetPods(c.Clientset, namespace)
}

// GetServices returns services in the given namespace
func (c *K8sClient) GetServices(namespace string) ([]resources.ServiceInfo, error) {
	return resources.GetServices(c.Clientset, namespace)
}

// GetPodDetail returns detailed info for a pod
func (c *K8sClient) GetPodDetail(namespace, name string) (string, error) {
	return resources.GetPodDetail(c.Clientset, namespace, name)
}

// GetServiceDetail returns detailed info for a service
func (c *K8sClient) GetServiceDetail(namespace, name string) (string, error) {
	return resources.GetServiceDetail(c.Clientset, namespace, name)
}

// GetCurrentContext returns the current Kubernetes context name
func (c *K8sClient) GetCurrentContext() (string, error) {
	// Load kubeconfig
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		// Try default location if not specified
		homeDir, err := os.UserHomeDir()
		if err == nil {
			kubeconfig = filepath.Join(homeDir, ".kube", "config")
		}
	}

	// Load client config
	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return "", fmt.Errorf("error loading kubeconfig: %v", err)
	}

	return config.CurrentContext, nil
}
