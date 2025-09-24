package service

import (
	"context"
	"fmt"
	"time"

	"github.com/noah-loop/backend/modules/notify/internal/domain"
	"github.com/noah-loop/backend/modules/notify/internal/domain/repository"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"go.uber.org/zap"
)

// NotificationService 通知应用服务
type NotificationService struct {
	notificationRepo repository.NotificationRepository
	recipientRepo    repository.RecipientRepository
	templateRepo     repository.TemplateRepository
	channelRepo      repository.ChannelRepository
	channelService   *ChannelService
	templateService  *TemplateService
	logger           infrastructure.Logger
}

// NewNotificationService 创建通知服务
func NewNotificationService(
	notificationRepo repository.NotificationRepository,
	recipientRepo repository.RecipientRepository,
	templateRepo repository.TemplateRepository,
	channelRepo repository.ChannelRepository,
	channelService *ChannelService,
	templateService *TemplateService,
	logger infrastructure.Logger,
) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		recipientRepo:    recipientRepo,
		templateRepo:     templateRepo,
		channelRepo:      channelRepo,
		channelService:   channelService,
		templateService:  templateService,
		logger:          logger,
	}
}

// CreateNotification 创建通知
func (s *NotificationService) CreateNotification(ctx context.Context, cmd *CreateNotificationCommand) (*domain.Notification, error) {
	s.logger.Info("Creating notification",
		zap.String("title", cmd.Title),
		zap.String("channel", string(cmd.Channel)),
		zap.String("created_by", cmd.CreatedBy))

	// 创建通知
	notification, err := domain.NewNotification(
		cmd.Title,
		cmd.Content,
		cmd.Type,
		cmd.Channel,
		cmd.CreatedBy,
	)
	if err != nil {
		return nil, err
	}

	// 设置可选属性
	if cmd.Priority != "" {
		notification.Priority = cmd.Priority
	}
	if cmd.TemplateID != "" {
		notification.TemplateID = cmd.TemplateID
	}
	if cmd.Variables != nil {
		notification.Variables = cmd.Variables
	}
	if cmd.Metadata != nil {
		notification.Metadata = *cmd.Metadata
	}
	if cmd.ScheduledAt != nil {
		notification.ScheduledAt = cmd.ScheduledAt
	}
	if cmd.MaxRetries > 0 {
		notification.MaxRetries = cmd.MaxRetries
	}

	// 添加接收者
	for _, recipientCmd := range cmd.Recipients {
		recipient, err := domain.NewRecipient(
			notification.ID,
			recipientCmd.Type,
			recipientCmd.Identifier,
			cmd.Channel,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create recipient: %w", err)
		}
		
		recipient.Name = recipientCmd.Name
		recipient.Address = recipientCmd.Address
		if recipientCmd.Variables != nil {
			recipient.Variables = recipientCmd.Variables
		}
		
		notification.AddRecipient(*recipient)
	}

	// 保存通知
	err = s.notificationRepo.Save(ctx, notification)
	if err != nil {
		s.logger.Error("Failed to save notification", zap.Error(err))
		return nil, err
	}

	// 保存接收者
	err = s.recipientRepo.SaveBatch(ctx, convertRecipientsToPointers(notification.Recipients))
	if err != nil {
		s.logger.Error("Failed to save recipients", zap.Error(err))
		return nil, err
	}

	// 如果不是定时通知，立即发送
	if !notification.IsScheduled() {
		go s.processNotificationAsync(context.Background(), notification.ID)
	}

	s.logger.Info("Notification created successfully", zap.String("id", notification.ID))
	return notification, nil
}

// CreateNotificationFromTemplate 从模板创建通知
func (s *NotificationService) CreateNotificationFromTemplate(ctx context.Context, cmd *CreateNotificationFromTemplateCommand) (*domain.Notification, error) {
	s.logger.Info("Creating notification from template",
		zap.String("template_id", cmd.TemplateID),
		zap.String("channel", string(cmd.Channel)))

	// 获取模板
	template, err := s.templateRepo.FindByID(ctx, cmd.TemplateID)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, domain.ErrTemplateNotFoundf(cmd.TemplateID)
	}

	// 渲染模板
	subject, content, err := template.RenderTemplate(cmd.Channel, cmd.Variables)
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	// 创建通知命令
	createCmd := &CreateNotificationCommand{
		Title:       subject,
		Content:     content,
		Type:        cmd.Type,
		Channel:     cmd.Channel,
		Priority:    cmd.Priority,
		TemplateID:  cmd.TemplateID,
		Variables:   cmd.Variables,
		Recipients:  cmd.Recipients,
		Metadata:    cmd.Metadata,
		ScheduledAt: cmd.ScheduledAt,
		MaxRetries:  cmd.MaxRetries,
		CreatedBy:   cmd.CreatedBy,
	}

	return s.CreateNotification(ctx, createCmd)
}

