package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/noah-loop/backend/modules/notify/internal/application/service"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"go.uber.org/zap"
)

// NotifyHandler 通知HTTP处理器
type NotifyHandler struct {
	notificationService *service.NotificationService
	templateService     *service.TemplateService
	channelService      *service.ChannelService
	logger             infrastructure.Logger
}

// NewNotifyHandler 创建通知处理器
func NewNotifyHandler(
	notificationService *service.NotificationService,
	templateService *service.TemplateService,
	channelService *service.ChannelService,
	logger infrastructure.Logger,
) *NotifyHandler {
	return &NotifyHandler{
		notificationService: notificationService,
		templateService:     templateService,
		channelService:      channelService,
		logger:             logger,
	}
}

// CreateNotification 创建通知
func (h *NotifyHandler) CreateNotification(c *gin.Context) {
	var cmd service.CreateNotificationCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notification, err := h.notificationService.CreateNotification(c.Request.Context(), &cmd)
	if err != nil {
		h.logger.Error("Failed to create notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"notification": notification,
		"message":      "Notification created successfully",
	})
}

// CreateNotificationFromTemplate 从模板创建通知
func (h *NotifyHandler) CreateNotificationFromTemplate(c *gin.Context) {
	var cmd service.CreateNotificationFromTemplateCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notification, err := h.notificationService.CreateNotificationFromTemplate(c.Request.Context(), &cmd)
	if err != nil {
		h.logger.Error("Failed to create notification from template", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"notification": notification,
		"message":      "Notification created from template successfully",
	})
}

// GetNotification 获取通知
func (h *NotifyHandler) GetNotification(c *gin.Context) {
	id := c.Param("id")
	notification, err := h.notificationService.GetNotification(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"notification": notification})
}

// ListNotifications 列出通知
func (h *NotifyHandler) ListNotifications(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	cmd := &service.ListNotificationsCommand{
		Status:    c.Query("status"),
		CreatedBy: c.Query("created_by"),
		Offset:    offset,
		Limit:     limit,
	}

	notifications, total, err := h.notificationService.ListNotifications(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"total":         total,
		"offset":        offset,
		"limit":         limit,
	})
}

// SendNotification 发送通知
func (h *NotifyHandler) SendNotification(c *gin.Context) {
	id := c.Param("id")
	err := h.notificationService.SendNotification(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification sent successfully"})
}

// CreateTemplate 创建模板
func (h *NotifyHandler) CreateTemplate(c *gin.Context) {
	var cmd service.CreateTemplateCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template, err := h.templateService.CreateTemplate(c.Request.Context(), &cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"template": template,
		"message":  "Template created successfully",
	})
}

// CreateChannelConfig 创建渠道配置
func (h *NotifyHandler) CreateChannelConfig(c *gin.Context) {
	var cmd service.CreateChannelConfigCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config, err := h.channelService.CreateChannelConfig(c.Request.Context(), &cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"config":  config,
		"message": "Channel config created successfully",
	})
}

// TestChannel 测试渠道
func (h *NotifyHandler) TestChannel(c *gin.Context) {
	var cmd service.TestChannelCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.channelService.TestChannel(c.Request.Context(), &cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Channel test successful"})
}

// Health 健康检查
func (h *NotifyHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "notify",
		"message": "Notify service is running normally",
	})
}
