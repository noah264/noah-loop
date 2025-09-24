package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/noah-loop/backend/shared/pkg/domain"
)

// TemplateType 模板类型
type TemplateType string

const (
	TemplateTypeText     TemplateType = "text"     // 纯文本
	TemplateTypeHTML     TemplateType = "html"     // HTML
	TemplateTypeMarkdown TemplateType = "markdown" // Markdown
	TemplateTypeJSON     TemplateType = "json"     // JSON格式
)

// TemplateStatus 模板状态
type TemplateStatus string

const (
	TemplateStatusDraft     TemplateStatus = "draft"     // 草稿
	TemplateStatusActive    TemplateStatus = "active"    // 活跃
	TemplateStatusInactive  TemplateStatus = "inactive"  // 非活跃
	TemplateStatusArchived  TemplateStatus = "archived"  // 已归档
)

// NotificationTemplate 通知模板聚合根
type NotificationTemplate struct {
	domain.Entity
	Name        string                         `gorm:"not null" json:"name"`
	Code        string                         `gorm:"not null;uniqueIndex:idx_template_code" json:"code"` // 模板代码
	Type        TemplateType                   `gorm:"not null" json:"type"`
	Status      TemplateStatus                 `gorm:"not null;default:'draft'" json:"status"`
	Category    string                         `json:"category"`    // 分类
	Description string                         `json:"description"` // 描述
	Variables   []TemplateVariable             `json:"variables"`   // 模板变量
	Versions    []TemplateVersion              `json:"versions"`    // 版本历史
	Channels    []TemplateChannel              `json:"channels"`    // 渠道配置
	Tags        []string                       `gorm:"serializer:json" json:"tags,omitempty"`
	CreatedBy   string                         `gorm:"not null;index" json:"created_by"`
	UpdatedBy   string                         `gorm:"index" json:"updated_by"`
	CreatedAt   time.Time                      `json:"created_at"`
	UpdatedAt   time.Time                      `json:"updated_at"`
}

// TemplateVariable 模板变量
type TemplateVariable struct {
	domain.Entity
	TemplateID   string `gorm:"not null;index" json:"template_id"`
	Name         string `gorm:"not null" json:"name"`         // 变量名
	DisplayName  string `json:"display_name"`                 // 显示名
	Type         string `json:"type"`                         // 数据类型
	DefaultValue string `json:"default_value"`                // 默认值
	Required     bool   `json:"required"`                     // 是否必需
	Description  string `json:"description"`                  // 描述
	Validation   string `json:"validation"`                   // 验证规则
}