// SendNotification 发送通知
func (s *NotificationService) SendNotification(ctx context.Context, notificationID string) error {
	s.logger.Info("Sending notification", zap.String("notification_id", notificationID))

	// 获取通知
	notification, err := s.notificationRepo.FindByID(ctx, notificationID)
	if err != nil {
		return err
	}
	if notification == nil {
		return domain.ErrNotificationNotFoundf(notificationID)
	}

	// 检查是否可以发送
	if !notification.ShouldSend() {
		return domain.NewDomainError("NOTIFICATION_NOT_READY", "notification is not ready to send")
	}

	// 更新状态为发送中
	err = notification.UpdateStatus(domain.NotificationStatusSending)
	if err != nil {
		return err
	}
	err = s.notificationRepo.Update(ctx, notification)
	if err != nil {
		return err
	}

	// 获取接收者
	recipients, err := s.recipientRepo.FindByNotificationID(ctx, notificationID)
	if err != nil {
		return err
	}

	// 获取渠道配置
	channelConfig, err := s.channelRepo.FindByChannelAndOwner(ctx, notification.Channel, notification.CreatedBy)
	if err != nil {
		return err
	}
	if channelConfig == nil {
		return domain.ErrChannelNotFoundf(string(notification.Channel))
	}

	// 验证渠道配置
	err = channelConfig.IsValidForSending()
	if err != nil {
		return err
	}

	// 发送给每个接收者
	var sendErrors []string
	successCount := 0

	for _, recipient := range recipients {
		if recipient.Status != domain.RecipientStatusPending {
			continue
		}

		// 更新接收者状态为发送中
		recipient.UpdateStatus(domain.RecipientStatusSending)
		s.recipientRepo.Update(ctx, recipient)

		// 发送通知
		err = s.channelService.SendToRecipient(ctx, notification, recipient, channelConfig)
		if err != nil {
			recipient.SetError(err)
			sendErrors = append(sendErrors, err.Error())
			s.logger.Error("Failed to send to recipient",
				zap.String("recipient_id", recipient.ID),
				zap.Error(err))
		} else {
			recipient.UpdateStatus(domain.RecipientStatusSent)
			successCount++
		}

		// 更新接收者状态
		s.recipientRepo.Update(ctx, recipient)
	}

	// 更新通知状态
	if successCount == 0 {
		notification.SetError(fmt.Errorf("failed to send to all recipients: %v", sendErrors))
	} else if successCount == len(recipients) {
		notification.UpdateStatus(domain.NotificationStatusSent)
	} else {
		// 部分成功，状态保持为已发送但记录错误
		notification.UpdateStatus(domain.NotificationStatusSent)
		notification.ErrorMessage = fmt.Sprintf("partial success: %d/%d sent", successCount, len(recipients))
	}

	err = s.notificationRepo.Update(ctx, notification)
	if err != nil {
		return err
	}

	s.logger.Info("Notification sending completed",
		zap.String("notification_id", notificationID),
		zap.Int("success_count", successCount),
		zap.Int("total_count", len(recipients)))

	return nil
}

// GetNotification 获取通知
func (s *NotificationService) GetNotification(ctx context.Context, notificationID string) (*domain.Notification, error) {
	notification, err := s.notificationRepo.FindByID(ctx, notificationID)
	if err != nil {
		return nil, err
	}
	if notification == nil {
		return nil, domain.ErrNotificationNotFoundf(notificationID)
	}

	// 加载接收者
	recipients, err := s.recipientRepo.FindByNotificationID(ctx, notificationID)
	if err != nil {
		return nil, err
	}

	notification.Recipients = convertPointersToRecipients(recipients)
	return notification, nil
}

