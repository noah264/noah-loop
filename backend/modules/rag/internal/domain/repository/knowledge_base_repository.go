package repository

import (
	"context"

	"github.com/noah-loop/backend/modules/rag/internal/domain"
)

// KnowledgeBaseRepository 知识库仓储接口
type KnowledgeBaseRepository interface {
	// 基本CRUD操作
	Save(ctx context.Context, knowledgeBase *domain.KnowledgeBase) error
	FindByID(ctx context.Context, id string) (*domain.KnowledgeBase, error)
	FindByName(ctx context.Context, name, ownerID string) (*domain.KnowledgeBase, error)
	Update(ctx context.Context, knowledgeBase *domain.KnowledgeBase) error
	Delete(ctx context.Context, id string) error

	// 查询操作
	FindByOwnerID(ctx context.Context, ownerID string) ([]*domain.KnowledgeBase, error)
	FindByStatus(ctx context.Context, status domain.KnowledgeBaseStatus) ([]*domain.KnowledgeBase, error)
	FindByTags(ctx context.Context, tagNames []string) ([]*domain.KnowledgeBase, error)

	// 分页查询
	FindWithPagination(ctx context.Context, offset, limit int) ([]*domain.KnowledgeBase, int64, error)
	FindByOwnerIDWithPagination(ctx context.Context, ownerID string, offset, limit int) ([]*domain.KnowledgeBase, int64, error)

	// 搜索操作
	SearchByName(ctx context.Context, query string, ownerID string, limit int) ([]*domain.KnowledgeBase, error)
	SearchByDescription(ctx context.Context, query string, limit int) ([]*domain.KnowledgeBase, error)

	// 统计操作
	CountByOwnerID(ctx context.Context, ownerID string) (int64, error)
	CountByStatus(ctx context.Context, status domain.KnowledgeBaseStatus) (int64, error)
	UpdateStatistics(ctx context.Context, knowledgeBaseID string, stats domain.KnowledgeBaseStats) error

	// 访问记录
	RecordQuery(ctx context.Context, knowledgeBaseID string, score float32) error
	GetQueryHistory(ctx context.Context, knowledgeBaseID string, limit int) ([]QueryRecord, error)

	// 权限相关
	CheckAccess(ctx context.Context, knowledgeBaseID, userID string) (bool, error)
	GrantAccess(ctx context.Context, knowledgeBaseID, userID string, permission Permission) error
	RevokeAccess(ctx context.Context, knowledgeBaseID, userID string) error
	ListAccessUsers(ctx context.Context, knowledgeBaseID string) ([]UserPermission, error)
}

// QueryRecord 查询记录
type QueryRecord struct {
	ID              string    `json:"id"`
	KnowledgeBaseID string    `json:"knowledge_base_id"`
	Query           string    `json:"query"`
	ResultCount     int       `json:"result_count"`
	AverageScore    float32   `json:"average_score"`
	Duration        int64     `json:"duration"` // 毫秒
	CreatedAt       string    `json:"created_at"`
}

// Permission 权限类型
type Permission string

const (
	PermissionRead   Permission = "read"   // 只读
	PermissionWrite  Permission = "write"  // 读写
	PermissionAdmin  Permission = "admin"  // 管理员
	PermissionOwner  Permission = "owner"  // 所有者
)

// UserPermission 用户权限
type UserPermission struct {
	UserID     string     `json:"user_id"`
	Permission Permission `json:"permission"`
	GrantedAt  string     `json:"granted_at"`
}
