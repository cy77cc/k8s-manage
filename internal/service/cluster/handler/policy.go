package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func (h *Handler) uidFromContext(c *gin.Context) uint {
	v, ok := c.Get("uid")
	if !ok {
		return 0
	}
	switch x := v.(type) {
	case uint:
		return x
	case uint64:
		return uint(x)
	case int:
		return uint(x)
	case int64:
		return uint(x)
	case float64:
		return uint(x)
	default:
		return 0
	}
}

func (h *Handler) teamIDFromHeader(c *gin.Context) uint {
	raw := strings.TrimSpace(c.GetHeader("X-Team-ID"))
	if raw == "" {
		return 0
	}
	n, _ := strconv.ParseUint(raw, 10, 64)
	return uint(n)
}

func (h *Handler) isAdminUser(userID uint) bool {
	if userID == 0 {
		return false
	}
	var u model.User
	if err := h.svcCtx.DB.Select("id", "username").Where("id = ?", userID).First(&u).Error; err == nil {
		if strings.EqualFold(strings.TrimSpace(u.Username), "admin") {
			return true
		}
	}
	var rows []struct {
		Code string `gorm:"column:code"`
	}
	if err := h.svcCtx.DB.Table("roles").
		Select("roles.code").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Scan(&rows).Error; err != nil {
		return false
	}
	for _, row := range rows {
		if strings.EqualFold(strings.TrimSpace(row.Code), "admin") {
			return true
		}
	}
	return false
}

func (h *Handler) hasAnyPermission(userID uint, codes ...string) bool {
	if userID == 0 {
		return false
	}
	if h.isAdminUser(userID) {
		return true
	}
	var rows []struct {
		Code string `gorm:"column:code"`
	}
	err := h.svcCtx.DB.Table("permissions").
		Select("permissions.code").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", userID).
		Scan(&rows).Error
	if err != nil {
		return false
	}
	set := map[string]struct{}{}
	for _, r := range rows {
		set[strings.TrimSpace(r.Code)] = struct{}{}
	}
	if _, ok := set["*:*"]; ok {
		return true
	}
	for _, code := range codes {
		if _, ok := set[code]; ok {
			return true
		}
		parts := strings.Split(code, ":")
		if len(parts) > 1 {
			if _, ok := set[parts[0]+":*"]; ok {
				return true
			}
		}
	}
	return false
}

func (h *Handler) authorize(c *gin.Context, codes ...string) bool {
	uid := h.uidFromContext(c)
	if h.hasAnyPermission(uid, codes...) {
		return true
	}
	c.JSON(http.StatusForbidden, gin.H{"code": 2004, "msg": "forbidden"})
	return false
}

func (h *Handler) namespaceWritable(c *gin.Context, clusterID uint, namespace string) bool {
	if namespace == "" {
		return true
	}
	uid := h.uidFromContext(c)
	if h.isAdminUser(uid) {
		return true
	}
	teamID := h.teamIDFromHeader(c)
	if teamID == 0 {
		c.JSON(http.StatusForbidden, gin.H{"code": 2004, "msg": "missing team header"})
		return false
	}
	var binding model.ClusterNamespaceBinding
	err := h.svcCtx.DB.Where("cluster_id = ? AND team_id = ? AND namespace = ?", clusterID, teamID, namespace).First(&binding).Error
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"code": 2004, "msg": "namespace not bound to team"})
		return false
	}
	if binding.Readonly {
		c.JSON(http.StatusForbidden, gin.H{"code": 2004, "msg": "namespace is readonly for team"})
		return false
	}
	return true
}

func (h *Handler) namespaceReadable(c *gin.Context, clusterID uint, namespace string) bool {
	if namespace == "" {
		return true
	}
	uid := h.uidFromContext(c)
	if h.isAdminUser(uid) {
		return true
	}
	teamID := h.teamIDFromHeader(c)
	if teamID == 0 {
		c.JSON(http.StatusForbidden, gin.H{"code": 2004, "msg": "missing team header"})
		return false
	}
	var count int64
	h.svcCtx.DB.Model(&model.ClusterNamespaceBinding{}).
		Where("cluster_id = ? AND team_id = ? AND namespace = ?", clusterID, teamID, namespace).
		Count(&count)
	if count == 0 {
		c.JSON(http.StatusForbidden, gin.H{"code": 2004, "msg": "namespace not bound to team"})
		return false
	}
	return true
}

