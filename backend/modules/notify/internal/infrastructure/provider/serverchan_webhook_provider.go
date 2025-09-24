package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/noah-loop/backend/modules/notify/internal/application/service"
	"github.com/noah-loop/backend/modules/notify/internal/domain"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"go.uber.org/zap"
)

// ServerChanWebhookProvider Server酱Webhook提供商
type ServerChanWebhookProvider struct {
	logger infrastructure.Logger
	client *http.Client
}

// NewServerChanWebhookProvider 创建Server酱Webhook提供商
func NewServerChanWebhookProvider(logger infrastructure.Logger) service.WebhookProvider {
	return &ServerChanWebhookProvider{
		logger: logger,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// SendWebhook 发送Webhook
func (p *ServerChanWebhookProvider) SendWebhook(ctx context.Context, data *service.WebhookData, config *domain.ChannelConfig) error {
	p.logger.Info("Sending webhook via ServerChan",
		zap.String("url", data.URL))

	// 根据不同的webhook类型处理
	if p.isServerChanURL(data.URL) {
		return p.sendServerChanMessage(ctx, data, config)
	}

	// 通用Webhook处理
	return p.sendGenericWebhook(ctx, data, config)
}

// isServerChanURL 判断是否是Server酱URL
func (p *ServerChanWebhookProvider) isServerChanURL(url string) bool {
	return strings.Contains(url, "sctapi.ftqq.com") || strings.Contains(url, "sc.ftqq.com")
}

// sendServerChanMessage 发送Server酱消息
func (p *ServerChanWebhookProvider) sendServerChanMessage(ctx context.Context, data *service.WebhookData, config *domain.ChannelConfig) error {
	// Server酱消息格式
	message := &ServerChanMessage{}
	
	// 从data.Data中提取标题和内容
	if title, ok := data.Data["title"].(string); ok {
		message.Title = title
	} else if title, ok := data.Data["text"].(string); ok {
		message.Title = title
	}
	
	if desp, ok := data.Data["desp"].(string); ok {
		message.Desp = desp
	} else if content, ok := data.Data["content"].(string); ok {
		message.Desp = content
	}
	
	// 如果没有从data中提取到，尝试从通知数据中获取
	if message.Title == "" {
		if notification, ok := data.Data["notification"]; ok {
			if notifyMap, ok := notification.(map[string]interface{}); ok {
				if title, ok := notifyMap["title"].(string); ok {
					message.Title = title
				}
				if content, ok := notifyMap["content"].(string); ok {
					message.Desp = content
				}
			}
		}
	}
	
	// 可选参数
	if short, ok := data.Data["short"].(string); ok {
		message.Short = short
	}
	if noip, ok := data.Data["noip"].(string); ok {
		message.NoIP = noip
	}
	if channel, ok := data.Data["channel"].(string); ok {
		message.Channel = channel
	}
	if openid, ok := data.Data["openid"].(string); ok {
		message.OpenID = openid
	}

	// 序列化消息
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal ServerChan message: %w", err)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", data.URL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	for key, value := range data.Headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := p.client.Do(req)
	if err != nil {
		p.logger.Error("Failed to send ServerChan webhook", zap.Error(err))
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook failed with status %d", resp.StatusCode)
	}

	// 解析响应
	var response ServerChanResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
		if response.Code != 0 {
			return fmt.Errorf("ServerChan webhook failed: code=%d, message=%s", response.Code, response.Message)
		}
		
		p.logger.Info("ServerChan webhook sent successfully",
			zap.String("pushid", response.Data.PushID))
	}

	return nil
}

// sendGenericWebhook 发送通用Webhook
func (p *ServerChanWebhookProvider) sendGenericWebhook(ctx context.Context, data *service.WebhookData, config *domain.ChannelConfig) error {
	method := data.Method
	if method == "" {
		method = "POST"
	}

	// 序列化数据
	var payload []byte
	var err error
	
	if method == "GET" {
		// GET请求不需要body
	} else {
		payload, err = json.Marshal(data.Data)
		if err != nil {
			return fmt.Errorf("failed to marshal webhook data: %w", err)
		}
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, data.URL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	if method != "GET" {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range data.Headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := p.client.Do(req)
	if err != nil {
		p.logger.Error("Failed to send generic webhook", zap.Error(err))
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook failed with status %d", resp.StatusCode)
	}

	p.logger.Info("Generic webhook sent successfully", zap.String("url", data.URL))
	return nil
}

// ValidateConfig 验证配置
func (p *ServerChanWebhookProvider) ValidateConfig(config *domain.ChannelConfig) error {
	if _, exists := config.GetConfig("url"); !exists {
		return domain.NewDomainError("MISSING_CONFIG", "missing required webhook config: url")
	}
	
	return nil
}

// GetProviderName 获取提供商名称
func (p *ServerChanWebhookProvider) GetProviderName() string {
	return "webhook"
}

// ServerChanMessage Server酱消息结构
type ServerChanMessage struct {
	Title   string `json:"title"`
	Desp    string `json:"desp,omitempty"`
	Short   string `json:"short,omitempty"`
	NoIP    string `json:"noip,omitempty"`
	Channel string `json:"channel,omitempty"`
	OpenID  string `json:"openid,omitempty"`
}

// ServerChanResponse Server酱响应结构
type ServerChanResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		PushID string `json:"pushid"`
		ReadKey string `json:"readkey"`
		Error  string `json:"error"`
		Errno  int    `json:"errno"`
	} `json:"data"`
}
