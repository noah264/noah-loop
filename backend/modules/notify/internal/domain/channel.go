package domain

import (
	"time"

	"github.com/noah-loop/backend/shared/pkg/domain"
)

// NotificationChannel 通知渠道
type NotificationChannel string

const (
	ChannelEmail     NotificationChannel = "email"      // 邮件
	ChannelSMS       NotificationChannel = "sms"        // 短信
	ChannelPush      NotificationChannel = "push"       // 推送通知
	ChannelWebhook   NotificationChannel = "webhook"    // Webhook
	ChannelBark      NotificationChannel = "bark"       // Bark推送
	ChannelServerChan NotificationChannel = "serverchan" // Server酱
	ChannelDingTalk  NotificationChannel = "dingtalk"   // 钉钉
	ChannelWeChat    NotificationChannel = "wechat"     // 微信
	ChannelSlack     NotificationChannel = "slack"      // Slack
	ChannelTelegram  NotificationChannel = "telegram"   // Telegram
	ChannelDiscord   NotificationChannel = "discord"    // Discord
)

// ChannelConfig 渠道配置实体
type ChannelConfig struct {
	domain.Entity
	Channel     NotificationChannel `gorm:"not null;uniqueIndex:idx_channel_owner" json:"channel"`
	Name        string              `gorm:"not null" json:"name"`
	Description string              `json:"description"`
	OwnerID     string              `gorm:"not null;uniqueIndex:idx_channel_owner" json:"owner_id"`
	Config      map[string]string   `gorm:"serializer:json" json:"config"`
	IsEnabled   bool                `gorm:"default:true" json:"is_enabled"`
	RateLimit   ChannelRateLimit    `gorm:"embedded" json:"rate_limit"`
	RetryConfig ChannelRetryConfig  `gorm:"embedded" json:"retry_config"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// ChannelRateLimit 渠道限流配置
type ChannelRateLimit struct {
	MaxPerMinute int `json:"max_per_minute" gorm:"default:60"`  // 每分钟最大发送数
	MaxPerHour   int `json:"max_per_hour" gorm:"default:1000"`  // 每小时最大发送数
	MaxPerDay    int `json:"max_per_day" gorm:"default:10000"`  // 每天最大发送数
}

// ChannelRetryConfig 渠道重试配置
type ChannelRetryConfig struct {
	MaxRetries    int           `json:"max_retries" gorm:"default:3"`       // 最大重试次数
	RetryInterval time.Duration `json:"retry_interval" gorm:"default:300"`  // 重试间隔（秒）
	BackoffFactor float64       `json:"backoff_factor" gorm:"default:2.0"`  // 退避因子
}

// UpdateConfig 更新渠道配置
func (c *ChannelConfig) UpdateConfig(config map[string]string) {
	if c.Config == nil {
		c.Config = make(map[string]string)
	}
	
	for key, value := range config {
		c.Config[key] = value
	}
	
	c.UpdatedAt = time.Now()
}

// GetConfig 获取配置项
func (c *ChannelConfig) GetConfig(key string) (string, bool) {
	if c.Config == nil {
		return "", false
	}
	
	value, exists := c.Config[key]
	return value, exists
}

// SetConfig 设置配置项
func (c *ChannelConfig) SetConfig(key, value string) {
	if c.Config == nil {
		c.Config = make(map[string]string)
	}
	
	c.Config[key] = value
	c.UpdatedAt = time.Now()
}

// Enable 启用渠道
func (c *ChannelConfig) Enable() {
	c.IsEnabled = true
	c.UpdatedAt = time.Now()
}

// Disable 禁用渠道
func (c *ChannelConfig) Disable() {
	c.IsEnabled = false
	c.UpdatedAt = time.Now()
}

// IsValidForSending 检查是否可以发送
func (c *ChannelConfig) IsValidForSending() error {
	if !c.IsEnabled {
		return NewDomainError("CHANNEL_DISABLED", "notification channel is disabled")
	}
	
	// 根据不同渠道验证必要的配置
	switch c.Channel {
	case ChannelEmail:
		return c.validateEmailConfig()
	case ChannelSMS:
		return c.validateSMSConfig()
	case ChannelBark:
		return c.validateBarkConfig()
	case ChannelServerChan:
		return c.validateServerChanConfig()
	case ChannelWebhook:
		return c.validateWebhookConfig()
	}
	
	return nil
}

// validateEmailConfig 验证邮件配置
func (c *ChannelConfig) validateEmailConfig() error {
	requiredFields := []string{"smtp_host", "smtp_port", "smtp_username", "smtp_password"}
	
	for _, field := range requiredFields {
		if _, exists := c.GetConfig(field); !exists {
			return NewDomainError("MISSING_CONFIG", "missing required config: "+field)
		}
	}
	
	return nil
}

// validateSMSConfig 验证短信配置
func (c *ChannelConfig) validateSMSConfig() error {
	requiredFields := []string{"access_key", "secret_key", "sign_name"}
	
	for _, field := range requiredFields {
		if _, exists := c.GetConfig(field); !exists {
			return NewDomainError("MISSING_CONFIG", "missing required config: "+field)
		}
	}
	
	return nil
}

// validateBarkConfig 验证Bark配置
func (c *ChannelConfig) validateBarkConfig() error {
	if _, exists := c.GetConfig("device_key"); !exists {
		return NewDomainError("MISSING_CONFIG", "missing required config: device_key")
	}
	
	return nil
}

// validateServerChanConfig 验证Server酱配置
func (c *ChannelConfig) validateServerChanConfig() error {
	if _, exists := c.GetConfig("send_key"); !exists {
		return NewDomainError("MISSING_CONFIG", "missing required config: send_key")
	}
	
	return nil
}

// validateWebhookConfig 验证Webhook配置
func (c *ChannelConfig) validateWebhookConfig() error {
	if _, exists := c.GetConfig("url"); !exists {
		return NewDomainError("MISSING_CONFIG", "missing required config: url")
	}
	
	return nil
}

// NewChannelConfig 创建新的渠道配置
func NewChannelConfig(channel NotificationChannel, name, ownerID string) (*ChannelConfig, error) {
	if name == "" {
		return nil, NewDomainError("INVALID_NAME", "channel name cannot be empty")
	}
	
	if ownerID == "" {
		return nil, NewDomainError("INVALID_OWNER_ID", "owner ID cannot be empty")
	}
	
	config := &ChannelConfig{
		Entity:      domain.NewEntity(),
		Channel:     channel,
		Name:        name,
		OwnerID:     ownerID,
		Config:      make(map[string]string),
		IsEnabled:   true,
		RateLimit: ChannelRateLimit{
			MaxPerMinute: 60,
			MaxPerHour:   1000,
			MaxPerDay:    10000,
		},
		RetryConfig: ChannelRetryConfig{
			MaxRetries:    3,
			RetryInterval: 5 * time.Minute,
			BackoffFactor: 2.0,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	return config, nil
}

// GetDefaultChannelConfigs 获取默认渠道配置示例
func GetDefaultChannelConfigs() map[NotificationChannel]map[string]string {
	return map[NotificationChannel]map[string]string{
		ChannelEmail: {
			"smtp_host":     "smtp.example.com",
			"smtp_port":     "587",
			"smtp_username": "user@example.com",
			"smtp_password": "password",
			"from_name":     "Noah-Loop System",
			"use_tls":       "true",
		},
		ChannelSMS: {
			"provider":    "aliyun", // 阿里云短信
			"access_key":  "your_access_key",
			"secret_key":  "your_secret_key",
			"sign_name":   "Noah-Loop",
			"region":      "cn-hangzhou",
		},
		ChannelBark: {
			"device_key":  "your_bark_device_key",
			"server_url":  "https://api.day.app", // 可选，默认官方服务器
			"sound":       "default",
			"group":       "Noah-Loop",
		},
		ChannelServerChan: {
			"send_key": "your_server_chan_send_key",
			"base_url": "https://sctapi.ftqq.com", // Server酱Turbo版
		},
		ChannelWebhook: {
			"url":            "https://your-webhook-endpoint.com",
			"method":         "POST",
			"content_type":   "application/json",
			"timeout":        "30",
			"secret":         "optional_webhook_secret",
		},
		ChannelDingTalk: {
			"webhook_url": "https://oapi.dingtalk.com/robot/send?access_token=xxx",
			"secret":      "optional_secret_for_signature",
			"at_mobiles":  "", // @指定手机号，多个用逗号分隔
			"at_all":      "false",
		},
	}
}
