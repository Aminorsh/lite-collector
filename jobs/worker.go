package jobs

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"lite-collector/models"
	"lite-collector/repository"
	"lite-collector/services"
)

// Worker polls for queued AI jobs and dispatches them by job type.
type Worker struct {
	aiJobRepo      repository.AIJobRepository
	submissionRepo repository.SubmissionRepository
	formRepo       repository.FormRepository
	deepseek       *services.DeepSeekClient
	formGenerator  *services.FormGenerator
	pollInterval   time.Duration
}

// NewWorker creates a new AI job worker.
func NewWorker(
	aiJobRepo repository.AIJobRepository,
	submissionRepo repository.SubmissionRepository,
	formRepo repository.FormRepository,
	deepseek *services.DeepSeekClient,
	formGenerator *services.FormGenerator,
) *Worker {
	return &Worker{
		aiJobRepo:      aiJobRepo,
		submissionRepo: submissionRepo,
		formRepo:       formRepo,
		deepseek:       deepseek,
		formGenerator:  formGenerator,
		pollInterval:   5 * time.Second,
	}
}

// Start begins the polling loop in a goroutine.
func (w *Worker) Start() {
	go func() {
		log.Println("[ai-worker] started")
		for {
			w.processOne()
			time.Sleep(w.pollInterval)
		}
	}()
}

func (w *Worker) processOne() {
	job, err := w.aiJobRepo.ClaimQueued()
	if err != nil {
		return
	}

	log.Printf("[ai-worker] processing job %d (type: %s)", job.ID, job.JobType)

	switch job.JobType {
	case "detect_anomaly":
		w.handleAnomaly(job)
	case "generate_report":
		w.handleReport(job)
	case "generate_form":
		w.handleFormGeneration(job)
	default:
		w.failJob(job, "unsupported job type: "+job.JobType)
	}
}

// --- Form generation ---

type formGenInput struct {
	Description string `json:"description"`
}

func (w *Worker) handleFormGeneration(job *models.AIJob) {
	if w.formGenerator == nil {
		w.failJob(job, "form generator not configured")
		return
	}

	var input formGenInput
	if err := json.Unmarshal([]byte(job.Input), &input); err != nil {
		w.failJob(job, "invalid job input: "+err.Error())
		return
	}

	result, err := w.formGenerator.Generate(input.Description)
	if err != nil {
		w.failJob(job, "form generation failed: "+err.Error())
		return
	}

	output, err := json.Marshal(map[string]string{
		"title":       result.Title,
		"description": result.Description,
		"schema":      result.Schema,
	})
	if err != nil {
		w.failJob(job, "failed to encode result: "+err.Error())
		return
	}

	w.completeJob(job, string(output))
	log.Printf("[ai-worker] form generation job %d done", job.ID)
}

// --- Anomaly detection ---

type anomalyJobInput struct {
	SubmissionID uint64 `json:"submission_id"`
}

type anomalyResult struct {
	IsAnomalous bool     `json:"is_anomalous"`
	Reasons     []string `json:"reasons"`
}

func (w *Worker) handleAnomaly(job *models.AIJob) {
	var input anomalyJobInput
	if err := json.Unmarshal([]byte(job.Input), &input); err != nil {
		w.failJob(job, "invalid job input: "+err.Error())
		return
	}

	submission, err := w.submissionRepo.FindByID(input.SubmissionID)
	if err != nil {
		w.failJob(job, "submission not found: "+err.Error())
		return
	}

	values, err := w.submissionRepo.FindValuesBySubmissionID(submission.ID)
	if err != nil {
		w.failJob(job, "failed to load submission values: "+err.Error())
		return
	}

	form, err := w.formRepo.FindByID(submission.FormID)
	if err != nil {
		w.failJob(job, "form not found: "+err.Error())
		return
	}

	valueMap := make(map[string]string, len(values))
	for _, v := range values {
		valueMap[v.FieldKey] = v.Value
	}

	valuesJSON, _ := json.Marshal(valueMap)

	systemPrompt := `你是一个数据质量检查员。你会收到一份表单结构（字段定义）和一份用户提交的数据。请分析数据是否存在异常：

1. 值与预期类型或格式不符（如数字字段出现字母）
2. 值不合理（如年龄=500、负数金额）
3. 字段之间不一致（如年龄与出生年份矛盾）

请仅返回一个 JSON 对象，不要使用 markdown，不要附加说明：
{"is_anomalous": true/false, "reasons": ["原因1", "原因2"]}

如果数据正常，返回：{"is_anomalous": false, "reasons": []}`

	userPrompt := fmt.Sprintf("表单结构:\n%s\n\n提交数据:\n%s", string(form.Schema), string(valuesJSON))

	reply, err := w.deepseek.Chat(systemPrompt, userPrompt)
	if err != nil {
		w.failJob(job, "deepseek call failed: "+err.Error())
		return
	}

	var result anomalyResult
	if err := json.Unmarshal([]byte(reply), &result); err != nil {
		w.completeJob(job, reply)
		submission.Status = 1
		_ = w.submissionRepo.Update(submission)
		return
	}

	if result.IsAnomalous {
		submission.Status = 2
	} else {
		submission.Status = 1
	}
	_ = w.submissionRepo.Update(submission)

	output, _ := json.Marshal(result)
	w.completeJob(job, string(output))
	log.Printf("[ai-worker] anomaly job %d done — anomalous: %v", job.ID, result.IsAnomalous)
}

