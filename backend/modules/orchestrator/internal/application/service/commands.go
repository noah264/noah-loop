package service

import (
	"errors"
	"time"
	
	"github.com/google/uuid"
	"github.com/noah-loop/backend/modules/orchestrator/internal/domain"
	"github.com/noah-loop/backend/shared/pkg/application"
)

// CreateWorkflowCommand 创建工作流命令
type CreateWorkflowCommand struct {
	application.BaseCommand
	Name        string                    `json:"name" binding:"required"`
	Description string                    `json:"description"`
	OwnerID     uuid.UUID                 `json:"owner_id" binding:"required"`
	Definition  map[string]interface{}    `json:"definition"`
	Variables   map[string]interface{}    `json:"variables"`
	Tags        []string                  `json:"tags"`
	IsTemplate  bool                      `json:"is_template"`
}

func NewCreateWorkflowCommand() *CreateWorkflowCommand {
	return &CreateWorkflowCommand{
		BaseCommand: application.BaseCommand{
			CommandID:   uuid.New(),
			CommandType: "create_workflow",
		},
		Definition: make(map[string]interface{}),
		Variables:  make(map[string]interface{}),
		Tags:       make([]string, 0),
		IsTemplate: false,
	}
}

func (c *CreateWorkflowCommand) Validate() error {
	if c.Name == "" {
		return errors.New("workflow name is required")
	}
	
	if c.OwnerID == uuid.Nil {
		return errors.New("owner ID is required")
	}
	
	return nil
}

// ExecuteWorkflowCommand 执行工作流命令
type ExecuteWorkflowCommand struct {
	application.BaseCommand
	WorkflowID uuid.UUID                 `json:"workflow_id" binding:"required"`
	TriggerID  uuid.UUID                 `json:"trigger_id"`
	Input      map[string]interface{}    `json:"input"`
	Context    map[string]interface{}    `json:"context"`
}

func NewExecuteWorkflowCommand() *ExecuteWorkflowCommand {
	return &ExecuteWorkflowCommand{
		BaseCommand: application.BaseCommand{
			CommandID:   uuid.New(),
			CommandType: "execute_workflow",
		},
		Input:   make(map[string]interface{}),
		Context: make(map[string]interface{}),
	}
}

func (c *ExecuteWorkflowCommand) Validate() error {
	if c.WorkflowID == uuid.Nil {
		return errors.New("workflow ID is required")
	}
	
	return nil
}

// AddStepCommand 添加步骤命令
type AddStepCommand struct {
	application.BaseCommand
	WorkflowID   uuid.UUID                 `json:"workflow_id" binding:"required"`
	Name         string                    `json:"name" binding:"required"`
	Type         domain.StepType           `json:"type" binding:"required"`
	Order        int                       `json:"order" binding:"required"`
	Description  string                    `json:"description"`
	Config       map[string]interface{}    `json:"config"`
	Input        map[string]interface{}    `json:"input"`
	Timeout      time.Duration             `json:"timeout"`
	MaxRetries   int                       `json:"max_retries"`
	Dependencies []uuid.UUID               `json:"dependencies"`
}

func NewAddStepCommand() *AddStepCommand {
	return &AddStepCommand{
		BaseCommand: application.BaseCommand{
			CommandID:   uuid.New(),
			CommandType: "add_step",
		},
		Config:       make(map[string]interface{}),
		Input:        make(map[string]interface{}),
		Timeout:      30 * time.Minute,
		MaxRetries:   3,
		Dependencies: make([]uuid.UUID, 0),
	}
}

func (c *AddStepCommand) Validate() error {
	if c.WorkflowID == uuid.Nil {
		return errors.New("workflow ID is required")
	}
	
	if c.Name == "" {
		return errors.New("step name is required")
	}
	
	if c.Order < 0 {
		return errors.New("step order must be non-negative")
	}
	
	// 验证步骤类型
	switch c.Type {
	case domain.StepTypeAction, domain.StepTypeCondition, domain.StepTypeLoop, 
		 domain.StepTypeParallel, domain.StepTypeWait, domain.StepTypeHuman, 
		 domain.StepTypeSubworkflow:
		// valid
	default:
		return errors.New("invalid step type")
	}
	
	if c.MaxRetries < 0 {
		return errors.New("max retries must be non-negative")
	}
	
	return nil
}

// AddTriggerCommand 添加触发器命令
type AddTriggerCommand struct {
	application.BaseCommand
	WorkflowID  uuid.UUID                 `json:"workflow_id" binding:"required"`
	Type        domain.TriggerType        `json:"type" binding:"required"`
	Name        string                    `json:"name" binding:"required"`
	Description string                    `json:"description"`
	Config      map[string]interface{}    `json:"config"`
	Schedule    string                    `json:"schedule"` // for schedule triggers
	Timezone    string                    `json:"timezone"` // for schedule triggers
	Conditions  []TriggerCondition        `json:"conditions"` // for condition triggers
}

type TriggerCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

