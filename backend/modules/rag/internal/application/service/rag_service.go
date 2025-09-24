package service

import (
	"context"
	"time"

	"github.com/noah-loop/backend/modules/rag/internal/domain"
	"github.com/noah-loop/backend/modules/rag/internal/domain/repository"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"go.uber.org/zap"
)

// RAGService RAG应用服务
type RAGService struct {
	kbRepo       repository.KnowledgeBaseRepository
	docRepo      repository.DocumentRepository
	chunkRepo    repository.ChunkRepository
	vectorRepo   repository.VectorRepository
	embeddingService EmbeddingService
	chunkingService  ChunkingService
	logger       infrastructure.Logger
}

// NewRAGService 创建RAG服务
func NewRAGService(
	kbRepo repository.KnowledgeBaseRepository,
	docRepo repository.DocumentRepository,
	chunkRepo repository.ChunkRepository,
	vectorRepo repository.VectorRepository,
	embeddingService EmbeddingService,
	chunkingService ChunkingService,
	logger infrastructure.Logger,
) *RAGService {
	return &RAGService{
		kbRepo:           kbRepo,
		docRepo:          docRepo,
		chunkRepo:        chunkRepo,
		vectorRepo:       vectorRepo,
		embeddingService: embeddingService,
		chunkingService:  chunkingService,
		logger:          logger,
	}
}

