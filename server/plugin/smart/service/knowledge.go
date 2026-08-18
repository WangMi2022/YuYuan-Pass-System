package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	smartModel "github.com/WangMi2022/mit-assets-admin/server/plugin/smart/model"
	"gorm.io/gorm"
)

const (
	knowledgeChunkRunes   = 800
	knowledgeChunkOverlap = 120
	knowledgeMaxTextBytes = 2 << 20
)

type KnowledgeSourceInput struct {
	SourceType string
	SourceID   string
	Title      string
	Content    string
}

type KnowledgeSearchResult struct {
	ChunkID    uint    `json:"chunkId"`
	SourceType string  `json:"sourceType"`
	SourceID   string  `json:"sourceId"`
	Title      string  `json:"title"`
	Snippet    string  `json:"snippet"`
	Score      float64 `json:"score"`
}

type KnowledgeStore struct{}

var Knowledge = new(KnowledgeStore)

func (s *KnowledgeStore) ReplaceSource(ctx context.Context, actor AssistantActor, input KnowledgeSourceInput) error {
	if actor.UserID == 0 || actor.AuthorityID == 0 {
		return errors.New("知识来源必须绑定已认证用户和角色")
	}
	input.SourceType = strings.TrimSpace(input.SourceType)
	input.SourceID = strings.TrimSpace(input.SourceID)
	input.Title = strings.TrimSpace(input.Title)
	input.Content = normalizeKnowledgeText(input.Content)
	if input.SourceType == "" || input.SourceID == "" || input.Title == "" || input.Content == "" {
		return errors.New("知识来源类型、ID、标题和内容不能为空")
	}
	if len([]byte(input.Content)) > knowledgeMaxTextBytes {
		return errors.New("单个知识来源正文不能超过 2MB")
	}
	chunks := splitKnowledgeText(input.Content, knowledgeChunkRunes, knowledgeChunkOverlap)
	if len(chunks) == 0 {
		return errors.New("知识来源没有可索引正文")
	}
	sourceHash := sha256.Sum256([]byte(input.Content))
	sourceVersion := hex.EncodeToString(sourceHash[:])
	rows := make([]smartModel.KnowledgeChunk, 0, len(chunks))
	for index, content := range chunks {
		contentHash := sha256.Sum256([]byte(content))
		rows = append(rows, smartModel.KnowledgeChunk{
			TenantID: actor.TenantID, DepartmentID: actor.DepartmentID,
			OwnerUserID: actor.UserID, AuthorityID: actor.AuthorityID,
			Visibility: smartModel.KnowledgeVisibilityPrivate,
			SourceType: input.SourceType, SourceID: input.SourceID, SourceVersion: sourceVersion,
			ChunkIndex: index, Title: input.Title, Content: content,
			ContentHash: hex.EncodeToString(contentHash[:]),
		})
	}
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where(
			"tenant_id = ? AND department_id = ? AND owner_user_id = ? AND authority_id = ? AND source_type = ? AND source_id = ?",
			actor.TenantID, actor.DepartmentID, actor.UserID, actor.AuthorityID, input.SourceType, input.SourceID,
		).Unscoped().Delete(&smartModel.KnowledgeChunk{}).Error; err != nil {
			return err
		}
		return tx.Create(&rows).Error
	})
}

func (s *KnowledgeStore) DeleteSource(ctx context.Context, actor AssistantActor, sourceType, sourceID string) error {
	if actor.UserID == 0 || actor.AuthorityID == 0 {
		return errors.New("知识来源必须绑定已认证用户和角色")
	}
	return global.GVA_DB.WithContext(ctx).Where(
		"tenant_id = ? AND department_id = ? AND owner_user_id = ? AND authority_id = ? AND source_type = ? AND source_id = ?",
		actor.TenantID, actor.DepartmentID, actor.UserID, actor.AuthorityID, strings.TrimSpace(sourceType), strings.TrimSpace(sourceID),
	).Delete(&smartModel.KnowledgeChunk{}).Error
}

type knowledgeRow struct {
	ID         uint
	SourceType string
	SourceID   string
	Title      string
	Content    string
	Score      float64
}

