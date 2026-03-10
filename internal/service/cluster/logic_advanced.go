package cluster

import (
	"context"
	"fmt"
	"strings"
	"time"

	sshclient "github.com/cy77cc/OpsPilot/internal/client/ssh"
	"github.com/cy77cc/OpsPilot/internal/httpx"
	"github.com/cy77cc/OpsPilot/internal/model"
	"github.com/gin-gonic/gin"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EventInfo represents cluster event information
type EventInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
	Reason    string `json:"reason"`
	Message   string `json:"message"`
	Source    string `json:"source"`
	Count     int32  `json:"count"`
	Age       string `json:"age"`
	FirstSeen string `json:"first_seen"`
	LastSeen  string `json:"last_seen"`
}

// HPAInfo represents Horizontal Pod Autoscaler information
type HPAInfo struct {
	Name        string          `json:"name"`
	Namespace   string          `json:"namespace"`
	Reference   string          `json:"reference"`
	MinReplicas int32           `json:"min_replicas"`
	MaxReplicas int32           `json:"max_replicas"`
	CurrentCPU  string          `json:"current_cpu"`
	TargetCPU   string          `json:"target_cpu"`
	CurrentMem  string          `json:"current_mem"`
	TargetMem   string          `json:"target_mem"`
	Replicas    int32           `json:"replicas"`
	Metrics     []HPAMetricInfo `json:"metrics"`
	Age         string          `json:"age"`
	CreatedAt   string          `json:"created_at"`
}

// HPAMetricInfo represents HPA metric information
type HPAMetricInfo struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Current string `json:"current"`
	Target  string `json:"target"`
}

// ResourceQuotaInfo represents ResourceQuota information
type ResourceQuotaInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Hard      map[string]string `json:"hard"`
	Used      map[string]string `json:"used"`
	Age       string            `json:"age"`
	CreatedAt string            `json:"created_at"`
}

// LimitRangeInfo represents LimitRange information
type LimitRangeInfo struct {
	Name      string           `json:"name"`
	Namespace string           `json:"namespace"`
	Type      string           `json:"type"`
	Limits    []LimitRangeItem `json:"limits"`
	Age       string           `json:"age"`
	CreatedAt string           `json:"created_at"`
}

// LimitRangeItem represents a limit range item
type LimitRangeItem struct {
	Type           string            `json:"type"`
	Max            map[string]string `json:"max"`
	Min            map[string]string `json:"min"`
	Default        map[string]string `json:"default"`
	DefaultRequest map[string]string `json:"default_request"`
}

// ClusterVersionInfo represents cluster version information
type ClusterVersionInfo struct {
	KubernetesVersion string `json:"kubernetes_version"`
	GitVersion        string `json:"git_version"`
	Platform          string `json:"platform"`
	GoVersion         string `json:"go_version"`
}

// ClusterUpgradePlan represents upgrade plan information
type ClusterUpgradePlan struct {
	CurrentVersion string   `json:"current_version"`
	TargetVersion  string   `json:"target_version"`
	Upgradable     bool     `json:"upgradable"`
	Steps          []string `json:"steps"`
	Warnings       []string `json:"warnings"`
}

// CertificateInfo represents certificate information
type CertificateInfo struct {
	Name           string   `json:"name"`
	ExpiresAt      string   `json:"expires_at"`
	DaysLeft       int      `json:"days_left"`
	CA             bool     `json:"ca"`
	AlternateNames []string `json:"alternate_names"`
}

// GetEvents returns cluster events
func (h *Handler) GetEvents(c *gin.Context) {
	id := httpx.UintFromParam(c, "id")
	ns := c.Query("namespace")
	if id == 0 {
		httpx.BindErr(c, nil)
		return
	}

	client, err := h.getClusterClient(c.Request.Context(), id)
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	var list *corev1.EventList
	if ns != "" {
		list, err = client.CoreV1().Events(ns).List(c.Request.Context(), metav1.ListOptions{})
	} else {
		list, err = client.CoreV1().Events("").List(c.Request.Context(), metav1.ListOptions{})
	}
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	items := make([]EventInfo, 0, len(list.Items))
	for _, event := range list.Items {
		source := ""
		if event.Source.Component != "" {
			source = event.Source.Component
		} else if event.ReportingController != "" {
			source = event.ReportingController
		}

		items = append(items, EventInfo{
			Name:      event.Name,
			Namespace: event.Namespace,
			Type:      event.Type,
			Reason:    event.Reason,
			Message:   event.Message,
			Source:    source,
			Count:     event.Count,
			Age:       getAge(event.CreationTimestamp),
			FirstSeen: event.FirstTimestamp.Format("2006-01-02 15:04:05"),
			LastSeen:  event.LastTimestamp.Format("2006-01-02 15:04:05"),
		})
	}

	httpx.OK(c, gin.H{"list": items, "total": len(items)})
}

