package http

import (
	"net/http"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/noah-loop/backend/modules/mcp/internal/application/service"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"github.com/noah-loop/backend/shared/pkg/utils"
	"go.uber.org/zap"
)

// MCPHandler MCP HTTP处理器
type MCPHandler struct {
	mcpService *service.MCPService
	logger     infrastructure.Logger
}

// NewMCPHandler 创建MCP处理器
func NewMCPHandler(mcpService *service.MCPService, logger infrastructure.Logger) *MCPHandler {
	return &MCPHandler{
		mcpService: mcpService,
		logger:     logger,
	}
}


// CreateSession 创建会话
func (h *MCPHandler) CreateSession(c *gin.Context) {
	cmd := service.NewCreateSessionCommand()
	if err := c.ShouldBindJSON(cmd); err != nil {
		h.logger.Warn("Invalid request body", zap.Error(err))
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("validation", err.Error()))
		return
	}
	
	result, err := h.mcpService.CreateSession(c.Request.Context(), cmd)
	if err != nil {
		h.logger.Error("Failed to create session", zap.Error(err))
		utils.ErrorResponse(c, utils.ErrInternalServer.WithCause(err))
		return
	}
	
	utils.CreatedResponse(c, result.Data, "Session created successfully")
}

// GetSessions 获取会话列表
func (h *MCPHandler) GetSessions(c *gin.Context) {
	query := service.NewGetSessionsQuery()
	if err := c.ShouldBindQuery(query); err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("validation", err.Error()))
		return
	}
	
	// TODO: 实现查询逻辑
	utils.SuccessResponse(c, []interface{}{}, "Sessions retrieved successfully")
}

// GetSession 获取单个会话
func (h *MCPHandler) GetSession(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("id", "invalid UUID format"))
		return
	}
	
	query := service.NewGetSessionQuery()
	query.SessionID = id
	
	// TODO: 实现获取单个会话逻辑
	utils.SuccessResponse(c, nil, "Session retrieved successfully")
}

// UpdateSession 更新会话
func (h *MCPHandler) UpdateSession(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("id", "invalid UUID format"))
		return
	}
	
	cmd := service.NewUpdateSessionCommand(id)
	if err := c.ShouldBindJSON(cmd); err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("validation", err.Error()))
		return
	}
	
	// TODO: 实现更新会话逻辑
	utils.SuccessResponse(c, nil, "Session updated successfully")
}

// DeleteSession 删除会话
func (h *MCPHandler) DeleteSession(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("id", "invalid UUID format"))
		return
	}
	
	// TODO: 实现删除会话逻辑
	_ = id
	utils.SuccessResponse(c, nil, "Session deleted successfully")
}