// TemplateVersion 模板版本
type TemplateVersion struct {
	domain.Entity
	TemplateID string    `gorm:"not null;index" json:"template_id"`
	Version    string    `gorm:"not null" json:"version"`    // 版本号
	Subject    string    `json:"subject"`                    // 标题模板
	Content    string    `gorm:"type:text;not null" json:"content"` // 内容模板
	IsActive   bool      `json:"is_active"`                  // 是否活跃版本
	ChangLog   string    `json:"change_log"`                 // 变更日志
	CreatedBy  string    `gorm:"not null;index" json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
}

// TemplateChannel 模板渠道配置
type TemplateChannel struct {
	domain.Entity
	TemplateID string              `gorm:"not null;index" json:"template_id"`
	Channel    NotificationChannel `gorm:"not null" json:"channel"`
	Subject    string              `json:"subject"`    // 渠道特定标题模板
	Content    string              `gorm:"type:text" json:"content"` // 渠道特定内容模板
	Config     map[string]string   `gorm:"serializer:json" json:"config,omitempty"` // 渠道特定配置
	IsEnabled  bool                `gorm:"default:true" json:"is_enabled"`
}

// AddVariable 添加模板变量
func (t *NotificationTemplate) AddVariable(variable TemplateVariable) error {
	// 检查变量名是否已存在
	for _, v := range t.Variables {
		if v.Name == variable.Name {
			return NewDomainError("VARIABLE_EXISTS", "variable name already exists")
		}
	}
	
	variable.TemplateID = t.ID
	t.Variables = append(t.Variables, variable)
	t.UpdatedAt = time.Now()
	
	return nil
}

// RemoveVariable 移除模板变量
func (t *NotificationTemplate) RemoveVariable(variableName string) {
	for i, v := range t.Variables {
		if v.Name == variableName {
			t.Variables = append(t.Variables[:i], t.Variables[i+1:]...)
			t.UpdatedAt = time.Now()
			break
		}
	}
}

// AddVersion 添加新版本
func (t *NotificationTemplate) AddVersion(version TemplateVersion) error {
	// 检查版本号是否已存在
	for _, v := range t.Versions {
		if v.Version == version.Version {
			return NewDomainError("VERSION_EXISTS", "version already exists")
		}
	}
	
	// 如果是活跃版本，先将其他版本设为非活跃
	if version.IsActive {
		for i := range t.Versions {
			t.Versions[i].IsActive = false
		}
	}
	
	version.TemplateID = t.ID
	version.CreatedAt = time.Now()
	t.Versions = append(t.Versions, version)
	t.UpdatedAt = time.Now()
	
	return nil
}

// GetActiveVersion 获取活跃版本
func (t *NotificationTemplate) GetActiveVersion() *TemplateVersion {
	for i, version := range t.Versions {
		if version.IsActive {
			return &t.Versions[i]
		}
	}
	
	// 如果没有活跃版本，返回最新版本
	if len(t.Versions) > 0 {
		return &t.Versions[len(t.Versions)-1]
	}
	
	return nil
}

// SetChannelTemplate 设置渠道模板
func (t *NotificationTemplate) SetChannelTemplate(channel NotificationChannel, subject, content string, config map[string]string) {
	// 查找现有渠道配置
	for i, tc := range t.Channels {
		if tc.Channel == channel {
			t.Channels[i].Subject = subject
			t.Channels[i].Content = content
			t.Channels[i].Config = config
			t.UpdatedAt = time.Now()
			return
		}
	}
	
	// 添加新的渠道配置
	channelTemplate := TemplateChannel{
		Entity:     domain.NewEntity(),
		TemplateID: t.ID,
		Channel:    channel,
		Subject:    subject,
		Content:    content,
		Config:     config,
		IsEnabled:  true,
	}
	
	t.Channels = append(t.Channels, channelTemplate)
	t.UpdatedAt = time.Now()
}

// GetChannelTemplate 获取渠道模板
func (t *NotificationTemplate) GetChannelTemplate(channel NotificationChannel) *TemplateChannel {
	for i, tc := range t.Channels {
		if tc.Channel == channel && tc.IsEnabled {
			return &t.Channels[i]
		}
	}
	
	return nil
}

// RenderTemplate 渲染模板
func (t *NotificationTemplate) RenderTemplate(channel NotificationChannel, variables map[string]string) (string, string, error) {
	// 获取活跃版本
	version := t.GetActiveVersion()
	if version == nil {
		return "", "", NewDomainError("NO_ACTIVE_VERSION", "no active version found")
	}
	
	// 获取渠道模板，如果没有则使用默认模板
	channelTemplate := t.GetChannelTemplate(channel)
	
	var subject, content string
	
	if channelTemplate != nil {
		subject = channelTemplate.Subject
		content = channelTemplate.Content
		
		// 如果渠道模板为空，使用默认模板
		if subject == "" {
			subject = version.Subject
		}
		if content == "" {
			content = version.Content
		}
	} else {
		subject = version.Subject
		content = version.Content
	}
	
	// 合并变量（默认值 + 传入值）
	allVariables := make(map[string]string)
	
	// 先设置默认值
	for _, variable := range t.Variables {
		if variable.DefaultValue != "" {
			allVariables[variable.Name] = variable.DefaultValue
		}
	}
	
	// 再设置传入的值
	for key, value := range variables {
		allVariables[key] = value
	}
	
	// 验证必需变量
	for _, variable := range t.Variables {
		if variable.Required {
			if _, exists := allVariables[variable.Name]; !exists {
				return "", "", NewDomainError("MISSING_REQUIRED_VARIABLE", "missing required variable: "+variable.Name)
			}
		}
	}
	
	// 渲染模板
	renderedSubject, err := renderString(subject, allVariables)
	if err != nil {
		return "", "", fmt.Errorf("failed to render subject: %w", err)
	}
	
	renderedContent, err := renderString(content, allVariables)
	if err != nil {
		return "", "", fmt.Errorf("failed to render content: %w", err)
	}
	
	return renderedSubject, renderedContent, nil
}

// UpdateStatus 更新模板状态
func (t *NotificationTemplate) UpdateStatus(status TemplateStatus) {
	t.Status = status
	t.UpdatedAt = time.Now()
}

// Activate 激活模板
func (t *NotificationTemplate) Activate() {
	t.Status = TemplateStatusActive
	t.UpdatedAt = time.Now()
}

// Deactivate 停用模板
func (t *NotificationTemplate) Deactivate() {
	t.Status = TemplateStatusInactive
	t.UpdatedAt = time.Now()
}

// Archive 归档模板
func (t *NotificationTemplate) Archive() {
	t.Status = TemplateStatusArchived
	t.UpdatedAt = time.Now()
}

// IsUsable 检查模板是否可用
func (t *NotificationTemplate) IsUsable() bool {
	return t.Status == TemplateStatusActive && len(t.Versions) > 0
}

// NewNotificationTemplate 创建新的通知模板
func NewNotificationTemplate(name, code string, templateType TemplateType, createdBy string) (*NotificationTemplate, error) {
	if name == "" {
		return nil, NewDomainError("INVALID_NAME", "template name cannot be empty")
	}
	
	if code == "" {
		return nil, NewDomainError("INVALID_CODE", "template code cannot be empty")
	}
	
	if createdBy == "" {
		return nil, NewDomainError("INVALID_CREATOR", "creator cannot be empty")
	}
	
	template := &NotificationTemplate{
		Entity:      domain.NewEntity(),
		Name:        name,
		Code:        code,
		Type:        templateType,
		Status:      TemplateStatusDraft,
		Variables:   make([]TemplateVariable, 0),
		Versions:    make([]TemplateVersion, 0),
		Channels:    make([]TemplateChannel, 0),
		Tags:        make([]string, 0),
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	return template, nil
}

// renderString 渲染字符串模板
func renderString(template string, variables map[string]string) (string, error) {
	result := template
	
	// 简单的变量替换 {{variable_name}}
	re := regexp.MustCompile(`\{\{(\w+)\}\}`)
	
	result = re.ReplaceAllStringFunc(result, func(match string) string {
		// 提取变量名
		varName := strings.Trim(match, "{}")
		varName = strings.TrimSpace(varName)
		
		if value, exists := variables[varName]; exists {
			return value
		}
		
		// 如果变量不存在，保留原样
		return match
	})
	
	return result, nil
}

// ValidateTemplate 验证模板语法
func ValidateTemplate(template string, variables []TemplateVariable) error {
	re := regexp.MustCompile(`\{\{(\w+)\}\}`)
	matches := re.FindAllStringSubmatch(template, -1)
	
	// 收集模板中使用的变量
	usedVars := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			varName := strings.TrimSpace(match[1])
			usedVars[varName] = true
		}
	}
	
	// 检查必需变量是否都在模板中使用
	for _, variable := range variables {
		if variable.Required && !usedVars[variable.Name] {
			return NewDomainError("UNUSED_REQUIRED_VARIABLE", "required variable not used in template: "+variable.Name)
		}
	}
	
	return nil
}