func NewAddTriggerCommand() *AddTriggerCommand {
	return &AddTriggerCommand{
		BaseCommand: application.BaseCommand{
			CommandID:   uuid.New(),
			CommandType: "add_trigger",
		},
		Config:     make(map[string]interface{}),
		Conditions: make([]TriggerCondition, 0),
	}
}

func (c *AddTriggerCommand) Validate() error {
	if c.WorkflowID == uuid.Nil {
		return errors.New("workflow ID is required")
	}
	
	if c.Name == "" {
		return errors.New("trigger name is required")
	}
	
	// 验证触发器类型
	switch c.Type {
	case domain.TriggerTypeManual, domain.TriggerTypeSchedule, domain.TriggerTypeEvent, 
		 domain.TriggerTypeWebhook, domain.TriggerTypeCondition:
		// valid
	default:
		return errors.New("invalid trigger type")
	}
	
	// 验证特定类型的配置
	if c.Type == domain.TriggerTypeSchedule && c.Schedule == "" {
		return errors.New("schedule is required for schedule triggers")
	}
	
	if c.Type == domain.TriggerTypeCondition && len(c.Conditions) == 0 {
		return errors.New("conditions are required for condition triggers")
	}
	
	return nil
}

// UpdateWorkflowCommand 更新工作流命令
type UpdateWorkflowCommand struct {
	application.BaseCommand
	WorkflowID  uuid.UUID                 `json:"workflow_id" binding:"required"`
	Name        *string                   `json:"name"`
	Description *string                   `json:"description"`
	Definition  map[string]interface{}    `json:"definition"`
	Variables   map[string]interface{}    `json:"variables"`
	Tags        []string                  `json:"tags"`
}

func NewUpdateWorkflowCommand(workflowID uuid.UUID) *UpdateWorkflowCommand {
	return &UpdateWorkflowCommand{
		BaseCommand: application.BaseCommand{
			CommandID:   uuid.New(),
			CommandType: "update_workflow",
		},
		WorkflowID: workflowID,
	}
}

func (c *UpdateWorkflowCommand) Validate() error {
	if c.WorkflowID == uuid.Nil {
		return errors.New("workflow ID is required")
	}
	
	return nil
}

// ActivateWorkflowCommand 激活工作流命令
type ActivateWorkflowCommand struct {
	application.BaseCommand
	WorkflowID uuid.UUID `json:"workflow_id" binding:"required"`
}

func NewActivateWorkflowCommand(workflowID uuid.UUID) *ActivateWorkflowCommand {
	return &ActivateWorkflowCommand{
		BaseCommand: application.BaseCommand{
			CommandID:   uuid.New(),
			CommandType: "activate_workflow",
		},
		WorkflowID: workflowID,
	}
}

func (c *ActivateWorkflowCommand) Validate() error {
	if c.WorkflowID == uuid.Nil {
		return errors.New("workflow ID is required")
	}
	
	return nil
}

// 查询对象

// GetWorkflowsQuery 获取工作流查询
type GetWorkflowsQuery struct {
	application.BaseQuery
	OwnerID    *uuid.UUID              `form:"owner_id"`
	Status     *domain.WorkflowStatus  `form:"status"`
	IsTemplate *bool                   `form:"is_template"`
	Tags       []string                `form:"tags"`
	Page       int                     `form:"page,default=1"`
	PageSize   int                     `form:"page_size,default=20"`
}

func NewGetWorkflowsQuery() *GetWorkflowsQuery {
	return &GetWorkflowsQuery{
		BaseQuery: application.BaseQuery{
			QueryID:   uuid.New(),
			QueryType: "get_workflows",
		},
		Tags:     make([]string, 0),
		Page:     1,
		PageSize: 20,
	}
}

func (q *GetWorkflowsQuery) Validate() error {
	if q.Page <= 0 {
		return errors.New("page must be greater than 0")
	}
	
	if q.PageSize <= 0 || q.PageSize > 100 {
		return errors.New("page size must be between 1 and 100")
	}
	
	return nil
}

// GetExecutionsQuery 获取执行记录查询
type GetExecutionsQuery struct {
	application.BaseQuery
	WorkflowID *uuid.UUID               `form:"workflow_id"`
	Status     *domain.ExecutionStatus  `form:"status"`
	Page       int                      `form:"page,default=1"`
	PageSize   int                      `form:"page_size,default=20"`
}

func NewGetExecutionsQuery() *GetExecutionsQuery {
	return &GetExecutionsQuery{
		BaseQuery: application.BaseQuery{
			QueryID:   uuid.New(),
			QueryType: "get_executions",
		},
		Page:     1,
		PageSize: 20,
	}
}

func (q *GetExecutionsQuery) Validate() error {
	if q.Page <= 0 {
		return errors.New("page must be greater than 0")
	}
	
	if q.PageSize <= 0 || q.PageSize > 100 {
		return errors.New("page size must be between 1 and 100")
	}
	
	return nil
}
