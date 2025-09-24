package domain

import "fmt"

// DomainError RAG领域错误
type DomainError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *DomainError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewDomainError 创建领域错误
func NewDomainError(code, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}

// NewDomainErrorWithDetails 创建带详情的领域错误
func NewDomainErrorWithDetails(code, message, details string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// 预定义错误代码
const (
	// 文档相关错误
	ErrDocumentNotFound         = "DOCUMENT_NOT_FOUND"
	ErrDocumentAlreadyExists    = "DOCUMENT_ALREADY_EXISTS"
	ErrDocumentInvalidContent   = "DOCUMENT_INVALID_CONTENT"
	ErrDocumentIndexingFailed   = "DOCUMENT_INDEXING_FAILED"
	ErrDocumentProcessingFailed = "DOCUMENT_PROCESSING_FAILED"

	// 知识库相关错误
	ErrKnowledgeBaseNotFound     = "KNOWLEDGE_BASE_NOT_FOUND"
	ErrKnowledgeBaseInactive     = "KNOWLEDGE_BASE_INACTIVE"
	ErrKnowledgeBaseMaxDocuments = "KNOWLEDGE_BASE_MAX_DOCUMENTS"
	ErrKnowledgeBaseDeleted      = "KNOWLEDGE_BASE_DELETED"

	// 分块相关错误
	ErrChunkNotFound       = "CHUNK_NOT_FOUND"
	ErrChunkTooLarge       = "CHUNK_TOO_LARGE"
	ErrChunkInvalidContent = "CHUNK_INVALID_CONTENT"
	ErrEmbeddingFailed     = "EMBEDDING_FAILED"

	// 搜索相关错误
	ErrSearchFailed        = "SEARCH_FAILED"
	ErrInvalidQuery        = "INVALID_QUERY"
	ErrSearchTimeout       = "SEARCH_TIMEOUT"
	ErrNoResults           = "NO_RESULTS"

	// 向量相关错误
	ErrVectorDimensionMismatch = "VECTOR_DIMENSION_MISMATCH"
	ErrVectorIndexNotFound     = "VECTOR_INDEX_NOT_FOUND"
	ErrVectorStoreFailed       = "VECTOR_STORE_FAILED"

	// 标签相关错误
	ErrTagNotFound      = "TAG_NOT_FOUND"
	ErrTagAlreadyExists = "TAG_ALREADY_EXISTS"
	ErrTagInUse         = "TAG_IN_USE"

	// 通用错误
	ErrInvalidInput    = "INVALID_INPUT"
	ErrPermissionDenied = "PERMISSION_DENIED"
	ErrResourceLocked  = "RESOURCE_LOCKED"
	ErrTimeout         = "TIMEOUT"
	ErrInternalError   = "INTERNAL_ERROR"
)

// 常用错误创建函数
func ErrDocumentNotFoundf(documentID string) *DomainError {
	return NewDomainErrorWithDetails(ErrDocumentNotFound, "Document not found", fmt.Sprintf("document_id: %s", documentID))
}

func ErrKnowledgeBaseNotFoundf(kbID string) *DomainError {
	return NewDomainErrorWithDetails(ErrKnowledgeBaseNotFound, "Knowledge base not found", fmt.Sprintf("knowledge_base_id: %s", kbID))
}

func ErrChunkNotFoundf(chunkID string) *DomainError {
	return NewDomainErrorWithDetails(ErrChunkNotFound, "Chunk not found", fmt.Sprintf("chunk_id: %s", chunkID))
}

func ErrTagNotFoundf(tagName string) *DomainError {
	return NewDomainErrorWithDetails(ErrTagNotFound, "Tag not found", fmt.Sprintf("tag_name: %s", tagName))
}

func ErrInvalidInputf(field, reason string) *DomainError {
	return NewDomainErrorWithDetails(ErrInvalidInput, "Invalid input", fmt.Sprintf("field: %s, reason: %s", field, reason))
}
