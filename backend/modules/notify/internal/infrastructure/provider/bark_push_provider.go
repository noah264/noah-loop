package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/noah-loop/backend/modules/notify/internal/application/service"
	"github.com/noah-loop/backend/modules/notify/internal/domain"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"go.uber.org/zap"
)

// BarkPushProvider Bark推送提供商
type BarkPushProvider struct {
	logger infrastructure.Logger
	client *http.Client
}

// NewBarkPushProvider 创建Bark推送提供商
func NewBarkPushProvider(logger infrastructure.Logger) service.PushProvider {
	return &BarkPushProvider{
		logger: logger,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// SendPush 发送推送通知
func (p *BarkPushProvider) SendPush(ctx context.Context, data *service.PushData, config *domain.ChannelConfig) error {
	p.logger.Info("Sending push notification via Bark",
		zap.String("device_token", data.DeviceToken),
		zap.String("title", data.Title))

	// 获取配置
	deviceKey, _ := config.GetConfig("device_key")
	serverURL, _ := config.GetConfig("server_url")
	
	if serverURL == "" {
		serverURL = "https://api.day.app"
	}

	// 如果data中没有device_token，使用配置中的device_key
	if data.DeviceToken == "" {
		data.DeviceToken = deviceKey
	}

	// 构建Bark消息
	message := &BarkMessage{
		Title: data.Title,
		Body:  data.Content,
		Sound: data.Sound,
		Group: data.Group,
	}

	// 从data.Data中获取额外参数
	if data.Data != nil {
		if sound, exists := data.Data["sound"]; exists && message.Sound == "" {
			message.Sound = sound
		}
		if group, exists := data.Data["group"]; exists && message.Group == "" {
			message.Group = group
		}
		if badge, exists := data.Data["badge"]; exists {
			message.Badge = badge
		}
		if url, exists := data.Data["url"]; exists {
			message.URL = url
		}
		if copy, exists := data.Data["copy"]; exists {
			message.Copy = copy
		}
		if autoCopy, exists := data.Data["autocopy"]; exists {
			message.AutoCopy = autoCopy
		}
		if level, exists := data.Data["level"]; exists {
			message.Level = level
		}
	}

	// 设置默认值
	if message.Sound == "" {
		message.Sound = "default"
	}
	if message.Group == "" {
		message.Group = "Noah-Loop"
	}

	// 构建请求URL
	url := fmt.Sprintf("%s/%s", serverURL, data.DeviceToken)

	// 序列化消息
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := p.client.Do(req)
	if err != nil {
		p.logger.Error("Failed to send Bark push request", zap.Error(err))
		return fmt.Errorf("failed to send push request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Bark push failed with status %d", resp.StatusCode)
	}

	// 解析响应（可选）
	var response BarkResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
		if response.Code != 200 {
			return fmt.Errorf("Bark push failed: code=%d, message=%s", response.Code, response.Message)
		}
	}

	p.logger.Info("Push notification sent successfully via Bark",
		zap.String("device_token", data.DeviceToken),
		zap.String("title", data.Title))

	return nil
}

// ValidateConfig 验证配置
func (p *BarkPushProvider) ValidateConfig(config *domain.ChannelConfig) error {
	if _, exists := config.GetConfig("device_key"); !exists {
		return domain.NewDomainError("MISSING_CONFIG", "missing required Bark config: device_key")
	}
	
	return nil
}

// GetProviderName 获取提供商名称
func (p *BarkPushProvider) GetProviderName() string {
	return "bark"
}

// BarkMessage Bark消息结构
type BarkMessage struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	Sound    string `json:"sound,omitempty"`
	Group    string `json:"group,omitempty"`
	Badge    string `json:"badge,omitempty"`
	URL      string `json:"url,omitempty"`
	Copy     string `json:"copy,omitempty"`
	AutoCopy string `json:"autocopy,omitempty"`
	Level    string `json:"level,omitempty"`
}

// BarkResponse Bark响应结构
type BarkResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}
