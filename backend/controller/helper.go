package controller

import (
	"strconv"
	"time"

	"github.com/yndxw/workbuddy-ai/backend/service"
	"github.com/gin-gonic/gin"
)

const timeFormat = "2006-01-02T15:04:05Z07:00"

// parseUintParam 将路径参数转换为 uint64。
func parseUintParam(c *gin.Context, name string) (uint64, error) {
	value := c.Param(name)
	return strconv.ParseUint(value, 10, 64)
}

// parseUintQuery 将查询参数转换为 uint64。
func parseUintQuery(c *gin.Context, name string) (uint64, error) {
	value := c.Query(name)
	if value == "" {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseUint(value, 10, 64)
}

// getUserIDFromHeader 从请求读取当前用户 ID。
// 修复点：优先从上下文读取（由 RequireAuth 校验 Token 后注入），防止伪造 X-User-Id。
func getUserIDFromHeader(c *gin.Context) uint {
	// 1. 优先从上下文获取（安全方式：RequireAuth 校验 Token 后设置）
	if v, ok := c.Get("user_id"); ok {
		if id, ok2 := v.(uint); ok2 {
			return id
		}
	}

	// 2. 降级：从请求头 X-User-Id 读取（非安全方式，建议生产环境仅允许 Token 模式）
	value := c.GetHeader("X-User-Id")
	if value == "" {
		return 0
	}
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0
	}
	return uint(id)
}

// formatTimeValue 按统一格式输出时间字符串。
func formatTimeValue(t time.Time) string {
	return t.Format(timeFormat)
}

// formatTimePointer 在指针为空时返回空字符串。
func formatTimePointer(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(timeFormat)
}

// getTraceID 从请求上下文读取 trace_id（由中间件注入）。
func getTraceID(c *gin.Context) string {
	if v, ok := c.Get("trace_id"); ok {
		if s, ok2 := v.(string); ok2 {
			return s
		}
	}
	return ""
}

// requirePermission 统一的权限校验（基于 X-User-Id）。
// 返回 true 表示允许继续；false 表示已输出错误响应。
func requirePermission(c *gin.Context, userSvc *service.UserService, perm string) bool {
	if userSvc == nil {
		c.JSON(500, gin.H{"error": "权限服务未初始化"})
		return false
	}
	userID := getUserIDFromHeader(c)
	if err := userSvc.CheckPermission(userID, perm); err != nil {
		// 未授权/无权限统一 403（避免泄露过多信息）
		c.JSON(403, gin.H{"error": err.Error()})
		return false
	}
	return true
}