func (h *Handler) listBoundNamespaces(clusterID, teamID uint) ([]string, error) {
	if teamID == 0 {
		return nil, nil
	}
	var rows []model.ClusterNamespaceBinding
	err := h.svcCtx.DB.Where("cluster_id = ? AND team_id = ?", clusterID, teamID).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.Namespace)
	}
	return out, nil
}

func (h *Handler) isProdNamespace(clusterID uint, namespace string) bool {
	lower := strings.ToLower(strings.TrimSpace(namespace))
	if strings.Contains(lower, "prod") || strings.Contains(lower, "production") {
		return true
	}
	var count int64
	h.svcCtx.DB.Model(&model.ClusterNamespaceBinding{}).
		Where("cluster_id = ? AND namespace = ? AND env = ?", clusterID, namespace, "production").
		Count(&count)
	return count > 0
}

func (h *Handler) requireProdApproval(c *gin.Context, clusterID uint, namespace, action, approvalToken string) bool {
	if !h.isProdNamespace(clusterID, namespace) {
		return true
	}
	uid := h.uidFromContext(c)
	if h.hasAnyPermission(uid, "k8s:approve", "kubernetes:approve") {
		return true
	}
	if strings.TrimSpace(approvalToken) == "" {
		ticket := fmt.Sprintf("k8s-appr-%d", time.Now().UnixNano())
		rec := model.ClusterDeployApproval{
			Ticket:    ticket,
			ClusterID: clusterID,
			Namespace: namespace,
			Action:    action,
			Status:    "pending",
			RequestBy: uid,
			ExpiresAt: time.Now().Add(30 * time.Minute),
		}
		_ = h.svcCtx.DB.Create(&rec).Error
		c.JSON(http.StatusForbidden, gin.H{"code": 2004, "msg": "production action requires k8s:approve or approval_token", "data": gin.H{"approval_required": true, "ticket": ticket, "expires_at": rec.ExpiresAt}})
		return false
	}
	var rec model.ClusterDeployApproval
	if err := h.svcCtx.DB.Where("ticket = ?", approvalToken).First(&rec).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"code": 2004, "msg": "invalid approval token"})
		return false
	}
	if rec.ClusterID != clusterID || rec.Namespace != namespace || rec.Action != action {
		c.JSON(http.StatusForbidden, gin.H{"code": 2004, "msg": "approval token scope mismatch"})
		return false
	}
	if rec.Status != "approved" {
		c.JSON(http.StatusForbidden, gin.H{"code": 2004, "msg": "approval token not approved"})
		return false
	}
	if !rec.ExpiresAt.IsZero() && time.Now().After(rec.ExpiresAt) {
		c.JSON(http.StatusForbidden, gin.H{"code": 2004, "msg": "approval token expired"})
		return false
	}
	return true
}

func (h *Handler) createAudit(clusterID uint, namespace, action, resource, resourceID, status, message string, operatorID uint) {
	rec := model.ClusterOperationAudit{
		ClusterID:  clusterID,
		Namespace:  namespace,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Status:     status,
		Message:    message,
		OperatorID: operatorID,
	}
	_ = h.svcCtx.DB.Create(&rec).Error
}

func (h *Handler) buildRESTConfig(cluster *model.Cluster) (*rest.Config, error) {
	if cluster != nil && strings.TrimSpace(cluster.KubeConfig) != "" {
		return clientcmd.RESTConfigFromKubeConfig([]byte(cluster.KubeConfig))
	}
	if home := homedir.HomeDir(); home != "" {
		cfg, err := clientcmd.BuildConfigFromFlags("", filepath.Join(home, ".kube", "config"))
		if err == nil {
			return cfg, nil
		}
	}
	return rest.InClusterConfig()
}

func (h *Handler) getClients(cluster *model.Cluster) (*kubernetes.Clientset, dynamic.Interface, error) {
	cfg, err := h.buildRESTConfig(cluster)
	if err != nil {
		return nil, nil, err
	}
	cli, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, nil, err
	}
	dc, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, nil, err
	}
	return cli, dc, nil
}

func (h *Handler) tx(fn func(tx *gorm.DB) error) error {
	return h.svcCtx.DB.Transaction(fn)
}
