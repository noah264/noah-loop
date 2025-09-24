package service

import (
	"context"

	"github.com/noah-loop/backend/modules/notify/internal/domain"
)

// EmailProvider 邮件提供商接口
type EmailProvider interface {
	SendEmail(ctx context.Context, data *EmailData, config *domain.ChannelConfig) error
	ValidateConfig(config *domain.ChannelConfig) error
	GetProviderName() string
}

// EmailData 邮件数据
type EmailData struct {
	To       []string          `json:"to"`
	CC       []string          `json:"cc,omitempty"`
	BCC      []string          `json:"bcc,omitempty"`
	From     string            `json:"from"`
	FromName string            `json:"from_name,omitempty"`
	Subject  string            `json:"subject"`
	Content  string            `json:"content"`
	HTML     bool              `json:"html"`
	Attachments []EmailAttachment `json:"attachments,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
}

// EmailAttachment 邮件附件
type EmailAttachment struct {
	Filename    string `json:"filename"`
	Content     []byte `json:"content"`
	ContentType string `json:"content_type"`
}

// SMSProvider 短信提供商接口
type SMSProvider interface {
	SendSMS(ctx context.Context, data *SMSData, config *domain.ChannelConfig) error
	ValidateConfig(config *domain.ChannelConfig) error
	GetProviderName() string
}

// SMSData 短信数据
type SMSData struct {
	Phone       string            `json:"phone"`
	Content     string            `json:"content"`
	TemplateID  string            `json:"template_id,omitempty"`
	Variables   map[string]string `json:"variables,omitempty"`
	SignName    string            `json:"sign_name,omitempty"`
}

// PushProvider 推送提供商接口（包括Bark等）
type PushProvider interface {
	SendPush(ctx context.Context, data *PushData, config *domain.ChannelConfig) error
	ValidateConfig(config *domain.ChannelConfig) error
	GetProviderName() string
}

// PushData 推送数据
type PushData struct {
	DeviceToken string            `json:"device_token"`
	Title       string            `json:"title"`
	Content     string            `json:"content"`
	Sound       string            `json:"sound,omitempty"`
	Badge       int               `json:"badge,omitempty"`
	Data        map[string]string `json:"data,omitempty"`
	URL         string            `json:"url,omitempty"`
	Group       string            `json:"group,omitempty"`
}

// WebhookProvider Webhook提供商接口（包括Server酱等）
type WebhookProvider interface {
	SendWebhook(ctx context.Context, data *WebhookData, config *domain.ChannelConfig) error
	ValidateConfig(config *domain.ChannelConfig) error
	GetProviderName() string
}

// WebhookData Webhook数据
type WebhookData struct {
	URL     string                 `json:"url"`
	Method  string                 `json:"method"`
	Headers map[string]string      `json:"headers"`
	Data    map[string]interface{} `json:"data"`
	Timeout int                    `json:"timeout"` // 秒
}

// DingTalkProvider 钉钉提供商接口
type DingTalkProvider interface {
	SendDingTalk(ctx context.Context, data *DingTalkData, config *domain.ChannelConfig) error
	ValidateConfig(config *domain.ChannelConfig) error
	GetProviderName() string
}

// DingTalkData 钉钉数据
type DingTalkData struct {
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	AtMobiles  []string `json:"at_mobiles,omitempty"`
	AtAll      bool     `json:"at_all,omitempty"`
	MessageType string  `json:"message_type"` // text, markdown, link等
}

// WeChatProvider 微信提供商接口
type WeChatProvider interface {
	SendWeChat(ctx context.Context, data *WeChatData, config *domain.ChannelConfig) error
	ValidateConfig(config *domain.ChannelConfig) error
	GetProviderName() string
}

// WeChatData 微信数据
type WeChatData struct {
	ToUser   string            `json:"to_user"`
	Template string            `json:"template"`
	Data     map[string]string `json:"data"`
	URL      string            `json:"url,omitempty"`
	Color    string            `json:"color,omitempty"`
}

// SlackProvider Slack提供商接口
type SlackProvider interface {
	SendSlack(ctx context.Context, data *SlackData, config *domain.ChannelConfig) error
	ValidateConfig(config *domain.ChannelConfig) error
	GetProviderName() string
}

// SlackData Slack数据
type SlackData struct {
	Channel     string                   `json:"channel"`
	Text        string                   `json:"text"`
	Username    string                   `json:"username,omitempty"`
	IconEmoji   string                   `json:"icon_emoji,omitempty"`
	Attachments []SlackAttachment        `json:"attachments,omitempty"`
	Blocks      []map[string]interface{} `json:"blocks,omitempty"`
}

// SlackAttachment Slack附件
type SlackAttachment struct {
	Color      string                     `json:"color,omitempty"`
	Title      string                     `json:"title,omitempty"`
	Text       string                     `json:"text,omitempty"`
	Fields     []SlackField               `json:"fields,omitempty"`
	Actions    []map[string]interface{}   `json:"actions,omitempty"`
}

// SlackField Slack字段
type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// TelegramProvider Telegram提供商接口
type TelegramProvider interface {
	SendTelegram(ctx context.Context, data *TelegramData, config *domain.ChannelConfig) error
	ValidateConfig(config *domain.ChannelConfig) error
	GetProviderName() string
}

// TelegramData Telegram数据
type TelegramData struct {
	ChatID      string `json:"chat_id"`
	Text        string `json:"text"`
	ParseMode   string `json:"parse_mode,omitempty"` // Markdown, HTML
	ReplyMarkup string `json:"reply_markup,omitempty"`
}

// DiscordProvider Discord提供商接口
type DiscordProvider interface {
	SendDiscord(ctx context.Context, data *DiscordData, config *domain.ChannelConfig) error
	ValidateConfig(config *domain.ChannelConfig) error
	GetProviderName() string
}

// DiscordData Discord数据
type DiscordData struct {
	Content   string                   `json:"content"`
	Username  string                   `json:"username,omitempty"`
	AvatarURL string                   `json:"avatar_url,omitempty"`
	Embeds    []map[string]interface{} `json:"embeds,omitempty"`
}

// ProviderRegistry 提供商注册表
type ProviderRegistry struct {
	emailProviders    map[string]EmailProvider
	smsProviders      map[string]SMSProvider
	pushProviders     map[string]PushProvider
	webhookProviders  map[string]WebhookProvider
}

// NewProviderRegistry 创建提供商注册表
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		emailProviders:   make(map[string]EmailProvider),
		smsProviders:     make(map[string]SMSProvider),
		pushProviders:    make(map[string]PushProvider),
		webhookProviders: make(map[string]WebhookProvider),
	}
}

// RegisterEmailProvider 注册邮件提供商
func (r *ProviderRegistry) RegisterEmailProvider(name string, provider EmailProvider) {
	r.emailProviders[name] = provider
}

// RegisterSMSProvider 注册短信提供商
func (r *ProviderRegistry) RegisterSMSProvider(name string, provider SMSProvider) {
	r.smsProviders[name] = provider
}

// RegisterPushProvider 注册推送提供商
func (r *ProviderRegistry) RegisterPushProvider(name string, provider PushProvider) {
	r.pushProviders[name] = provider
}

// RegisterWebhookProvider 注册Webhook提供商
func (r *ProviderRegistry) RegisterWebhookProvider(name string, provider WebhookProvider) {
	r.webhookProviders[name] = provider
}

// GetEmailProvider 获取邮件提供商
func (r *ProviderRegistry) GetEmailProvider(name string) EmailProvider {
	return r.emailProviders[name]
}

// GetSMSProvider 获取短信提供商
func (r *ProviderRegistry) GetSMSProvider(name string) SMSProvider {
	return r.smsProviders[name]
}

// GetPushProvider 获取推送提供商
func (r *ProviderRegistry) GetPushProvider(name string) PushProvider {
	return r.pushProviders[name]
}

// GetWebhookProvider 获取Webhook提供商
func (r *ProviderRegistry) GetWebhookProvider(name string) WebhookProvider {
	return r.webhookProviders[name]
}
