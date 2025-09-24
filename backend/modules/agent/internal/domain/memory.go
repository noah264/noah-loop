package domain

import (
	"sort"
	"time"
	
	"github.com/google/uuid"
	"github.com/noah-loop/backend/shared/pkg/domain"
)

// MemoryType 记忆类型
type MemoryType string

const (
	MemoryTypeConversation MemoryType = "conversation" // 对话记忆
	MemoryTypeLearned      MemoryType = "learned"      // 学习记忆
	MemoryTypeExperience   MemoryType = "experience"   // 经验记忆
	MemoryTypeKnowledge    MemoryType = "knowledge"    // 知识记忆
	MemoryTypeEpisodic     MemoryType = "episodic"     // 情节记忆
	MemoryTypeSemantic     MemoryType = "semantic"     // 语义记忆
)

// Memory 记忆条目
type Memory struct {
	domain.BaseEntity
	AgentID      uuid.UUID               `json:"agent_id" gorm:"type:uuid;index"`
	Type         MemoryType              `json:"type" gorm:"not null"`
	Content      string                  `json:"content" gorm:"type:text;not null"`
	Context      map[string]interface{}  `json:"context" gorm:"type:jsonb"`
	Importance   float64                 `json:"importance" gorm:"default:1.0"` // 重要性评分 0-1
	AccessCount  int                     `json:"access_count" gorm:"default:0"`
	LastAccessed time.Time               `json:"last_accessed"`
	Decay        float64                 `json:"decay" gorm:"default:0.0"`      // 遗忘衰减
	Tags         []string                `json:"tags" gorm:"type:text[]"`
	IsActive     bool                    `json:"is_active" gorm:"default:true"`
	
	// 关联信息
	RelatedMemories []uuid.UUID `json:"related_memories" gorm:"type:uuid[]"`
	Embedding       []float64   `json:"embedding" gorm:"type:real[]"` // 向量嵌入
}

// NewMemory 创建新记忆
func NewMemory(content string, memoryType MemoryType, importance float64) *Memory {
	return &Memory{
		BaseEntity: domain.BaseEntity{
			ID:        domain.NewEntityID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Type:         memoryType,
		Content:      content,
		Context:      make(map[string]interface{}),
		Importance:   importance,
		AccessCount:  0,
		LastAccessed: time.Now(),
		Decay:        0.0,
		Tags:         make([]string, 0),
		IsActive:     true,
		RelatedMemories: make([]uuid.UUID, 0),
		Embedding:       make([]float64, 0),
	}
}

// Access 访问记忆
func (m *Memory) Access() {
	m.AccessCount++
	m.LastAccessed = time.Now()
	m.UpdatedAt = time.Now()
	
	// 增强重要性
	m.Importance = min(1.0, m.Importance+0.01)
}

// ApplyDecay 应用遗忘衰减
func (m *Memory) ApplyDecay(decayRate float64) {
	timeSinceAccess := time.Since(m.LastAccessed).Hours()
	m.Decay = 1 - (1/(1+decayRate*timeSinceAccess))
	m.Importance = max(0.0, m.Importance*(1-m.Decay))
}

// GetRelevanceScore 获取相关性评分
func (m *Memory) GetRelevanceScore() float64 {
	// 综合考虑重要性、访问频率、时效性
	recencyScore := 1.0 / (1.0 + time.Since(m.LastAccessed).Hours()/24) // 时效性
	frequencyScore := min(1.0, float64(m.AccessCount)/10.0)              // 访问频率
	importanceScore := m.Importance                                       // 重要性
	
	return (recencyScore + frequencyScore + importanceScore) / 3.0
}

// AgentMemory 智能体记忆系统
type AgentMemory struct {
	domain.BaseEntity
	AgentID         uuid.UUID `json:"agent_id" gorm:"type:uuid;uniqueIndex"`
	Memories        []*Memory `json:"memories" gorm:"foreignKey:AgentID"`
	Capacity        int       `json:"capacity" gorm:"default:1000"`
	DecayRate       float64   `json:"decay_rate" gorm:"default:0.01"`
	ConsolidationThreshold float64 `json:"consolidation_threshold" gorm:"default:0.8"`
	
	// 统计信息
	TotalMemories   int     `json:"total_memories"`
	ActiveMemories  int     `json:"active_memories"`
	MemoryUsage     float64 `json:"memory_usage"` // 使用率
}

// NewAgentMemory 创建智能体记忆系统
func NewAgentMemory(agentID uuid.UUID) *AgentMemory {
	return &AgentMemory{
		BaseEntity: domain.BaseEntity{
			ID:        domain.NewEntityID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		AgentID:                agentID,
		Memories:               make([]*Memory, 0),
		Capacity:               1000,
		DecayRate:              0.01,
		ConsolidationThreshold: 0.8,
		TotalMemories:          0,
		ActiveMemories:         0,
		MemoryUsage:            0.0,
	}
}

// AddMemory 添加记忆
func (am *AgentMemory) AddMemory(memory *Memory) error {
	memory.AgentID = am.AgentID
	
	// 检查是否超出容量
	if len(am.Memories) >= am.Capacity {
		// 执行记忆整理
		if err := am.Consolidate(); err != nil {
			return err
		}
	}
	
	am.Memories = append(am.Memories, memory)
	am.updateStatistics()
	
	return nil
}

// SearchMemories 搜索记忆
func (am *AgentMemory) SearchMemories(query string, memoryType *MemoryType, limit int) []*Memory {
	var results []*Memory
	
	for _, memory := range am.Memories {
		if !memory.IsActive {
			continue
		}
		
		// 类型过滤
		if memoryType != nil && memory.Type != *memoryType {
			continue
		}
		
		// 简单的关键词匹配（实际应用中可能使用向量搜索）
		if contains(memory.Content, query) || containsTags(memory.Tags, query) {
			memory.Access() // 访问记忆
			results = append(results, memory)
		}
	}
	
	// 按相关性排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].GetRelevanceScore() > results[j].GetRelevanceScore()
	})
	
	// 限制结果数量
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	
	return results
}