// ExtendSession 延长会话
func (h *MCPHandler) ExtendSession(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("id", "invalid UUID format"))
		return
	}
	
	var req struct {
		Duration string `json:"duration" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("validation", err.Error()))
		return
	}
	
	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("duration", "invalid duration format"))
		return
	}
	
	// TODO: 实现延长会话逻辑
	_ = id
	_ = duration
	utils.SuccessResponse(c, nil, "Session extended successfully")
}

// AddContext 添加上下文
func (h *MCPHandler) AddContext(c *gin.Context) {
	cmd := service.NewAddContextCommand()
	if err := c.ShouldBindJSON(cmd); err != nil {
		h.logger.Warn("Invalid request body", zap.Error(err))
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("validation", err.Error()))
		return
	}
	
	result, err := h.mcpService.AddContext(c.Request.Context(), cmd)
	if err != nil {
		h.logger.Error("Failed to add context", zap.Error(err))
		utils.ErrorResponse(c, utils.ErrInternalServer.WithCause(err))
		return
	}
	
	utils.CreatedResponse(c, result.Data, "Context added successfully")
}

// GetContext 获取上下文
func (h *MCPHandler) GetContext(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("id", "invalid UUID format"))
		return
	}
	
	query := service.NewGetContextQuery()
	query.ContextID = id
	if err := c.ShouldBindQuery(query); err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("validation", err.Error()))
		return
	}
	
	result, err := h.mcpService.GetContext(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to get context", zap.Error(err))
		utils.ErrorResponse(c, utils.ErrInternalServer.WithCause(err))
		return
	}
	
	utils.SuccessResponse(c, result.Data, "Context retrieved successfully")
}

// UpdateContext 更新上下文
func (h *MCPHandler) UpdateContext(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("id", "invalid UUID format"))
		return
	}
	
	cmd := service.NewUpdateContextCommand(id)
	if err := c.ShouldBindJSON(cmd); err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("validation", err.Error()))
		return
	}
	
	// TODO: 实现更新上下文逻辑
	utils.SuccessResponse(c, nil, "Context updated successfully")
}

// DeleteContext 删除上下文
func (h *MCPHandler) DeleteContext(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("id", "invalid UUID format"))
		return
	}
	
	// TODO: 实现删除上下文逻辑
	_ = id
	utils.SuccessResponse(c, nil, "Context deleted successfully")
}

// GetSessionContexts 获取会话上下文
func (h *MCPHandler) GetSessionContexts(c *gin.Context) {
	sessionIDParam := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDParam)
	if err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("session_id", "invalid UUID format"))
		return
	}
	
	query := service.NewGetSessionContextsQuery()
	query.SessionID = sessionID
	if err := c.ShouldBindQuery(query); err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("validation", err.Error()))
		return
	}
	
	result, err := h.mcpService.GetSessionContexts(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to get session contexts", zap.Error(err))
		utils.ErrorResponse(c, utils.ErrInternalServer.WithCause(err))
		return
	}
	
	utils.SuccessResponse(c, result.Data, "Session contexts retrieved successfully")
}

// AddContextToSession 向会话添加上下文
func (h *MCPHandler) AddContextToSession(c *gin.Context) {
	sessionIDParam := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDParam)
	if err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("session_id", "invalid UUID format"))
		return
	}
	
	cmd := service.NewAddContextCommand()
	cmd.SessionID = sessionID
	if err := c.ShouldBindJSON(cmd); err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("validation", err.Error()))
		return
	}
	
	result, err := h.mcpService.AddContext(c.Request.Context(), cmd)
	if err != nil {
		h.logger.Error("Failed to add context to session", zap.Error(err))
		utils.ErrorResponse(c, utils.ErrInternalServer.WithCause(err))
		return
	}
	
	utils.CreatedResponse(c, result.Data, "Context added to session successfully")
}

// CleanupExpiredSessions 清理过期会话
func (h *MCPHandler) CleanupExpiredSessions(c *gin.Context) {
	if err := h.mcpService.CleanupExpiredSessions(c.Request.Context()); err != nil {
		h.logger.Error("Failed to cleanup expired sessions", zap.Error(err))
		utils.ErrorResponse(c, utils.ErrInternalServer.WithCause(err))
		return
	}
	
	utils.SuccessResponse(c, nil, "Expired sessions cleaned up successfully")
}

// ManageIdleSessions 管理空闲会话
func (h *MCPHandler) ManageIdleSessions(c *gin.Context) {
	var req struct {
		IdleThreshold string `json:"idle_threshold" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("validation", err.Error()))
		return
	}
	
	idleThreshold, err := time.ParseDuration(req.IdleThreshold)
	if err != nil {
		utils.ErrorResponse(c, utils.ErrInvalidInput.WithDetail("idle_threshold", "invalid duration format"))
		return
	}
	
	if err := h.mcpService.ManageIdleSessions(c.Request.Context(), idleThreshold); err != nil {
		h.logger.Error("Failed to manage idle sessions", zap.Error(err))
		utils.ErrorResponse(c, utils.ErrInternalServer.WithCause(err))
		return
	}
	
	utils.SuccessResponse(c, nil, "Idle sessions managed successfully")
}