// GetHPAs returns Horizontal Pod Autoscalers in a namespace
func (h *Handler) GetHPAs(c *gin.Context) {
	id := httpx.UintFromParam(c, "id")
	ns := c.Param("namespace")
	if id == 0 || ns == "" {
		httpx.BindErr(c, nil)
		return
	}

	client, err := h.getClusterClient(c.Request.Context(), id)
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	list, err := client.AutoscalingV2().HorizontalPodAutoscalers(ns).List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	items := make([]HPAInfo, 0, len(list.Items))
	for _, hpa := range list.Items {
		info := HPAInfo{
			Name:        hpa.Name,
			Namespace:   hpa.Namespace,
			Reference:   fmt.Sprintf("%s/%s", hpa.Spec.ScaleTargetRef.Kind, hpa.Spec.ScaleTargetRef.Name),
			MinReplicas: *hpa.Spec.MinReplicas,
			MaxReplicas: hpa.Spec.MaxReplicas,
			Replicas:    hpa.Status.CurrentReplicas,
			Age:         getAge(hpa.CreationTimestamp),
			CreatedAt:   hpa.CreationTimestamp.Format("2006-01-02 15:04:05"),
			Metrics:     make([]HPAMetricInfo, 0),
		}

		// Parse metrics
		for _, metric := range hpa.Spec.Metrics {
			metricInfo := HPAMetricInfo{}
			switch metric.Type {
			case autoscalingv2.ResourceMetricSourceType:
				metricInfo.Name = string(metric.Resource.Name)
				metricInfo.Type = "Resource"
				if metric.Resource.Target.AverageUtilization != nil {
					metricInfo.Target = fmt.Sprintf("%d%%", *metric.Resource.Target.AverageUtilization)
				} else if metric.Resource.Target.AverageValue != nil {
					metricInfo.Target = metric.Resource.Target.AverageValue.String()
				}
			case autoscalingv2.PodsMetricSourceType:
				metricInfo.Name = metric.Pods.Metric.Name
				metricInfo.Type = "Pods"
			case autoscalingv2.ObjectMetricSourceType:
				metricInfo.Name = metric.Object.Metric.Name
				metricInfo.Type = "Object"
			case autoscalingv2.ExternalMetricSourceType:
				metricInfo.Name = metric.External.Metric.Name
				metricInfo.Type = "External"
			}

			// Get current value from status
			for _, current := range hpa.Status.CurrentMetrics {
				if current.Type == metric.Type {
					if current.Resource != nil && current.Resource.Current.AverageUtilization != nil {
						metricInfo.Current = fmt.Sprintf("%d%%", *current.Resource.Current.AverageUtilization)
					}
					break
				}
			}

			info.Metrics = append(info.Metrics, metricInfo)
		}

		// Set target CPU if exists (for backward compatibility display)
		for _, metric := range hpa.Spec.Metrics {
			if metric.Type == autoscalingv2.ResourceMetricSourceType && metric.Resource.Name == corev1.ResourceCPU {
				if metric.Resource.Target.AverageUtilization != nil {
					info.TargetCPU = fmt.Sprintf("%d%%", *metric.Resource.Target.AverageUtilization)
				}
			}
			if metric.Type == autoscalingv2.ResourceMetricSourceType && metric.Resource.Name == corev1.ResourceMemory {
				if metric.Resource.Target.AverageUtilization != nil {
					info.TargetMem = fmt.Sprintf("%d%%", *metric.Resource.Target.AverageUtilization)
				}
			}
		}

		// Set current CPU/Memory from status
		for _, current := range hpa.Status.CurrentMetrics {
			if current.Type == autoscalingv2.ResourceMetricSourceType {
				if current.Resource.Name == corev1.ResourceCPU && current.Resource.Current.AverageUtilization != nil {
					info.CurrentCPU = fmt.Sprintf("%d%%", *current.Resource.Current.AverageUtilization)
				}
				if current.Resource.Name == corev1.ResourceMemory && current.Resource.Current.AverageUtilization != nil {
					info.CurrentMem = fmt.Sprintf("%d%%", *current.Resource.Current.AverageUtilization)
				}
			}
		}

		items = append(items, info)
	}

	httpx.OK(c, gin.H{"list": items, "total": len(items)})
}

