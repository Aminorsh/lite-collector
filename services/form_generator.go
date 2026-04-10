package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"lite-collector/utils"
)

// GeneratedForm is the result of AI form generation.
type GeneratedForm struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Schema      string `json:"schema"`
}

// FormGenerator uses DeepSeek to generate form schemas from natural language.
type FormGenerator struct {
	deepseek *DeepSeekClient
}

// NewFormGenerator creates a new FormGenerator.
func NewFormGenerator(deepseek *DeepSeekClient) *FormGenerator {
	return &FormGenerator{deepseek: deepseek}
}

// Generate creates a form schema from a natural language description.
func (g *FormGenerator) Generate(description string) (*GeneratedForm, error) {
	if g.deepseek == nil {
		return nil, utils.ErrAINotConfigured
	}

	systemPrompt := `你是一个表单设计师。用户会用自然语言描述他们需要的表单，你需要生成表单的标题、描述和字段结构。

请仅返回一个 JSON 对象，不要使用 markdown，不要附加说明：
{
  "title": "表单标题",
  "description": "表单描述（一句话说明用途）",
  "schema": {
    "fields": [
      {"key": "f_001", "label": "字段名", "type": "text", "required": true},
      {"key": "f_002", "label": "字段名", "type": "number", "required": false}
    ]
  }
}

可用字段类型：text, number, date, select, textarea
如果是 select 类型，需要额外添加 "options": ["选项1", "选项2", ...]

字段 key 使用 f_001, f_002... 格式自动编号。required 根据字段重要性判断。`

	userPrompt := fmt.Sprintf("请为我生成一个表单：%s", description)

	reply, err := g.deepseek.Chat(systemPrompt, userPrompt)
	if err != nil {
		return nil, utils.ErrAIGenerateFail
	}

	// Strip markdown wrappers if present
	reply = strings.TrimSpace(reply)
	reply = strings.TrimPrefix(reply, "```json")
	reply = strings.TrimPrefix(reply, "```")
	reply = strings.TrimSuffix(reply, "```")
	reply = strings.TrimSpace(reply)

	// Parse and validate the response structure
	var raw struct {
		Title       string          `json:"title"`
		Description string          `json:"description"`
		Schema      json.RawMessage `json:"schema"`
	}
	if err := json.Unmarshal([]byte(reply), &raw); err != nil {
		return nil, utils.ErrAIGenerateFail
	}

	if raw.Title == "" || len(raw.Schema) == 0 {
		return nil, utils.ErrAIGenerateFail
	}

	return &GeneratedForm{
		Title:       raw.Title,
		Description: raw.Description,
		Schema:      string(raw.Schema),
	}, nil
}