// CreateKnowledgeBase 创建知识库
func (s *RAGService) CreateKnowledgeBase(ctx context.Context, cmd *CreateKnowledgeBaseCommand) (*domain.KnowledgeBase, error) {
	s.logger.Info("Creating knowledge base",
		zap.String("name", cmd.Name),
		zap.String("owner_id", cmd.OwnerID))

	// 检查知识库名称是否已存在
	existing, err := s.kbRepo.FindByName(ctx, cmd.Name, cmd.OwnerID)
	if err == nil && existing != nil {
		return nil, domain.NewDomainError("KNOWLEDGE_BASE_EXISTS", "knowledge base name already exists")
	}

	// 创建知识库
	kb, err := domain.NewKnowledgeBase(cmd.Name, cmd.Description, cmd.OwnerID)
	if err != nil {
		return nil, err
	}

	// 设置自定义设置
	if cmd.Settings != nil {
		err = kb.UpdateSettings(*cmd.Settings)
		if err != nil {
			return nil, err
		}
	}

	// 保存知识库
	err = s.kbRepo.Save(ctx, kb)
	if err != nil {
		s.logger.Error("Failed to save knowledge base", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Knowledge base created successfully", zap.String("id", kb.ID))
	return kb, nil
}

// UpdateKnowledgeBase 更新知识库
func (s *RAGService) UpdateKnowledgeBase(ctx context.Context, cmd *UpdateKnowledgeBaseCommand) (*domain.KnowledgeBase, error) {
	kb, err := s.kbRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	if kb == nil {
		return nil, domain.ErrKnowledgeBaseNotFoundf(cmd.ID)
	}

	// 更新基本信息
	if cmd.Name != "" {
		kb.Name = cmd.Name
	}
	if cmd.Description != "" {
		kb.Description = cmd.Description
	}

	// 更新设置
	if cmd.Settings != nil {
		err = kb.UpdateSettings(*cmd.Settings)
		if err != nil {
			return nil, err
		}
	}

	// 更新状态
	if cmd.Status != "" {
		err = kb.UpdateStatus(cmd.Status)
		if err != nil {
			return nil, err
		}
	}

	// 保存更新
	err = s.kbRepo.Update(ctx, kb)
	if err != nil {
		s.logger.Error("Failed to update knowledge base", zap.Error(err))
		return nil, err
	}

	return kb, nil
}

// AddDocument 添加文档到知识库
func (s *RAGService) AddDocument(ctx context.Context, cmd *AddDocumentCommand) (*domain.Document, error) {
	s.logger.Info("Adding document to knowledge base",
		zap.String("title", cmd.Title),
		zap.String("knowledge_base_id", cmd.KnowledgeBaseID))

	// 检查知识库是否存在
	kb, err := s.kbRepo.FindByID(ctx, cmd.KnowledgeBaseID)
	if err != nil {
		return nil, err
	}
	if kb == nil {
		return nil, domain.ErrKnowledgeBaseNotFoundf(cmd.KnowledgeBaseID)
	}

	// 创建文档
	doc, err := domain.NewDocument(cmd.Title, cmd.Content, cmd.Type, cmd.Source)
	if err != nil {
		return nil, err
	}

	doc.KnowledgeBaseID = cmd.KnowledgeBaseID
	if cmd.Language != "" {
		doc.Language = cmd.Language
	}

	// 设置元数据
	if cmd.Metadata != nil {
		doc.Metadata = *cmd.Metadata
	}

	// 保存文档
	err = s.docRepo.Save(ctx, doc)
	if err != nil {
		s.logger.Error("Failed to save document", zap.Error(err))
		return nil, err
	}

	// 异步处理文档索引
	go s.processDocumentAsync(context.Background(), doc.ID)

	s.logger.Info("Document added successfully", zap.String("id", doc.ID))
	return doc, nil
}

// ProcessDocument 处理文档（分块和向量化）
func (s *RAGService) ProcessDocument(ctx context.Context, documentID string) error {
	s.logger.Info("Processing document", zap.String("document_id", documentID))

	// 获取文档
	doc, err := s.docRepo.FindByID(ctx, documentID)
	if err != nil {
		return err
	}
	if doc == nil {
		return domain.ErrDocumentNotFoundf(documentID)
	}

	// 更新状态为索引中
	err = doc.UpdateStatus(domain.DocumentStatusIndexing)
	if err != nil {
		return err
	}
	err = s.docRepo.Update(ctx, doc)
	if err != nil {
		return err
	}

	// 分块处理
	chunks, err := s.chunkingService.ChunkDocument(ctx, doc)
	if err != nil {
		s.logger.Error("Failed to chunk document", zap.Error(err))
		doc.UpdateStatus(domain.DocumentStatusFailed)
		s.docRepo.Update(ctx, doc)
		return err
	}

	// 保存分块
	err = s.chunkRepo.SaveBatch(ctx, chunks)
	if err != nil {
		s.logger.Error("Failed to save chunks", zap.Error(err))
		doc.UpdateStatus(domain.DocumentStatusFailed)
		s.docRepo.Update(ctx, doc)
		return err
	}

	// 生成向量嵌入
	err = s.generateEmbeddings(ctx, chunks)
	if err != nil {
		s.logger.Error("Failed to generate embeddings", zap.Error(err))
		doc.UpdateStatus(domain.DocumentStatusFailed)
		s.docRepo.Update(ctx, doc)
		return err
	}

	// 标记为已索引
	err = doc.MarkAsIndexed(chunks)
	if err != nil {
		return err
	}
	err = s.docRepo.Update(ctx, doc)
	if err != nil {
		return err
	}

	s.logger.Info("Document processed successfully", zap.String("document_id", documentID))
	return nil
}

// Search 搜索相关内容
func (s *RAGService) Search(ctx context.Context, query *domain.SearchQuery) (*domain.SearchResults, error) {
	s.logger.Info("Searching knowledge base",
		zap.String("query", query.Query),
		zap.String("knowledge_base_id", query.KnowledgeBaseID))

	start := time.Now()

	// 检查知识库
	kb, err := s.kbRepo.FindByID(ctx, query.KnowledgeBaseID)
	if err != nil {
		return nil, err
	}
	if kb == nil {
		return nil, domain.ErrKnowledgeBaseNotFoundf(query.KnowledgeBaseID)
	}
	if !kb.CanBeQueried() {
		return nil, domain.NewDomainError("KNOWLEDGE_BASE_NOT_QUERYABLE", "knowledge base cannot be queried")
	}

	// 生成查询向量
	queryVector, err := s.embeddingService.GenerateEmbedding(ctx, query.Query)
	if err != nil {
		s.logger.Error("Failed to generate query embedding", zap.Error(err))
		return nil, err
	}

	// 构建向量查询
	vectorQuery := repository.NewVectorQuery(
		s.getIndexName(query.KnowledgeBaseID),
		queryVector,
		query.TopK,
	).WithScoreThreshold(query.ScoreThreshold)

	// 添加过滤条件
	if len(query.Filters.DocumentTypes) > 0 {
		vectorQuery.WithFilter("document_type", query.Filters.DocumentTypes[0])
	}

	// 执行向量搜索
	vectorResult, err := s.vectorRepo.Search(ctx, vectorQuery)
	if err != nil {
		s.logger.Error("Failed to search vectors", zap.Error(err))
		return nil, err
	}

	// 转换搜索结果
	results := domain.NewSearchResults(*query)
	for _, match := range vectorResult.Results {
		chunk, err := s.chunkRepo.FindByID(ctx, match.ID)
		if err != nil {
			continue
		}

		result := domain.NewSearchResult(
			chunk.ID,
			chunk.Content,
			chunk.Metadata.Title,
			match.Metadata["source"],
			match.Score,
			domain.SearchResultTypeChunk,
		)

		// 设置分块信息
		result.SetChunkInfo(&domain.ChunkInfo{
			Position:   chunk.Position,
			StartIndex: chunk.StartIndex,
			EndIndex:   chunk.EndIndex,
			TokenCount: chunk.TokenCount,
			ChunkType:  string(chunk.Type),
		})

		results.AddResult(*result)
	}

	// 过滤低分结果
	results.FilterByScore(query.ScoreThreshold)

	// 记录查询统计
	avgScore := float32(0)
	if len(results.Results) > 0 {
		totalScore := float32(0)
		for _, result := range results.Results {
			totalScore += result.Score
		}
		avgScore = totalScore / float32(len(results.Results))
	}
	kb.RecordQuery(avgScore)
	s.kbRepo.Update(ctx, kb)

	results.Duration = time.Since(start)
	s.logger.Info("Search completed",
		zap.Int("result_count", len(results.Results)),
		zap.Duration("duration", results.Duration))

	return results, nil
}

// DeleteDocument 删除文档
func (s *RAGService) DeleteDocument(ctx context.Context, documentID string) error {
	doc, err := s.docRepo.FindByID(ctx, documentID)
	if err != nil {
		return err
	}
	if doc == nil {
		return domain.ErrDocumentNotFoundf(documentID)
	}

	// 删除向量索引
	chunks, err := s.chunkRepo.FindByDocumentID(ctx, doc.ID)
	if err == nil {
		chunkIDs := make([]string, len(chunks))
		for i, chunk := range chunks {
			chunkIDs[i] = chunk.ID
		}
		indexName := s.getIndexName(doc.KnowledgeBaseID)
		s.vectorRepo.Delete(ctx, indexName, chunkIDs)
	}

	// 删除分块
	err = s.chunkRepo.DeleteByDocumentID(ctx, doc.ID)
	if err != nil {
		s.logger.Error("Failed to delete chunks", zap.Error(err))
	}

	// 删除文档
	err = s.docRepo.Delete(ctx, documentID)
	if err != nil {
		s.logger.Error("Failed to delete document", zap.Error(err))
		return err
	}

	return nil
}

// processDocumentAsync 异步处理文档
func (s *RAGService) processDocumentAsync(ctx context.Context, documentID string) {
	err := s.ProcessDocument(ctx, documentID)
	if err != nil {
		s.logger.Error("Failed to process document asynchronously",
			zap.String("document_id", documentID),
			zap.Error(err))
	}
}

// generateEmbeddings 生成向量嵌入
func (s *RAGService) generateEmbeddings(ctx context.Context, chunks []*domain.Chunk) error {
	indexName := ""
	if len(chunks) > 0 {
		doc, err := s.docRepo.FindByID(ctx, chunks[0].DocumentID)
		if err != nil {
			return err
		}
		indexName = s.getIndexName(doc.KnowledgeBaseID)
	}

	// 批量生成嵌入
	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = chunk.Content
	}

	embeddings, err := s.embeddingService.GenerateEmbeddings(ctx, texts)
	if err != nil {
		return err
	}

	// 更新分块的嵌入向量
	vectorRecords := make([]repository.VectorRecord, len(chunks))
	for i, chunk := range chunks {
		err = chunk.SetEmbedding(embeddings[i])
		if err != nil {
			return err
		}

		vectorRecords[i] = repository.VectorRecord{
			ID:     chunk.ID,
			Vector: embeddings[i],
			Metadata: map[string]string{
				"document_id": chunk.DocumentID,
				"chunk_type":  string(chunk.Type),
				"position":    string(rune(chunk.Position)),
			},
		}
	}

	// 保存向量到向量数据库
	err = s.vectorRepo.Insert(ctx, indexName, vectorRecords)
	if err != nil {
		return err
	}

	// 更新分块
	return s.chunkRepo.UpdateBatch(ctx, chunks)
}

// getIndexName 获取索引名称
func (s *RAGService) getIndexName(knowledgeBaseID string) string {
	return "kb_" + knowledgeBaseID
}
