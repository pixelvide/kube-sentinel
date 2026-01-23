package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pixelvide/cloud-sentinel-k8s/pkg/cluster"
	openai "github.com/sashabaranov/go-openai"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

type ClientSetKey struct{}

func GetClientSet(ctx context.Context) (*cluster.ClientSet, error) {
	cs, ok := ctx.Value("cluster_client").(*cluster.ClientSet)
	if !ok || cs == nil {
		klog.Warningf("K8s Tool: Kubernetes client not found in context (key: cluster_client)")
		return nil, fmt.Errorf("kubernetes client not found in context")
	}
	klog.V(2).Infof("K8s Tool: Found client for cluster %s", cs.Name)
	return cs, nil
}

// --- List Pods Tool ---

type ListPodsTool struct{}

func (t *ListPodsTool) Name() string { return "list_pods" }

func (t *ListPodsTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "list_pods",
			Description: "List pods in a namespace, optionally filtered by status",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"namespace": {
						"type": "string",
						"description": "The namespace to list pods from. If empty, lists from all namespaces."
					},
					"status_filter": {
						"type": "string",
						"enum": ["Running", "Pending", "Failed", "Succeeded", "Unknown"],
						"description": "Filter pods by status phase."
					}
				}
			}`),
		},
	}
}

func (t *ListPodsTool) Execute(ctx context.Context, args string) (string, error) {
	var params struct {
		Namespace    string `json:"namespace"`
		StatusFilter string `json:"status_filter"`
	}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", err
	}

	cs, err := GetClientSet(ctx)
	if err != nil {
		return "", err
	}

	pods, err := cs.K8sClient.ClientSet.CoreV1().Pods(params.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	var results []string
	for _, pod := range pods.Items {
		if params.StatusFilter != "" && string(pod.Status.Phase) != params.StatusFilter {
			continue
		}

		restarts := 0
		for _, status := range pod.Status.ContainerStatuses {
			restarts += int(status.RestartCount)
		}

		results = append(results, fmt.Sprintf("%s (Status: %s, Restarts: %d, IP: %s)",
			pod.Name, pod.Status.Phase, restarts, pod.Status.PodIP))
	}

	if len(results) == 0 {
		return "No pods found.", nil
	}

	// Limit output to prevent token overflow
	if len(results) > 50 {
		return strings.Join(results[:50], "\n") + fmt.Sprintf("\n... and %d more", len(results)-50), nil
	}

	return strings.Join(results, "\n"), nil
}

// --- Get Pod Logs Tool ---

type GetPodLogsTool struct{}

func (t *GetPodLogsTool) Name() string { return "get_pod_logs" }

func (t *GetPodLogsTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "get_pod_logs",
			Description: "Get logs from a specific pod",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"namespace": {
						"type": "string",
						"description": "The namespace of the pod."
					},
					"pod_name": {
						"type": "string",
						"description": "The name of the pod."
					},
					"container": {
						"type": "string",
						"description": "Optional container name."
					},
					"lines": {
						"type": "integer",
						"description": "Number of lines to retrieve (max 100). Defaults to 50."
					}
				},
				"required": ["namespace", "pod_name"]
			}`),
		},
	}
}

func (t *GetPodLogsTool) Execute(ctx context.Context, args string) (string, error) {
	var params struct {
		Namespace string `json:"namespace"`
		PodName   string `json:"pod_name"`
		Container string `json:"container"`
		Lines     int64  `json:"lines"`
	}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", err
	}

	if params.Lines <= 0 {
		params.Lines = 50
	}
	if params.Lines > 100 {
		params.Lines = 100
	}

	cs, err := GetClientSet(ctx)
	if err != nil {
		return "", err
	}

	opts := &corev1.PodLogOptions{
		TailLines: &params.Lines,
	}
	if params.Container != "" {
		opts.Container = params.Container
	}

	req := cs.K8sClient.ClientSet.CoreV1().Pods(params.Namespace).GetLogs(params.PodName, opts)
	logs, err := req.DoRaw(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %w", err)
	}

	return string(logs), nil
}

// --- Describe Resource Tool ---

type DescribeResourceTool struct{}

func (t *DescribeResourceTool) Name() string { return "describe_resource" }

