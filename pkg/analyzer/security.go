package analyzer

import (
	"context"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/prometheus"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ImmutableTagAnalyzer detects usage of 'latest' tag or missing tags
type ImmutableTagAnalyzer struct{}

func (a *ImmutableTagAnalyzer) Name() string { return "ImmutableTags" }

func (a *ImmutableTagAnalyzer) Analyze(ctx context.Context, c client.Client, promClient *prometheus.Client, obj client.Object) ([]Anomaly, error) {
	var containers []corev1.Container
	var initContainers []corev1.Container
	var name string

	switch o := obj.(type) {
	case *appsv1.Deployment:
		containers = o.Spec.Template.Spec.Containers
		initContainers = o.Spec.Template.Spec.InitContainers
		name = o.Name
	case *appsv1.StatefulSet:
		containers = o.Spec.Template.Spec.Containers
		initContainers = o.Spec.Template.Spec.InitContainers
		name = o.Name
	case *appsv1.DaemonSet:
		containers = o.Spec.Template.Spec.Containers
		initContainers = o.Spec.Template.Spec.InitContainers
		name = o.Name
	case *corev1.Pod:
		containers = o.Spec.Containers
		initContainers = o.Spec.InitContainers
		name = o.Name
	default:
		return nil, nil
	}

	var anomalies []Anomaly
	var allContainers []corev1.Container
	allContainers = append(allContainers, containers...)
	allContainers = append(allContainers, initContainers...)

	for _, c := range allContainers {
		image := c.Image
		// Check if tag is 'latest' or missing (which implies latest)
		// Examples: "nginx", "nginx:latest", "my-registry.io/img:latest"
		isLatest := false
		if strings.HasSuffix(image, ":latest") {
			isLatest = true
		} else if !strings.Contains(image, ":") {
			// No tag specified usually means latest, unless it has a digest (sha256:...)
			// "ubuntu@sha256:..." is valid and immutable.
			if !strings.Contains(image, "@") {
				isLatest = true
			}
		}

		if isLatest {
			anomalies = append(anomalies, Anomaly{
				Severity:    SeverityMedium,
				Title:       "Mutable Image Tag Detected",
				Message:     fmt.Sprintf("Container '%s' in '%s' is using a mutable image tag: '%s'.", c.Name, name, image),
				Remediation: "Use a specific version tag (e.g., :v1.0.0) or digest (@sha256:...) to ensure immutability and reproducible deployments.",
				RuleID:      "SEC-001",
			})
		}
	}

	return anomalies, nil
}

// PrivilegedContainerAnalyzer detects containers running with privileged: true
type PrivilegedContainerAnalyzer struct{}

func (a *PrivilegedContainerAnalyzer) Name() string { return "PrivilegedContainer" }

func (a *PrivilegedContainerAnalyzer) Analyze(ctx context.Context, c client.Client, promClient *prometheus.Client, obj client.Object) ([]Anomaly, error) {
	var containers []corev1.Container
	var initContainers []corev1.Container
	var name string

	switch o := obj.(type) {
	case *appsv1.Deployment:
		containers = o.Spec.Template.Spec.Containers
		initContainers = o.Spec.Template.Spec.InitContainers
		name = o.Name
	case *appsv1.StatefulSet:
		containers = o.Spec.Template.Spec.Containers
		initContainers = o.Spec.Template.Spec.InitContainers
		name = o.Name
	case *appsv1.DaemonSet:
		containers = o.Spec.Template.Spec.Containers
		initContainers = o.Spec.Template.Spec.InitContainers
		name = o.Name
	case *corev1.Pod:
		containers = o.Spec.Containers
		initContainers = o.Spec.InitContainers
		name = o.Name
	default:
		return nil, nil
	}

	var anomalies []Anomaly
	var allContainers []corev1.Container
	allContainers = append(allContainers, containers...)
	allContainers = append(allContainers, initContainers...)

	for _, c := range allContainers {
		if c.SecurityContext != nil && c.SecurityContext.Privileged != nil && *c.SecurityContext.Privileged {
			anomalies = append(anomalies, Anomaly{
				Severity:    SeverityHigh,
				Title:       "Privileged Container Detected",
				Message:     fmt.Sprintf("Container '%s' in '%s' is running in privileged mode.", c.Name, name),
				Remediation: "Avoid running containers as privileged unless absolutely necessary. Grant specific capabilities instead.",
				RuleID:      "SEC-002",
			})
		}
	}

	return anomalies, nil
}

// RootUserAnalyzer detects containers running as root (UID 0)
type RootUserAnalyzer struct{}

func (a *RootUserAnalyzer) Name() string { return "RootUser" }

func (a *RootUserAnalyzer) Analyze(ctx context.Context, c client.Client, promClient *prometheus.Client, obj client.Object) ([]Anomaly, error) {
	var containers []corev1.Container
	var initContainers []corev1.Container
	var podSecurityContext *corev1.PodSecurityContext
	var name string

	switch o := obj.(type) {
	case *appsv1.Deployment:
		containers = o.Spec.Template.Spec.Containers
		initContainers = o.Spec.Template.Spec.InitContainers
		podSecurityContext = o.Spec.Template.Spec.SecurityContext
		name = o.Name
	case *appsv1.StatefulSet:
		containers = o.Spec.Template.Spec.Containers
		initContainers = o.Spec.Template.Spec.InitContainers
		podSecurityContext = o.Spec.Template.Spec.SecurityContext
		name = o.Name
	case *appsv1.DaemonSet:
		containers = o.Spec.Template.Spec.Containers
		initContainers = o.Spec.Template.Spec.InitContainers
		podSecurityContext = o.Spec.Template.Spec.SecurityContext
		name = o.Name
	case *corev1.Pod:
		containers = o.Spec.Containers
		initContainers = o.Spec.InitContainers
		podSecurityContext = o.Spec.SecurityContext
		name = o.Name
	default:
		return nil, nil
	}

	var anomalies []Anomaly
	var allContainers []corev1.Container
	allContainers = append(allContainers, containers...)
	allContainers = append(allContainers, initContainers...)

	podRunAsNonRoot := false
	podRunAsUser := int64(-1)

	if podSecurityContext != nil {
		if podSecurityContext.RunAsNonRoot != nil && *podSecurityContext.RunAsNonRoot {
			podRunAsNonRoot = true
		}
		if podSecurityContext.RunAsUser != nil {
			podRunAsUser = *podSecurityContext.RunAsUser
		}
	}

	for _, c := range allContainers {
		// Container Security Context overrides Pod Security Context
		runAsNonRoot := podRunAsNonRoot
		runAsUser := podRunAsUser

		if c.SecurityContext != nil {
			if c.SecurityContext.RunAsNonRoot != nil {
				runAsNonRoot = *c.SecurityContext.RunAsNonRoot
			}
			if c.SecurityContext.RunAsUser != nil {
				runAsUser = *c.SecurityContext.RunAsUser
			}
		}

		isRoot := false
		if runAsUser == 0 {
			isRoot = true
		} else if runAsUser == -1 && !runAsNonRoot {
			// If no user specified and not enforced non-root, it acts as root (UID 0) in standard Docker images.
			isRoot = true
		}

		if isRoot {
			msg := fmt.Sprintf("Container '%s' may be running as root.", c.Name)
			if runAsUser == 0 {
				msg = fmt.Sprintf("Container '%s' is explicitly configured to run as root (UID 0).", c.Name)
			}

			anomalies = append(anomalies, Anomaly{
				Severity:    SeverityMedium,
				Title:       "Container Running as Root",
				Message:     fmt.Sprintf("%s (Resource: '%s')", msg, name),
				Remediation: "Configure 'securityContext.runAsUser' to a non-zero ID or set 'securityContext.runAsNonRoot: true' to improve security isolation.",
				RuleID:      "SEC-003",
			})
		}
	}

	return anomalies, nil
}

// HostPathAnalyzer detects usage of hostPath volumes
type HostPathAnalyzer struct{}

func (a *HostPathAnalyzer) Name() string { return "HostPathVolume" }

func (a *HostPathAnalyzer) Analyze(ctx context.Context, c client.Client, promClient *prometheus.Client, obj client.Object) ([]Anomaly, error) {
	var volumes []corev1.Volume
	var name string

	switch o := obj.(type) {
	case *appsv1.Deployment:
		volumes = o.Spec.Template.Spec.Volumes
		name = o.Name
	case *appsv1.StatefulSet:
		volumes = o.Spec.Template.Spec.Volumes
		name = o.Name
	case *appsv1.DaemonSet:
		volumes = o.Spec.Template.Spec.Volumes
		name = o.Name
	case *corev1.Pod:
		volumes = o.Spec.Volumes
		name = o.Name
	default:
		return nil, nil
	}

	var anomalies []Anomaly

	for _, v := range volumes {
		if v.HostPath != nil {
			anomalies = append(anomalies, Anomaly{
				Severity:    SeverityHigh,
				Title:       "HostPath Volume Detected",
				Message:     fmt.Sprintf("Volume '%s' in '%s' is using hostPath.", v.Name, name),
				Remediation: "Avoid using hostPath volumes as they present security risks and limit pod portability. Use PVCs or emptyDir instead.",
				RuleID:      "SEC-004",
			})
		}
	}

	return anomalies, nil
}

func init() {
	Register(&ImmutableTagAnalyzer{})
	Register(&PrivilegedContainerAnalyzer{})
	Register(&RootUserAnalyzer{})
	Register(&HostPathAnalyzer{})
}
