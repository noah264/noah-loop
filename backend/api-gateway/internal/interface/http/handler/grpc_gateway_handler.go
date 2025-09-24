package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	agentv1 "github.com/noah-loop/backend/api-gateway/proto/agent/v1"
	llmv1 "github.com/noah-loop/backend/api-gateway/proto/llm/v1"
	"github.com/noah-loop/backend/api-gateway/internal/application/service"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// GRPCGatewayHandler gRPC网关处理器 - 将HTTP请求转换为gRPC调用
type GRPCGatewayHandler struct {
	grpcProxyService *service.GRPCProxyService
	logger           infrastructure.Logger
}

// NewGRPCGatewayHandler 创建gRPC网关处理器
func NewGRPCGatewayHandler(grpcProxyService *service.GRPCProxyService, logger infrastructure.Logger) *GRPCGatewayHandler {
	return &GRPCGatewayHandler{
		grpcProxyService: grpcProxyService,
		logger:           logger,
	}
}

// Agent相关处理器

// CreateAgent 创建智能体
func (h *GRPCGatewayHandler) CreateAgent(c *gin.Context) {
	var req agentv1.CreateAgentRequest
	
	// 绑定JSON到Protobuf消息
	if err := h.bindJSONToProto(c, &req); err != nil {
		h.handleError(c, "Failed to bind request", err)
		return
	}
	
	// 调用gRPC服务
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	
	resp, err := h.grpcProxyService.CreateAgent(ctx, &req)
	if err != nil {
		h.handleGRPCError(c, "CreateAgent", err)
		return
	}
	
	// 返回响应
	h.respondWithProto(c, http.StatusOK, resp)
}

// GetAgent 获取智能体
func (h *GRPCGatewayHandler) GetAgent(c *gin.Context) {
	agentID := c.Param("id")
	if agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Missing agent ID",
			"error":   "invalid_request",
		})
		return
	}
	
	req := &agentv1.GetAgentRequest{
		AgentId: agentID,
	}
	
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	
	resp, err := h.grpcProxyService.GetAgent(ctx, req)
	if err != nil {
		h.handleGRPCError(c, "GetAgent", err)
		return
	}
	
	h.respondWithProto(c, http.StatusOK, resp)
}

// ListAgents 获取智能体列表
func (h *GRPCGatewayHandler) ListAgents(c *gin.Context) {
	req := &agentv1.ListAgentsRequest{}
	
	// 解析查询参数
	if ownerID := c.Query("owner_id"); ownerID != "" {
		req.OwnerId = ownerID
	}
	if page := c.Query("page"); page != "" {
		// 简化处理，实际需要错误处理
		req.Page = 1 // 默认值
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		req.PageSize = 20 // 默认值
	}
	
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	
	resp, err := h.grpcProxyService.ListAgents(ctx, req)
	if err != nil {
		h.handleGRPCError(c, "ListAgents", err)
		return
	}
	
	h.respondWithProto(c, http.StatusOK, resp)
}

// UpdateAgent 更新智能体
func (h *GRPCGatewayHandler) UpdateAgent(c *gin.Context) {
	agentID := c.Param("id")
	if agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Missing agent ID",
		})
		return
	}
	
	var req agentv1.UpdateAgentRequest
	if err := h.bindJSONToProto(c, &req); err != nil {
		h.handleError(c, "Failed to bind request", err)
		return
	}
	
	req.AgentId = agentID // 确保ID正确
	
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	
	resp, err := h.grpcProxyService.UpdateAgent(ctx, &req)
	if err != nil {
		h.handleGRPCError(c, "UpdateAgent", err)
		return
	}
	
	h.respondWithProto(c, http.StatusOK, resp)
}

// DeleteAgent 删除智能体
func (h *GRPCGatewayHandler) DeleteAgent(c *gin.Context) {
	agentID := c.Param("id")
	if agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Missing agent ID",
		})
		return
	}
	
	req := &agentv1.DeleteAgentRequest{
		AgentId: agentID,
	}
	
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	
	resp, err := h.grpcProxyService.DeleteAgent(ctx, req)
	if err != nil {
		h.handleGRPCError(c, "DeleteAgent", err)
		return
	}
	
	h.respondWithProto(c, http.StatusOK, resp)
}

