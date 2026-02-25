package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"
)

var rolloutGVR = schema.GroupVersionResource{Group: "argoproj.io", Version: "v1alpha1", Resource: "rollouts"}

func (h *Handler) Namespaces(c *gin.Context) {
	if !h.authorize(c, "k8s:read", "kubernetes:read") {
		return
	}
	cluster, ok := h.mustCluster(c)
	if !ok {
		return
	}
	cli, _, err := h.getClients(cluster)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": []any{}, "total": 0, "data_source": "none"}})
		return
	}
	nsList, err := cli.CoreV1().Namespaces().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	uid := h.uidFromContext(c)
	teamID := h.teamIDFromHeader(c)
	allowed := map[string]struct{}{}
	if !h.isAdminUser(uid) && teamID > 0 {
		bound, _ := h.listBoundNamespaces(cluster.ID, teamID)
		for _, b := range bound {
			allowed[b] = struct{}{}
		}
	}
	list := make([]gin.H, 0, len(nsList.Items))
	for _, ns := range nsList.Items {
		if len(allowed) > 0 {
			if _, ok := allowed[ns.Name]; !ok {
				continue
			}
		}
		list = append(list, gin.H{"name": ns.Name, "status": string(ns.Status.Phase), "labels": ns.Labels, "created_at": ns.CreationTimestamp.Time})
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": list, "total": len(list)}})
}

