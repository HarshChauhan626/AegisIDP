package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Permission constants used in route registration.
const (
	PermEnvironmentCreate = "environment:create"
	PermEnvironmentDelete = "environment:delete"
	PermWorkflowView      = "workflow:view"
	PermWorkflowCancel    = "workflow:cancel"
	PermWorkflowRetry     = "workflow:retry"
	PermTemplateManage    = "template:manage"
	PermScheduleManage    = "schedule:manage"
	PermMetricsView       = "metrics:view"
	PermAuditView         = "audit:view"
	PermUserManage        = "user:manage"
)

// rolePermissions maps each role to its set of allowed permissions.
var rolePermissions = map[string]map[string]bool{
	"admin": {
		PermEnvironmentCreate: true,
		PermEnvironmentDelete: true,
		PermWorkflowView:      true,
		PermWorkflowCancel:    true,
		PermWorkflowRetry:     true,
		PermTemplateManage:    true,
		PermScheduleManage:    true,
		PermMetricsView:       true,
		PermAuditView:         true,
		PermUserManage:        true,
	},
	"developer": {
		PermEnvironmentCreate: true,
		PermEnvironmentDelete: true,
		PermWorkflowView:      true,
		PermWorkflowCancel:    true,
		PermWorkflowRetry:     true,
		PermTemplateManage:    true,
		PermScheduleManage:    true,
		PermMetricsView:       true,
	},
	"viewer": {
		PermWorkflowView: true,
		PermMetricsView:  true,
	},
}

// RequirePermission returns a middleware that enforces a specific permission for a route.
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := GetRole(c)
		if role == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": gin.H{
				"code":    "FORBIDDEN",
				"message": "No role assigned",
			}})
			return
		}

		perms, ok := rolePermissions[role]
		if !ok || !perms[permission] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": gin.H{
				"code":    "FORBIDDEN",
				"message": "Insufficient permissions",
			}})
			return
		}
		c.Next()
	}
}

// HasPermission checks whether a role has a given permission (for use in handlers).
func HasPermission(role, permission string) bool {
	perms, ok := rolePermissions[role]
	return ok && perms[permission]
}
