package service

import (
	"errors"
	
	"github.com/google/uuid"
	"github.com/noah-loop/backend/modules/agent/internal/domain"
	"github.com/noah-loop/backend/shared/pkg/application"
)

// CreateAgentCommand 创建智能体命令
type CreateAgentCommand struct {
	application.BaseCommand
	Name           string                        `json:"name" binding:"required"`
	Type           domain.AgentType              `json:"type" binding:"required"`
	OwnerID        uuid.UUID                     `json:"owner_id" binding:"required"`
	Description    string                        `json:"description"`
	SystemPrompt   string                        `json:"system_prompt"`
	Config         map[string]interface{}        `json:"config"`
	Capabilities   []string                      `json:"capabilities"`
	MemoryCapacity int                           `json:"memory_capacity"`
}

func NewCreateAgentCommand() *CreateAgentCommand {
	return &CreateAgentCommand{
		BaseCommand: application.BaseCommand{
			CommandID:   uuid.New(),
			CommandType: "create_agent",
		},
		Config:         make(map[string]interface{}),
		Capabilities:   make([]string, 0),
		MemoryCapacity: 1000,
	}
}

func (c *CreateAgentCommand) Validate() error {
	if c.Name == "" {
		return errors.New("agent name is required")
	}
	
	if c.OwnerID == uuid.Nil {
		return errors.New("owner ID is required")
	}
	
	// 验证智能体类型
	switch c.Type {
	case domain.AgentTypeConversational, domain.AgentTypeTask, domain.AgentTypeReflective, domain.AgentTypePlanning, domain.AgentTypeMultiModal:
		// valid
	default:
		return errors.New("invalid agent type")
	}
	
	if c.MemoryCapacity <= 0 {
		c.MemoryCapacity = 1000 // 设置默认值
	}
	
	return nil
}

// UpdateAgentCommand 更新智能体命令
type UpdateAgentCommand struct {
	application.BaseCommand
	AgentID      uuid.UUID                 `json:"agent_id" binding:"required"`
	Name         *string                   `json:"name"`
	Description  *string                   `json:"description"`
	SystemPrompt *string                   `json:"system_prompt"`
	Config       map[string]interface{}    `json:"config"`
	IsActive     *bool                     `json:"is_active"`
}

func NewUpdateAgentCommand(agentID uuid.UUID) *UpdateAgentCommand {
	return &UpdateAgentCommand{
		BaseCommand: application.BaseCommand{
			CommandID:   uuid.New(),
			CommandType: "update_agent",
		},
		AgentID: agentID,
	}
}

func (c *UpdateAgentCommand) Validate() error {
	if c.AgentID == uuid.Nil {
		return errors.New("agent ID is required")
	}
	return nil
}

// ExecuteToolCommand 执行工具命令
type ExecuteToolCommand struct {
	application.BaseCommand
	AgentID uuid.UUID                 `json:"agent_id" binding:"required"`
	ToolID  uuid.UUID                 `json:"tool_id" binding:"required"`
	Input   map[string]interface{}    `json:"input" binding:"required"`
	Context map[string]interface{}    `json:"context"`
}

func NewExecuteToolCommand() *ExecuteToolCommand {
	return &ExecuteToolCommand{
		BaseCommand: application.BaseCommand{
			CommandID:   uuid.New(),
			CommandType: "execute_tool",
		},
		Input:   make(map[string]interface{}),
		Context: make(map[string]interface{}),
	}
}

func (c *ExecuteToolCommand) Validate() error {
	if c.AgentID == uuid.Nil {
		return errors.New("agent ID is required")
	}
	
	if c.ToolID == uuid.Nil {
		return errors.New("tool ID is required")
	}
	
	if len(c.Input) == 0 {
		return errors.New("input is required")
	}
	
	return nil
}

// ChatCommand 对话命令
type ChatCommand struct {
	application.BaseCommand
	AgentID   uuid.UUID `json:"agent_id" binding:"required"`
	SessionID uuid.UUID `json:"session_id"`
	Message   string    `json:"message" binding:"required"`
	Context   map[string]interface{} `json:"context"`
}

func NewChatCommand() *ChatCommand {
	return &ChatCommand{
		BaseCommand: application.BaseCommand{
			CommandID:   uuid.New(),
			CommandType: "chat",
		},
		SessionID: uuid.New(), // 默认生成新会话
		Context:   make(map[string]interface{}),
	}
}

func (c *ChatCommand) Validate() error {
	if c.AgentID == uuid.Nil {
		return errors.New("agent ID is required")
	}
	
	if c.Message == "" {
		return errors.New("message is required")
	}
	
	return nil
}

// CreateToolCommand 创建工具命令
type CreateToolCommand struct {
	application.BaseCommand
	Name          string                        `json:"name" binding:"required"`
	Type          domain.ToolType               `json:"type" binding:"required"`
	OwnerID       uuid.UUID                     `json:"owner_id" binding:"required"`
	Description   string                        `json:"description"`
	Schema        map[string]interface{}        `json:"schema"`
	Config        map[string]interface{}        `json:"config"`
	ExecutionMode domain.ToolExecutionMode      `json:"execution_mode"`
	IsPublic      bool                          `json:"is_public"`
}

func NewCreateToolCommand() *CreateToolCommand {
	return &CreateToolCommand{
		BaseCommand: application.BaseCommand{
			CommandID:   uuid.New(),
			CommandType: "create_tool",
		},
		Schema:        make(map[string]interface{}),
		Config:        make(map[string]interface{}),
		ExecutionMode: domain.ExecutionModeSync,
		IsPublic:      false,
	}
}

