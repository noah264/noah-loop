package service

import (
	"context"

	"github.com/noah-loop/backend/modules/notify/internal/domain"
	"github.com/noah-loop/backend/modules/notify/internal/domain/repository"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"go.uber.org/zap"
)

// TemplateService 模板服务
type TemplateService struct {
	templateRepo repository.TemplateRepository
	logger       infrastructure.Logger
}

// NewTemplateService 创建模板服务
func NewTemplateService(
	templateRepo repository.TemplateRepository,
	logger infrastructure.Logger,
) *TemplateService {
	return &TemplateService{
		templateRepo: templateRepo,
		logger:       logger,
	}
}

// CreateTemplate 创建模板
func (s *TemplateService) CreateTemplate(ctx context.Context, cmd *CreateTemplateCommand) (*domain.NotificationTemplate, error) {
	s.logger.Info("Creating template",
		zap.String("name", cmd.Name),
		zap.String("code", cmd.Code),
		zap.String("created_by", cmd.CreatedBy))

	// 检查模板代码是否已存在
	existing, err := s.templateRepo.FindByCode(ctx, cmd.Code)
	if err == nil && existing != nil {
		return nil, domain.NewDomainError("TEMPLATE_CODE_EXISTS", "template code already exists")
	}

	// 创建模板
	template, err := domain.NewNotificationTemplate(cmd.Name, cmd.Code, cmd.Type, cmd.CreatedBy)
	if err != nil {
		return nil, err
	}

	template.Category = cmd.Category
	template.Description = cmd.Description
	template.Tags = cmd.Tags

	// 添加变量
	for _, varCmd := range cmd.Variables {
		variable := domain.TemplateVariable{
			Name:         varCmd.Name,
			DisplayName:  varCmd.DisplayName,
			Type:         varCmd.Type,
			DefaultValue: varCmd.DefaultValue,
			Required:     varCmd.Required,
			Description:  varCmd.Description,
			Validation:   varCmd.Validation,
		}
		
		err = template.AddVariable(variable)
		if err != nil {
			return nil, err
		}
	}

	// 添加默认版本
	version := domain.TemplateVersion{
		Version:   "1.0.0",
		Subject:   cmd.Subject,
		Content:   cmd.Content,
		IsActive:  true,
		CreatedBy: cmd.CreatedBy,
	}

	err = template.AddVersion(version)
	if err != nil {
		return nil, err
	}

	// 验证模板语法
	err = domain.ValidateTemplate(cmd.Content, template.Variables)
	if err != nil {
		return nil, err
	}

	// 保存模板
	err = s.templateRepo.Save(ctx, template)
	if err != nil {
		s.logger.Error("Failed to save template", zap.Error(err))
		return nil, err
	}

	// 保存变量
	if len(template.Variables) > 0 {
		variables := make([]*domain.TemplateVariable, len(template.Variables))
		for i := range template.Variables {
			variables[i] = &template.Variables[i]
		}
		err = s.templateRepo.SaveVariables(ctx, variables)
		if err != nil {
			s.logger.Error("Failed to save template variables", zap.Error(err))
			return nil, err
		}
	}

	s.logger.Info("Template created successfully", zap.String("id", template.ID))
	return template, nil
}

// GetTemplate 获取模板
func (s *TemplateService) GetTemplate(ctx context.Context, templateID string) (*domain.NotificationTemplate, error) {
	template, err := s.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, domain.ErrTemplateNotFoundf(templateID)
	}

	// 加载变量
	variables, err := s.templateRepo.FindVariablesByTemplateID(ctx, templateID)
	if err == nil {
		template.Variables = convertPointersToVariables(variables)
	}

	// 加载版本
	versions, err := s.templateRepo.FindVersionsByTemplateID(ctx, templateID)
	if err == nil {
		template.Versions = convertPointersToVersions(versions)
	}

	// 加载渠道模板
	channels, err := s.templateRepo.FindChannelTemplates(ctx, templateID)
	if err == nil {
		template.Channels = convertPointersToChannels(channels)
	}

	return template, nil
}

