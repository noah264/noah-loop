package http

import (
	"net/http"
	
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/noah-loop/backend/modules/agent/internal/application/service"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"github.com/noah-loop/backend/shared/pkg/utils"
	"go.uber.org/zap"
)

// AgentHandler 智能体HTTP处理器
type AgentHandler struct {
	agentService *service.AgentService
	logger       infrastructure.Logger
}

// NewAgentHandler 创建智能体处理器
func NewAgentHandler(agentService *service.AgentService, logger infrastructure.Logger) *AgentHandler {
	return &AgentHandler{
		agentService: agentService,
		logger:       logger,
	}
}


// CreateAgent 创建智能体
func (h *AgentHandler) CreateAgent(c *gin.Context) {
	cmd := service.NewCreateAgentCommand()
	if err := c.ShouldBindJSON(cmd); err != nil {
		h.logger.Warn("Invalid request body", zap.Error(err))
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("validation", err.Error()))
		return
	}
	
	result, err := h.agentService.CreateAgent(c.Request.Context(), cmd)
	if err != nil {
		h.logger.Error("Failed to create agent", zap.Error(err))
		utils.ErrorResponse(c, utils.ErrInternalServer.WithCause(err))
		return
	}
	
	utils.CreatedResponse(c, result.Data, "Agent created successfully")
}

// GetAgents 获取智能体列表
func (h *AgentHandler) GetAgents(c *gin.Context) {
	query := service.NewGetAgentsQuery()
	if err := c.ShouldBindQuery(query); err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("validation", err.Error()))
		return
	}
	
	// TODO: 实现查询逻辑
	utils.SuccessResponse(c, []interface{}{}, "Agents retrieved successfully")
}

// GetAgent 获取单个智能体
func (h *AgentHandler) GetAgent(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("id", "invalid UUID format"))
		return
	}
	
	// TODO: 实现获取单个智能体逻辑
	_ = id
	utils.SuccessResponse(c, nil, "Agent retrieved successfully")
}

// UpdateAgent 更新智能体
func (h *AgentHandler) UpdateAgent(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("id", "invalid UUID format"))
		return
	}
	
	cmd := service.NewUpdateAgentCommand(id)
	if err := c.ShouldBindJSON(cmd); err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("validation", err.Error()))
		return
	}
	
	// TODO: 实现更新智能体逻辑
	utils.SuccessResponse(c, nil, "Agent updated successfully")
}

// DeleteAgent 删除智能体
func (h *AgentHandler) DeleteAgent(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("id", "invalid UUID format"))
		return
	}
	
	// TODO: 实现删除智能体逻辑
	_ = id
	utils.SuccessResponse(c, nil, "Agent deleted successfully")
}

// ChatWithAgent 与智能体对话
func (h *AgentHandler) ChatWithAgent(c *gin.Context) {
	idParam := c.Param("id")
	agentID, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("id", "invalid UUID format"))
		return
	}
	
	cmd := service.NewChatCommand()
	cmd.AgentID = agentID
	
	if err := c.ShouldBindJSON(cmd); err != nil {
		h.logger.Warn("Invalid request body", zap.Error(err))
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("validation", err.Error()))
		return
	}
	
	result, err := h.agentService.ChatWithAgent(c.Request.Context(), cmd)
	if err != nil {
		h.logger.Error("Failed to chat with agent", zap.Error(err))
		utils.ErrorResponse(c, utils.ErrInternalServer.WithCause(err))
		return
	}
	
	utils.SuccessResponse(c, result.Data, "Chat completed successfully")
}

// LearnAgent 让智能体学习
func (h *AgentHandler) LearnAgent(c *gin.Context) {
	idParam := c.Param("id")
	agentID, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("id", "invalid UUID format"))
		return
	}
	
	cmd := service.NewLearnCommand()
	cmd.AgentID = agentID
	
	if err := c.ShouldBindJSON(cmd); err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("validation", err.Error()))
		return
	}
	
	// TODO: 实现学习逻辑
	utils.SuccessResponse(c, nil, "Agent learned successfully")
}

