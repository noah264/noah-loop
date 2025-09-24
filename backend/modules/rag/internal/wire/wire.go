package wire

import (
	"github.com/google/wire"
	"github.com/noah-loop/backend/modules/rag/internal/application/service"
	"github.com/noah-loop/backend/modules/rag/internal/domain/repository"
	"github.com/noah-loop/backend/modules/rag/internal/infrastructure/embedding"
	infraRepo "github.com/noah-loop/backend/modules/rag/internal/infrastructure/repository"
	"github.com/noah-loop/backend/modules/rag/internal/infrastructure/vector"
	"github.com/noah-loop/backend/modules/rag/internal/interface/http"
	"github.com/noah-loop/backend/modules/rag/internal/interface/http/handler"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"github.com/noah-loop/backend/shared/pkg/infrastructure/etcd"
	"github.com/noah-loop/backend/shared/pkg/infrastructure/tracing"
	"gorm.io/gorm"
)

// RAGApp RAG应用结构
type RAGApp struct {
	RAGService      *service.RAGService
	Handler         *handler.RAGHandler
	Router          *http.Router
	Config          *infrastructure.Config
	Logger          infrastructure.Logger
	Metrics         *infrastructure.MetricsRegistry
	Database        *gorm.DB

	// etcd相关组件
	EtcdClient      *etcd.Client
	ServiceRegistry *etcd.ServiceRegistry
	ServiceDiscovery *etcd.ServiceDiscovery
	ConfigManager   *etcd.ConfigManager
	SecretManager   *etcd.SecretManager

	// 链路追踪组件
	TracerManager   *tracing.TracerManager
	TracingWrapper  *tracing.TracingWrapper

	// RAG特定组件
	EmbeddingService service.EmbeddingService
	ChunkingService  service.ChunkingService
}

// RAGRepositoryProviderSet RAG仓储提供者集合
var RAGRepositoryProviderSet = wire.NewSet(
	infraRepo.NewGormDocumentRepository,
	infraRepo.NewGormKnowledgeBaseRepository,
	infraRepo.NewGormChunkRepository,
	wire.Bind(new(repository.DocumentRepository), new(*infraRepo.GormDocumentRepository)),
	wire.Bind(new(repository.KnowledgeBaseRepository), new(*infraRepo.GormKnowledgeBaseRepository)),
	wire.Bind(new(repository.ChunkRepository), new(*infraRepo.GormChunkRepository)),
)

// RAGVectorProviderSet RAG向量提供者集合
var RAGVectorProviderSet = wire.NewSet(
	NewMilvusConfig,
	vector.NewMilvusVectorRepository,
	wire.Bind(new(repository.VectorRepository), new(*vector.MilvusVectorRepository)),
)

// RAGServiceProviderSet RAG服务提供者集合
var RAGServiceProviderSet = wire.NewSet(
	// 嵌入服务
	NewEmbeddingConfig,
	embedding.NewOpenAIEmbeddingService,
	wire.Bind(new(service.EmbeddingService), new(*embedding.OpenAIEmbeddingService)),

	// 分块服务
	NewChunkingConfig,
	service.NewDefaultChunkingService,
	wire.Bind(new(service.ChunkingService), new(*service.DefaultChunkingService)),

	// 主服务
	service.NewRAGService,
)

// RAGHandlerProviderSet RAG处理器提供者集合
var RAGHandlerProviderSet = wire.NewSet(
	handler.NewRAGHandler,
	http.NewRouter,
)

// InitializeRAGApp 初始化RAG应用
func InitializeRAGApp() (*RAGApp, func(), error) {
	wire.Build(
		// 基础设施
		infrastructure.InfrastructureProviderSet,

		// etcd相关
		infrastructure.EtcdProviderSet,

		// 链路追踪相关
		infrastructure.TracingProviderSet,

		// RAG特定组件
		RAGRepositoryProviderSet,
		RAGVectorProviderSet,
		RAGServiceProviderSet,
		RAGHandlerProviderSet,

		// 应用结构
		wire.Struct(new(RAGApp), "*"),

		// 提供服务名称
		wire.Value("rag-service"),
	)

	return &RAGApp{}, nil, nil
}

// NewEmbeddingConfig 创建嵌入配置
func NewEmbeddingConfig(config *infrastructure.Config, secretManager *etcd.SecretManager) *service.EmbeddingConfig {
	embeddingConfig := service.DefaultEmbeddingConfig()

	// 从etcd获取OpenAI API密钥
	if secretManager != nil {
		if apiKey, err := secretManager.GetSecret("openai_api_key"); err == nil && apiKey != "" {
			embeddingConfig.APIKey = apiKey
		}
	}

	// 可以从配置文件覆盖设置
	// embeddingConfig.Model = config.RAG.EmbeddingModel
	// embeddingConfig.Dimension = config.RAG.EmbeddingDimension

	return embeddingConfig
}

// NewChunkingConfig 创建分块配置
func NewChunkingConfig(config *infrastructure.Config) *service.ChunkingConfig {
	chunkingConfig := service.DefaultChunkingConfig()

	// 可以从配置文件覆盖设置
	// chunkingConfig.ChunkSize = config.RAG.ChunkSize
	// chunkingConfig.ChunkOverlap = config.RAG.ChunkOverlap

	return chunkingConfig
}

// NewMilvusConfig 创建Milvus配置
func NewMilvusConfig(config *infrastructure.Config) *vector.MilvusConfig {
	return &vector.MilvusConfig{
		Host:       "localhost",
		Port:       19530,
		Database:   "default",
		Timeout:    30,
		MaxRetries: 3,
	}

	// 可以从配置文件覆盖设置
	// return &vector.MilvusConfig{
	//     Host:       config.Vector.Host,
	//     Port:       config.Vector.Port,
	//     Username:   config.Vector.Username,
	//     Password:   config.Vector.Password,
	//     Database:   config.Vector.Database,
	//     Timeout:    config.Vector.Timeout,
	//     MaxRetries: config.Vector.MaxRetries,
	// }
}