// ListNotifications 列出通知
func (s *NotificationService) ListNotifications(ctx context.Context, cmd *ListNotificationsCommand) ([]*domain.Notification, int64, error) {
	var notifications []*domain.Notification
	var total int64
	var err error

	if cmd.Status != "" {
		status := domain.NotificationStatus(cmd.Status)
		notifications, total, err = s.notificationRepo.FindByStatusWithPagination(ctx, status, cmd.Offset, cmd.Limit)
	} else if cmd.CreatedBy != "" {
		notifications, total, err = s.notificationRepo.FindByCreatedByWithPagination(ctx, cmd.CreatedBy, cmd.Offset, cmd.Limit)
	} else {
		notifications, total, err = s.notificationRepo.FindWithPagination(ctx, cmd.Offset, cmd.Limit)
	}

	return notifications, total, err
}

// CancelNotification 取消通知
func (s *NotificationService) CancelNotification(ctx context.Context, notificationID string) error {
	notification, err := s.notificationRepo.FindByID(ctx, notificationID)
	if err != nil {
		return err
	}
	if notification == nil {
		return domain.ErrNotificationNotFoundf(notificationID)
	}

	// 只有待发送或发送中的通知可以取消
	if notification.Status != domain.NotificationStatusPending && notification.Status != domain.NotificationStatusSending {
		return domain.NewDomainError("CANNOT_CANCEL", "notification cannot be cancelled")
	}

	// 更新状态
	err = notification.UpdateStatus(domain.NotificationStatusCancelled)
	if err != nil {
		return err
	}

	// 取消所有待发送的接收者
	recipients, err := s.recipientRepo.FindByNotificationID(ctx, notificationID)
	if err == nil {
		for _, recipient := range recipients {
			if recipient.Status == domain.RecipientStatusPending || recipient.Status == domain.RecipientStatusSending {
				recipient.UpdateStatus(domain.RecipientStatusSkipped)
				s.recipientRepo.Update(ctx, recipient)
			}
		}
	}

	return s.notificationRepo.Update(ctx, notification)
}

// RetryNotification 重试通知
func (s *NotificationService) RetryNotification(ctx context.Context, notificationID string) error {
	notification, err := s.notificationRepo.FindByID(ctx, notificationID)
	if err != nil {
		return err
	}
	if notification == nil {
		return domain.ErrNotificationNotFoundf(notificationID)
	}

	// 检查是否可以重试
	if !notification.CanRetry() {
		return domain.NewDomainError("CANNOT_RETRY", "notification cannot be retried")
	}

	// 重置状态为待发送
	err = notification.UpdateStatus(domain.NotificationStatusPending)
	if err != nil {
		return err
	}

	err = s.notificationRepo.Update(ctx, notification)
	if err != nil {
		return err
	}

	// 异步发送
	go s.processNotificationAsync(context.Background(), notificationID)

	return nil
}

// ProcessScheduledNotifications 处理定时通知
func (s *NotificationService) ProcessScheduledNotifications(ctx context.Context) error {
	// 获取应该发送的定时通知
	notifications, err := s.notificationRepo.FindScheduledNotifications(ctx, time.Now().Unix())
	if err != nil {
		return err
	}

	for _, notification := range notifications {
		// 异步处理每个通知
		go s.processNotificationAsync(context.Background(), notification.ID)
	}

	return nil
}

// ProcessRetryNotifications 处理重试通知
func (s *NotificationService) ProcessRetryNotifications(ctx context.Context) error {
	notifications, err := s.notificationRepo.FindRetryableNotifications(ctx, 100)
	if err != nil {
		return err
	}

	for _, notification := range notifications {
		go s.processNotificationAsync(context.Background(), notification.ID)
	}

	return nil
}

// GetNotificationStats 获取通知统计
func (s *NotificationService) GetNotificationStats(ctx context.Context, cmd *GetNotificationStatsCommand) (*repository.NotificationStats, error) {
	return s.notificationRepo.GetStatsByDateRange(ctx, cmd.StartDate, cmd.EndDate)
}

// processNotificationAsync 异步处理通知
func (s *NotificationService) processNotificationAsync(ctx context.Context, notificationID string) {
	err := s.SendNotification(ctx, notificationID)
	if err != nil {
		s.logger.Error("Failed to process notification asynchronously",
			zap.String("notification_id", notificationID),
			zap.Error(err))
	}
}

// 辅助函数
func convertRecipientsToPointers(recipients []domain.Recipient) []*domain.Recipient {
	result := make([]*domain.Recipient, len(recipients))
	for i := range recipients {
		result[i] = &recipients[i]
	}
	return result
}

func convertPointersToRecipients(recipients []*domain.Recipient) []domain.Recipient {
	result := make([]domain.Recipient, len(recipients))
	for i, r := range recipients {
		result[i] = *r
	}
	return result
}
