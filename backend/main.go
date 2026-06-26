package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yndxw/workbuddy-ai/backend/controller"
	"github.com/yndxw/workbuddy-ai/backend/infra"
	"github.com/yndxw/workbuddy-ai/backend/infra/mcp"
	infra_search "github.com/yndxw/workbuddy-ai/backend/infra/search"
	"github.com/yndxw/workbuddy-ai/backend/middleware"
	"github.com/yndxw/workbuddy-ai/backend/models"
	"github.com/yndxw/workbuddy-ai/backend/repository"
	appRouter "github.com/yndxw/workbuddy-ai/backend/router"
	"github.com/yndxw/workbuddy-ai/backend/service"
	"github.com/yndxw/workbuddy-ai/backend/service/embedding"
	"github.com/yndxw/workbuddy-ai/backend/service/rag"
	"github.com/yndxw/workbuddy-ai/backend/websocket"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	milvus "github.com/milvus-io/milvus-sdk-go/v2/client"
	"golang.org/x/crypto/bcrypt"
)

// 初始化默认管理员账号（如果不存在）
// 用户名从环境变量中读取（默认值：管理员）
// 密码从环境变量中读取（必须设置）
func initDefaultAdmin(userRepo *repository.UserRepository) {
	// 从环境变量读取管理员用户名和密码
	adminUsername := os.Getenv("ADMIN_USERNAME")
	if adminUsername == "" {
		adminUsername = "admin" // 默认用户名
	}

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		log.Println("⚠️ 警告：未设置管理员密码环境变量，跳过创建默认管理员账号")
		log.Println("   请在环境变量文件中设置密码后重启服务")
		return
	}

	// 检查管理员账号是否已存在
	if _, err := userRepo.FindByUsername(adminUsername); err == nil {
		log.Printf("✅ 管理员账号 '%s' 已存在", adminUsername)
		return
	}

	// 对密码进行加密处理
	hash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("⚠️ 创建默认管理员失败：密码加密错误 %v", err)
		return
	}

	admin := &models.User{
		Username: adminUsername,
		Password: string(hash),
		Role:     "admin",
	}

	if err := userRepo.Create(admin); err != nil {
		log.Printf("⚠️ 创建默认管理员失败：%v", err)
		return
	}

	log.Printf("✅ 默认管理员账号创建成功")
	log.Printf("   用户名: %s", adminUsername)
	log.Println("   ⚠️ 请首次登录后立即修改密码！")
}

// 将向量数据库启动相关事件写入系统日志，供前端日志中心查询
func logVectorStartup(sys *service.SystemLogService, level, event, message string, meta map[string]interface{}) {
	if sys == nil {
		return
	}
	if meta == nil {
		meta = map[string]interface{}{}
	}
	if err := sys.Create(service.CreateSystemLogInput{
		Level:    level,
		Category: "vector",
		Event:    event,
		Source:   "backend",
		Message:  message,
		Meta:     meta,
	}); err != nil {
		log.Printf("写入系统日志失败 (事件=%s): %v", event, err)
	}
}

// 在启动阶段先写入一条向量错误日志，再执行强制退出
func fatalVectorStartup(sys *service.SystemLogService, event, message string, meta map[string]interface{}) {
	logVectorStartup(sys, "error", event, message, meta)
	log.Fatalf("%s", message)
}

