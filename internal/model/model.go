package model

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"zvelocity/k8s-cli/internal/client"
	"zvelocity/k8s-cli/internal/resources"
	"zvelocity/k8s-cli/internal/ui"
)

// Model is the main application model
type Model struct {
	// UI State
	spinner      spinner.Model
	loading      bool
	currentView  resources.ViewType
	selectedItem int
	width        int
	height       int
	message      string
	error        string

	// Data
	client        *client.K8sClient
	namespaces    []string
	currentNS     string
	context       string
	resourceData  resources.ResourceData
	detailContent string
}

// New creates a new model
func New() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = ui.StatusStyle

	return Model{
		spinner:      s,
		loading:      true,
		currentView:  resources.PodView,
		selectedItem: 0,
		currentNS:    "default",
		message:      "Connecting to Kubernetes cluster...",
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		initK8sClient,
	)
}

// Update handles messages and updates model state
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "p":
			if !m.loading {
				m.currentView = resources.PodView
				m.selectedItem = 0
			}

		case "s":
			if !m.loading {
				m.currentView = resources.ServiceView
				m.selectedItem = 0
			}

		case "esc":
			if m.currentView == resources.DetailView {
				m.currentView = resources.PodView
			} else if m.currentView == resources.NamespaceView {
				m.currentView = resources.PodView
			}

		case "up", "k":
			if !m.loading {
				if m.selectedItem > 0 {
					m.selectedItem--
				}
			}

		case "down", "j":
			if !m.loading {
				switch m.currentView {
				case resources.PodView:
					if m.selectedItem < len(m.resourceData.Pods)-1 {
						m.selectedItem++
					}
				case resources.ServiceView:
					if m.selectedItem < len(m.resourceData.Services)-1 {
						m.selectedItem++
					}
				case resources.NamespaceView:
					if m.selectedItem < len(m.namespaces)-1 {
						m.selectedItem++
					}
				}
			}

		case "enter":
			if !m.loading {
				switch m.currentView {
				case resources.PodView:
					if len(m.resourceData.Pods) > 0 {
						m.currentView = resources.DetailView
						m.loading = true
						selectedPod := m.resourceData.Pods[m.selectedItem]
						return m, tea.Batch(
							m.spinner.Tick,
							getPodDetail(m.client, selectedPod.Namespace, selectedPod.Name),
						)
					}
				case resources.ServiceView:
					if len(m.resourceData.Services) > 0 {
						m.currentView = resources.DetailView
						m.loading = true
						selectedSvc := m.resourceData.Services[m.selectedItem]
						return m, tea.Batch(
							m.spinner.Tick,
							getServiceDetail(m.client, selectedSvc.Namespace, selectedSvc.Name),
						)
					}
				case resources.NamespaceView:
					if len(m.namespaces) > 0 {
						m.currentNS = m.namespaces[m.selectedItem]
						m.currentView = resources.PodView
						m.loading = true
						m.message = fmt.Sprintf("Switching to namespace: %s", m.currentNS)
						return m, tea.Batch(
							m.spinner.Tick,
							getResources(m.client, m.currentNS),
						)
					}
				}
			}

		case "r":
			if !m.loading {
				m.loading = true
				m.message = "Refreshing resources..."
				return m, tea.Batch(
					m.spinner.Tick,
					getResources(m.client, m.currentNS),
				)
			}

		case "n":
			if !m.loading {
				m.currentView = resources.NamespaceView
				// Find current namespace in list
				for i, ns := range m.namespaces {
					if ns == m.currentNS {
						m.selectedItem = i
						break
					}
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case k8sClientMsg:
		if msg.err != nil {
			m.loading = false
			m.error = fmt.Sprintf("Error connecting to Kubernetes: %v", msg.err)
			return m, nil
		}
		m.client = msg.client
		m.message = "Getting context information..."
		return m, getContextInfo(m.client)

	case contextInfoMsg:
		if msg.err != nil {
			m.context = "unknown-context"
		} else {
			m.context = msg.context
		}
		m.message = "Fetching namespaces..."
		return m, getNamespaces(m.client)

	case namespacesMsg:
		if msg.err != nil {
			m.loading = false
			m.error = fmt.Sprintf("Error fetching namespaces: %v", msg.err)
			return m, nil
		}
		m.namespaces = msg.namespaces
		m.message = "Fetching resources..."
		return m, getResources(m.client, m.currentNS)

	case resourcesMsg:
		m.loading = false
		if msg.err != nil {
			m.error = fmt.Sprintf("Error fetching resources: %v", msg.err)
			return m, nil
		}
		m.resourceData = msg.data
		return m, nil

	case podDetailMsg:
		m.loading = false
		if msg.err != nil {
			m.error = fmt.Sprintf("Error fetching pod details: %v", msg.err)
			return m, nil
		}
		m.detailContent = msg.detail
		return m, nil

	case serviceDetailMsg:
		m.loading = false
		if msg.err != nil {
			m.error = fmt.Sprintf("Error fetching service details: %v", msg.err)
			return m, nil
		}
		m.detailContent = msg.detail
		return m, nil
	}

	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

// View renders the current view
func (m Model) View() string {
	if m.loading {
		return ui.RenderLoadingView(m.spinner.View(), m.message)
	}

	if m.error != "" {
		return ui.RenderErrorView(m.error)
	}

	// Add context information to title
	contextInfo := fmt.Sprintf(" (Context: %s)", m.context)

	switch m.currentView {
	case resources.PodView:
		return ui.RenderPodsView(m.resourceData.Pods, m.selectedItem, m.currentNS) + contextInfo
	case resources.ServiceView:
		return ui.RenderServicesView(m.resourceData.Services, m.selectedItem, m.currentNS) + contextInfo
	case resources.DetailView:
		return ui.RenderPodDetailView(m.detailContent)
	case resources.NamespaceView:
		return ui.RenderNamespacesView(m.namespaces, m.selectedItem)
	default:
		return "Unknown view"
	}
}

// Message types and commands
type k8sClientMsg struct {
	client *client.K8sClient
	err    error
}

func initK8sClient() tea.Msg {
	client, err := client.New()
	return k8sClientMsg{client, err}
}

type contextInfoMsg struct {
	context string
	err     error
}

func getContextInfo(client *client.K8sClient) tea.Cmd {
	return func() tea.Msg {
		context, err := client.GetCurrentContext()
		return contextInfoMsg{context, err}
	}
}

type namespacesMsg struct {
	namespaces []string
	err        error
}

func getNamespaces(client *client.K8sClient) tea.Cmd {
	return func() tea.Msg {
		namespaces, err := client.GetNamespaces()
		return namespacesMsg{namespaces, err}
	}
}

type resourcesMsg struct {
	data resources.ResourceData
	err  error
}

func getResources(client *client.K8sClient, namespace string) tea.Cmd {
	return func() tea.Msg {
		data := resources.ResourceData{}

		// Get pods
		pods, err := client.GetPods(namespace)
		if err != nil {
			return resourcesMsg{data, err}
		}
		data.Pods = pods

		// Get services
		services, err := client.GetServices(namespace)
		if err != nil {
			return resourcesMsg{data, err}
		}
		data.Services = services

		return resourcesMsg{data, nil}
	}
}

type podDetailMsg struct {
	detail string
	err    error
}

func getPodDetail(client *client.K8sClient, namespace, name string) tea.Cmd {
	return func() tea.Msg {
		detail, err := client.GetPodDetail(namespace, name)
		return podDetailMsg{detail, err}
	}
}

type serviceDetailMsg struct {
	detail string
	err    error
}

func getServiceDetail(client *client.K8sClient, namespace, name string) tea.Cmd {
	return func() tea.Msg {
		detail, err := client.GetServiceDetail(namespace, name)
		return serviceDetailMsg{detail, err}
	}
}
