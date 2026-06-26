package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/yndxw/workbuddy-ai/backend/service"
	"github.com/yndxw/workbuddy-ai/backend/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func newTraceID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b[:])
}

// TraceID 为每个请求注入 trace_id，便于链路排障。
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-Id")
		if traceID == "" {
			traceID = newTraceID()
		}
		c.Set("trace_id", traceID)
		c.Writer.Header().Set("X-Trace-Id", traceID)
		c.Next()
	}
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Printf("[GIN] %s %s %d %s",
			c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(start))
	}
}

// StructuredHTTPLogger 将 HTTP 请求结构化落库（分类: http）。
func StructuredHTTPLogger(logSvc *service.SystemLogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		if logSvc == nil {
			return
		}
		latencyMs := time.Since(start).Milliseconds()
		status := c.Writer.Status()
		level := "info"
		if status >= 500 {
			level = "error"
		} else if status >= 400 || latencyMs >= 2000 {
			level = "warn"
		}
		if !logSvc.ShouldPersistLevel(level) {
			return
		}
		var userID *uint
		if v := c.GetHeader("X-User-Id"); v != "" {
			if id, err := strconv.ParseUint(v, 10, 64); err == nil && id > 0 {
				t := uint(id)
				userID = &t
			}
		}
		traceID := ""
		if v, ok := c.Get("trace_id"); ok {
			if s, ok2 := v.(string); ok2 {
				traceID = s
			}
		}
		_ = logSvc.Create(service.CreateSystemLogInput{
			Level:   level,
			Category: "http",
			Event:   "http_request",
			Source:  "backend",
			TraceID: traceID,
			UserID:  userID,
			Message: c.Request.Method + " " + c.Request.URL.Path,
			Meta: map[string]interface{}{
				"status":     status,
				"latency_ms": latencyMs,
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
				"query":      c.Request.URL.RawQuery,
			},
		})
	}
}

func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "X-User-Id", "X-Trace-Id", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// RequireAuth 认证中间件：要求请求头中包含有效的令牌 (Bearer Token) 或有效的 X-User-Id
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		var userID uint

		// 1. 优先尝试从 Authorization Header 获取 Bearer Token
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token := authHeader[7:]
			uid, parseErr := utils.ParseWSToken(token)
			if parseErr == nil {
				userID = uid
			} else {
				log.Printf("⚠️ 令牌校验失败: %v", parseErr)
			}
		}

		// 2. 降级逻辑：如果令牌无效，且没有令牌，则检查 X-User-Id (仅用于开发调试)
		if userID == 0 {
			userIDStr := c.GetHeader("X-User-Id")
			if userIDStr != "" {
				id, parseErr := strconv.ParseUint(userIDStr, 10, 64)
				if parseErr == nil && id > 0 {
					userID = uint(id)
				}
			}
		}

		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问，请重新登录"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
