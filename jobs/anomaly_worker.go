package jobs

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"lite-collector/models"
	"lite-collector/repository"
	"lite-collector/services"
)

// AnomalyWorker polls for queued detect_anomaly jobs and processes them.
type AnomalyWorker struct {
	aiJobRepo      repository.AIJobRepository
	submissionRepo repository.SubmissionRepository
	formRepo       repository.FormRepository
	deepseek       *services.DeepSeekClient
	pollInterval   time.Duration
}

// NewAnomalyWorker creates a new anomaly detection worker.
func NewAnomalyWorker(
	aiJobRepo repository.AIJobRepository,
	submissionRepo repository.SubmissionRepository,
	formRepo repository.FormRepository,
	deepseek *services.DeepSeekClient,
) *AnomalyWorker {
	return &AnomalyWorker{
		aiJobRepo:      aiJobRepo,
		submissionRepo: submissionRepo,
		formRepo:       formRepo,
		deepseek:       deepseek,
		pollInterval:   5 * time.Second,
	}
}

// anomalyJobInput is the JSON input stored in ai_jobs.input
type anomalyJobInput struct {
	SubmissionID uint64 `json:"submission_id"`
}

// anomalyResult is what we ask DeepSeek to return
type anomalyResult struct {
	IsAnomalous bool     `json:"is_anomalous"`
	Reasons     []string `json:"reasons"`
}

// Start begins the polling loop in a goroutine. It runs until the process exits.
func (w *AnomalyWorker) Start() {
	go func() {
		log.Println("[anomaly-worker] started")
		for {
			w.processOne()
			time.Sleep(w.pollInterval)
		}
	}()
}

func (w *AnomalyWorker) processOne() {
	job, err := w.aiJobRepo.ClaimQueued()
	if err != nil {
		return // no queued jobs
	}

	if job.JobType != "detect_anomaly" {
		// Skip non-anomaly jobs for now; mark as failed so they don't block the queue
		w.failJob(job, "unsupported job type: "+job.JobType)
		return
	}

	log.Printf("[anomaly-worker] processing job %d", job.ID)

	var input anomalyJobInput
	if err := json.Unmarshal([]byte(job.Input), &input); err != nil {
		w.failJob(job, "invalid job input: "+err.Error())
		return
	}

	// Fetch submission and its values
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

	// Fetch form schema for context
	form, err := w.formRepo.FindByID(submission.FormID)
	if err != nil {
		w.failJob(job, "form not found: "+err.Error())
		return
	}

	// Build value map
	valueMap := make(map[string]string, len(values))
	for _, v := range values {
		valueMap[v.FieldKey] = v.Value
	}

	valuesJSON, _ := json.Marshal(valueMap)
	schemaStr := string(form.Schema)

	// Call DeepSeek
	systemPrompt := `You are a data quality inspector. You will receive a form schema (field definitions) and a user's submission values. Analyze the values for anomalies:

1. Values that don't match the expected type or format (e.g. letters in a number field)
2. Values that are unrealistic (e.g. age = 500, negative counts)
3. Cross-field inconsistencies (e.g. age doesn't match birth year)

Respond ONLY with a JSON object, no markdown, no explanation:
{"is_anomalous": true/false, "reasons": ["reason1", "reason2"]}

If the data looks normal, return: {"is_anomalous": false, "reasons": []}`

	userPrompt := fmt.Sprintf("Form schema:\n%s\n\nSubmission values:\n%s", schemaStr, string(valuesJSON))

	reply, err := w.deepseek.Chat(systemPrompt, userPrompt)
	if err != nil {
		w.failJob(job, "deepseek call failed: "+err.Error())
		return
	}

	// Parse DeepSeek's response
	var result anomalyResult
	if err := json.Unmarshal([]byte(reply), &result); err != nil {
		// Try to salvage — store raw reply and mark done anyway
		w.completeJob(job, reply)
		submission.Status = 1 // treat parse failure as normal
		_ = w.submissionRepo.Update(submission)
		return
	}

	// Update submission status
	if result.IsAnomalous {
		submission.Status = 2 // has_anomaly
	} else {
		submission.Status = 1 // normal
	}
	_ = w.submissionRepo.Update(submission)

	// Store result and mark job done
	output, _ := json.Marshal(result)
	w.completeJob(job, string(output))
	log.Printf("[anomaly-worker] job %d done — anomalous: %v", job.ID, result.IsAnomalous)
}

func (w *AnomalyWorker) failJob(job *models.AIJob, reason string) {
	job.Status = 3 // failed
	job.Output = reason
	_ = w.aiJobRepo.Update(job)
	log.Printf("[anomaly-worker] job %d failed: %s", job.ID, reason)
}

func (w *AnomalyWorker) completeJob(job *models.AIJob, output string) {
	job.Status = 2 // done
	job.Output = output
	_ = w.aiJobRepo.Update(job)
}
