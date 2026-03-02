package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/cy77cc/k8s-manage/internal/httpx"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// NamespaceInfo represents namespace information
type NamespaceInfo struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Labels      map[string]string `json:"labels"`
	CreatedAt   string `json:"created_at"`
}

// PodInfo represents pod information
type PodInfo struct {
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace"`
	Status       string            `json:"status"`
	PodIP        string            `json:"pod_ip"`
	NodeName     string            `json:"node_name"`
	Ready        string            `json:"ready"`
	Restarts     int32             `json:"restarts"`
	Age          string            `json:"age"`
	Labels       map[string]string `json:"labels"`
	CreatedAt    string            `json:"created_at"`
}

// DeploymentInfo represents deployment information
type DeploymentInfo struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Replicas    int32  `json:"replicas"`
	Ready       int32  `json:"ready"`
	Updated     int32  `json:"updated"`
	Available   int32  `json:"available"`
	Age         string `json:"age"`
	CreatedAt   string `json:"created_at"`
}

// StatefulSetInfo represents statefulset information
type StatefulSetInfo struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Replicas    int32  `json:"replicas"`
	Ready       int32  `json:"ready"`
	Age         string `json:"age"`
	CreatedAt   string `json:"created_at"`
}

// DaemonSetInfo represents daemonset information
type DaemonSetInfo struct {
	Name          string `json:"name"`
	Namespace     string `json:"namespace"`
	Desired       int32  `json:"desired"`
	Ready         int32  `json:"ready"`
	Age           string `json:"age"`
	CreatedAt     string `json:"created_at"`
}

// JobInfo represents job information
type JobInfo struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Completions int32  `json:"completions"`
	Succeeded   int32  `json:"succeeded"`
	Failed      int32  `json:"failed"`
	Status      string `json:"status"`
	Age         string `json:"age"`
	CreatedAt   string `json:"created_at"`
}

// ServiceInfo represents service information
type ServiceInfo struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Type       string            `json:"type"`
	ClusterIP  string            `json:"cluster_ip"`
	Ports      []ServicePort     `json:"ports"`
	Selector   map[string]string `json:"selector"`
	Age        string            `json:"age"`
	CreatedAt  string            `json:"created_at"`
}

// ServicePort represents service port
type ServicePort struct {
	Name       string `json:"name"`
	Port       int32  `json:"port"`
	TargetPort string `json:"target_port"`
	Protocol   string `json:"protocol"`
}

// IngressInfo represents ingress information
type IngressInfo struct {
	Name       string           `json:"name"`
	Namespace  string           `json:"namespace"`
	Hosts      []IngressHost    `json:"hosts"`
	Age        string           `json:"age"`
	CreatedAt  string           `json:"created_at"`
}

// IngressHost represents ingress host
type IngressHost struct {
	Host  string   `json:"host"`
	Paths []string `json:"paths"`
}

// ConfigMapInfo represents configmap information
type ConfigMapInfo struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	DataKeys   []string `json:"data_keys"`
	Age        string `json:"age"`
	CreatedAt  string `json:"created_at"`
}

// SecretInfo represents secret information (metadata only)
type SecretInfo struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Type       string `json:"type"`
	DataKeys   []string `json:"data_keys"`
	Age        string `json:"age"`
	CreatedAt  string `json:"created_at"`
}

// PVCInfo represents PVC information
type PVCInfo struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Status      string `json:"status"`
	Capacity    string `json:"capacity"`
	AccessModes string `json:"access_modes"`
	StorageClass string `json:"storage_class"`
	VolumeName  string `json:"volume_name"`
	Age         string `json:"age"`
	CreatedAt   string `json:"created_at"`
}

// PVInfo represents PV information
type PVInfo struct {
	Name          string `json:"name"`
	Status        string `json:"status"`
	Capacity      string `json:"capacity"`
	AccessModes   string `json:"access_modes"`
	StorageClass  string `json:"storage_class"`
	ClaimRef      string `json:"claim_ref"`
	Age           string `json:"age"`
	CreatedAt     string `json:"created_at"`
}