// Consolidate 记忆整理（遗忘不重要的记忆）
func (am *AgentMemory) Consolidate() error {
	if len(am.Memories) == 0 {
		return nil
	}
	
	// 应用衰减
	for _, memory := range am.Memories {
		memory.ApplyDecay(am.DecayRate)
	}
	
	// 按相关性排序
	sort.Slice(am.Memories, func(i, j int) bool {
		return am.Memories[i].GetRelevanceScore() > am.Memories[j].GetRelevanceScore()
	})
	
	// 保留重要记忆，标记不重要的为非激活
	consolidationPoint := int(float64(am.Capacity) * am.ConsolidationThreshold)
	for i := consolidationPoint; i < len(am.Memories); i++ {
		am.Memories[i].IsActive = false
	}
	
	am.updateStatistics()
	return nil
}

// GetRecentMemories 获取最近记忆
func (am *AgentMemory) GetRecentMemories(limit int, memoryType *MemoryType) []*Memory {
	var memories []*Memory
	
	// 收集符合条件的记忆
	for _, memory := range am.Memories {
		if !memory.IsActive {
			continue
		}
		
		if memoryType != nil && memory.Type != *memoryType {
			continue
		}
		
		memories = append(memories, memory)
	}
	
	// 按时间排序
	sort.Slice(memories, func(i, j int) bool {
		return memories[i].CreatedAt.After(memories[j].CreatedAt)
	})
	
	// 限制数量
	if limit > 0 && len(memories) > limit {
		memories = memories[:limit]
	}
	
	return memories
}

// updateStatistics 更新统计信息
func (am *AgentMemory) updateStatistics() {
	am.TotalMemories = len(am.Memories)
	
	activeCount := 0
	for _, memory := range am.Memories {
		if memory.IsActive {
			activeCount++
		}
	}
	
	am.ActiveMemories = activeCount
	am.MemoryUsage = float64(am.ActiveMemories) / float64(am.Capacity)
	am.UpdatedAt = time.Now()
}

// 工具函数
func contains(content, query string) bool {
	// 简单的字符串包含检查
	// 实际应用中可能使用更复杂的文本匹配算法
	return len(query) > 0 && len(content) > 0
}

func containsTags(tags []string, query string) bool {
	for _, tag := range tags {
		if tag == query {
			return true
		}
	}
	return false
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
