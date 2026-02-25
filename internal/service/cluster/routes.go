package cluster

import (
	"github.com/cy77cc/k8s-manage/internal/middleware"
	clusterhandler "github.com/cy77cc/k8s-manage/internal/service/cluster/handler"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

func RegisterClusterHandlers(v1 *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	h := clusterhandler.NewHandler(svcCtx)
	g := v1.Group("/clusters", middleware.JWTAuth())
	{
		g.GET("", h.List)
		g.POST("", h.Create)
		g.GET("/:id", h.Get)
		g.GET("/:id/namespaces", h.Namespaces)
		g.POST("/:id/namespaces", h.CreateNamespace)
		g.DELETE("/:id/namespaces/:name", h.DeleteNamespace)
		g.GET("/:id/namespaces/bindings", h.ListNamespaceBindings)
		g.PUT("/:id/namespaces/bindings/:teamId", h.PutNamespaceBindings)
		g.GET("/:id/nodes", h.Nodes)
		g.GET("/:id/deployments", h.Deployments)
		g.GET("/:id/pods", h.Pods)
		g.GET("/:id/services", h.Services)
		g.GET("/:id/ingresses", h.Ingresses)
		g.GET("/:id/events", h.Events)
		g.GET("/:id/logs", h.Logs)
		g.GET("/:id/hpa", h.ListHPA)
		g.POST("/:id/hpa", h.CreateHPA)
		g.PUT("/:id/hpa/:name", h.UpdateHPA)
		g.DELETE("/:id/hpa/:name", h.DeleteHPA)
		g.GET("/:id/quotas", h.ListQuotas)
		g.POST("/:id/quotas", h.CreateOrUpdateQuota)
		g.PUT("/:id/quotas/:name", h.CreateOrUpdateQuota)
		g.DELETE("/:id/quotas/:name", h.DeleteQuota)
		g.GET("/:id/limit-ranges", h.ListLimitRanges)
		g.POST("/:id/limit-ranges", h.CreateLimitRange)
		g.POST("/:id/rollouts/preview", h.RolloutPreview)
		g.POST("/:id/rollouts/apply", h.RolloutApply)
		g.GET("/:id/rollouts", h.ListRollouts)
		g.POST("/:id/rollouts/:name/promote", h.RolloutPromote)
		g.POST("/:id/rollouts/:name/abort", h.RolloutAbort)
		g.POST("/:id/rollouts/:name/rollback", h.RolloutRollback)
		g.POST("/:id/approvals", h.CreateApproval)
		g.POST("/:id/approvals/:ticket/confirm", h.ConfirmApproval)
		g.POST("/:id/connect/test", h.ConnectTest)
		g.POST("/:id/deploy/preview", h.DeployPreview)
		g.POST("/:id/deploy/apply", h.DeployApply)
	}
}