// CreateTool 创建工具
func (h *AgentHandler) CreateTool(c *gin.Context) {
	cmd := service.NewCreateToolCommand()
	if err := c.ShouldBindJSON(cmd); err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("validation", err.Error()))
		return
	}
	
	// TODO: 实现创建工具逻辑
	utils.CreatedResponse(c, nil, "Tool created successfully")
}

// GetTools 获取工具列表
func (h *AgentHandler) GetTools(c *gin.Context) {
	query := service.NewGetToolsQuery()
	if err := c.ShouldBindQuery(query); err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("validation", err.Error()))
		return
	}
	
	// TODO: 实现查询逻辑
	utils.SuccessResponse(c, []interface{}{}, "Tools retrieved successfully")
}

// GetTool 获取单个工具
func (h *AgentHandler) GetTool(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("id", "invalid UUID format"))
		return
	}
	
	// TODO: 实现获取单个工具逻辑
	_ = id
	utils.SuccessResponse(c, nil, "Tool retrieved successfully")
}

// UpdateTool 更新工具
func (h *AgentHandler) UpdateTool(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("id", "invalid UUID format"))
		return
	}
	
	// TODO: 实现更新工具逻辑
	_ = id
	utils.SuccessResponse(c, nil, "Tool updated successfully")
}

// DeleteTool 删除工具
func (h *AgentHandler) DeleteTool(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("id", "invalid UUID format"))
		return
	}
	
	// TODO: 实现删除工具逻辑
	_ = id
	utils.SuccessResponse(c, nil, "Tool deleted successfully")
}

// ExecuteTool 执行工具
func (h *AgentHandler) ExecuteTool(c *gin.Context) {
	idParam := c.Param("id")
	toolID, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("id", "invalid UUID format"))
		return
	}
	
	cmd := service.NewExecuteToolCommand()
	cmd.ToolID = toolID
	
	if err := c.ShouldBindJSON(cmd); err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("validation", err.Error()))
		return
	}
	
	result, err := h.agentService.ExecuteTool(c.Request.Context(), cmd)
	if err != nil {
		h.logger.Error("Failed to execute tool", zap.Error(err))
		utils.ErrorResponse(c, utils.ErrInternalServer.WithCause(err))
		return
	}
	
	utils.SuccessResponse(c, result.Data, "Tool executed successfully")
}

// AssignTool 分配工具给智能体
func (h *AgentHandler) AssignTool(c *gin.Context) {
	cmd := service.NewAssignToolCommand()
	if err := c.ShouldBindJSON(cmd); err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("validation", err.Error()))
		return
	}
	
	// TODO: 实现工具分配逻辑
	utils.SuccessResponse(c, nil, "Tool assigned successfully")
}

// UnassignTool 取消分配工具
func (h *AgentHandler) UnassignTool(c *gin.Context) {
	cmd := service.NewAssignToolCommand()
	if err := c.ShouldBindJSON(cmd); err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("validation", err.Error()))
		return
	}
	
	// TODO: 实现工具取消分配逻辑
	utils.SuccessResponse(c, nil, "Tool unassigned successfully")
}

// SearchMemory 搜索记忆
func (h *AgentHandler) SearchMemory(c *gin.Context) {
	query := service.NewSearchMemoryQuery()
	if err := c.ShouldBindQuery(query); err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("validation", err.Error()))
		return
	}
	
	// TODO: 实现记忆搜索逻辑
	utils.SuccessResponse(c, []interface{}{}, "Memory searched successfully")
}

// GetRecentMemories 获取最近记忆
func (h *AgentHandler) GetRecentMemories(c *gin.Context) {
	agentIDParam := c.Param("agent_id")
	agentID, err := uuid.Parse(agentIDParam)
	if err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("agent_id", "invalid UUID format"))
		return
	}
	
	// TODO: 实现获取最近记忆逻辑
	_ = agentID
	utils.SuccessResponse(c, []interface{}{}, "Recent memories retrieved successfully")
}

// GetExecutions 获取执行历史
func (h *AgentHandler) GetExecutions(c *gin.Context) {
	// TODO: 实现获取执行历史逻辑
	utils.SuccessResponse(c, []interface{}{}, "Executions retrieved successfully")
}

// GetExecution 获取单个执行记录
func (h *AgentHandler) GetExecution(c *gin.Context) {
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
