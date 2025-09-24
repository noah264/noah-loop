package http

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/noah-loop/backend/modules/orchestrator/internal/application/service"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"github.com/noah-loop/backend/shared/pkg/utils"
	"go.uber.org/zap"
)

// OrchestratorHandler 编排器HTTP处理器
type OrchestratorHandler struct {
	orchestratorService *service.OrchestratorService
	logger              infrastructure.Logger
}

// NewOrchestratorHandler 创建编排器处理器
func NewOrchestratorHandler(orchestratorService *service.OrchestratorService, logger infrastructure.Logger) *OrchestratorHandler {
	return &OrchestratorHandler{
		orchestratorService: orchestratorService,
		logger:              logger,
	}
}

// CreateWorkflow 创建工作流
func (h *OrchestratorHandler) CreateWorkflow(c *gin.Context) {
	cmd := service.NewCreateWorkflowCommand()
	if err := c.ShouldBindJSON(cmd); err != nil {
		h.logger.Warn("Invalid request body", zap.Error(err))
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("validation", err.Error()))
		return
	}

	result, err := h.orchestratorService.CreateWorkflow(c.Request.Context(), cmd)
	if err != nil {
		h.logger.Error("Failed to create workflow", zap.Error(err))
		utils.ErrorResponse(c, utils.ErrInternalServer.WithCause(err))
		return
	}

	utils.CreatedResponse(c, result.Data, "Workflow created successfully")
}

// GetWorkflows 获取工作流列表
func (h *OrchestratorHandler) GetWorkflows(c *gin.Context) {
	query := service.NewGetWorkflowsQuery()
	if err := c.ShouldBindQuery(query); err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("validation", err.Error()))
		return
	}

	// TODO: 实现查询逻辑
	utils.SuccessResponse(c, []interface{}{}, "Workflows retrieved successfully")
}

// GetWorkflow 获取单个工作流
func (h *OrchestratorHandler) GetWorkflow(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("id", "invalid UUID format"))
		return
	}

	// TODO: 实现获取单个工作流逻辑
	_ = id
	utils.SuccessResponse(c, nil, "Workflow retrieved successfully")
}

// UpdateWorkflow 更新工作流
func (h *OrchestratorHandler) UpdateWorkflow(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("id", "invalid UUID format"))
		return
	}

	cmd := service.NewUpdateWorkflowCommand(id)
	if err := c.ShouldBindJSON(cmd); err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("validation", err.Error()))
		return
	}

	// TODO: 实现更新工作流逻辑
	utils.SuccessResponse(c, nil, "Workflow updated successfully")
}

// DeleteWorkflow 删除工作流
func (h *OrchestratorHandler) DeleteWorkflow(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("id", "invalid UUID format"))
		return
	}

	// TODO: 实现删除工作流逻辑
	_ = id
	utils.SuccessResponse(c, nil, "Workflow deleted successfully")
}

// ExecuteWorkflow 执行工作流
func (h *OrchestratorHandler) ExecuteWorkflow(c *gin.Context) {
	idParam := c.Param("id")
	workflowID, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("id", "invalid UUID format"))
		return
	}

	cmd := service.NewExecuteWorkflowCommand()
	cmd.WorkflowID = workflowID

	if err := c.ShouldBindJSON(cmd); err != nil {
		h.logger.Warn("Invalid request body", zap.Error(err))
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("validation", err.Error()))
		return
	}

	result, err := h.orchestratorService.ExecuteWorkflow(c.Request.Context(), cmd)
	if err != nil {
		h.logger.Error("Failed to execute workflow", zap.Error(err))
		utils.ErrorResponse(c, utils.ErrInternalServer.WithCause(err))
		return
	}

	utils.SuccessResponse(c, result.Data, "Workflow executed successfully")
}

// CreateTrigger 创建触发器
func (h *OrchestratorHandler) CreateTrigger(c *gin.Context) {
	cmd := service.NewCreateTriggerCommand()
	if err := c.ShouldBindJSON(cmd); err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("validation", err.Error()))
		return
	}

	// TODO: 实现创建触发器逻辑
	utils.CreatedResponse(c, nil, "Trigger created successfully")
}

// GetTriggers 获取触发器列表
func (h *OrchestratorHandler) GetTriggers(c *gin.Context) {
	query := service.NewGetTriggersQuery()
	if err := c.ShouldBindQuery(query); err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("validation", err.Error()))
		return
	}

	// TODO: 实现查询逻辑
	utils.SuccessResponse(c, []interface{}{}, "Triggers retrieved successfully")
}

// GetExecutions 获取执行历史
func (h *OrchestratorHandler) GetExecutions(c *gin.Context) {
	// TODO: 实现获取执行历史逻辑
	utils.SuccessResponse(c, []interface{}{}, "Executions retrieved successfully")
}

// GetExecution 获取单个执行记录
func (h *OrchestratorHandler) GetExecution(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("id", "invalid UUID format"))
		return
	}

	// TODO: 实现获取单个执行记录逻辑
	_ = id
	utils.SuccessResponse(c, nil, "Execution retrieved successfully")
}