// Chat 与智能体对话
func (h *GRPCGatewayHandler) Chat(c *gin.Context) {
	agentID := c.Param("id")
	if agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Missing agent ID",
		})
		return
	}
	
	var req agentv1.ChatRequest
	if err := h.bindJSONToProto(c, &req); err != nil {
		h.handleError(c, "Failed to bind request", err)
		return
	}
	
	req.AgentId = agentID
	
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	
	resp, err := h.grpcProxyService.Chat(ctx, &req)
	if err != nil {
		h.handleGRPCError(c, "Chat", err)
		return
	}
	
	h.respondWithProto(c, http.StatusOK, resp)
}

// LLM相关处理器

// CreateModel 创建模型
func (h *GRPCGatewayHandler) CreateModel(c *gin.Context) {
	var req llmv1.CreateModelRequest
	
	if err := h.bindJSONToProto(c, &req); err != nil {
		h.handleError(c, "Failed to bind request", err)
		return
	}
	
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	
	resp, err := h.grpcProxyService.CreateModel(ctx, &req)
	if err != nil {
		h.handleGRPCError(c, "CreateModel", err)
		return
	}
	
	h.respondWithProto(c, http.StatusOK, resp)
}

// GetModel 获取模型
func (h *GRPCGatewayHandler) GetModel(c *gin.Context) {
	modelID := c.Param("id")
	if modelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Missing model ID",
		})
		return
	}
	
	req := &llmv1.GetModelRequest{
		ModelId: modelID,
	}
	
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	
	resp, err := h.grpcProxyService.GetModel(ctx, req)
	if err != nil {
		h.handleGRPCError(c, "GetModel", err)
		return
	}
	
	h.respondWithProto(c, http.StatusOK, resp)
}

// ProcessRequest 处理请求
func (h *GRPCGatewayHandler) ProcessRequest(c *gin.Context) {
	var req llmv1.ProcessRequestRequest
	
	if err := h.bindJSONToProto(c, &req); err != nil {
		h.handleError(c, "Failed to bind request", err)
		return
	}
	
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()
	
	resp, err := h.grpcProxyService.ProcessRequest(ctx, &req)
	if err != nil {
		h.handleGRPCError(c, "ProcessRequest", err)
		return
	}
	
	h.respondWithProto(c, http.StatusOK, resp)
}

// 工具方法

// bindJSONToProto 将JSON绑定到Protobuf消息
func (h *GRPCGatewayHandler) bindJSONToProto(c *gin.Context, msg proto.Message) error {
	jsonData, err := c.GetRawData()
	if err != nil {
		return err
	}
	
	// 使用protojson进行JSON到Protobuf的转换
	unmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	
	return unmarshaler.Unmarshal(jsonData, msg)
}

// respondWithProto 使用Protobuf消息响应
func (h *GRPCGatewayHandler) respondWithProto(c *gin.Context, statusCode int, msg proto.Message) {
	// 将Protobuf消息转换为JSON
	marshaler := protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: true,
	}
	
	jsonData, err := marshaler.Marshal(msg)
	if err != nil {
		h.logger.Error("Failed to marshal proto message", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Internal server error",
		})
		return
	}
	
	// 解析JSON以便返回结构化响应
	var data interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		h.logger.Error("Failed to unmarshal JSON", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Internal server error",
		})
		return
	}
	
	c.JSON(statusCode, gin.H{
		"success": true,
		"data":    data,
	})
}

// handleError 处理一般错误
func (h *GRPCGatewayHandler) handleError(c *gin.Context, message string, err error) {
	h.logger.Error(message, 
		zap.Error(err),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method))
	
	c.JSON(http.StatusBadRequest, gin.H{
		"success":    false,
		"message":    message,
		"error":      err.Error(),
		"request_id": c.GetString("request_id"),
	})
}

// handleGRPCError 处理gRPC错误
func (h *GRPCGatewayHandler) handleGRPCError(c *gin.Context, method string, err error) {
	h.logger.Error("gRPC call failed",
		zap.String("method", method),
		zap.Error(err),
		zap.String("path", c.Request.URL.Path))
	
	// 根据gRPC错误类型返回相应的HTTP状态码
	// 这里简化处理，实际需要根据grpc.Code判断
	c.JSON(http.StatusInternalServerError, gin.H{
		"success":    false,
		"message":    "Service call failed",
		"error":      err.Error(),
		"method":     method,
		"request_id": c.GetString("request_id"),
	})
}