func main() {

	// 加载环境变量文件（统一配置源：优先当前目录，其次上级目录）
	wd, _ := os.Getwd()
	candidates := []string{
		filepath.Join(wd, ".env"),
		filepath.Join(wd, "..", ".env"),
	}
	envPath := ""
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			envPath = p
			break
		}
	}
	if envPath == "" {
		log.Printf("⚠️ 未找到环境变量文件（已检查位置: %v）", candidates)
		log.Println("将仅使用操作系统环境变量")
	} else {
		log.Printf("✅ 找到环境变量文件: %s", envPath)
	}

	// 尝试加载环境变量
	if envPath != "" {
		if err := godotenv.Load(envPath); err != nil {
			log.Printf("❌ 加载环境变量文件失败: %v", err)
			log.Println("⚠️ 提示：如果看到字符错误，可能是文件编码问题")
			log.Println("将使用操作系统环境变量")
		} else {
			log.Println("✅ 环境变量加载成功")
		}
	}

	db, err := infra.NewDB()
	if err != nil {
		log.Fatalf("数据库连接失败：%v", err)
	}

	// 根据结构体定义自动创建和更新数据库表
	if err := db.AutoMigrate(&models.User{}, &models.Conversation{}, &models.Message{}, &models.AIConfig{}, &models.FAQ{}, &models.KnowledgeBase{}, &models.Document{}, &models.EmbeddingConfig{}, &models.PromptConfig{}, &models.WidgetOpenEvent{}, &models.SystemLog{}, &models.AppSetting{}); err != nil {
		log.Fatalf("自动创建表失败： %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	conversationRepo := repository.NewConversationRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	aiConfigRepo := repository.NewAIConfigRepository(db)
	faqRepo := repository.NewFAQRepository(db)
	kbRepo := repository.NewKnowledgeBaseRepository(db)
	docRepo := repository.NewDocumentRepository(db)
	embeddingConfigRepo := repository.NewEmbeddingConfigRepository(db)
	promptConfigRepo := repository.NewPromptConfigRepository(db)
	systemLogRepo := repository.NewSystemLogRepository(db)
	appSettingRepo := repository.NewAppSettingRepository(db)
	systemLogMin := service.SystemLogMinPersistLevelFromEnv()
	systemLogService := service.NewSystemLogService(systemLogRepo, systemLogMin)
	if row, err := appSettingRepo.Get(models.AppSettingKeySystemLogMinLevel); err == nil && row != nil && strings.TrimSpace(row.Value) != "" {
		dbRank := service.ParseSystemLogMinPersistLevel(row.Value)
		systemLogService.SetMinPersistLevelRank(dbRank)
		log.Printf("ℹ️ 结构化日志最低保存级别: %s（数据库配置覆盖，环境默认 %s）",
			service.SystemLogMinLevelLabel(dbRank), service.SystemLogMinLevelLabel(systemLogMin))
	} else if systemLogMin == -1 {
		log.Println("ℹ️ 日志级别设置为关闭，已停止向数据库写入日志记录")
	} else {
		log.Printf("ℹ️ 结构化日志最低保存级别: %s", service.SystemLogMinLevelLabel(systemLogMin))
	}

	// 初始化默认管理员账号
	initDefaultAdmin(userRepo)

	// 超文本传输协议框架初始化
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 追踪标识、结构化超文本协议日志以及控制台日志
	r.Use(middleware.TraceID())
	r.Use(middleware.StructuredHTTPLogger(systemLogService))
	r.Use(middleware.Logger())

	// 跨域资源共享配置
	r.Use(middleware.CORS())

	// 初始化本地文件存储服务
	uploadDir := filepath.Join(wd, "uploads")
	publicPath := "/uploads"
	storageService := infra.NewLocalStorageService(uploadDir, publicPath)

	// 初始化向量数据库
	milvusDisabled := infra.IsMilvusDisabled()
	milvusRequired := infra.IsMilvusRequired()
	var milvusClient milvus.Client
	defer func() {
		if milvusClient != nil {
			if err := milvusClient.Close(); err != nil {
				log.Printf("关闭向量数据库客户端失败: %v", err)
			}
		}
	}()
	var vectorStore *infra.VectorStore
	milvusCfg := infra.GetMilvusConfig()
	milvusMeta := map[string]interface{}{
		"milvus_host":     milvusCfg.Host,
		"milvus_port":     milvusCfg.Port,
		"milvus_required": milvusRequired,
		"milvus_disabled": milvusDisabled,
	}

	if milvusDisabled {
		log.Println("ℹ️ 已禁用向量数据库；知识库检索与向量化功能不可用")
		logVectorStartup(systemLogService, "info", "milvus_disabled",
			"已禁用向量数据库；知识库检索与向量化功能不可用，启用后需重启",
			milvusMeta)
	} else {
		c, err := infra.NewMilvusClient()
		if err != nil {
			if milvusRequired {
				m := map[string]interface{}{}
				for k, v := range milvusMeta {
					m[k] = v
				}
				m["error"] = err.Error()
				fatalVectorStartup(systemLogService, "milvus_required_connect_failed",
					"连接向量数据库失败（已设置为必须连接）", m)
			}
			log.Printf("⚠️ 连接向量数据库失败，将以无向量库模式启动: %v", err)
			m := map[string]interface{}{}
			for k, v := range milvusMeta {
				m[k] = v
			}
			m["error"] = err.Error()
			logVectorStartup(systemLogService, "warn", "milvus_connect_failed",
				"连接向量数据库失败，已降级为无向量库模式启动", m)
		} else {
			milvusClient = c
			if err := infra.HealthCheck(milvusClient); err != nil {
				_ = milvusClient.Close()
				milvusClient = nil
				if milvusRequired {
					m := map[string]interface{}{}
					for k, v := range milvusMeta {
						m[k] = v
					}
					m["error"] = err.Error()
					fatalVectorStartup(systemLogService, "milvus_required_health_check_failed",
						"向量数据库健康检查失败（已设置为必须通过）", m)
				}
				log.Printf("⚠️ 向量数据库健康检查失败，将以无向量库模式启动: %v", err)
				m := map[string]interface{}{}
				for k, v := range milvusMeta {
					m[k] = v
				}
				m["error"] = err.Error()
				logVectorStartup(systemLogService, "warn", "milvus_health_check_failed",
					"向量数据库健康检查失败，已降级为无向量库模式启动", m)
			} else {
				log.Println("✅ 向量数据库连接成功")
			}
		}
	}

	// 嵌入服务按需从数据库配置获取
	embeddingConfigService := service.NewEmbeddingConfigService(embeddingConfigRepo, userRepo)
	promptConfigService := service.NewPromptConfigService(promptConfigRepo, userRepo)
	embeddingFactory := embedding.NewEmbeddingFactory()
	embeddingProvider := service.NewConfigBackedEmbeddingProvider(embeddingConfigService, embeddingFactory)

	// 启动时获取一次维度用于创建或校验向量集合
	initCtx := context.Background()
	initSvc, _ := embeddingProvider.Get(initCtx)
	if initSvc != nil {
		log.Printf("✅ 嵌入服务已从数据库加载，模型: %s (维度: %d)", initSvc.GetModelName(), initSvc.GetDimension())
	} else {
		log.Printf("⚠️ 未配置嵌入服务；请在设置中配置后使用")
	}
	dimension := 1536
	if initSvc != nil {
		dimension = initSvc.GetDimension()
	}

	// 向量存储服务初始化
	getEmbedding := func(ctx context.Context) (infra.EmbeddingService, error) {
		svc, err := embeddingProvider.Get(ctx)
		if err != nil || svc == nil {
			return nil, err
		}
		return svc, nil
	}
	if milvusClient != nil {
		vs, err := infra.NewVectorStore(milvusClient, "documents", dimension, getEmbedding)
		if err != nil {
			_ = milvusClient.Close()
			milvusClient = nil
			if milvusRequired {
				m := map[string]interface{}{}
				for k, v := range milvusMeta {
					m[k] = v
				}
				m["error"] = err.Error()
				fatalVectorStartup(systemLogService, "milvus_required_vector_store_init_failed",
					"创建向量存储失败", m)
			}
			log.Printf("⚠️ 创建向量存储失败，将以无向量库模式启动: %v", err)
			m := map[string]interface{}{}
			for k, v := range milvusMeta {
				m[k] = v
			}
			m["error"] = err.Error()
			logVectorStartup(systemLogService, "warn", "milvus_vector_store_init_failed",
				"创建向量存储失败，已降级为无向量库模式启动", m)
		} else {
			vectorStore = vs
		}
	}
	if vectorStore != nil {
		okMeta := map[string]interface{}{}
		for k, v := range milvusMeta {
			okMeta[k] = v
		}
		okMeta["collection"] = "documents"
		logVectorStartup(systemLogService, "info", "milvus_ready",
			"向量数据库已连接且向量集合可用", okMeta)
	}
	vectorStoreService := rag.NewVectorStoreService(vectorStore)

	// 文档向量化与检索服务初始化
	documentEmbeddingService := rag.NewDocumentEmbeddingService(vectorStoreService, embeddingProvider)
	retrievalService := rag.NewRetrievalService(vectorStoreService, embeddingProvider, docRepo, kbRepo)
	retrievalService.EnableCache(5 * time.Minute)
	healthChecker := rag.NewHealthChecker(embeddingProvider, vectorStoreService)

	// 联网搜索服务初始化（可选）
	var webSearchProvider infra_search.WebSearchProvider
	if mcpURL := os.Getenv("SERPER_MCP_URL"); mcpURL != "" {
		mcpClient := mcp.NewClient(mcpURL)
		if err := mcpClient.Connect(initCtx); err != nil {
			log.Printf("⚠️ 联网搜索工具连接失败: %v", err)
		} else {
			webSearchProvider = mcp.NewSerperWebSearchProvider(mcpClient)
			log.Println("✅ 联网搜索已接入")
		}
	}
	if webSearchProvider == nil {
		if apiKey := os.Getenv("SERPER_API_KEY"); apiKey != "" {
			webSearchProvider = infra_search.NewSerperProvider(apiKey)
			log.Println("✅ 联网搜索已通过接口接入")
		}
	}

	// 初始化核心服务层
	authService := service.NewAuthService(userRepo)
	conversationService := service.NewConversationService(conversationRepo, messageRepo, aiConfigRepo, userRepo, systemLogService)
	profileService := service.NewProfileService(userRepo, storageService)
	aiConfigService := service.NewAIConfigService(aiConfigRepo, userRepo)
	aiService := service.NewAIService(aiConfigRepo, messageRepo, conversationRepo, retrievalService, webSearchProvider, embeddingConfigService, promptConfigService, storageService, systemLogService)
	userService := service.NewUserService(userRepo, aiConfigRepo)
	faqService := service.NewFAQService(faqRepo, retrievalService, documentEmbeddingService)
	documentService := service.NewDocumentService(docRepo, kbRepo, documentEmbeddingService, retrievalService)
	knowledgeBaseService := service.NewKnowledgeBaseService(kbRepo, docRepo)
	importService := service.NewImportService(docRepo, kbRepo, documentService, documentEmbeddingService)

	// 声明网络套接字中心变量
	var wsHub *websocket.Hub

	// 处理客户端连接与断开事件的回调逻辑
	onConnect := func(conversationID uint, isVisitor bool, visitorCount int, agentID uint) {
		if isVisitor {
			if err := conversationService.UpdateVisitorOnlineStatus(conversationID, true); err != nil {
				log.Printf("更新访客在线状态失败: %v", err)
				return
			}
			// 广播状态更新到所有客服端
			wsHub.BroadcastToAllAgents("visitor_status_update", map[string]interface{}{
				"conversation_id": conversationID,
				"is_online":       true,
				"visitor_count":   visitorCount,
			})
		} else if agentID > 0 {
			// 客服加入会话处理
			agent, err := userRepo.GetByID(agentID)
			if err != nil {
				log.Printf("获取客服信息失败: %v", err)
				return
			}
			agentName := agent.Nickname
			if agentName == "" {
				agentName = agent.Username
			}
			hasJoinMessage, err := messageRepo.HasAgentJoinMessage(conversationID, agentID, agentName)
			if err != nil {
				log.Printf("检查客服加入消息失败: %v", err)
				return
			}
			if hasJoinMessage {
				return
			}
			conv, err := conversationRepo.GetByID(conversationID)
			if err != nil {
				log.Printf("获取对话信息失败: %v", err)
				return
			}
			now := time.Now()
			chatMode := conv.ChatMode
			if chatMode == "" {
				chatMode = "human"
			}
			systemMessage := &models.Message{
				ConversationID: conversationID,
				SenderID:       agentID,
				SenderIsAgent:  true,
				Content:        agentName + "加入了会话",
				MessageType:    "system_message",
				ChatMode:       chatMode,
				IsRead:         true,
				ReadAt:         &now,
			}
			if err := messageRepo.Create(systemMessage); err != nil {
				log.Printf("创建系统消息失败: %v", err)
				return
			}
			go func() {
				time.Sleep(100 * time.Millisecond)
				wsHub.BroadcastMessage(conversationID, "new_message", systemMessage)
			}()
		}
	}

	onDisconnect := func(conversationID uint, isVisitor bool, visitorCount int) {
		if isVisitor {
			if visitorCount == 0 {
				if err := conversationService.UpdateVisitorOnlineStatus(conversationID, false); err != nil {
					log.Printf("更新访客离线状态失败: %v", err)
					return
				}
				wsHub.BroadcastToAllAgents("visitor_status_update", map[string]interface{}{
					"conversation_id": conversationID,
					"is_online":       false,
					"visitor_count":   0,
				})
			} else {
				if err := conversationService.UpdateLastSeenAt(conversationID); err != nil {
					log.Printf("更新最后活跃时间失败: %v", err)
					return
				}
			}
		}
	}

	// 创建网络套接字分发中心
	wsBus, wsBusErr := websocket.NewRedisBusFromEnv()
	if wsBusErr != nil {
		log.Printf("⚠️ 分布式消息总线初始化失败: %v", wsBusErr)
	}
	if wsBus != nil {
		defer func() {
			if err := wsBus.Close(); err != nil {
				log.Printf("关闭消息总线失败: %v", err)
			}
		}()
		log.Println("✅ 已启用分布式消息广播")
	}
	wsHub = websocket.NewHub(onConnect, onDisconnect, wsBus)
	go wsHub.Run()

	messageService := service.NewMessageService(db, conversationRepo, messageRepo, wsHub, aiService)
	visitorService := service.NewVisitorService(userRepo, wsHub)

	// 初始化控制器
	authController := controller.NewAuthController(authService)
	conversationController := controller.NewConversationController(conversationService, aiConfigService, userService)
	messageController := controller.NewMessageController(messageService, conversationService, userService, storageService)
	adminController := controller.NewAdminController(authService, userService)
	profileController := controller.NewProfileController(profileService)
	aiConfigController := controller.NewAIConfigController(aiConfigService, userService)
	faqController := controller.NewFAQController(faqService, userService)
	documentController := controller.NewDocumentController(documentService, embeddingConfigService, userService)
	embeddingConfigController := controller.NewEmbeddingConfigController(embeddingConfigService, userService)
	promptConfigController := controller.NewPromptConfigController(promptConfigService, userService)
	knowledgeBaseController := controller.NewKnowledgeBaseController(knowledgeBaseService, embeddingConfigService, userService)
	importController := controller.NewImportController(importService, embeddingConfigService, userService)
	visitorController := controller.NewVisitorController(visitorService, embeddingConfigService)
	healthController := controller.NewHealthController(healthChecker, retrievalService)

	widgetOpenRepo := repository.NewWidgetOpenRepository(db)
	analyticsService := service.NewAnalyticsService(db, widgetOpenRepo)
	analyticsController := controller.NewAnalyticsController(analyticsService, userService)
	systemLogController := controller.NewSystemLogController(systemLogService, userService, appSettingRepo)

	appRouter.RegisterRoutes(
		r,
		appRouter.ControllerSet{
			Auth:            authController,
			Conversation:    conversationController,
			Message:         messageController,
			Admin:           adminController,
			Profile:         profileController,
			AIConfig:        aiConfigController,
			EmbeddingConfig: embeddingConfigController,
			PromptConfig:    promptConfigController,
			FAQ:             faqController,
			Document:        documentController,
			KnowledgeBase:   knowledgeBaseController,
			Import:          importController,
			Visitor:         visitorController,
			Health:          healthController,
			Analytics:       analyticsController,
			SystemLog:       systemLogController,
		},
		websocket.HandleWebSocket(wsHub, userRepo),
	)

	r.Static("/uploads", uploadDir)

	// 启动服务器并监听指定网络接口与端口
	host := os.Getenv("SERVER_HOST")
	if host == "" {
		host = "0.0.0.0"
	}
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "18080"
	}
	addr := host + ":" + port
	log.Println("🚀 服务器启动成功，监听地址: " + addr)
	r.Run(addr)
}
