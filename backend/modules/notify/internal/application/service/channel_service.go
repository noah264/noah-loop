package service

import (
	"context"
	"fmt"

	"github.com/noah-loop/backend/modules/notify/internal/domain"
	"github.com/noah-loop/backend/modules/notify/internal/domain/repository"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"go.uber.org/zap"
)

// ChannelService 渠道服务
type ChannelService struct {
	channelRepo     repository.ChannelRepository
	emailProvider   EmailProvider
	smsProvider     SMSProvider
	pushProvider    PushProvider
	webhookProvider WebhookProvider
	logger          infrastructure.Logger
}

// NewChannelService 创建渠道服务
func NewChannelService(
	channelRepo repository.ChannelRepository,
	emailProvider EmailProvider,
	smsProvider SMSProvider,
	pushProvider PushProvider,
	webhookProvider WebhookProvider,
	logger infrastructure.Logger,
) *ChannelService {
	return &ChannelService{
		channelRepo:     channelRepo,
		emailProvider:   emailProvider,
		smsProvider:     smsProvider,
		pushProvider:    pushProvider,
		webhookProvider: webhookProvider,
		logger:          logger,
	}
}

// CreateChannelConfig 创建渠道配置
func (s *ChannelService) CreateChannelConfig(ctx context.Context, cmd *CreateChannelConfigCommand) (*domain.ChannelConfig, error) {
	s.logger.Info("Creating channel config",
		zap.String("channel", string(cmd.Channel)),
		zap.String("name", cmd.Name),
		zap.String("owner_id", cmd.OwnerID))

	// 检查是否已存在
	existing, err := s.channelRepo.FindByChannelAndOwner(ctx, cmd.Channel, cmd.OwnerID)
	if err == nil && existing != nil {
		return nil, domain.NewDomainError("CHANNEL_CONFIG_EXISTS", "channel config already exists for this owner")
	}

	// 创建渠道配置
	config, err := domain.NewChannelConfig(cmd.Channel, cmd.Name, cmd.OwnerID)
	if err != nil {
		return nil, err
	}

	config.Description = cmd.Description
	config.UpdateConfig(cmd.Config)

	// 验证配置
	err = config.IsValidForSending()
	if err != nil {
		return nil, err
	}

	// 保存配置
	err = s.channelRepo.Save(ctx, config)
	if err != nil {
		s.logger.Error("Failed to save channel config", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Channel config created successfully", zap.String("id", config.ID))
	return config, nil
}

// UpdateChannelConfig 更新渠道配置
func (s *ChannelService) UpdateChannelConfig(ctx context.Context, cmd *UpdateChannelConfigCommand) (*domain.ChannelConfig, error) {
	config, err := s.channelRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, domain.ErrChannelNotFoundf(cmd.ID)
	}

	// 更新字段
	if cmd.Name != "" {
		config.Name = cmd.Name
	}
	if cmd.Description != "" {
		config.Description = cmd.Description
	}
	if cmd.Config != nil {
		config.UpdateConfig(cmd.Config)
	}
	if cmd.IsEnabled != nil {
		if *cmd.IsEnabled {
			config.Enable()
		} else {
			config.Disable()
		}
	}

	// 验证配置
	if config.IsEnabled {
		err = config.IsValidForSending()
		if err != nil {
			return nil, err
		}
	}

	// 保存更新
	err = s.channelRepo.Update(ctx, config)
	if err != nil {
		s.logger.Error("Failed to update channel config", zap.Error(err))
		return nil, err
	}

	return config, nil
}

// GetChannelConfig 获取渠道配置
func (s *ChannelService) GetChannelConfig(ctx context.Context, id string) (*domain.ChannelConfig, error) {
	return s.channelRepo.FindByID(ctx, id)
}

// ListChannelConfigs 列出渠道配置
func (s *ChannelService) ListChannelConfigs(ctx context.Context, cmd *ListChannelConfigsCommand) ([]*domain.ChannelConfig, int64, error) {
	if cmd.OwnerID != "" {
		return s.channelRepo.FindByOwnerWithPagination(ctx, cmd.OwnerID, cmd.Offset, cmd.Limit)
	}
	
	return s.channelRepo.FindWithPagination(ctx, cmd.Offset, cmd.Limit)
}

// TestChannel 测试渠道
func (s *ChannelService) TestChannel(ctx context.Context, cmd *TestChannelCommand) error {
	config, err := s.channelRepo.FindByID(ctx, cmd.ChannelID)
	if err != nil {
		return err
	}
	if config == nil {
		return domain.ErrChannelNotFoundf(cmd.ChannelID)
	}

	// 验证配置
	err = config.IsValidForSending()
	if err != nil {
		return err
	}

	// 创建测试通知
	testNotification, err := domain.NewNotification(
		"Test Notification",
		"This is a test notification to verify channel configuration.",
		domain.NotificationTypeSystem,
		config.Channel,
		"system",
	)
	if err != nil {
		return err
	}

	// 创建测试接收者
	testRecipient, err := domain.NewRecipient(
		testNotification.ID,
		domain.RecipientTypeEmail, // 默认类型
		"test@example.com",
		config.Channel,
	)
	if err != nil {
		return err
	}

	// 如果提供了测试数据，使用测试数据
	if cmd.TestData != nil {
		if email, exists := cmd.TestData["email"]; exists {
			testRecipient.Type = domain.RecipientTypeEmail
			testRecipient.Address = email
		}
		if phone, exists := cmd.TestData["phone"]; exists {
			testRecipient.Type = domain.RecipientTypePhone
			testRecipient.Address = phone
		}
	}

	// 发送测试通知
	return s.SendToRecipient(ctx, testNotification, testRecipient, config)
}

// SendToRecipient 发送通知给接收者
func (s *ChannelService) SendToRecipient(ctx context.Context, notification *domain.Notification, recipient *domain.Recipient, config *domain.ChannelConfig) error {
	s.logger.Info("Sending notification to recipient",
		zap.String("notification_id", notification.ID),
		zap.String("recipient_id", recipient.ID),
		zap.String("channel", string(config.Channel)))

	switch config.Channel {
	case domain.ChannelEmail:
		return s.sendEmail(ctx, notification, recipient, config)
	case domain.ChannelSMS:
		return s.sendSMS(ctx, notification, recipient, config)
	case domain.ChannelPush:
		return s.sendPush(ctx, notification, recipient, config)
	case domain.ChannelWebhook:
		return s.sendWebhook(ctx, notification, recipient, config)
	case domain.ChannelBark:
		return s.sendBark(ctx, notification, recipient, config)
	case domain.ChannelServerChan:
		return s.sendServerChan(ctx, notification, recipient, config)
	default:
		return domain.NewDomainError("UNSUPPORTED_CHANNEL", "unsupported notification channel")
	}
}

// sendEmail 发送邮件
func (s *ChannelService) sendEmail(ctx context.Context, notification *domain.Notification, recipient *domain.Recipient, config *domain.ChannelConfig) error {
	if s.emailProvider == nil {
		return domain.NewDomainError("EMAIL_PROVIDER_NOT_CONFIGURED", "email provider is not configured")
	}

	// 构建邮件数据
	emailData := &EmailData{
		To:      []string{recipient.GetEffectiveAddress()},
		Subject: notification.Title,
		Content: notification.Content,
		From:    config.Config["smtp_username"],
	}

	if fromName, exists := config.GetConfig("from_name"); exists {
		emailData.FromName = fromName
	}

	// 发送邮件
	return s.emailProvider.SendEmail(ctx, emailData, config)
}

// sendSMS 发送短信
func (s *ChannelService) sendSMS(ctx context.Context, notification *domain.Notification, recipient *domain.Recipient, config *domain.ChannelConfig) error {
	if s.smsProvider == nil {
		return domain.NewDomainError("SMS_PROVIDER_NOT_CONFIGURED", "SMS provider is not configured")
	}

	// 构建短信数据
	smsData := &SMSData{
		Phone:   recipient.GetEffectiveAddress(),
		Content: notification.Content,
	}

	// 发送短信
	return s.smsProvider.SendSMS(ctx, smsData, config)
}

// sendPush 发送推送通知
func (s *ChannelService) sendPush(ctx context.Context, notification *domain.Notification, recipient *domain.Recipient, config *domain.ChannelConfig) error {
	if s.pushProvider == nil {
		return domain.NewDomainError("PUSH_PROVIDER_NOT_CONFIGURED", "push provider is not configured")
	}

	// 构建推送数据
	pushData := &PushData{
		DeviceToken: recipient.GetEffectiveAddress(),
		Title:       notification.Title,
		Content:     notification.Content,
		Data:        notification.Variables,
	}

	// 发送推送
	return s.pushProvider.SendPush(ctx, pushData, config)
}

// sendWebhook 发送Webhook
func (s *ChannelService) sendWebhook(ctx context.Context, notification *domain.Notification, recipient *domain.Recipient, config *domain.ChannelConfig) error {
	if s.webhookProvider == nil {
		return domain.NewDomainError("WEBHOOK_PROVIDER_NOT_CONFIGURED", "webhook provider is not configured")
	}

	// 构建Webhook数据
	webhookData := &WebhookData{
		URL:     config.Config["url"],
		Method:  config.Config["method"],
		Headers: map[string]string{},
		Data: map[string]interface{}{
			"notification_id": notification.ID,
			"title":           notification.Title,
			"content":         notification.Content,
			"type":            notification.Type,
			"priority":        notification.Priority,
			"recipient":       recipient,
			"metadata":        notification.Metadata,
		},
	}

	// 设置Content-Type
	if contentType, exists := config.GetConfig("content_type"); exists {
		webhookData.Headers["Content-Type"] = contentType
	} else {
		webhookData.Headers["Content-Type"] = "application/json"
	}

	// 设置认证
	if secret, exists := config.GetConfig("secret"); exists {
		webhookData.Headers["X-Webhook-Secret"] = secret
	}

	// 发送Webhook
	return s.webhookProvider.SendWebhook(ctx, webhookData, config)
}

// sendBark 发送Bark通知
func (s *ChannelService) sendBark(ctx context.Context, notification *domain.Notification, recipient *domain.Recipient, config *domain.ChannelConfig) error {
	if s.pushProvider == nil {
		return domain.NewDomainError("BARK_PROVIDER_NOT_CONFIGURED", "Bark provider is not configured")
	}

	// 构建Bark数据
	barkData := &PushData{
		DeviceToken: config.Config["device_key"],
		Title:       notification.Title,
		Content:     notification.Content,
		Data: map[string]string{
			"sound": config.Config["sound"],
			"group": config.Config["group"],
		},
	}

	// 使用Push provider发送Bark通知
	return s.pushProvider.SendPush(ctx, barkData, config)
}

// sendServerChan 发送Server酱通知
func (s *ChannelService) sendServerChan(ctx context.Context, notification *domain.Notification, recipient *domain.Recipient, config *domain.ChannelConfig) error {
	if s.webhookProvider == nil {
		return domain.NewDomainError("SERVERCHAN_PROVIDER_NOT_CONFIGURED", "Server酱 provider is not configured")
	}

	// 构建Server酱数据
	baseURL := config.Config["base_url"]
	if baseURL == "" {
		baseURL = "https://sctapi.ftqq.com"
	}

	url := fmt.Sprintf("%s/SCT%s.send", baseURL, config.Config["send_key"])

	webhookData := &WebhookData{
		URL:    url,
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Data: map[string]interface{}{
			"title": notification.Title,
			"desp":  notification.Content,
		},
	}

	// 发送Server酱通知
	return s.webhookProvider.SendWebhook(ctx, webhookData, config)
}