// GetTemplateByCode 根据代码获取模板
func (s *TemplateService) GetTemplateByCode(ctx context.Context, code string) (*domain.NotificationTemplate, error) {
	template, err := s.templateRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, domain.NewDomainError("TEMPLATE_NOT_FOUND", "template not found")
	}

	return s.GetTemplate(ctx, template.ID)
}

// UpdateTemplate 更新模板
func (s *TemplateService) UpdateTemplate(ctx context.Context, cmd *UpdateTemplateCommand) (*domain.NotificationTemplate, error) {
	template, err := s.templateRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, domain.ErrTemplateNotFoundf(cmd.ID)
	}

	// 更新字段
	if cmd.Name != "" {
		template.Name = cmd.Name
	}
	if cmd.Category != "" {
		template.Category = cmd.Category
	}
	if cmd.Description != "" {
		template.Description = cmd.Description
	}
	if cmd.Status != "" {
		template.UpdateStatus(cmd.Status)
	}
	if cmd.Tags != nil {
		template.Tags = cmd.Tags
	}

	// 保存更新
	err = s.templateRepo.Update(ctx, template)
	if err != nil {
		s.logger.Error("Failed to update template", zap.Error(err))
		return nil, err
	}

	return template, nil
}

// CreateTemplateVersion 创建模板版本
func (s *TemplateService) CreateTemplateVersion(ctx context.Context, cmd *CreateTemplateVersionCommand) (*domain.TemplateVersion, error) {
	template, err := s.templateRepo.FindByID(ctx, cmd.TemplateID)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, domain.ErrTemplateNotFoundf(cmd.TemplateID)
	}

	// 加载变量用于验证
	variables, err := s.templateRepo.FindVariablesByTemplateID(ctx, cmd.TemplateID)
	if err != nil {
		return nil, err
	}

	templateVariables := convertPointersToVariables(variables)

	// 验证模板语法
	err = domain.ValidateTemplate(cmd.Content, templateVariables)
	if err != nil {
		return nil, err
	}

	// 创建版本
	version := domain.TemplateVersion{
		TemplateID: cmd.TemplateID,
		Version:    cmd.Version,
		Subject:    cmd.Subject,
		Content:    cmd.Content,
		IsActive:   cmd.IsActive,
		ChangeLog:  cmd.ChangeLog,
		CreatedBy:  cmd.CreatedBy,
	}

	// 添加版本到模板
	err = template.AddVersion(version)
	if err != nil {
		return nil, err
	}

	// 保存版本
	err = s.templateRepo.SaveVersion(ctx, &version)
	if err != nil {
		s.logger.Error("Failed to save template version", zap.Error(err))
		return nil, err
	}

	// 如果是活跃版本，更新其他版本状态
	if cmd.IsActive {
		err = s.templateRepo.UpdateVersionStatus(ctx, cmd.TemplateID, cmd.Version, true)
		if err != nil {
			s.logger.Error("Failed to update version status", zap.Error(err))
			return nil, err
		}
	}

	return &version, nil
}

// RenderTemplate 渲染模板
func (s *TemplateService) RenderTemplate(ctx context.Context, cmd *RenderTemplateCommand) (string, string, error) {
	template, err := s.GetTemplate(ctx, cmd.TemplateID)
	if err != nil {
		return "", "", err
	}

	if !template.IsUsable() {
		return "", "", domain.NewDomainError("TEMPLATE_NOT_USABLE", "template is not usable")
	}

	return template.RenderTemplate(cmd.Channel, cmd.Variables)
}