func (t *DescribeResourceTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "describe_resource",
			Description: "Get details (JSON) of a specific resource",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"namespace": {
						"type": "string",
						"description": "The namespace of the resource."
					},
					"kind": {
						"type": "string",
						"description": "The kind of resource (Pod, Deployment, Service, etc)."
					},
					"name": {
						"type": "string",
						"description": "The name of the resource."
					}
				},
				"required": ["namespace", "kind", "name"]
			}`),
		},
	}
}

func (t *DescribeResourceTool) Execute(ctx context.Context, args string) (string, error) {
	var params struct {
		Namespace string `json:"namespace"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
	}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", err
	}

	cs, err := GetClientSet(ctx)
	if err != nil {
		return "", err
	}

	var obj interface{}
	var getErr error

	switch strings.ToLower(params.Kind) {
	case "pod":
		obj, getErr = cs.K8sClient.ClientSet.CoreV1().Pods(params.Namespace).Get(ctx, params.Name, metav1.GetOptions{})
	case "deployment":
		obj, getErr = cs.K8sClient.ClientSet.AppsV1().Deployments(params.Namespace).Get(ctx, params.Name, metav1.GetOptions{})
	case "service":
		obj, getErr = cs.K8sClient.ClientSet.CoreV1().Services(params.Namespace).Get(ctx, params.Name, metav1.GetOptions{})
	case "node":
		obj, getErr = cs.K8sClient.ClientSet.CoreV1().Nodes().Get(ctx, params.Name, metav1.GetOptions{})
	default:
		return "", fmt.Errorf("unsupported resource kind: %s", params.Kind)
	}

	if getErr != nil {
		return "", getErr
	}

	// Serialize to JSON for the LLM
	bytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// --- List Resources Tool ---

type ListResourcesTool struct{}

func (t *ListResourcesTool) Name() string { return "list_resources" }