func (h *Handler) CreateNamespace(c *gin.Context) {
	if !h.authorize(c, "k8s:write", "kubernetes:write") {
		return
	}
	cluster, ok := h.mustCluster(c)
	if !ok {
		return
	}
	var req struct {
		Name   string            `json:"name" binding:"required"`
		Env    string            `json:"env"`
		Labels map[string]string `json:"labels"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	if !h.namespaceWritable(c, cluster.ID, req.Name) {
		uid := h.uidFromContext(c)
		if !h.isAdminUser(uid) {
			return
		}
	}
	cli, _, err := h.getClients(cluster)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	labels := req.Labels
	if labels == nil {
		labels = map[string]string{}
	}
	if strings.TrimSpace(req.Env) != "" {
		labels["environment"] = req.Env
	}
	nsObj := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: req.Name, Labels: labels}}
	ns, err := cli.CoreV1().Namespaces().Create(c.Request.Context(), nsObj, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	if ns == nil {
		ns, _ = cli.CoreV1().Namespaces().Get(c.Request.Context(), req.Name, metav1.GetOptions{})
	}
	uid := h.uidFromContext(c)
	teamID := h.teamIDFromHeader(c)
	if teamID > 0 {
		var existing model.ClusterNamespaceBinding
		if h.svcCtx.DB.Where("cluster_id = ? AND team_id = ? AND namespace = ?", cluster.ID, teamID, req.Name).First(&existing).Error != nil {
			_ = h.svcCtx.DB.Create(&model.ClusterNamespaceBinding{ClusterID: cluster.ID, TeamID: teamID, Namespace: req.Name, Env: req.Env, Readonly: false, CreatedBy: uid}).Error
		}
	}
	h.createAudit(cluster.ID, req.Name, "namespace.create", "namespace", req.Name, "success", "namespace created", uid)
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": ns})
}

func (h *Handler) DeleteNamespace(c *gin.Context) {
	if !h.authorize(c, "k8s:write", "kubernetes:write") {
		return
	}
	cluster, ok := h.mustCluster(c)
	if !ok {
		return
	}
	name := strings.TrimSpace(c.Param("name"))
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": "namespace required"})
		return
	}
	if name == "kube-system" || name == "kube-public" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": "system namespace cannot be deleted"})
		return
	}
	if !h.namespaceWritable(c, cluster.ID, name) {
		return
	}
	cli, _, err := h.getClients(cluster)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	if err := cli.CoreV1().Namespaces().Delete(c.Request.Context(), name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	h.createAudit(cluster.ID, name, "namespace.delete", "namespace", name, "success", "namespace deleted", h.uidFromContext(c))
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": nil})
}

func (h *Handler) ListNamespaceBindings(c *gin.Context) {
	if !h.authorize(c, "k8s:namespace:bind", "k8s:read", "kubernetes:read") {
		return
	}
	cluster, ok := h.mustCluster(c)
	if !ok {
		return
	}
	query := h.svcCtx.DB.Model(&model.ClusterNamespaceBinding{}).Where("cluster_id = ?", cluster.ID)
	if teamRaw := strings.TrimSpace(c.Query("team_id")); teamRaw != "" {
		query = query.Where("team_id = ?", teamRaw)
	}
	var rows []model.ClusterNamespaceBinding
	if err := query.Order("team_id ASC, namespace ASC").Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": rows, "total": len(rows)}})
}

func (h *Handler) PutNamespaceBindings(c *gin.Context) {
	if !h.authorize(c, "k8s:namespace:bind", "kubernetes:write") {
		return
	}
	cluster, ok := h.mustCluster(c)
	if !ok {
		return
	}
	teamID64, err := strconv.ParseUint(c.Param("teamId"), 10, 64)
	if err != nil || teamID64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": "invalid team id"})
		return
	}
	var req struct {
		Bindings []struct {
			Namespace string `json:"namespace" binding:"required"`
			Env       string `json:"env"`
			Readonly  bool   `json:"readonly"`
		} `json:"bindings" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	uid := h.uidFromContext(c)
	teamID := uint(teamID64)
	if err := h.tx(func(tx *gorm.DB) error {
		if err := tx.Where("cluster_id = ? AND team_id = ?", cluster.ID, teamID).Delete(&model.ClusterNamespaceBinding{}).Error; err != nil {
			return err
		}
		for _, item := range req.Bindings {
			if strings.TrimSpace(item.Namespace) == "" {
				continue
			}
			row := model.ClusterNamespaceBinding{ClusterID: cluster.ID, TeamID: teamID, Namespace: strings.TrimSpace(item.Namespace), Env: strings.TrimSpace(item.Env), Readonly: item.Readonly, CreatedBy: uid}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	h.createAudit(cluster.ID, "", "namespace.bindings.update", "binding", strconv.FormatUint(teamID64, 10), "success", "namespace bindings updated", uid)
	h.ListNamespaceBindings(c)
}

func (h *Handler) ListHPA(c *gin.Context) {
	if !h.authorize(c, "k8s:read", "k8s:hpa", "kubernetes:read") {
		return
	}
	cluster, ok := h.mustCluster(c)
	if !ok {
		return
	}
	cli, _, err := h.getClients(cluster)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": []any{}, "total": 0}})
		return
	}
	namespace := strings.TrimSpace(c.Query("namespace"))
	if namespace == "" {
		namespace = corev1.NamespaceAll
	}
	if namespace != corev1.NamespaceAll && !h.namespaceReadable(c, cluster.ID, namespace) {
		return
	}
	items, err := cli.AutoscalingV2().HorizontalPodAutoscalers(namespace).List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	out := make([]gin.H, 0, len(items.Items))
	for _, item := range items.Items {
		cpu, mem := int32(0), int32(0)
		for _, m := range item.Spec.Metrics {
			if m.Type != autoscalingv2.ResourceMetricSourceType || m.Resource == nil || m.Resource.Target.AverageUtilization == nil {
				continue
			}
			if m.Resource.Name == corev1.ResourceCPU {
				cpu = *m.Resource.Target.AverageUtilization
			}
			if m.Resource.Name == corev1.ResourceMemory {
				mem = *m.Resource.Target.AverageUtilization
			}
		}
		out = append(out, gin.H{
			"name":               item.Name,
			"namespace":          item.Namespace,
			"target_ref_kind":    item.Spec.ScaleTargetRef.Kind,
			"target_ref_name":    item.Spec.ScaleTargetRef.Name,
			"min_replicas":       ptr.Deref(item.Spec.MinReplicas, 1),
			"max_replicas":       item.Spec.MaxReplicas,
			"cpu_utilization":    cpu,
			"memory_utilization": mem,
		})
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": out, "total": len(out)}})
}

func (h *Handler) CreateHPA(c *gin.Context) {
	if !h.authorize(c, "k8s:hpa", "k8s:write", "kubernetes:write") {
		return
	}
	h.applyHPA(c, false)
}

func (h *Handler) UpdateHPA(c *gin.Context) {
	if !h.authorize(c, "k8s:hpa", "k8s:write", "kubernetes:write") {
		return
	}
	h.applyHPA(c, true)
}

func (h *Handler) applyHPA(c *gin.Context, update bool) {
	cluster, ok := h.mustCluster(c)
	if !ok {
		return
	}
	var req struct {
		Namespace         string `json:"namespace" binding:"required"`
		Name              string `json:"name" binding:"required"`
		TargetRefKind     string `json:"target_ref_kind" binding:"required"`
		TargetRefName     string `json:"target_ref_name" binding:"required"`
		MinReplicas       int32  `json:"min_replicas"`
		MaxReplicas       int32  `json:"max_replicas" binding:"required"`
		CPUUtilization    *int32 `json:"cpu_utilization"`
		MemoryUtilization *int32 `json:"memory_utilization"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	if !h.namespaceWritable(c, cluster.ID, req.Namespace) {
		return
	}
	if req.MinReplicas <= 0 {
		req.MinReplicas = 1
	}
	cli, _, err := h.getClients(cluster)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	metrics := make([]autoscalingv2.MetricSpec, 0, 2)
	if req.CPUUtilization != nil {
		metrics = append(metrics, autoscalingv2.MetricSpec{Type: autoscalingv2.ResourceMetricSourceType, Resource: &autoscalingv2.ResourceMetricSource{Name: corev1.ResourceCPU, Target: autoscalingv2.MetricTarget{Type: autoscalingv2.UtilizationMetricType, AverageUtilization: req.CPUUtilization}}})
	}
	if req.MemoryUtilization != nil {
		metrics = append(metrics, autoscalingv2.MetricSpec{Type: autoscalingv2.ResourceMetricSourceType, Resource: &autoscalingv2.ResourceMetricSource{Name: corev1.ResourceMemory, Target: autoscalingv2.MetricTarget{Type: autoscalingv2.UtilizationMetricType, AverageUtilization: req.MemoryUtilization}}})
	}
	if len(metrics) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": "at least one metric is required"})
		return
	}
	obj := &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{Name: req.Name, Namespace: req.Namespace},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{APIVersion: "apps/v1", Kind: req.TargetRefKind, Name: req.TargetRefName},
			MinReplicas:    ptr.To(req.MinReplicas),
			MaxReplicas:    req.MaxReplicas,
			Metrics:        metrics,
		},
	}
	if update {
		if _, err := cli.AutoscalingV2().HorizontalPodAutoscalers(req.Namespace).Update(c.Request.Context(), obj, metav1.UpdateOptions{}); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
			return
		}
	} else {
		if _, err := cli.AutoscalingV2().HorizontalPodAutoscalers(req.Namespace).Create(c.Request.Context(), obj, metav1.CreateOptions{}); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
			return
		}
	}
	raw, _ := json.Marshal(req)
	_ = h.svcCtx.DB.Save(&model.ClusterHPAPolicy{ClusterID: cluster.ID, Namespace: req.Namespace, Name: req.Name, TargetRefKind: req.TargetRefKind, TargetRefName: req.TargetRefName, MinReplicas: req.MinReplicas, MaxReplicas: req.MaxReplicas, CPUUtilization: req.CPUUtilization, MemoryUtilization: req.MemoryUtilization, RawPolicyJSON: string(raw)}).Error
	h.createAudit(cluster.ID, req.Namespace, "hpa.apply", "hpa", req.Name, "success", "hpa policy applied", h.uidFromContext(c))
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": obj})
}

func (h *Handler) DeleteHPA(c *gin.Context) {
	if !h.authorize(c, "k8s:hpa", "k8s:write", "kubernetes:write") {
		return
	}
	cluster, ok := h.mustCluster(c)
	if !ok {
		return
	}
	namespace := strings.TrimSpace(c.Query("namespace"))
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": "namespace required"})
		return
	}
	name := strings.TrimSpace(c.Param("name"))
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": "name required"})
		return
	}
	if !h.namespaceWritable(c, cluster.ID, namespace) {
		return
	}
	cli, _, err := h.getClients(cluster)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	if err := cli.AutoscalingV2().HorizontalPodAutoscalers(namespace).Delete(c.Request.Context(), name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	_ = h.svcCtx.DB.Where("cluster_id = ? AND namespace = ? AND name = ?", cluster.ID, namespace, name).Delete(&model.ClusterHPAPolicy{}).Error
	h.createAudit(cluster.ID, namespace, "hpa.delete", "hpa", name, "success", "hpa policy deleted", h.uidFromContext(c))
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": nil})
}

func (h *Handler) ListQuotas(c *gin.Context) {
	if !h.authorize(c, "k8s:read", "k8s:quota", "kubernetes:read") {
		return
	}
	cluster, ok := h.mustCluster(c)
	if !ok {
		return
	}
	cli, _, err := h.getClients(cluster)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": []any{}, "total": 0}})
		return
	}
	namespace := strings.TrimSpace(c.Query("namespace"))
	if namespace == "" {
		namespace = corev1.NamespaceAll
	}
	if namespace != corev1.NamespaceAll && !h.namespaceReadable(c, cluster.ID, namespace) {
		return
	}
	items, err := cli.CoreV1().ResourceQuotas(namespace).List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	list := make([]gin.H, 0, len(items.Items))
	for _, item := range items.Items {
		hard := map[string]string{}
		for k, v := range item.Spec.Hard {
			hard[string(k)] = v.String()
		}
		list = append(list, gin.H{"name": item.Name, "namespace": item.Namespace, "hard": hard})
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": list, "total": len(list)}})
}

func (h *Handler) CreateOrUpdateQuota(c *gin.Context) {
	if !h.authorize(c, "k8s:quota", "k8s:write", "kubernetes:write") {
		return
	}
	cluster, ok := h.mustCluster(c)
	if !ok {
		return
	}
	var req struct {
		Namespace string            `json:"namespace" binding:"required"`
		Name      string            `json:"name" binding:"required"`
		Hard      map[string]string `json:"hard" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	if !h.namespaceWritable(c, cluster.ID, req.Namespace) {
		return
	}
	cli, _, err := h.getClients(cluster)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	hard := corev1.ResourceList{}
	for k, v := range req.Hard {
		q, err := resource.ParseQuantity(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": fmt.Sprintf("invalid quota %s: %v", k, err)})
			return
		}
		hard[corev1.ResourceName(k)] = q
	}
	obj := &corev1.ResourceQuota{ObjectMeta: metav1.ObjectMeta{Name: req.Name, Namespace: req.Namespace}, Spec: corev1.ResourceQuotaSpec{Hard: hard}}
	if _, err := cli.CoreV1().ResourceQuotas(req.Namespace).Get(c.Request.Context(), req.Name, metav1.GetOptions{}); err == nil {
		if _, err := cli.CoreV1().ResourceQuotas(req.Namespace).Update(c.Request.Context(), obj, metav1.UpdateOptions{}); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
			return
		}
	} else {
		if _, err := cli.CoreV1().ResourceQuotas(req.Namespace).Create(c.Request.Context(), obj, metav1.CreateOptions{}); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
			return
		}
	}
	raw, _ := json.Marshal(req)
	_ = h.svcCtx.DB.Save(&model.ClusterQuotaPolicy{ClusterID: cluster.ID, Namespace: req.Namespace, Name: req.Name, Type: "resourcequota", SpecJSON: string(raw)}).Error
	h.createAudit(cluster.ID, req.Namespace, "quota.apply", "resourcequota", req.Name, "success", "quota applied", h.uidFromContext(c))
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": obj})
}

func (h *Handler) DeleteQuota(c *gin.Context) {
	if !h.authorize(c, "k8s:quota", "k8s:write", "kubernetes:write") {
		return
	}
	cluster, ok := h.mustCluster(c)
	if !ok {
		return
	}
	namespace := strings.TrimSpace(c.Query("namespace"))
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": "namespace required"})
		return
	}
	if !h.namespaceWritable(c, cluster.ID, namespace) {
		return
	}
	name := strings.TrimSpace(c.Param("name"))
	cli, _, err := h.getClients(cluster)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	if err := cli.CoreV1().ResourceQuotas(namespace).Delete(c.Request.Context(), name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	_ = h.svcCtx.DB.Where("cluster_id = ? AND namespace = ? AND name = ? AND type = ?", cluster.ID, namespace, name, "resourcequota").Delete(&model.ClusterQuotaPolicy{}).Error
	h.createAudit(cluster.ID, namespace, "quota.delete", "resourcequota", name, "success", "quota deleted", h.uidFromContext(c))
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": nil})
}

func (h *Handler) ListLimitRanges(c *gin.Context) {
	if !h.authorize(c, "k8s:read", "k8s:quota", "kubernetes:read") {
		return
	}
	cluster, ok := h.mustCluster(c)
	if !ok {
		return
	}
	cli, _, err := h.getClients(cluster)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": []any{}, "total": 0}})
		return
	}
	namespace := strings.TrimSpace(c.Query("namespace"))
	if namespace == "" {
		namespace = corev1.NamespaceAll
	}
	if namespace != corev1.NamespaceAll && !h.namespaceReadable(c, cluster.ID, namespace) {
		return
	}
	items, err := cli.CoreV1().LimitRanges(namespace).List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	list := make([]gin.H, 0, len(items.Items))
	for _, item := range items.Items {
		list = append(list, gin.H{"name": item.Name, "namespace": item.Namespace, "limits": item.Spec.Limits})
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": list, "total": len(list)}})
}

func (h *Handler) CreateLimitRange(c *gin.Context) {
	if !h.authorize(c, "k8s:quota", "k8s:write", "kubernetes:write") {
		return
	}
	cluster, ok := h.mustCluster(c)
	if !ok {
		return
	}
	var req struct {
		Namespace      string            `json:"namespace" binding:"required"`
		Name           string            `json:"name" binding:"required"`
		Default        map[string]string `json:"default"`
		DefaultRequest map[string]string `json:"default_request"`
		Min            map[string]string `json:"min"`
		Max            map[string]string `json:"max"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	if !h.namespaceWritable(c, cluster.ID, req.Namespace) {
		return
	}
	buildRes := func(m map[string]string) (corev1.ResourceList, error) {
		out := corev1.ResourceList{}
		for k, v := range m {
			q, err := resource.ParseQuantity(v)
			if err != nil {
				return nil, err
			}
			out[corev1.ResourceName(k)] = q
		}
		return out, nil
	}
	def, err := buildRes(req.Default)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	defReq, err := buildRes(req.DefaultRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	minRes, err := buildRes(req.Min)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	maxRes, err := buildRes(req.Max)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	obj := &corev1.LimitRange{ObjectMeta: metav1.ObjectMeta{Name: req.Name, Namespace: req.Namespace}, Spec: corev1.LimitRangeSpec{Limits: []corev1.LimitRangeItem{{Type: corev1.LimitTypeContainer, Default: def, DefaultRequest: defReq, Min: minRes, Max: maxRes}}}}
	cli, _, err := h.getClients(cluster)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	if _, err := cli.CoreV1().LimitRanges(req.Namespace).Get(c.Request.Context(), req.Name, metav1.GetOptions{}); err == nil {
		if _, err := cli.CoreV1().LimitRanges(req.Namespace).Update(c.Request.Context(), obj, metav1.UpdateOptions{}); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
			return
		}
	} else {
		if _, err := cli.CoreV1().LimitRanges(req.Namespace).Create(c.Request.Context(), obj, metav1.CreateOptions{}); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
			return
		}
	}
	raw, _ := json.Marshal(req)
	_ = h.svcCtx.DB.Save(&model.ClusterQuotaPolicy{ClusterID: cluster.ID, Namespace: req.Namespace, Name: req.Name, Type: "limitrange", SpecJSON: string(raw)}).Error
	h.createAudit(cluster.ID, req.Namespace, "limitrange.apply", "limitrange", req.Name, "success", "limitrange applied", h.uidFromContext(c))
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": obj})
}

func (h *Handler) RolloutPreview(c *gin.Context) {
	if !h.authorize(c, "k8s:deploy", "k8s:write", "kubernetes:write") {
		return
	}
	cluster, ok := h.mustCluster(c)
	if !ok {
		return
	}
	var req rolloutApplyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	if !h.namespaceWritable(c, cluster.ID, req.Namespace) {
		return
	}
	manifest := buildRolloutManifest(req)
	out, _ := yaml.Marshal(manifest.Object)
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"manifest": string(out), "strategy": req.Strategy}})
}

type rolloutApplyReq struct {
	Namespace     string            `json:"namespace" binding:"required"`
	Name          string            `json:"name" binding:"required"`
	Image         string            `json:"image" binding:"required"`
	Replicas      int32             `json:"replicas"`
	Strategy      string            `json:"strategy"`
	Labels        map[string]string `json:"labels"`
	CanarySteps   []map[string]any  `json:"canary_steps"`
	ActiveService string            `json:"active_service"`
	PreviewSvc    string            `json:"preview_service"`
	ApprovalToken string            `json:"approval_token"`
}

func buildRolloutManifest(req rolloutApplyReq) *unstructured.Unstructured {
	if req.Replicas <= 0 {
		req.Replicas = 1
	}
	if req.Strategy == "" {
		req.Strategy = "rolling"
	}
	labels := req.Labels
	if labels == nil {
		labels = map[string]string{}
	}
	if _, ok := labels["app"]; !ok {
		labels["app"] = req.Name
	}
	strategy := map[string]any{}
	switch strings.ToLower(req.Strategy) {
	case "blue-green", "bluegreen":
		strategy["blueGreen"] = map[string]any{
			"activeService":  defaultString(req.ActiveService, req.Name),
			"previewService": defaultString(req.PreviewSvc, req.Name+"-preview"),
		}
	case "canary":
		steps := req.CanarySteps
		if len(steps) == 0 {
			steps = []map[string]any{{"setWeight": 20}, {"pause": map[string]any{"duration": "30s"}}, {"setWeight": 50}, {"pause": map[string]any{"duration": "30s"}}}
		}
		strategy["canary"] = map[string]any{"steps": steps}
	default:
		strategy["canary"] = map[string]any{"steps": []map[string]any{{"setWeight": 100}}}
	}
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "argoproj.io/v1alpha1",
		"kind":       "Rollout",
		"metadata": map[string]any{
			"name":      req.Name,
			"namespace": req.Namespace,
			"labels":    labels,
		},
		"spec": map[string]any{
			"replicas": req.Replicas,
			"selector": map[string]any{"matchLabels": map[string]any{"app": req.Name}},
			"template": map[string]any{"metadata": map[string]any{"labels": map[string]any{"app": req.Name}}, "spec": map[string]any{"containers": []map[string]any{{"name": req.Name, "image": req.Image}}}},
			"strategy": strategy,
		},
	}}
}

func defaultString(v, d string) string {
	if strings.TrimSpace(v) == "" {
		return d
	}
	return v
}

func (h *Handler) RolloutApply(c *gin.Context) {
	if !h.authorize(c, "k8s:deploy", "k8s:write", "kubernetes:write") {
		return
	}
	cluster, ok := h.mustCluster(c)
	if !ok {
		return
	}
	var req rolloutApplyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	if !h.namespaceWritable(c, cluster.ID, req.Namespace) {
		return
	}
	if !h.requireProdApproval(c, cluster.ID, req.Namespace, "deploy", req.ApprovalToken) {
		return
	}
	_, dc, err := h.getClients(cluster)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	obj := buildRolloutManifest(req)
	resource := dc.Resource(rolloutGVR).Namespace(req.Namespace)
	existing, err := resource.Get(c.Request.Context(), req.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			if _, err := resource.Create(c.Request.Context(), obj, metav1.CreateOptions{}); err != nil {
				if isRolloutCRDMissing(err) {
					c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": "rollout_crd_missing"})
					return
				}
				c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
				return
			}
		} else {
			if isRolloutCRDMissing(err) {
				c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": "rollout_crd_missing"})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
			return
		}
	} else {
		obj.SetResourceVersion(existing.GetResourceVersion())
		if _, err := resource.Update(c.Request.Context(), obj, metav1.UpdateOptions{}); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
			return
		}
	}
	raw, _ := json.Marshal(req)
	rec := model.ClusterReleaseRecord{ClusterID: cluster.ID, Namespace: req.Namespace, App: req.Name, Strategy: req.Strategy, RolloutName: req.Name, Revision: int(req.Replicas), Status: "applied", Operator: strconv.FormatUint(uint64(h.uidFromContext(c)), 10), PayloadJSON: string(raw)}
	_ = h.svcCtx.DB.Create(&rec).Error
	h.createAudit(cluster.ID, req.Namespace, "rollout.apply", "rollout", req.Name, "success", "rollout applied", h.uidFromContext(c))
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"applied": true, "name": req.Name}})
}

func isRolloutCRDMissing(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "the server could not find the requested resource") || strings.Contains(msg, "no matches for kind") || strings.Contains(msg, "argoproj.io")
}

func (h *Handler) ListRollouts(c *gin.Context) {
	if !h.authorize(c, "k8s:read", "k8s:deploy", "kubernetes:read") {
		return
	}
	cluster, ok := h.mustCluster(c)
	if !ok {
		return
	}
	_, dc, err := h.getClients(cluster)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": []any{}, "total": 0}})
		return
	}
	ns := strings.TrimSpace(c.Query("namespace"))
	if ns == "" {
		ns = corev1.NamespaceAll
	}
	if ns != corev1.NamespaceAll && !h.namespaceReadable(c, cluster.ID, ns) {
		return
	}
	items, err := dc.Resource(rolloutGVR).Namespace(ns).List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		if isRolloutCRDMissing(err) {
			c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": []any{}, "total": 0, "diagnostics": []string{"rollout_crd_missing"}}})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	list := make([]gin.H, 0, len(items.Items))
	for _, item := range items.Items {
		strategy := "rolling"
		if _, ok, _ := unstructured.NestedMap(item.Object, "spec", "strategy", "canary"); ok {
			strategy = "canary"
		}
		if _, ok, _ := unstructured.NestedMap(item.Object, "spec", "strategy", "blueGreen"); ok {
			strategy = "blue-green"
		}
		ready, _, _ := unstructured.NestedInt64(item.Object, "status", "readyReplicas")
		replicas, _, _ := unstructured.NestedInt64(item.Object, "status", "replicas")
		phase, _, _ := unstructured.NestedString(item.Object, "status", "phase")
		list = append(list, gin.H{"name": item.GetName(), "namespace": item.GetNamespace(), "strategy": strategy, "phase": phase, "ready_replicas": ready, "replicas": replicas, "created_at": item.GetCreationTimestamp().Time})
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": list, "total": len(list)}})
}

func (h *Handler) RolloutPromote(c *gin.Context) {
	h.rolloutAction(c, "promote", "k8s:deploy")
}

func (h *Handler) RolloutAbort(c *gin.Context) {
	h.rolloutAction(c, "abort", "k8s:deploy")
}

func (h *Handler) RolloutRollback(c *gin.Context) {
	h.rolloutAction(c, "undo", "k8s:rollback")
}

func (h *Handler) rolloutAction(c *gin.Context, action, perm string) {
	if !h.authorize(c, perm, "k8s:write", "kubernetes:write") {
		return
	}
	cluster, ok := h.mustCluster(c)
	if !ok {
		return
	}
	name := strings.TrimSpace(c.Param("name"))
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": "rollout name required"})
		return
	}
	var req struct {
		Namespace     string `json:"namespace"`
		ApprovalToken string `json:"approval_token"`
		Full          bool   `json:"full"`
	}
	_ = c.ShouldBindJSON(&req)
	if strings.TrimSpace(req.Namespace) == "" {
		req.Namespace = strings.TrimSpace(c.Query("namespace"))
	}
	if req.Namespace == "" {
		req.Namespace = "default"
	}
	if !h.namespaceWritable(c, cluster.ID, req.Namespace) {
		return
	}
	if !h.requireProdApproval(c, cluster.ID, req.Namespace, map[string]string{"promote": "deploy", "abort": "rollback", "undo": "rollback"}[action], req.ApprovalToken) {
		return
	}
	if err := h.execRolloutCLI(c.Request.Context(), cluster, req.Namespace, name, action, req.Full); err != nil {
		if isRolloutCRDMissing(err) {
			c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": "rollout_crd_missing"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	h.createAudit(cluster.ID, req.Namespace, "rollout."+action, "rollout", name, "success", "rollout action executed", h.uidFromContext(c))
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"action": action, "name": name}})
}

func (h *Handler) execRolloutCLI(ctx context.Context, cluster *model.Cluster, namespace, name, action string, full bool) error {
	if strings.TrimSpace(cluster.KubeConfig) == "" {
		return errors.New("cluster kubeconfig required for rollout action")
	}
	kubeFile, err := os.CreateTemp("", "cluster-kubeconfig-*.yaml")
	if err != nil {
		return err
	}
	defer os.Remove(kubeFile.Name())
	if _, err := kubeFile.WriteString(cluster.KubeConfig); err != nil {
		_ = kubeFile.Close()
		return err
	}
	_ = kubeFile.Close()

	args := []string{"argo", "rollouts", action, name, "-n", namespace, "--kubeconfig", kubeFile.Name()}
	if action == "promote" && full {
		args = append(args, "--full")
	}
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	msg := strings.ToLower(string(out) + err.Error())
	if strings.Contains(msg, "unknown command") || strings.Contains(msg, "argo") && strings.Contains(msg, "not found") {
		return fmt.Errorf("rollout_cli_missing: kubectl argo rollouts plugin is required")
	}
	if strings.Contains(msg, "no matches for kind") || strings.Contains(msg, "argoproj.io") {
		return fmt.Errorf("rollout_crd_missing")
	}
	return fmt.Errorf("rollout action failed: %s", strings.TrimSpace(string(out)))
}

func (h *Handler) CreateApproval(c *gin.Context) {
	if !h.authorize(c, "k8s:deploy", "k8s:rollback", "kubernetes:write") {
		return
	}
	cluster, ok := h.mustCluster(c)
	if !ok {
		return
	}
	var req struct {
		Namespace string `json:"namespace" binding:"required"`
		Action    string `json:"action" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	ticket := fmt.Sprintf("k8s-appr-%d", time.Now().UnixNano())
	rec := model.ClusterDeployApproval{Ticket: ticket, ClusterID: cluster.ID, Namespace: req.Namespace, Action: req.Action, Status: "pending", RequestBy: h.uidFromContext(c), ExpiresAt: time.Now().Add(30 * time.Minute)}
	if err := h.svcCtx.DB.Create(&rec).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": rec})
}

func (h *Handler) ConfirmApproval(c *gin.Context) {
	if !h.authorize(c, "k8s:approve", "kubernetes:approve") {
		return
	}
	cluster, ok := h.mustCluster(c)
	if !ok {
		return
	}
	ticket := strings.TrimSpace(c.Param("ticket"))
	if ticket == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": "ticket required"})
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	_ = c.ShouldBindJSON(&req)
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "approved"
	}
	if status != "approved" && status != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": "status must be approved or rejected"})
		return
	}
	uid := h.uidFromContext(c)
	result := h.svcCtx.DB.Model(&model.ClusterDeployApproval{}).
		Where("ticket = ? AND cluster_id = ?", ticket, cluster.ID).
		Updates(map[string]any{"status": status, "review_by": uid, "updated_at": time.Now()})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 3000, "msg": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 3004, "msg": "approval ticket not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"ticket": ticket, "status": status}})
}

func (h *Handler) legacyDeployWithApproval(c *gin.Context, cli *kubernetes.Clientset, cluster *model.Cluster, req struct {
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Image     string `json:"image" binding:"required"`
	Replicas  int32  `json:"replicas"`
}) bool {
	if !h.namespaceWritable(c, cluster.ID, req.Namespace) {
		return false
	}
	var body struct {
		ApprovalToken string `json:"approval_token"`
	}
	_ = c.ShouldBindJSON(&body)
	if !h.requireProdApproval(c, cluster.ID, req.Namespace, "deploy", body.ApprovalToken) {
		return false
	}
	return true
}

func deploymentFromRollout(req rolloutApplyReq) *appsv1.Deployment {
	replicas := req.Replicas
	if replicas <= 0 {
		replicas = 1
	}
	labels := map[string]string{"app": req.Name}
	return &appsv1.Deployment{
		TypeMeta:   metav1.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{Name: req.Name, Namespace: req.Namespace, Labels: labels},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{Labels: labels}, Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: req.Name, Image: req.Image}}}},
		},
	}
}

func (h *Handler) patchRolloutPaused(c *gin.Context, cluster *model.Cluster, namespace, name string, paused bool) error {
	_, dc, err := h.getClients(cluster)
	if err != nil {
		return err
	}
	patch := map[string]any{"spec": map[string]any{"paused": paused}}
	raw, _ := json.Marshal(patch)
	_, err = dc.Resource(rolloutGVR).Namespace(namespace).Patch(c.Request.Context(), name, types.MergePatchType, raw, metav1.PatchOptions{})
	return err
}
