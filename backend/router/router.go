package router

import (
	"github.com/yndxw/workbuddy-ai/backend/controller"
	"github.com/yndxw/workbuddy-ai/backend/middleware"
	"github.com/gin-gonic/gin"
)

// 控制器集合 用于收集路由需要的控制器实例。
type ControllerSet struct {
	Auth             *controller.AuthController
	Conversation     *controller.ConversationController
	Message          *controller.MessageController
	Admin            *controller.AdminController
	Profile          *controller.ProfileController
	AIConfig         *controller.AIConfigController
	EmbeddingConfig  *controller.EmbeddingConfigController
	PromptConfig     *controller.PromptConfigController
	FAQ              *controller.FAQController
	Document         *controller.DocumentController
	KnowledgeBase    *controller.KnowledgeBaseController
	Import           *controller.ImportController
	Visitor          *controller.VisitorController
	Health           *controller.HealthController
	Analytics        *controller.AnalyticsController
	SystemLog        *controller.SystemLogController
}

// 注册路由 函数用于注册所有的超文本传输协议路由。
func RegisterRoutes(r *gin.Engine, controllers ControllerSet, wsHandler gin.HandlerFunc) {
	// 公开路由注册逻辑
	registerPublic := func(routes gin.IRoutes) {
		// 身份认证相关
		routes.POST("/login", controllers.Auth.Login)
		routes.POST("/logout", controllers.Auth.Logout)

		// 访客相关接口
		routes.POST("/conversation/init", controllers.Conversation.InitConversation)
		routes.GET("/conversations/:id", controllers.Conversation.GetConversationDetail)
		routes.GET("/conversations/ai-models", controllers.Conversation.GetPublicAIModels)
		routes.POST("/messages", controllers.Message.CreateMessage)
		routes.POST("/messages/upload", controllers.Message.UploadFile)
		routes.PUT("/messages/read", controllers.Message.MarkMessagesRead)
		routes.GET("/messages", controllers.Message.ListMessages)

		routes.GET("/visitor/online-agents", controllers.Visitor.GetOnlineAgents)
		routes.GET("/visitor/widget-config", controllers.Visitor.GetWidgetConfig)
		routes.POST("/visitor/analytics/widget-open", controllers.Analytics.PostWidgetOpen)

		// 系统健康状态检查
		routes.GET("/health", controllers.Health.HealthCheck)
		routes.GET("/health/metrics", controllers.Health.Metrics)

		// 实时通信接口
		routes.GET("/ws", wsHandler)
	}

	// 受保护路由注册逻辑
	registerProtected := func(routes gin.IRoutes) {
		// 会话管理
		routes.GET("/conversations", controllers.Conversation.ListConversations)
		routes.POST("/conversations/internal", controllers.Conversation.InitInternalConversation)
		routes.POST("/conversations/:id/close", controllers.Conversation.CloseConversation)
		routes.PUT("/conversations/:id/contact", controllers.Conversation.UpdateContactInfo)
		routes.GET("/conversations/search", controllers.Conversation.SearchConversations)

		// 管理员权限：用户管理
		routes.GET("/admin/users", controllers.Admin.ListUsers)
		routes.GET("/admin/users/:id", controllers.Admin.GetUser)
		routes.POST("/admin/users", controllers.Admin.CreateUser)
		routes.PUT("/admin/users/:id", controllers.Admin.UpdateUser)
		routes.DELETE("/admin/users/:id", controllers.Admin.DeleteUser)
		routes.PUT("/admin/users/:id/password", controllers.Admin.UpdateUserPassword)
		routes.POST("/admin/agents", controllers.Admin.CreateAgent)

		// 个人资料管理
		routes.GET("/agent/profile/:user_id", controllers.Profile.GetProfile)
		routes.PUT("/agent/profile/:user_id", controllers.Profile.UpdateProfile)
		routes.POST("/agent/avatar/:user_id", controllers.Profile.UploadAvatar)

		// 人工智能配置管理
		routes.POST("/agent/ai-config/:user_id", controllers.AIConfig.CreateAIConfig)
		routes.GET("/agent/ai-config/:user_id", controllers.AIConfig.ListAIConfigs)
		routes.GET("/agent/ai-config/:user_id/:id", controllers.AIConfig.GetAIConfig)
		routes.PUT("/agent/ai-config/:user_id/:id", controllers.AIConfig.UpdateAIConfig)
		routes.DELETE("/agent/ai-config/:user_id/:id", controllers.AIConfig.DeleteAIConfig)

		// 嵌入服务与提示词管理
		routes.GET("/agent/embedding-config", controllers.EmbeddingConfig.Get)
		routes.PUT("/agent/embedding-config", controllers.EmbeddingConfig.Update)
		routes.GET("/agent/prompts", controllers.PromptConfig.Get)
		routes.PUT("/agent/prompts", controllers.PromptConfig.Update)

		// 知识库与文档管理
		routes.GET("/faqs", controllers.FAQ.ListFAQs)
		routes.GET("/faqs/:id", controllers.FAQ.GetFAQ)
		routes.POST("/faqs", controllers.FAQ.CreateFAQ)
		routes.PUT("/faqs/:id", controllers.FAQ.UpdateFAQ)
		routes.DELETE("/faqs/:id", controllers.FAQ.DeleteFAQ)

		routes.GET("/documents", controllers.Document.ListDocuments)
		routes.GET("/documents/:id", controllers.Document.GetDocument)
		routes.POST("/documents", controllers.Document.CreateDocument)
		routes.PUT("/documents/:id", controllers.Document.UpdateDocument)
		routes.DELETE("/documents/:id", controllers.Document.DeleteDocument)
		routes.GET("/documents/search", controllers.Document.SearchDocuments)
		routes.GET("/documents/hybrid-search", controllers.Document.HybridSearchDocuments)
		routes.PUT("/documents/:id/status", controllers.Document.UpdateDocumentStatus)
		routes.POST("/documents/:id/publish", controllers.Document.PublishDocument)
		routes.POST("/documents/:id/unpublish", controllers.Document.UnpublishDocument)

		routes.GET("/knowledge-bases", controllers.KnowledgeBase.ListKnowledgeBases)
		routes.GET("/knowledge-bases/:id", controllers.KnowledgeBase.GetKnowledgeBase)
		routes.POST("/knowledge-bases", controllers.KnowledgeBase.CreateKnowledgeBase)
		routes.PUT("/knowledge-bases/:id", controllers.KnowledgeBase.UpdateKnowledgeBase)
		routes.PATCH("/knowledge-bases/:id/rag-enabled", controllers.KnowledgeBase.UpdateKnowledgeBaseRAGEnabled)
		routes.DELETE("/knowledge-bases/:id", controllers.KnowledgeBase.DeleteKnowledgeBase)
		routes.GET("/knowledge-bases/:id/documents", controllers.KnowledgeBase.ListDocumentsByKnowledgeBase)

		// 数据导入
		routes.POST("/import/documents", controllers.Import.ImportDocuments)
		routes.POST("/import/urls", controllers.Import.ImportFromURLs)

		// 数据分析与系统日志
		routes.GET("/agent/analytics/summary", controllers.Analytics.GetSummary)
		routes.GET("/agent/logs/api", controllers.SystemLog.GetLogs)
		routes.GET("/agent/logs/min-level", controllers.SystemLog.GetLogMinLevel)
		routes.PUT("/agent/logs/min-level", controllers.SystemLog.PutLogMinLevel)
		routes.DELETE("/agent/logs/min-level", controllers.SystemLog.DeleteLogMinLevel)
		routes.POST("/agent/logs/frontend", controllers.SystemLog.ReportFrontendLog)
	}

	// 注册公开接口
	registerPublic(r)
	registerPublic(r.Group("/api"))

	// 注册需要令牌校验的私有接口
	protected := r.Group("/", middleware.RequireAuth())
	registerProtected(protected)

	protectedAPI := r.Group("/api", middleware.RequireAuth())
	registerProtected(protectedAPI)
}
