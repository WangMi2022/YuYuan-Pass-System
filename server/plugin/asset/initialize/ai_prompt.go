package initialize

import (
	"context"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/ai"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const assetRecognitionPromptKey = "asset-draft"

const assetRecognitionOutputSchemaV1 = `{
  "type": "object",
  "additionalProperties": false,
  "required": ["name", "brand", "model", "serialNumber", "specifications", "productionDate", "recommendedCategoryCode", "recommendedUnit", "recommendedWarrantyMonths", "rawText", "fieldConfidences"],
  "properties": {
    "name": {"type": "string", "maxLength": 150},
    "brand": {"type": "string", "maxLength": 100},
    "model": {"type": "string", "maxLength": 120},
    "serialNumber": {"type": "string", "maxLength": 120},
    "specifications": {"type": "string", "maxLength": 1000},
    "productionDate": {"type": "string", "maxLength": 30},
    "recommendedCategoryCode": {"type": "string", "maxLength": 50},
    "recommendedUnit": {"type": "string", "maxLength": 30},
    "recommendedWarrantyMonths": {"type": "integer", "minimum": 0, "maximum": 120},
    "rawText": {"type": "string", "maxLength": 10000},
    "fieldConfidences": {
      "type": "object",
      "maxProperties": 32,
      "additionalProperties": {"type": "number", "minimum": 0, "maximum": 1}
    }
  }
}`

const assetRecognitionOutputSchemaV2 = `{
  "type": "object",
  "additionalProperties": false,
  "minProperties": 1,
  "properties": {
    "name": {"type": "string", "maxLength": 150},
    "brand": {"type": "string", "maxLength": 100},
    "model": {"type": "string", "maxLength": 120},
    "serialNumber": {"type": "string", "maxLength": 120},
    "specifications": {"type": "string", "maxLength": 1000},
    "productionDate": {"type": "string", "maxLength": 30},
    "recommendedCategoryCode": {"type": "string", "maxLength": 50},
    "recommendedUnit": {"type": "string", "maxLength": 30},
    "recommendedWarrantyMonths": {"type": "integer", "minimum": 0, "maximum": 120},
    "rawText": {"type": "string", "maxLength": 10000},
    "fieldConfidences": {
      "type": "object",
      "maxProperties": 32,
      "additionalProperties": {"type": "number", "minimum": 0, "maximum": 1}
    }
  }
}`

const assetRecognitionOutputSchema = `{
  "type": "object",
  "additionalProperties": false,
  "minProperties": 1,
  "properties": {
    "name": {"type": "string", "maxLength": 150},
    "productName": {"type": "string", "maxLength": 150},
    "brand": {"type": "string", "maxLength": 100},
    "manufacturer": {"type": "string", "maxLength": 100},
    "model": {"type": "string", "maxLength": 120},
    "serialNumber": {"type": "string", "maxLength": 120},
    "specifications": {"type": "string", "maxLength": 1000},
    "productionDate": {"type": "string", "maxLength": 30},
    "recommendedCategoryCode": {"type": "string", "maxLength": 50},
    "recommendedUnit": {"type": "string", "maxLength": 30},
    "recommendedWarrantyMonths": {"type": "integer", "minimum": 0, "maximum": 120},
    "warrantyMonths": {"type": "integer", "minimum": 0, "maximum": 120},
    "rawText": {"type": "string", "maxLength": 10000},
    "fieldConfidences": {
      "type": "object",
      "maxProperties": 32,
      "additionalProperties": {"type": "number", "minimum": 0, "maximum": 1}
    }
  }
}`

const assetRecognitionPrompt = `你是企业资产铭牌识别助手。请只根据图片中可见内容和本次提供的分类选项提取资产草稿。
要求：
1. 只输出符合 JSON Schema 的 JSON，不要输出 Markdown 或解释文字。
2. 优先使用 name、brand、recommendedWarrantyMonths 字段；兼容字段 productName、manufacturer、warrantyMonths 仅在无法按标准字段输出时使用。无法确认的字符串返回空字符串，无法确认的建议质保月数返回 0，不得编造。部分字段缺失时服务端按空值处理并交给人工补充。
3. productionDate 使用 YYYY-MM-DD；只有年月时可使用该月第一天；完全无法确认则返回空字符串。
4. recommendedCategoryCode 必须从本次提供的 categories.code 中选择，无法匹配则返回空字符串。
5. fieldConfidences 使用 0 到 1，至少覆盖所有非空识别字段。
6. rawText 保留铭牌上与型号、序列号、规格、生产日期和厂商相关的原始文字。
7. 不生成资产编号，不推测价格、供应商、购置日期或当前估值。`

func seedAssetRecognitionPrompt(ctx context.Context) {
	if err := ai.ValidateOutputSchema(assetRecognitionOutputSchema); err != nil {
		global.GVA_LOG.Error("资产智能建档 Prompt Schema 不合法", zap.Error(err))
		return
	}
	if err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var templates []ai.PromptTemplate
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("prompt_key = ?", assetRecognitionPromptKey).
			Order("version ASC").
			Find(&templates).Error; err != nil {
			return err
		}
		if len(templates) == 0 {
			now := time.Now()
			return tx.Create(&ai.PromptTemplate{
				PromptKey: assetRecognitionPromptKey, Version: 1, Content: assetRecognitionPrompt,
				OutputSchema: assetRecognitionOutputSchema, Status: ai.PromptStatusActive, ActivatedAt: &now,
			}).Error
		}

		activeIndex := -1
		for index := range templates {
			if templates[index].Status != ai.PromptStatusActive {
				continue
			}
			if activeIndex >= 0 {
				return nil
			}
			activeIndex = index
		}
		if activeIndex < 0 {
			return nil
		}
		activeSchema := strings.TrimSpace(templates[activeIndex].OutputSchema)
		if activeSchema == strings.TrimSpace(assetRecognitionOutputSchema) {
			return nil
		}
		if activeSchema != strings.TrimSpace(assetRecognitionOutputSchemaV1) &&
			activeSchema != strings.TrimSpace(assetRecognitionOutputSchemaV2) {
			return nil
		}

		now := time.Now()
		latestVersion := templates[len(templates)-1].Version
		upgraded := ai.PromptTemplate{
			PromptKey: assetRecognitionPromptKey, Version: latestVersion + 1, Content: assetRecognitionPrompt,
			OutputSchema: assetRecognitionOutputSchema, Status: ai.PromptStatusActive, ActivatedAt: &now,
		}
		if err := tx.Create(&upgraded).Error; err != nil {
			return err
		}
		return tx.Model(&ai.PromptTemplate{}).
			Where("prompt_key = ? AND id = ? AND status = ?", assetRecognitionPromptKey, templates[activeIndex].ID, ai.PromptStatusActive).
			Updates(map[string]any{"status": ai.PromptStatusRetired, "activated_at": nil}).Error
	}); err != nil {
		global.GVA_LOG.Warn("预置资产智能建档 Prompt 失败", zap.Error(err))
	}
}
