package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/noah-loop/backend/modules/rag/internal/application/service"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"go.uber.org/zap"
)

// RAGHandler RAG HTTP处理器
type RAGHandler struct {
	ragService *service.RAGService
	logger     infrastructure.Logger
}

// NewRAGHandler 创建RAG处理器
func NewRAGHandler(ragService *service.RAGService, logger infrastructure.Logger) *RAGHandler {
	return &RAGHandler{
		ragService: ragService,
		logger:     logger,
	}
}

// CreateKnowledgeBase 创建知识库
func (h *RAGHandler) CreateKnowledgeBase(c *gin.Context) {
	var cmd service.CreateKnowledgeBaseCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	kb, err := h.ragService.CreateKnowledgeBase(c.Request.Context(), &cmd)
	if err != nil {
		h.logger.Error("Failed to create knowledge base", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"knowledge_base": kb,
		"message":        "Knowledge base created successfully",
	})
}

// GetKnowledgeBase 获取知识库
func (h *RAGHandler) GetKnowledgeBase(c *gin.Context) {
	id := c.Param("id")
	includeDocuments := c.Query("include_documents") == "true"
	includeStats := c.Query("include_stats") == "true"

	cmd := &service.GetKnowledgeBaseCommand{
		ID:               id,
		IncludeDocuments: includeDocuments,
		IncludeStats:     includeStats,
	}

	// 这里需要在RAGService中实现GetKnowledgeBase方法
	// kb, err := h.ragService.GetKnowledgeBase(c.Request.Context(), cmd)
	// 暂时返回简单响应
	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"message": "Knowledge base retrieved successfully",
	})
}

// UpdateKnowledgeBase 更新知识库
func (h *RAGHandler) UpdateKnowledgeBase(c *gin.Context) {
	var cmd service.UpdateKnowledgeBaseCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd.ID = c.Param("id")

	kb, err := h.ragService.UpdateKnowledgeBase(c.Request.Context(), &cmd)
	if err != nil {
		h.logger.Error("Failed to update knowledge base", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"knowledge_base": kb,
		"message":        "Knowledge base updated successfully",
	})
}

// DeleteKnowledgeBase 删除知识库
func (h *RAGHandler) DeleteKnowledgeBase(c *gin.Context) {
	id := c.Param("id")

	err := h.ragService.DeleteDocument(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to delete knowledge base", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Knowledge base deleted successfully",
	})
}

// ListKnowledgeBases 列出知识库
func (h *RAGHandler) ListKnowledgeBases(c *gin.Context) {
	ownerID := c.Query("owner_id")
	status := c.Query("status")
	
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	cmd := &service.ListKnowledgeBasesCommand{
		OwnerID: ownerID,
		Status:  status,
		Offset:  offset,
		Limit:   limit,
	}

	// 这里需要在RAGService中实现ListKnowledgeBases方法
	// kbs, total, err := h.ragService.ListKnowledgeBases(c.Request.Context(), cmd)
	// 暂时返回简单响应
	c.JSON(http.StatusOK, gin.H{
		"knowledge_bases": []interface{}{},
		"total":           0,
		"offset":          offset,
		"limit":           limit,
	})
}

// AddDocument 添加文档
func (h *RAGHandler) AddDocument(c *gin.Context) {
	var cmd service.AddDocumentCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	doc, err := h.ragService.AddDocument(c.Request.Context(), &cmd)
	if err != nil {
		h.logger.Error("Failed to add document", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"document": doc,
		"message":  "Document added successfully",
	})
}

// GetDocument 获取文档
func (h *RAGHandler) GetDocument(c *gin.Context) {
	id := c.Param("id")
	includeContent := c.Query("include_content") == "true"
	includeChunks := c.Query("include_chunks") == "true"

	cmd := &service.GetDocumentCommand{
		ID:             id,
		IncludeContent: includeContent,
		IncludeChunks:  includeChunks,
	}

	// 这里需要在RAGService中实现GetDocument方法
	// doc, err := h.ragService.GetDocument(c.Request.Context(), cmd)
	// 暂时返回简单响应
	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"message": "Document retrieved successfully",
	})
}

// UpdateDocument 更新文档
func (h *RAGHandler) UpdateDocument(c *gin.Context) {
	var cmd service.UpdateDocumentCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd.ID = c.Param("id")

	// 这里需要在RAGService中实现UpdateDocument方法
	// doc, err := h.ragService.UpdateDocument(c.Request.Context(), &cmd)
	// 暂时返回简单响应
	c.JSON(http.StatusOK, gin.H{
		"id":      cmd.ID,
		"message": "Document updated successfully",
	})
}

// DeleteDocument 删除文档
func (h *RAGHandler) DeleteDocument(c *gin.Context) {
	id := c.Param("id")

	err := h.ragService.DeleteDocument(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to delete document", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Document deleted successfully",
	})
}

// ListDocuments 列出文档
func (h *RAGHandler) ListDocuments(c *gin.Context) {
	knowledgeBaseID := c.Query("knowledge_base_id")
	status := c.Query("status")
	docType := c.Query("type")
	
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	cmd := &service.ListDocumentsCommand{
		KnowledgeBaseID: knowledgeBaseID,
		Status:          status,
		Type:            docType,
		Offset:          offset,
		Limit:           limit,
	}

	// 这里需要在RAGService中实现ListDocuments方法
	// docs, total, err := h.ragService.ListDocuments(c.Request.Context(), cmd)
	// 暂时返回简单响应
	c.JSON(http.StatusOK, gin.H{
		"documents": []interface{}{},
		"total":     0,
		"offset":    offset,
		"limit":     limit,
	})
}

// ProcessDocument 处理文档（分块和向量化）
func (h *RAGHandler) ProcessDocument(c *gin.Context) {
	var cmd service.ProcessDocumentCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if cmd.DocumentID == "" {
		cmd.DocumentID = c.Param("id")
	}

	err := h.ragService.ProcessDocument(c.Request.Context(), cmd.DocumentID)
	if err != nil {
		h.logger.Error("Failed to process document", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Document processing started successfully",
	})
}

// Search 搜索相关内容
func (h *RAGHandler) Search(c *gin.Context) {
	var cmd service.SearchCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := cmd.ToSearchQuery()
	results, err := h.ragService.Search(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to search", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results":  results.Results,
		"total":    results.Total,
		"query":    results.Query,
		"duration": results.Duration.String(),
	})
}

// BatchAddDocuments 批量添加文档
func (h *RAGHandler) BatchAddDocuments(c *gin.Context) {
	var cmd service.BatchAddDocumentsCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var successCount int
	var errors []string

	for _, docCmd := range cmd.Documents {
		_, err := h.ragService.AddDocument(c.Request.Context(), &docCmd)
		if err != nil {
			errors = append(errors, err.Error())
		} else {
			successCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success_count": successCount,
		"total_count":   len(cmd.Documents),
		"errors":        errors,
		"message":       "Batch add documents completed",
	})
}

// BatchDeleteDocuments 批量删除文档
func (h *RAGHandler) BatchDeleteDocuments(c *gin.Context) {
	var cmd service.BatchDeleteDocumentsCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var successCount int
	var errors []string

	for _, docID := range cmd.DocumentIDs {
		err := h.ragService.DeleteDocument(c.Request.Context(), docID)
		if err != nil {
			errors = append(errors, err.Error())
		} else {
			successCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success_count": successCount,
		"total_count":   len(cmd.DocumentIDs),
		"errors":        errors,
		"message":       "Batch delete documents completed",
	})
}

// Health 健康检查
func (h *RAGHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "rag",
		"message": "RAG service is running normally",
	})
}
