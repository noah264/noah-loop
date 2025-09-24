package domain

import (
	"time"

	"github.com/noah-loop/backend/shared/pkg/domain"
)

// TagType 标签类型
type TagType string

const (
	TagTypeCategory   TagType = "category"   // 分类标签
	TagTypeKeyword    TagType = "keyword"    // 关键词标签
	TagTypeSource     TagType = "source"     // 来源标签
	TagTypeLanguage   TagType = "language"   // 语言标签
	TagTypeDomain     TagType = "domain"     // 领域标签
	TagTypeCustom     TagType = "custom"     // 自定义标签
)

// Tag 标签实体
type Tag struct {
	domain.Entity
	Name        string    `gorm:"not null;uniqueIndex:idx_tag_name_type" json:"name"`
	Type        TagType   `gorm:"not null;uniqueIndex:idx_tag_name_type" json:"type"`
	Description string    `json:"description"`
	Color       string    `json:"color"`       // 标签颜色
	Icon        string    `json:"icon"`        // 标签图标
	ParentID    string    `gorm:"index" json:"parent_id,omitempty"` // 父标签ID
	Children    []Tag     `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	UsageCount  int       `json:"usage_count"` // 使用次数
}

// UpdateUsageCount 更新使用次数
func (t *Tag) UpdateUsageCount(delta int) {
	t.UsageCount += delta
	if t.UsageCount < 0 {
		t.UsageCount = 0
	}
	t.UpdatedAt = time.Now()
}

// AddChild 添加子标签
func (t *Tag) AddChild(child *Tag) error {
	if child.ID == t.ID {
		return NewDomainError("CIRCULAR_REFERENCE", "tag cannot be its own child")
	}
	
	child.ParentID = t.ID
	t.Children = append(t.Children, *child)
	t.UpdatedAt = time.Now()
	
	return nil
}

// RemoveChild 移除子标签
func (t *Tag) RemoveChild(childID string) {
	for i, child := range t.Children {
		if child.ID == childID {
			t.Children = append(t.Children[:i], t.Children[i+1:]...)
			t.UpdatedAt = time.Now()
			break
		}
	}
}

// IsParentOf 检查是否是指定标签的父标签
func (t *Tag) IsParentOf(tagID string) bool {
	for _, child := range t.Children {
		if child.ID == tagID {
			return true
		}
	}
	return false
}

// GetPath 获取标签路径
func (t *Tag) GetPath() string {
	// TODO: 实现获取完整路径的逻辑
	return t.Name
}

// NewTag 创建新标签
func NewTag(name string, tagType TagType) (*Tag, error) {
	if name == "" {
		return nil, NewDomainError("INVALID_NAME", "tag name cannot be empty")
	}
	
	tag := &Tag{
		Entity:      domain.NewEntity(),
		Name:        name,
		Type:        tagType,
		Color:       generateDefaultColor(tagType),
		Children:    make([]Tag, 0),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		UsageCount:  0,
	}
	
	return tag, nil
}

// generateDefaultColor 生成默认颜色
func generateDefaultColor(tagType TagType) string {
	colorMap := map[TagType]string{
		TagTypeCategory: "#3b82f6", // 蓝色
		TagTypeKeyword:  "#10b981", // 绿色
		TagTypeSource:   "#f59e0b", // 黄色
		TagTypeLanguage: "#8b5cf6", // 紫色
		TagTypeDomain:   "#ef4444", // 红色
		TagTypeCustom:   "#6b7280", // 灰色
	}
	
	if color, exists := colorMap[tagType]; exists {
		return color
	}
	
	return "#6b7280" // 默认灰色
}