// GetResourceQuotas returns ResourceQuotas in a namespace
func (h *Handler) GetResourceQuotas(c *gin.Context) {
	id := httpx.UintFromParam(c, "id")
	ns := c.Param("namespace")
	if id == 0 || ns == "" {
		httpx.BindErr(c, nil)
		return
	}

	client, err := h.getClusterClient(c.Request.Context(), id)
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	list, err := client.CoreV1().ResourceQuotas(ns).List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	items := make([]ResourceQuotaInfo, 0, len(list.Items))
	for _, quota := range list.Items {
		hard := make(map[string]string)
		used := make(map[string]string)

		for k, v := range quota.Status.Hard {
			hard[string(k)] = v.String()
		}
		for k, v := range quota.Status.Used {
			used[string(k)] = v.String()
		}

		items = append(items, ResourceQuotaInfo{
			Name:      quota.Name,
			Namespace: quota.Namespace,
			Hard:      hard,
			Used:      used,
			Age:       getAge(quota.CreationTimestamp),
			CreatedAt: quota.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}

	httpx.OK(c, gin.H{"list": items, "total": len(items)})
}

// GetLimitRanges returns LimitRanges in a namespace
func (h *Handler) GetLimitRanges(c *gin.Context) {
	id := httpx.UintFromParam(c, "id")
	ns := c.Param("namespace")
	if id == 0 || ns == "" {
		httpx.BindErr(c, nil)
		return
	}

	client, err := h.getClusterClient(c.Request.Context(), id)
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	list, err := client.CoreV1().LimitRanges(ns).List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	items := make([]LimitRangeInfo, 0, len(list.Items))
	for _, lr := range list.Items {
		limits := make([]LimitRangeItem, 0, len(lr.Spec.Limits))
		for _, item := range lr.Spec.Limits {
			limitItem := LimitRangeItem{
				Type:           string(item.Type),
				Max:            make(map[string]string),
				Min:            make(map[string]string),
				Default:        make(map[string]string),
				DefaultRequest: make(map[string]string),
			}
			for k, v := range item.Max {
				limitItem.Max[string(k)] = v.String()
			}
			for k, v := range item.Min {
				limitItem.Min[string(k)] = v.String()
			}
			for k, v := range item.Default {
				limitItem.Default[string(k)] = v.String()
			}
			for k, v := range item.DefaultRequest {
				limitItem.DefaultRequest[string(k)] = v.String()
			}
			limits = append(limits, limitItem)
		}

		items = append(items, LimitRangeInfo{
			Name:      lr.Name,
			Namespace: lr.Namespace,
			Type:      string(lr.Spec.Limits[0].Type),
			Limits:    limits,
			Age:       getAge(lr.CreationTimestamp),
			CreatedAt: lr.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}

	httpx.OK(c, gin.H{"list": items, "total": len(items)})
}

// GetClusterVersion returns cluster version information
func (h *Handler) GetClusterVersion(c *gin.Context) {
	id := httpx.UintFromParam(c, "id")
	if id == 0 {
		httpx.BindErr(c, nil)
		return
	}

	client, err := h.getClusterClient(c.Request.Context(), id)
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	version, err := client.Discovery().ServerVersion()
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	info := ClusterVersionInfo{
		KubernetesVersion: version.GitVersion,
		GitVersion:        version.GitVersion,
		Platform:          version.Platform,
		GoVersion:         version.GoVersion,
	}

	httpx.OK(c, info)
}

// GetCertificates returns cluster certificates information
func (h *Handler) GetCertificates(c *gin.Context) {
	id := httpx.UintFromParam(c, "id")
	if id == 0 {
		httpx.BindErr(c, nil)
		return
	}

	client, err := h.getClusterClient(c.Request.Context(), id)
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	// Try to get certificates from kube-system namespace
	// This is typically available on kubeadm-managed clusters
	secrets, err := client.CoreV1().Secrets("kube-system").List(c.Request.Context(), metav1.ListOptions{
		LabelSelector: "kubeadm.io/component",
	})
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	items := make([]CertificateInfo, 0)
	for _, secret := range secrets.Items {
		if secret.Type != corev1.SecretTypeTLS {
			continue
		}

		// Parse certificate expiry from annotation
		expiresAt := ""
		daysLeft := 0
		if exp, ok := secret.Annotations["kubeadm.io/expiration"]; ok {
			expiresAt = exp
			if t, err := time.Parse(time.RFC3339, exp); err == nil {
				daysLeft = int(time.Until(t).Hours() / 24)
			}
		}

		altNames := []string{}
		if names, ok := secret.Annotations["kubeadm.io/alt-names"]; ok {
			altNames = splitAltNames(names)
		}

		items = append(items, CertificateInfo{
			Name:           secret.Name,
			ExpiresAt:      expiresAt,
			DaysLeft:       daysLeft,
			CA:             secret.Name == "ca" || secret.Name == "etcd-ca",
			AlternateNames: altNames,
		})
	}

	httpx.OK(c, gin.H{"list": items, "total": len(items)})
}

// GetUpgradePlan returns cluster upgrade plan
func (h *Handler) GetUpgradePlan(c *gin.Context) {
	id := httpx.UintFromParam(c, "id")
	if id == 0 {
		httpx.BindErr(c, nil)
		return
	}

	client, err := h.getClusterClient(c.Request.Context(), id)
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	version, err := client.Discovery().ServerVersion()
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	currentVersion := version.GitVersion

	// Get available versions from cluster model
	var cluster model.Cluster
	if err := h.svcCtx.DB.WithContext(c.Request.Context()).First(&cluster, id).Error; err != nil {
		httpx.ServerErr(c, err)
		return
	}

	// Build upgrade plan
	plan := ClusterUpgradePlan{
		CurrentVersion: currentVersion,
		TargetVersion:  "",
		Upgradable:     false,
		Steps:          []string{},
		Warnings:       []string{},
	}

	// Check if this is a self-hosted cluster
	if cluster.Source == "platform_managed" {
		plan.Upgradable = true
		plan.Steps = []string{
			"1. Backup etcd data and cluster state",
			"2. Upgrade control plane nodes one by one",
			"3. Upgrade worker nodes",
			"4. Verify cluster health after upgrade",
		}
		plan.Warnings = []string{
			"Ensure you have a valid backup before proceeding",
			"Upgrade should be done during maintenance window",
			"Check compatibility of workloads with new version",
		}
	} else {
		plan.Warnings = []string{
			"Only platform-managed clusters support managed upgrades",
			"Imported clusters should be upgraded using their native tools",
		}
	}

	httpx.OK(c, plan)
}

// UpgradeClusterReq represents a cluster upgrade request
type UpgradeClusterReq struct {
	TargetVersion string `json:"target_version" binding:"required"`
}

// UpgradeClusterResult represents the result of a cluster upgrade
type UpgradeClusterResult struct {
	ClusterID    uint     `json:"cluster_id"`
	FromVersion  string   `json:"from_version"`
	ToVersion    string   `json:"to_version"`
	Status       string   `json:"status"`
	Message      string   `json:"message"`
	UpgradeSteps []string `json:"upgrade_steps"`
}

// UpgradeCluster upgrades a platform-managed cluster
func (h *Handler) UpgradeCluster(c *gin.Context) {
	if !httpx.Authorize(c, h.svcCtx.DB, "cluster:write") {
		return
	}

	id := httpx.UintFromParam(c, "id")
	if id == 0 {
		httpx.BindErr(c, nil)
		return
	}

	var req UpgradeClusterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BindErr(c, err)
		return
	}

	// Get cluster
	var cluster model.Cluster
	if err := h.svcCtx.DB.WithContext(c.Request.Context()).First(&cluster, id).Error; err != nil {
		httpx.NotFound(c, "cluster not found")
		return
	}

	// Check if platform managed
	if cluster.Source != "platform_managed" {
		httpx.BadRequest(c, "only platform-managed clusters can be upgraded through this API")
		return
	}

	// Get current version
	client, err := h.getClusterClient(c.Request.Context(), id)
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	version, err := client.Discovery().ServerVersion()
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	// For now, return a preview response - actual upgrade would require
	// SSH access to nodes and careful orchestration
	result := UpgradeClusterResult{
		ClusterID:   id,
		FromVersion: version.GitVersion,
		ToVersion:   req.TargetVersion,
		Status:      "preview",
		Message:     "Cluster upgrade would require SSH access to all nodes. This is a preview.",
		UpgradeSteps: []string{
			fmt.Sprintf("1. Drain and cordon control plane nodes"),
			fmt.Sprintf("2. Upgrade kubeadm to v%s on control plane", req.TargetVersion),
			fmt.Sprintf("3. Run 'kubeadm upgrade apply v%s' on control plane", req.TargetVersion),
			fmt.Sprintf("4. Upgrade kubelet and kubectl on control plane"),
			fmt.Sprintf("5. Uncordon control plane nodes"),
			fmt.Sprintf("6. Repeat steps 1-5 for worker nodes"),
			"7. Verify cluster health",
		},
	}

	httpx.OK(c, result)
}

// RenewCertificates renews cluster certificates
func (h *Handler) RenewCertificates(c *gin.Context) {
	if !httpx.Authorize(c, h.svcCtx.DB, "cluster:write") {
		return
	}

	id := httpx.UintFromParam(c, "id")
	if id == 0 {
		httpx.BindErr(c, nil)
		return
	}

	// Get cluster
	var cluster model.Cluster
	if err := h.svcCtx.DB.WithContext(c.Request.Context()).First(&cluster, id).Error; err != nil {
		httpx.NotFound(c, "cluster not found")
		return
	}

	// Check if platform managed
	if cluster.Source != "platform_managed" {
		httpx.BadRequest(c, "only platform-managed clusters can renew certificates through this API")
		return
	}

	// Get control plane nodes
	var controlPlaneNodes []model.ClusterNode
	if err := h.svcCtx.DB.WithContext(c.Request.Context()).
		Where("cluster_id = ? AND role = ?", id, "control-plane").
		Find(&controlPlaneNodes).Error; err != nil {
		httpx.ServerErr(c, err)
		return
	}

	if len(controlPlaneNodes) == 0 {
		httpx.BadRequest(c, "no control plane nodes found")
		return
	}

	// Execute certificate renewal on each control plane node via SSH
	results := make([]map[string]interface{}, 0, len(controlPlaneNodes))
	for _, node := range controlPlaneNodes {
		if node.HostID == nil {
			results = append(results, map[string]interface{}{
				"node_name": node.Name,
				"success":   false,
				"message":   "no associated host for SSH access",
			})
			continue
		}

		var host model.Node
		if err := h.svcCtx.DB.WithContext(c.Request.Context()).First(&host, *node.HostID).Error; err != nil {
			results = append(results, map[string]interface{}{
				"node_name": node.Name,
				"success":   false,
				"message":   "host not found",
			})
			continue
		}

		// Execute kubeadm certs renew all via SSH
		err := h.executeCertRenewal(c.Request.Context(), &host)
		if err != nil {
			results = append(results, map[string]interface{}{
				"node_name": node.Name,
				"host_name": host.Name,
				"success":   false,
				"message":   err.Error(),
			})
		} else {
			results = append(results, map[string]interface{}{
				"node_name": node.Name,
				"host_name": host.Name,
				"success":   true,
				"message":   "certificates renewed successfully",
			})
		}
	}

	httpx.OK(c, gin.H{
		"cluster_id": id,
		"results":    results,
		"message":    fmt.Sprintf("Processed %d control plane nodes", len(controlPlaneNodes)),
	})
}

// executeCertRenewal executes certificate renewal on a host
func (h *Handler) executeCertRenewal(ctx context.Context, host *model.Node) error {
	privateKey, passphrase, err := h.loadNodePrivateKey(ctx, host)
	if err != nil {
		return err
	}

	password := strings.TrimSpace(host.SSHPassword)
	if strings.TrimSpace(privateKey) != "" {
		password = ""
	}

	cli, err := sshclient.NewSSHClient(host.SSHUser, password, host.IP, host.Port, privateKey, passphrase)
	if err != nil {
		return err
	}
	defer cli.Close()

	// Execute kubeadm certs renew all
	_, err = sshclient.RunCommand(cli, "sudo kubeadm certs renew all")
	return err
}

// Helper function to parse alternate names
func splitAltNames(names string) []string {
	result := []string{}
	for _, name := range splitString(names, ",") {
		if name != "" {
			result = append(result, name)
		}
	}
	return result
}

func splitString(s, sep string) []string {
	if s == "" {
		return nil
	}
	var result []string
	start := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}