func (c *CreateToolCommand) Validate() error {
	if c.Name == "" {
		return errors.New("tool name is required")
	}
	
	if c.OwnerID == uuid.Nil {
		return errors.New("owner ID is required")
	}
	
	// 验证工具类型
	switch c.Type {
	case domain.ToolTypeFunction, domain.ToolTypeAPI, domain.ToolTypeDatabase, domain.ToolTypeFile, domain.ToolTypeCalculator, domain.ToolTypeWeb, domain.ToolTypeCustom:
		// valid
	default:
		return errors.New("invalid tool type")
	}
	
	// 验证执行模式
	switch c.ExecutionMode {
	case domain.ExecutionModeSync, domain.ExecutionModeAsync, domain.ExecutionModeStream:
		// valid
	default:
		return errors.New("invalid execution mode")
	}
	
	return nil
}

// AssignToolCommand 分配工具命令
type AssignToolCommand struct {
	application.BaseCommand
	AgentID uuid.UUID `json:"agent_id" binding:"required"`
	ToolID  uuid.UUID `json:"tool_id" binding:"required"`
}

func NewAssignToolCommand() *AssignToolCommand {
	return &AssignToolCommand{
		BaseCommand: application.BaseCommand{
			CommandID:   uuid.New(),
			CommandType: "assign_tool",
		},
	}
}

func (c *AssignToolCommand) Validate() error {
	if c.AgentID == uuid.Nil {
		return errors.New("agent ID is required")
	}
	
	if c.ToolID == uuid.Nil {
		return errors.New("tool ID is required")
	}
	
	return nil
}

// LearnCommand 学习命令
type LearnCommand struct {
	application.BaseCommand
	AgentID    uuid.UUID `json:"agent_id" binding:"required"`
	Knowledge  string    `json:"knowledge" binding:"required"`
	Importance float64   `json:"importance"`
	Tags       []string  `json:"tags"`
}

func NewLearnCommand() *LearnCommand {
	return &LearnCommand{
		BaseCommand: application.BaseCommand{
			CommandID:   uuid.New(),
			CommandType: "learn",
		},
		Importance: 0.5, // 默认重要性
		Tags:       make([]string, 0),
	}
}

func (c *LearnCommand) Validate() error {
	if c.AgentID == uuid.Nil {
		return errors.New("agent ID is required")
	}
	
	if c.Knowledge == "" {
		return errors.New("knowledge is required")
	}
	
	if c.Importance < 0 || c.Importance > 1 {
		return errors.New("importance must be between 0 and 1")
	}
	
	return nil
}

// 查询对象

// GetAgentsQuery 获取智能体查询
type GetAgentsQuery struct {
	application.BaseQuery
	OwnerID  *uuid.UUID        `form:"owner_id"`
	Type     *domain.AgentType `form:"type"`
	Status   *domain.AgentStatus `form:"status"`
	IsActive *bool             `form:"is_active"`
	Page     int               `form:"page,default=1"`
	PageSize int               `form:"page_size,default=20"`
}

func NewGetAgentsQuery() *GetAgentsQuery {
	return &GetAgentsQuery{
		BaseQuery: application.BaseQuery{
			QueryID:   uuid.New(),
			QueryType: "get_agents",
		},
		Page:     1,
		PageSize: 20,
	}
}

func (q *GetAgentsQuery) Validate() error {
	if q.Page <= 0 {
		return errors.New("page must be greater than 0")
	}
	
	if q.PageSize <= 0 || q.PageSize > 100 {
		return errors.New("page size must be between 1 and 100")
	}
	
	return nil
}

// GetToolsQuery 获取工具查询
type GetToolsQuery struct {
	application.BaseQuery
	OwnerID   *uuid.UUID       `form:"owner_id"`
	Type      *domain.ToolType `form:"type"`
	IsEnabled *bool            `form:"is_enabled"`
	IsPublic  *bool            `form:"is_public"`
	Page      int              `form:"page,default=1"`
	PageSize  int              `form:"page_size,default=20"`
}

func NewGetToolsQuery() *GetToolsQuery {
	return &GetToolsQuery{
		BaseQuery: application.BaseQuery{
			QueryID:   uuid.New(),
			QueryType: "get_tools",
		},
		Page:     1,
		PageSize: 20,
	}
}

func (q *GetToolsQuery) Validate() error {
	if q.Page <= 0 {
		return errors.New("page must be greater than 0")
	}
	
	if q.PageSize <= 0 || q.PageSize > 100 {
		return errors.New("page size must be between 1 and 100")
	}
	
	return nil
}

// SearchMemoryQuery 搜索记忆查询
type SearchMemoryQuery struct {
	application.BaseQuery
	AgentID    uuid.UUID           `form:"agent_id" binding:"required"`
	Query      string              `form:"query" binding:"required"`
	Type       *domain.MemoryType  `form:"type"`
	Limit      int                 `form:"limit,default=10"`
}

func NewSearchMemoryQuery() *SearchMemoryQuery {
	return &SearchMemoryQuery{
		BaseQuery: application.BaseQuery{
			QueryID:   uuid.New(),
			QueryType: "search_memory",
		},
		Limit: 10,
	}
}

func (q *SearchMemoryQuery) Validate() error {
	if q.AgentID == uuid.Nil {
		return errors.New("agent ID is required")
	}
	
	if q.Query == "" {
		return errors.New("query is required")
	}
	
	if q.Limit <= 0 || q.Limit > 100 {
		return errors.New("limit must be between 1 and 100")
	}
	
	return nil
}
