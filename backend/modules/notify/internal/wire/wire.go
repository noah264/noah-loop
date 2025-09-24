package wire

import (
	"github.com/google/wire"
	"github.com/noah-loop/backend/modules/notify/internal/application/service"
	"github.com/noah-loop/backend/modules/notify/internal/domain/repository"
	"github.com/noah-loop/backend/modules/notify/internal/infrastructure/provider"
	infraRepo "github.com/noah-loop/backend/modules/notify/internal/infrastructure/repository"
	"github.com/noah-loop/backend/modules/notify/internal/interface/http"
	"github.com/noah-loop/backend/modules/notify/internal/interface/http/handler"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"github.com/noah-loop/backend/shared/pkg/infrastructure/etcd"
	"github.com/noah-loop/backend/shared/pkg/infrastructure/tracing"
	"gorm.io/gorm"
)

// NotifyApp 通知应用结构
type NotifyApp struct {
	NotificationService *service.NotificationService
	TemplateService     *service.TemplateService
	ChannelService      *service.ChannelService
	Handler             *handler.NotifyHandler
	Router              *http.Router
	Config              *infrastructure.Config
	Logger              infrastructure.Logger
	Metrics             *infrastructure.MetricsRegistry
	Database            *gorm.DB

	// etcd相关组件
	EtcdClient       *etcd.Client
	ServiceRegistry  *etcd.ServiceRegistry
	ServiceDiscovery *etcd.ServiceDiscovery
	ConfigManager    *etcd.ConfigManager
	SecretManager    *etcd.SecretManager

	// 链路追踪组件
	TracerManager  *tracing.TracerManager
	TracingWrapper *tracing.TracingWrapper

	// 通知提供商
	EmailProvider   service.EmailProvider
	SMSProvider     service.SMSProvider
	PushProvider    service.PushProvider
	WebhookProvider service.WebhookProvider
}

// NotifyRepositoryProviderSet 通知仓储提供者集合
var NotifyRepositoryProviderSet = wire.NewSet(
	infraRepo.NewGormNotificationRepository,
	// TODO: 添加其他仓储实现
	wire.Bind(new(repository.NotificationRepository), new(*infraRepo.GormNotificationRepository)),
)

// NotifyProviderSet 通知提供商集合
var NotifyProviderSet = wire.NewSet(
	provider.NewSMTPEmailProvider,
	provider.NewAliyunSMSProvider,
	provider.NewBarkPushProvider,
	provider.NewServerChanWebhookProvider,
	wire.Bind(new(service.EmailProvider), new(*provider.SMTPEmailProvider)),
	wire.Bind(new(service.SMSProvider), new(*provider.AliyunSMSProvider)),
	wire.Bind(new(service.PushProvider), new(*provider.BarkPushProvider)),
	wire.Bind(new(service.WebhookProvider), new(*provider.ServerChanWebhookProvider)),
)

// NotifyServiceProviderSet 通知服务提供者集合
var NotifyServiceProviderSet = wire.NewSet(
	service.NewNotificationService,
	service.NewTemplateService,
	service.NewChannelService,
)

// NotifyHandlerProviderSet 通知处理器提供者集合
var NotifyHandlerProviderSet = wire.NewSet(
	handler.NewNotifyHandler,
	http.NewRouter,
)

// InitializeNotifyApp 初始化通知应用
func InitializeNotifyApp() (*NotifyApp, func(), error) {
	wire.Build(
		// 基础设施
		infrastructure.InfrastructureProviderSet,

		// etcd相关
		infrastructure.EtcdProviderSet,

		// 链路追踪相关
		infrastructure.TracingProviderSet,

		// 通知特定组件
		NotifyRepositoryProviderSet,
		NotifyProviderSet,
		NotifyServiceProviderSet,
		NotifyHandlerProviderSet,

		// 应用结构
		wire.Struct(new(NotifyApp), "*"),

		// 提供服务名称
		wire.Value("notify-service"),
	)

	return &NotifyApp{}, nil, nil
}