// ListTemplates 列出模板
func (s *TemplateService) ListTemplates(ctx context.Context, cmd *ListTemplatesCommand) ([]*domain.NotificationTemplate, int64, error) {
	var templates []*domain.NotificationTemplate
	var total int64
	var err error

	if cmd.Status != "" {
		status := domain.TemplateStatus(cmd.Status)
		templates, total, err = s.templateRepo.FindByStatusWithPagination(ctx, status, cmd.Offset, cmd.Limit)
	} else if cmd.CreatedBy != "" {
		templates, total, err = s.templateRepo.FindByCreatedByWithPagination(ctx, cmd.CreatedBy, cmd.Offset, cmd.Limit)
	} else {
		templates, total, err = s.templateRepo.FindWithPagination(ctx, cmd.Offset, cmd.Limit)
	}

	return templates, total, err
}

// SearchTemplates 搜索模板
func (s *TemplateService) SearchTemplates(ctx context.Context, cmd *SearchTemplatesCommand) ([]*domain.NotificationTemplate, error) {
	return s.templateRepo.SearchByName(ctx, cmd.Query, cmd.Limit)
}

// ActivateTemplate 激活模板
func (s *TemplateService) ActivateTemplate(ctx context.Context, templateID string) error {
	template, err := s.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return err
	}
	if template == nil {
		return domain.ErrTemplateNotFoundf(templateID)
	}

	template.Activate()
	return s.templateRepo.Update(ctx, template)
}

// DeactivateTemplate 停用模板
func (s *TemplateService) DeactivateTemplate(ctx context.Context, templateID string) error {
	template, err := s.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return err
	}
	if template == nil {
		return domain.ErrTemplateNotFoundf(templateID)
	}

	template.Deactivate()
	return s.templateRepo.Update(ctx, template)
}

// ArchiveTemplate 归档模板
func (s *TemplateService) ArchiveTemplate(ctx context.Context, templateID string) error {
	template, err := s.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return err
	}
	if template == nil {
		return domain.ErrTemplateNotFoundf(templateID)
	}

	template.Archive()
	return s.templateRepo.Update(ctx, template)
}

// DeleteTemplate 删除模板
func (s *TemplateService) DeleteTemplate(ctx context.Context, templateID string) error {
	// 先删除相关数据
	s.templateRepo.DeleteVariablesByTemplateID(ctx, templateID)
	
	// 删除模板
	return s.templateRepo.Delete(ctx, templateID)
}

// SetChannelTemplate 设置渠道模板
func (s *TemplateService) SetChannelTemplate(ctx context.Context, templateID string, channel domain.NotificationChannel, subject, content string, config map[string]string) error {
	template, err := s.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return err
	}
	if template == nil {
		return domain.ErrTemplateNotFoundf(templateID)
	}

	// 设置渠道模板
	template.SetChannelTemplate(channel, subject, content, config)

	// 保存渠道模板
	channelTemplate := template.GetChannelTemplate(channel)
	if channelTemplate != nil {
		return s.templateRepo.SaveChannelTemplate(ctx, channelTemplate)
	}

	return nil
}

// GetTemplateUsageStats 获取模板使用统计
func (s *TemplateService) GetTemplateUsageStats(ctx context.Context, templateID string) (*repository.TemplateUsageStats, error) {
	return s.templateRepo.GetUsageStats(ctx, templateID)
}

// 辅助函数
func convertPointersToVariables(variables []*domain.TemplateVariable) []domain.TemplateVariable {
	result := make([]domain.TemplateVariable, len(variables))
	for i, v := range variables {
		result[i] = *v
	}
	return result
}

func convertPointersToVersions(versions []*domain.TemplateVersion) []domain.TemplateVersion {
	result := make([]domain.TemplateVersion, len(versions))
	for i, v := range versions {
		result[i] = *v
	}
	return result
}

func convertPointersToChannels(channels []*domain.TemplateChannel) []domain.TemplateChannel {
	result := make([]domain.TemplateChannel, len(channels))
	for i, c := range channels {
		result[i] = *c
	}
	return result
}
