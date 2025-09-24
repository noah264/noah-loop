package provider

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"

	"github.com/noah-loop/backend/modules/notify/internal/application/service"
	"github.com/noah-loop/backend/modules/notify/internal/domain"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"go.uber.org/zap"
)

// SMTPEmailProvider SMTP邮件提供商
type SMTPEmailProvider struct {
	logger infrastructure.Logger
}

// NewSMTPEmailProvider 创建SMTP邮件提供商
func NewSMTPEmailProvider(logger infrastructure.Logger) service.EmailProvider {
	return &SMTPEmailProvider{
		logger: logger,
	}
}

// SendEmail 发送邮件
func (p *SMTPEmailProvider) SendEmail(ctx context.Context, data *service.EmailData, config *domain.ChannelConfig) error {
	p.logger.Info("Sending email via SMTP",
		zap.Strings("to", data.To),
		zap.String("subject", data.Subject))

	// 获取SMTP配置
	host, _ := config.GetConfig("smtp_host")
	portStr, _ := config.GetConfig("smtp_port")
	username, _ := config.GetConfig("smtp_username")
	password, _ := config.GetConfig("smtp_password")
	useTLS, _ := config.GetConfig("use_tls")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid SMTP port: %w", err)
	}

	// 建立连接
	addr := fmt.Sprintf("%s:%d", host, port)
	
	var c *smtp.Client
	if useTLS == "true" {
		// 使用TLS连接
		tlsConfig := &tls.Config{
			ServerName: host,
		}
		
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP server with TLS: %w", err)
		}
		
		c, err = smtp.NewClient(conn, host)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
	} else {
		// 使用普通连接
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP server: %w", err)
		}
		
		c, err = smtp.NewClient(conn, host)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
		
		// 尝试STARTTLS
		if ok, _ := c.Extension("STARTTLS"); ok {
			config := &tls.Config{ServerName: host}
			if err = c.StartTLS(config); err != nil {
				p.logger.Warn("STARTTLS failed, continuing without TLS", zap.Error(err))
			}
		}
	}
	
	defer c.Quit()

	// 认证
	if username != "" && password != "" {
		auth := smtp.PlainAuth("", username, password, host)
		if err = c.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	// 设置发送者
	from := data.From
	if from == "" {
		from = username
	}
	
	if err = c.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// 设置接收者
	allRecipients := append(data.To, data.CC...)
	allRecipients = append(allRecipients, data.BCC...)
	
	for _, recipient := range allRecipients {
		if err = c.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	// 发送邮件内容
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	// 构建邮件头
	message := p.buildEmailMessage(data, config)
	
	if _, err = w.Write([]byte(message)); err != nil {
		return fmt.Errorf("failed to write email content: %w", err)
	}

	if err = w.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	p.logger.Info("Email sent successfully via SMTP",
		zap.Strings("to", data.To),
		zap.String("subject", data.Subject))

	return nil
}

// buildEmailMessage 构建邮件消息
func (p *SMTPEmailProvider) buildEmailMessage(data *service.EmailData, config *domain.ChannelConfig) string {
	var message strings.Builder
	
	// From
	fromName := data.FromName
	if fromName == "" {
		if name, exists := config.GetConfig("from_name"); exists {
			fromName = name
		}
	}
	
	if fromName != "" {
		message.WriteString(fmt.Sprintf("From: %s <%s>\r\n", fromName, data.From))
	} else {
		message.WriteString(fmt.Sprintf("From: %s\r\n", data.From))
	}
	
	// To
	message.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(data.To, ", ")))
	
	// CC
	if len(data.CC) > 0 {
		message.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(data.CC, ", ")))
	}
	
	// Subject
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", data.Subject))
	
	// MIME version
	message.WriteString("MIME-Version: 1.0\r\n")
	
	// Content type
	if data.HTML {
		message.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	} else {
		message.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	}
	
	// Custom headers
	for key, value := range data.Headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	
	// Empty line before body
	message.WriteString("\r\n")
	
	// Body
	message.WriteString(data.Content)
	
	return message.String()
}

// ValidateConfig 验证配置
func (p *SMTPEmailProvider) ValidateConfig(config *domain.ChannelConfig) error {
	requiredFields := []string{"smtp_host", "smtp_port", "smtp_username", "smtp_password"}
	
	for _, field := range requiredFields {
		if _, exists := config.GetConfig(field); !exists {
			return domain.NewDomainError("MISSING_CONFIG", "missing required SMTP config: "+field)
		}
	}
	
	// 验证端口
	portStr, _ := config.GetConfig("smtp_port")
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 || port > 65535 {
		return domain.NewDomainError("INVALID_CONFIG", "invalid SMTP port")
	}
	
	return nil
}

// GetProviderName 获取提供商名称
func (p *SMTPEmailProvider) GetProviderName() string {
	return "smtp"
}