// getClusterClient returns a kubernetes client for the cluster
func (h *Handler) getClusterClient(ctx context.Context, clusterID uint) (*kubernetes.Clientset, error) {
	var cred model.ClusterCredential
	if err := h.svcCtx.DB.WithContext(ctx).
		Where("cluster_id = ?", clusterID).
		First(&cred).Error; err != nil {
		return nil, fmt.Errorf("credential not found: %w", err)
	}

	restConfig, err := h.buildRestConfigFromCredential(&cred)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(restConfig)
}

// GetNamespaces returns namespaces in the cluster
func (h *Handler) GetNamespaces(c *gin.Context) {
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

	list, err := client.CoreV1().Namespaces().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	items := make([]NamespaceInfo, 0, len(list.Items))
	for _, ns := range list.Items {
		items = append(items, NamespaceInfo{
			Name:      ns.Name,
			Status:    string(ns.Status.Phase),
			Labels:    ns.Labels,
			CreatedAt: ns.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}

	httpx.OK(c, gin.H{"list": items, "total": len(items)})
}

// GetPods returns pods in a namespace
func (h *Handler) GetPods(c *gin.Context) {
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

	list, err := client.CoreV1().Pods(ns).List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	items := make([]PodInfo, 0, len(list.Items))
	for _, pod := range list.Items {
		ready, restarts := getPodReadyAndRestarts(&pod)
		items = append(items, PodInfo{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Status:    string(pod.Status.Phase),
			PodIP:     pod.Status.PodIP,
			NodeName:  pod.Spec.NodeName,
			Ready:     ready,
			Restarts:  restarts,
			Age:       getAge(pod.CreationTimestamp),
			Labels:    pod.Labels,
			CreatedAt: pod.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}

	httpx.OK(c, gin.H{"list": items, "total": len(items)})
}

// GetDeployments returns deployments in a namespace
func (h *Handler) GetDeployments(c *gin.Context) {
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

	list, err := client.AppsV1().Deployments(ns).List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	items := make([]DeploymentInfo, 0, len(list.Items))
	for _, dep := range list.Items {
		replicas := int32(0)
		if dep.Spec.Replicas != nil {
			replicas = *dep.Spec.Replicas
		}
		items = append(items, DeploymentInfo{
			Name:      dep.Name,
			Namespace: dep.Namespace,
			Replicas:  replicas,
			Ready:     dep.Status.ReadyReplicas,
			Updated:   dep.Status.UpdatedReplicas,
			Available: dep.Status.AvailableReplicas,
			Age:       getAge(dep.CreationTimestamp),
			CreatedAt: dep.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}

	httpx.OK(c, gin.H{"list": items, "total": len(items)})
}

// GetStatefulSets returns statefulsets in a namespace
func (h *Handler) GetStatefulSets(c *gin.Context) {
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

	list, err := client.AppsV1().StatefulSets(ns).List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	items := make([]StatefulSetInfo, 0, len(list.Items))
	for _, sts := range list.Items {
		replicas := int32(0)
		if sts.Spec.Replicas != nil {
			replicas = *sts.Spec.Replicas
		}
		items = append(items, StatefulSetInfo{
			Name:      sts.Name,
			Namespace: sts.Namespace,
			Replicas:  replicas,
			Ready:     sts.Status.ReadyReplicas,
			Age:       getAge(sts.CreationTimestamp),
			CreatedAt: sts.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}

	httpx.OK(c, gin.H{"list": items, "total": len(items)})
}

// GetDaemonSets returns daemonsets in a namespace
func (h *Handler) GetDaemonSets(c *gin.Context) {
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

	list, err := client.AppsV1().DaemonSets(ns).List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	items := make([]DaemonSetInfo, 0, len(list.Items))
	for _, ds := range list.Items {
		items = append(items, DaemonSetInfo{
			Name:      ds.Name,
			Namespace: ds.Namespace,
			Desired:   ds.Status.DesiredNumberScheduled,
			Ready:     ds.Status.NumberReady,
			Age:       getAge(ds.CreationTimestamp),
			CreatedAt: ds.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}

	httpx.OK(c, gin.H{"list": items, "total": len(items)})
}

// GetJobs returns jobs in a namespace
func (h *Handler) GetJobs(c *gin.Context) {
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

	list, err := client.BatchV1().Jobs(ns).List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	items := make([]JobInfo, 0, len(list.Items))
	for _, job := range list.Items {
		completions := int32(0)
		if job.Spec.Completions != nil {
			completions = *job.Spec.Completions
		}
		status := "Running"
		if job.Status.Succeeded > 0 {
			status = "Completed"
		} else if job.Status.Failed > 0 {
			status = "Failed"
		}
		items = append(items, JobInfo{
			Name:        job.Name,
			Namespace:   job.Namespace,
			Completions: completions,
			Succeeded:   job.Status.Succeeded,
			Failed:      job.Status.Failed,
			Status:      status,
			Age:         getAge(job.CreationTimestamp),
			CreatedAt:   job.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}

	httpx.OK(c, gin.H{"list": items, "total": len(items)})
}

// GetServices returns services in a namespace
func (h *Handler) GetServices(c *gin.Context) {
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

	list, err := client.CoreV1().Services(ns).List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	items := make([]ServiceInfo, 0, len(list.Items))
	for _, svc := range list.Items {
		ports := make([]ServicePort, 0, len(svc.Spec.Ports))
		for _, p := range svc.Spec.Ports {
			targetPort := ""
			if p.TargetPort.IntVal != 0 {
				targetPort = fmt.Sprintf("%d", p.TargetPort.IntVal)
			} else if p.TargetPort.StrVal != "" {
				targetPort = p.TargetPort.StrVal
			}
			ports = append(ports, ServicePort{
				Name:       p.Name,
				Port:       p.Port,
				TargetPort: targetPort,
				Protocol:   string(p.Protocol),
			})
		}
		items = append(items, ServiceInfo{
			Name:      svc.Name,
			Namespace: svc.Namespace,
			Type:      string(svc.Spec.Type),
			ClusterIP: svc.Spec.ClusterIP,
			Ports:     ports,
			Selector:  svc.Spec.Selector,
			Age:       getAge(svc.CreationTimestamp),
			CreatedAt: svc.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}

	httpx.OK(c, gin.H{"list": items, "total": len(items)})
}

// GetIngresses returns ingresses in a namespace
func (h *Handler) GetIngresses(c *gin.Context) {
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

	list, err := client.NetworkingV1().Ingresses(ns).List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	items := make([]IngressInfo, 0, len(list.Items))
	for _, ing := range list.Items {
		hosts := make([]IngressHost, 0)
		for _, rule := range ing.Spec.Rules {
			paths := make([]string, 0)
			if rule.HTTP != nil {
				for _, p := range rule.HTTP.Paths {
					paths = append(paths, p.Path)
				}
			}
			hosts = append(hosts, IngressHost{
				Host:  rule.Host,
				Paths: paths,
			})
		}
		items = append(items, IngressInfo{
			Name:      ing.Name,
			Namespace: ing.Namespace,
			Hosts:     hosts,
			Age:       getAge(ing.CreationTimestamp),
			CreatedAt: ing.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}

	httpx.OK(c, gin.H{"list": items, "total": len(items)})
}

// GetConfigMaps returns configmaps in a namespace
func (h *Handler) GetConfigMaps(c *gin.Context) {
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

	list, err := client.CoreV1().ConfigMaps(ns).List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	items := make([]ConfigMapInfo, 0, len(list.Items))
	for _, cm := range list.Items {
		keys := make([]string, 0, len(cm.Data))
		for k := range cm.Data {
			keys = append(keys, k)
		}
		items = append(items, ConfigMapInfo{
			Name:      cm.Name,
			Namespace: cm.Namespace,
			DataKeys:  keys,
			Age:       getAge(cm.CreationTimestamp),
			CreatedAt: cm.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}

	httpx.OK(c, gin.H{"list": items, "total": len(items)})
}

// GetSecrets returns secrets metadata in a namespace
func (h *Handler) GetSecrets(c *gin.Context) {
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

	list, err := client.CoreV1().Secrets(ns).List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	items := make([]SecretInfo, 0, len(list.Items))
	for _, sec := range list.Items {
		keys := make([]string, 0, len(sec.Data))
		for k := range sec.Data {
			keys = append(keys, k)
		}
		items = append(items, SecretInfo{
			Name:      sec.Name,
			Namespace: sec.Namespace,
			Type:      string(sec.Type),
			DataKeys:  keys,
			Age:       getAge(sec.CreationTimestamp),
			CreatedAt: sec.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}

	httpx.OK(c, gin.H{"list": items, "total": len(items)})
}

// GetPVCs returns PVCs in a namespace
func (h *Handler) GetPVCs(c *gin.Context) {
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

	list, err := client.CoreV1().PersistentVolumeClaims(ns).List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	items := make([]PVCInfo, 0, len(list.Items))
	for _, pvc := range list.Items {
		capacity := ""
		if pvc.Status.Capacity != nil {
			if q, ok := pvc.Status.Capacity[corev1.ResourceStorage]; ok {
				capacity = q.String()
			}
		}
		accessModes := ""
		for i, am := range pvc.Spec.AccessModes {
			if i > 0 {
				accessModes += ","
			}
			accessModes += string(am)
		}
		items = append(items, PVCInfo{
			Name:         pvc.Name,
			Namespace:    pvc.Namespace,
			Status:       string(pvc.Status.Phase),
			Capacity:     capacity,
			AccessModes:  accessModes,
			StorageClass: *pvc.Spec.StorageClassName,
			VolumeName:   pvc.Spec.VolumeName,
			Age:          getAge(pvc.CreationTimestamp),
			CreatedAt:    pvc.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}

	httpx.OK(c, gin.H{"list": items, "total": len(items)})
}

// GetPVs returns PVs in the cluster
func (h *Handler) GetPVs(c *gin.Context) {
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

	list, err := client.CoreV1().PersistentVolumes().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	items := make([]PVInfo, 0, len(list.Items))
	for _, pv := range list.Items {
		capacity := ""
		if pv.Spec.Capacity != nil {
			if q, ok := pv.Spec.Capacity[corev1.ResourceStorage]; ok {
				capacity = q.String()
			}
		}
		accessModes := ""
		for i, am := range pv.Spec.AccessModes {
			if i > 0 {
				accessModes += ","
			}
			accessModes += string(am)
		}
		claimRef := ""
		if pv.Spec.ClaimRef != nil {
			claimRef = fmt.Sprintf("%s/%s", pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name)
		}
		items = append(items, PVInfo{
			Name:         pv.Name,
			Status:       string(pv.Status.Phase),
			Capacity:     capacity,
			AccessModes:  accessModes,
			StorageClass: pv.Spec.StorageClassName,
			ClaimRef:     claimRef,
			Age:          getAge(pv.CreationTimestamp),
			CreatedAt:    pv.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}

	httpx.OK(c, gin.H{"list": items, "total": len(items)})
}

// Helper functions

func getPodReadyAndRestarts(pod *corev1.Pod) (string, int32) {
	var ready, total int
	var restarts int32
	for _, cs := range pod.Status.ContainerStatuses {
		total++
		if cs.Ready {
			ready++
		}
		restarts += cs.RestartCount
	}
	return fmt.Sprintf("%d/%d", ready, total), restarts
}

func getAge(t metav1.Time) string {
	if t.IsZero() {
		return "-"
	}
	d := metav1.Now().Sub(t.Time)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}
