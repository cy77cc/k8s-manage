package ai

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ListPodsRequest struct {
	Namespace string `json:"namespace" jsonschema:"description=The namespace to list pods from, default is 'default'"`
}

type ListPodsResponse struct {
	Pods []string `json:"pods"`
}

func NewK8sTools(clientset *kubernetes.Clientset) ([]tool.BaseTool, error) {
	listPodsFunc := func(ctx context.Context, req *ListPodsRequest) (*ListPodsResponse, error) {
		if clientset == nil {
			return nil, fmt.Errorf("kubernetes client not initialized")
		}
		ns := req.Namespace
		if ns == "" {
			ns = "default"
		}

		pods, err := clientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		var podNames []string
		for _, p := range pods.Items {
			status := string(p.Status.Phase)
			podNames = append(podNames, fmt.Sprintf("%s (%s)", p.Name, status))
		}
		return &ListPodsResponse{Pods: podNames}, nil
	}

	listPodsTool, err := utils.InferTool("list_pods", "List all pods in a specified namespace. Use this to check what applications are running.", listPodsFunc)
	if err != nil {
		return nil, err
	}

	// Describe Pod Tool
	describePodFunc := func(ctx context.Context, req *DescribePodRequest) (*DescribePodResponse, error) {
		if clientset == nil {
			return nil, fmt.Errorf("kubernetes client not initialized")
		}
		ns := req.Namespace
		if ns == "" {
			ns = "default"
		}

		pod, err := clientset.CoreV1().Pods(ns).Get(ctx, req.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		// Simple summary for now
		summary := fmt.Sprintf("Name: %s\nPhase: %s\nIP: %s\nNode: %s\nStart Time: %s",
			pod.Name, pod.Status.Phase, pod.Status.PodIP, pod.Spec.NodeName, pod.Status.StartTime)

		return &DescribePodResponse{Description: summary}, nil
	}

	describePodTool, err := utils.InferTool("describe_pod", "Get detailed information about a specific pod.", describePodFunc)
	if err != nil {
		return nil, err
	}

	return []tool.BaseTool{listPodsTool, describePodTool}, nil
}

type DescribePodRequest struct {
	Namespace string `json:"namespace" jsonschema:"description=The namespace of the pod"`
	Name      string `json:"name" jsonschema:"description=The name of the pod"`
}

type DescribePodResponse struct {
	Description string `json:"description"`
}