func (s *KnowledgeStore) Search(ctx context.Context, actor AssistantActor, query string, limit int) ([]KnowledgeSearchResult, error) {
	if actor.UserID == 0 || actor.AuthorityID == 0 {
		return nil, errors.New("知识检索必须绑定已认证用户和角色")
	}
	query = normalizeKnowledgeQuery(query)
	if query == "" {
		return []KnowledgeSearchResult{}, nil
	}
	if limit <= 0 || limit > 10 {
		limit = 5
	}
	pattern := "%" + strings.ToLower(query) + "%"
	db := global.GVA_DB.WithContext(ctx).Model(&smartModel.KnowledgeChunk{}).
		Where("tenant_id = ? AND department_id = ? AND owner_user_id = ? AND authority_id = ? AND visibility = ?",
			actor.TenantID, actor.DepartmentID, actor.UserID, actor.AuthorityID, smartModel.KnowledgeVisibilityPrivate)
	var rows []knowledgeRow
	if global.GVA_DB.Dialector.Name() == "postgres" {
		vector := "to_tsvector('simple', COALESCE(title, '') || ' ' || COALESCE(content, ''))"
		tsquery := "plainto_tsquery('simple', ?)"
		selectSQL := fmt.Sprintf("id, source_type, source_id, title, content, CASE WHEN LOWER(title) LIKE ? THEN 2 ELSE ts_rank(%s, %s) END AS score", vector, tsquery)
		whereSQL := fmt.Sprintf("(%s @@ %s OR LOWER(title) LIKE ? OR LOWER(content) LIKE ?)", vector, tsquery)
		db = db.Select(selectSQL, pattern, query).Where(whereSQL, query, pattern, pattern).Order("score DESC, id ASC")
	} else {
		db = db.Select("id, source_type, source_id, title, content, CASE WHEN LOWER(title) LIKE ? THEN 2 ELSE 1 END AS score", pattern).
			Where("LOWER(title) LIKE ? OR LOWER(content) LIKE ?", pattern, pattern).
			Order("score DESC, id ASC")
	}
	if err := db.Limit(limit).Scan(&rows).Error; err != nil {
		return nil, err
	}
	results := make([]KnowledgeSearchResult, 0, len(rows))
	for _, row := range rows {
		results = append(results, KnowledgeSearchResult{
			ChunkID: row.ID, SourceType: row.SourceType, SourceID: row.SourceID,
			Title: row.Title, Snippet: knowledgeSnippet(row.Content, query, 220), Score: row.Score,
		})
	}
	return results, nil
}

func normalizeKnowledgeQuery(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	for _, token := range []string{"请问", "帮我", "查询", "查找", "搜索", "知识库里的", "知识库中", "知识库", "文档里的", "文档中", "文档里", "制度里的", "制度里", "里的", "关于", "相关", "是什么", "怎么规定", "如何规定", "？", "?"} {
		value = strings.ReplaceAll(value, token, " ")
	}
	return strings.Join(strings.Fields(value), " ")
}

func normalizeKnowledgeText(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	lines := strings.Split(value, "\n")
	normalized := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			normalized = append(normalized, line)
		}
	}
	return strings.Join(normalized, "\n")
}

func splitKnowledgeText(value string, chunkSize, overlap int) []string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) == 0 || chunkSize <= 0 {
		return nil
	}
	if overlap < 0 || overlap >= chunkSize {
		overlap = 0
	}
	chunks := make([]string, 0, (len(runes)+chunkSize-1)/chunkSize)
	for start := 0; start < len(runes); {
		end := start + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		if end < len(runes) {
			for candidate := end; candidate > start+chunkSize/2; candidate-- {
				if unicode.IsSpace(runes[candidate-1]) || strings.ContainsRune("。！？；\n", runes[candidate-1]) {
					end = candidate
					break
				}
			}
		}
		chunk := strings.TrimSpace(string(runes[start:end]))
		if chunk != "" && utf8.ValidString(chunk) {
			chunks = append(chunks, chunk)
		}
		if end == len(runes) {
			break
		}
		start = end - overlap
	}
	return chunks
}

func knowledgeSnippet(content, query string, maxRunes int) string {
	runes := []rune(content)
	if len(runes) <= maxRunes {
		return content
	}
	lowerContent := strings.ToLower(content)
	byteIndex := strings.Index(lowerContent, strings.ToLower(query))
	runeIndex := 0
	if byteIndex > 0 {
		runeIndex = utf8.RuneCountInString(content[:byteIndex])
	}
	start := runeIndex - maxRunes/3
	if start < 0 {
		start = 0
	}
	end := start + maxRunes
	if end > len(runes) {
		end = len(runes)
		start = end - maxRunes
	}
	prefix, suffix := "", ""
	if start > 0 {
		prefix = "…"
	}
	if end < len(runes) {
		suffix = "…"
	}
	return prefix + string(runes[start:end]) + suffix
}