// --- Report generation ---

type reportJobInput struct {
	FormID uint64 `json:"form_id"`
}

func (w *Worker) handleReport(job *models.AIJob) {
	var input reportJobInput
	if err := json.Unmarshal([]byte(job.Input), &input); err != nil {
		w.failJob(job, "invalid job input: "+err.Error())
		return
	}

	form, err := w.formRepo.FindByID(input.FormID)
	if err != nil {
		w.failJob(job, "form not found: "+err.Error())
		return
	}

	submissions, err := w.submissionRepo.FindByFormID(form.ID)
	if err != nil {
		w.failJob(job, "failed to load submissions: "+err.Error())
		return
	}

	if len(submissions) == 0 {
		w.completeJob(job, `{"summary":"暂无提交数据，无法生成报告。"}`)
		return
	}

	// Build all submission data for the prompt
	var allData []map[string]string
	anomalyCount := 0
	for _, sub := range submissions {
		values, err := w.submissionRepo.FindValuesBySubmissionID(sub.ID)
		if err != nil {
			continue
		}
		row := make(map[string]string, len(values)+1)
		for _, v := range values {
			row[v.FieldKey] = v.Value
		}
		switch sub.Status {
		case 1:
			row["_status"] = "正常"
		case 2:
			row["_status"] = "异常"
			anomalyCount++
		default:
			row["_status"] = "待检测"
		}
		allData = append(allData, row)
	}

	dataJSON, _ := json.Marshal(allData)

	systemPrompt := `你是一个数据分析师。你会收到一份表单的结构定义和所有提交数据。请生成一份简洁的汇总报告，包含：

1. **概况**：提交总数、正常数、异常数
2. **关键统计**：对数值字段给出最小值、最大值、平均值；对文本字段给出常见值分布
3. **异常汇总**：列出所有被标记为异常的数据及其问题
4. **建议**：基于数据给出改进建议（如有）

请用 markdown 格式返回报告内容，不要包含代码块标记。`

	userPrompt := fmt.Sprintf("表单名称: %s\n表单结构:\n%s\n\n共 %d 条提交（其中 %d 条异常），数据如下:\n%s",
		form.Title, string(form.Schema), len(submissions), anomalyCount, string(dataJSON))

	reply, err := w.deepseek.Chat(systemPrompt, userPrompt)
	if err != nil {
		w.failJob(job, "deepseek call failed: "+err.Error())
		return
	}

	// Strip markdown code block wrappers if present
	reply = strings.TrimSpace(reply)
	reply = strings.TrimPrefix(reply, "```markdown")
	reply = strings.TrimPrefix(reply, "```")
	reply = strings.TrimSuffix(reply, "```")
	reply = strings.TrimSpace(reply)

	w.completeJob(job, reply)
	log.Printf("[ai-worker] report job %d done — %d submissions analyzed", job.ID, len(submissions))
}

// --- Helpers ---

func (w *Worker) failJob(job *models.AIJob, reason string) {
	job.Status = 3
	job.Output = reason
	_ = w.aiJobRepo.Update(job)
	log.Printf("[ai-worker] job %d failed: %s", job.ID, reason)
}

func (w *Worker) completeJob(job *models.AIJob, output string) {
	job.Status = 2
	job.Output = output
	_ = w.aiJobRepo.Update(job)
}
