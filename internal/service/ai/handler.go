package ai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/cloudwego/eino/schema"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ChatRequest struct {
	Message string `json:"message" binding:"required"`
}

func ChatHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		if svcCtx.AI == nil || svcCtx.AI.Runnable == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI service not enabled"})
			return
		}

		var req ChatRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx := c.Request.Context()

		// Prepare input with context
		clusterInfo := getClusterInfo(ctx, svcCtx.Clientset)
		systemPrompt := fmt.Sprintf("You are a Kubernetes Expert Assistant.\n"+
			"Your goal is to help users manage and troubleshoot their Kubernetes clusters.\n\n"+
			"Current Cluster Context:\n"+
			"%s\n\n"+
			"You have access to tools to query the cluster state (list_pods, etc.). "+
			"Use them when necessary to answer the user's question accurately. "+
			"If you use a tool, answer based on the tool's output.", clusterInfo)

		input := []*schema.Message{
			schema.SystemMessage(systemPrompt),
			schema.UserMessage(req.Message),
		}

		// Set headers for SSE
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")

		s, err := svcCtx.AI.Runnable.Stream(ctx, input)
		if err != nil {
			c.SSEvent("error", err.Error())
			return
		}
		defer s.Close()

		for {
			chunk, err := s.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				c.SSEvent("error", err.Error())
				break
			}

			c.SSEvent("message", chunk.Content)
			c.Writer.Flush()
		}
	}
}

func getClusterInfo(ctx context.Context, clientset *kubernetes.Clientset) string {
	if clientset == nil {
		return "Kubernetes Cluster: Not Connected (Clientset is nil)"
	}

	var sb strings.Builder
	sb.WriteString("Cluster Status Overview:\n")

	// Get Server Version
	if v, err := clientset.Discovery().ServerVersion(); err == nil {
		sb.WriteString(fmt.Sprintf("- Kubernetes Version: %s\n", v.String()))
	} else {
		sb.WriteString("- Kubernetes Version: Unknown (Error)\n")
	}

	// Get Node Count & Details
	if nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{}); err == nil {
		sb.WriteString(fmt.Sprintf("- Total Nodes: %d\n", len(nodes.Items)))
		for i, node := range nodes.Items {
			if i >= 5 {
				sb.WriteString(fmt.Sprintf("  ... and %d more nodes\n", len(nodes.Items)-5))
				break
			}
			// Check Ready status
			ready := "NotReady"
			for _, cond := range node.Status.Conditions {
				if cond.Type == "Ready" && cond.Status == "True" {
					ready = "Ready"
					break
				}
			}
			sb.WriteString(fmt.Sprintf("  - Node: %s [%s]\n", node.Name, ready))
		}
	} else {
		sb.WriteString("- Nodes: Unknown (Error listing nodes)\n")
	}

	return sb.String()
}