func (t *ListResourcesTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "list_resources",
			Description: "List Kubernetes resources of a specific kind in a namespace, optionally filtered by labels.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"kind": {
						"type": "string",
						"description": "The kind of resource to list (Pod, Service, Deployment, ReplicaSet, StatefulSet, DaemonSet, Job, CronJob, ConfigMap, Secret, Namespace, Node, Ingress, Event)."
					},
					"namespace": {
						"type": "string",
						"description": "The namespace to list resources from. If empty, lists from all namespaces (if applicable)."
					},
					"label_selector": {
						"type": "string",
						"description": "Optional label selector to filter results (e.g., 'app=nginx')."
					}
				},
				"required": ["kind"]
			}`),
		},
	}
}

func (t *ListResourcesTool) Execute(ctx context.Context, args string) (string, error) {
	var params struct {
		Kind          string `json:"kind"`
		Namespace     string `json:"namespace"`
		LabelSelector string `json:"label_selector"`
	}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", err
	}

	cs, err := GetClientSet(ctx)
	if err != nil {
		return "", err
	}

	opts := metav1.ListOptions{
		LabelSelector: params.LabelSelector,
	}

	var results []string
	kind := strings.ToLower(params.Kind)

	switch kind {
	case "pod", "pods":
		list, err := cs.K8sClient.ClientSet.CoreV1().Pods(params.Namespace).List(ctx, opts)
		if err == nil {
			for _, item := range list.Items {
				results = append(results, fmt.Sprintf("%s/%s (Status: %s)", item.Namespace, item.Name, item.Status.Phase))
			}
		} else {
			return "", err
		}
	case "service", "services", "svc":
		list, err := cs.K8sClient.ClientSet.CoreV1().Services(params.Namespace).List(ctx, opts)
		if err == nil {
			for _, item := range list.Items {
				results = append(results, fmt.Sprintf("%s/%s (Type: %s, ClusterIP: %s)", item.Namespace, item.Name, item.Spec.Type, item.Spec.ClusterIP))
			}
		} else {
			return "", err
		}
	case "deployment", "deployments", "deploy":
		list, err := cs.K8sClient.ClientSet.AppsV1().Deployments(params.Namespace).List(ctx, opts)
		if err == nil {
			for _, item := range list.Items {
				results = append(results, fmt.Sprintf("%s/%s (Ready: %d/%d)", item.Namespace, item.Name, item.Status.ReadyReplicas, item.Status.Replicas))
			}
		} else {
			return "", err
		}
	case "replicaset", "replicasets", "rs":
		list, err := cs.K8sClient.ClientSet.AppsV1().ReplicaSets(params.Namespace).List(ctx, opts)
		if err == nil {
			for _, item := range list.Items {
				results = append(results, fmt.Sprintf("%s/%s (Replicas: %d)", item.Namespace, item.Name, item.Status.Replicas))
			}
		} else {
			return "", err
		}
	case "statefulset", "statefulsets", "sts":
		list, err := cs.K8sClient.ClientSet.AppsV1().StatefulSets(params.Namespace).List(ctx, opts)
		if err == nil {
			for _, item := range list.Items {
				results = append(results, fmt.Sprintf("%s/%s (Ready: %d/%d)", item.Namespace, item.Name, item.Status.ReadyReplicas, item.Status.Replicas))
			}
		} else {
			return "", err
		}
	case "daemonset", "daemonsets", "ds":
		list, err := cs.K8sClient.ClientSet.AppsV1().DaemonSets(params.Namespace).List(ctx, opts)
		if err == nil {
			for _, item := range list.Items {
				results = append(results, fmt.Sprintf("%s/%s (Desired: %d, Ready: %d)", item.Namespace, item.Name, item.Status.DesiredNumberScheduled, item.Status.NumberReady))
			}
		} else {
			return "", err
		}
	case "job", "jobs":
		list, err := cs.K8sClient.ClientSet.BatchV1().Jobs(params.Namespace).List(ctx, opts)
		if err == nil {
			for _, item := range list.Items {
				results = append(results, fmt.Sprintf("%s/%s (Succeeded: %d)", item.Namespace, item.Name, item.Status.Succeeded))
			}
		} else {
			return "", err
		}
	case "cronjob", "cronjobs":
		list, err := cs.K8sClient.ClientSet.BatchV1().CronJobs(params.Namespace).List(ctx, opts)
		if err == nil {
			for _, item := range list.Items {
				results = append(results, fmt.Sprintf("%s/%s (Schedule: %s)", item.Namespace, item.Name, item.Spec.Schedule))
			}
		} else {
			return "", err
		}
	case "configmap", "configmaps", "cm":
		list, err := cs.K8sClient.ClientSet.CoreV1().ConfigMaps(params.Namespace).List(ctx, opts)
		if err == nil {
			for _, item := range list.Items {
				results = append(results, fmt.Sprintf("%s/%s", item.Namespace, item.Name))
			}
		} else {
			return "", err
		}
	case "secret", "secrets":
		list, err := cs.K8sClient.ClientSet.CoreV1().Secrets(params.Namespace).List(ctx, opts)
		if err == nil {
			for _, item := range list.Items {
				results = append(results, fmt.Sprintf("%s/%s (Type: %s)", item.Namespace, item.Name, item.Type))
			}
		} else {
			return "", err
		}
	case "namespace", "namespaces", "ns":
		list, err := cs.K8sClient.ClientSet.CoreV1().Namespaces().List(ctx, opts)
		if err == nil {
			for _, item := range list.Items {
				results = append(results, fmt.Sprintf("%s (Status: %s)", item.Name, item.Status.Phase))
			}
		} else {
			return "", err
		}
	case "node", "nodes", "no":
		list, err := cs.K8sClient.ClientSet.CoreV1().Nodes().List(ctx, opts)
		if err == nil {
			for _, item := range list.Items {
				status := "NotReady"
				for _, cond := range item.Status.Conditions {
					if cond.Type == corev1.NodeReady && cond.Status == corev1.ConditionTrue {
						status = "Ready"
						break
					}
				}
				results = append(results, fmt.Sprintf("%s (Status: %s, Version: %s)", item.Name, status, item.Status.NodeInfo.KubeletVersion))
			}
		} else {
			return "", err
		}
	case "ingress", "ingresses", "ing":
		list, err := cs.K8sClient.ClientSet.NetworkingV1().Ingresses(params.Namespace).List(ctx, opts)
		if err == nil {
			for _, item := range list.Items {
				results = append(results, fmt.Sprintf("%s/%s", item.Namespace, item.Name))
			}
		} else {
			return "", err
		}
	case "event", "events", "ev":
		list, err := cs.K8sClient.ClientSet.CoreV1().Events(params.Namespace).List(ctx, opts)
		if err == nil {
			for _, item := range list.Items {
				results = append(results, fmt.Sprintf("%s/%s (%s: %s)", item.Namespace, item.InvolvedObject.Name, item.Type, item.Message))
			}
		} else {
			return "", err
		}
	default:
		return "", fmt.Errorf("unsupported resource kind for listing: %s", params.Kind)
	}

	if len(results) == 0 {
		return fmt.Sprintf("No %s found.", params.Kind), nil
	}

	// Limit output to prevent token overflow
	if len(results) > 100 {
		return strings.Join(results[:100], "\n") + fmt.Sprintf("\n... and %d more", len(results)-100), nil
	}

	return strings.Join(results, "\n"), nil
}
