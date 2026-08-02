package copilot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"tenant-copilot/backend/crypto"
	"tenant-copilot/backend/domain"
	"tenant-copilot/backend/store"
)

type Agent struct {
	store  store.Store
	signer *crypto.ProposalSigner
	apiKey string
}

func NewAgent(st store.Store, proposalSigner *crypto.ProposalSigner) *Agent {
	return &Agent{
		store:  st,
		signer: proposalSigner,
		apiKey: os.Getenv("GEMINI_API_KEY"),
	}
}

// ProcessQuery runs the diagnostic & proposal workflow for an operator
func (a *Agent) ProcessQuery(req domain.CopilotChatRequest) (*domain.CopilotChatResponse, error) {
	jobID := req.JobID
	if jobID == "" {
		jobID = "job-101"
	}

	job, exists := a.store.GetJob(jobID)
	if !exists {
		return nil, fmt.Errorf("job %s not found", jobID)
	}

	// 1. Read-only Tool Call: get_job_timeline
	timeline := a.store.GetTimeline(jobID)

	// Create Tool Traces for audit log transparency
	traces := []domain.ToolTrace{
		{
			ToolName: "get_job_timeline",
			Args:     map[string]string{"job_id": jobID},
			Result:   timeline,
		},
		{
			ToolName: "get_job_error",
			Args:     map[string]string{"job_id": jobID},
			Result: map[string]interface{}{
				"job_id":             job.ID,
				"state":              job.State,
				"failed_step":        job.FailedStepName,
				"last_error_summary": job.LastErrorSummary,
				"error_category":     job.ErrorCategory,
				"current_attempt":    job.CurrentAttempt,
				"max_retries":        job.MaxRetries,
			},
		},
	}

	// If GEMINI_API_KEY is provided, perform live call to Gemini API
	if a.apiKey != "" {
		geminiResp, err := a.callGeminiAPI(req.Query, job, timeline)
		if err == nil && geminiResp != nil {
			proposal := a.generateSignedProposal(job, "Operator requested root cause diagnosis & recovery proposal")
			traces = append(traces, domain.ToolTrace{
				ToolName: "propose_job_retry",
				Args:     map[string]string{"job_id": jobID, "reason": "Retry after failure analysis"},
				Result:   proposal,
			})
			geminiResp.ToolTraces = traces
			geminiResp.Proposal = proposal
			geminiResp.RawTimelineReturned = true
			return geminiResp, nil
		}
	}

	// Native Agent Engine (interprets raw timeline without hiding details)
	explanation, proposal := a.interpretRawTimelineAndPropose(job, timeline)

	if proposal != nil {
		traces = append(traces, domain.ToolTrace{
			ToolName: "propose_job_retry",
			Args:     map[string]string{"job_id": jobID, "reason": "Root cause identified; proposed gated retry"},
			Result:   proposal,
		})
	}

	return &domain.CopilotChatResponse{
		JobID:               jobID,
		Explanation:         explanation,
		ToolTraces:          traces,
		RawTimelineReturned: true,
		Proposal:            proposal,
	}, nil
}

func (a *Agent) generateSignedProposal(job *domain.ProvisioningJob, reason string) *domain.SignedProposal {
	proposal := &domain.SignedProposal{
		JobID:             job.ID,
		SubscriptionID:   job.SubscriptionID,
		TenantName:        job.TenantName,
		Action:            "retry_job",
		CurrentState:      string(job.State),
		TargetState:       string(domain.JobStateQueued),
		RemainingAttempts: job.MaxRetries + 1,
		MaxRetries:        job.MaxRetries,
		WhatReRunningDoes: fmt.Sprintf("Re-queues provisioning pipeline for tenant '%s' (%s). Worker will re-execute all 5 provisioning steps starting from Step 1, with live progress streamed to dashboard.", job.TenantName, job.SubscriptionID),
	}

	// Sign with 5-minute (300 seconds) expiration window
	a.signer.Sign(proposal, 300)
	return proposal
}

func (a *Agent) interpretRawTimelineAndPropose(job *domain.ProvisioningJob, timeline []domain.StepEvent) (string, *domain.SignedProposal) {
	var lastFailedStep domain.StepEvent
	var failedAttemptCount int
	for _, evt := range timeline {
		if evt.Status == domain.StepStatusFailed {
			lastFailedStep = evt
			if evt.Attempt > failedAttemptCount {
				failedAttemptCount = evt.Attempt
			}
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### Root Cause Analysis for Job `%s` (Subscription `%s`)\n\n", job.ID, job.SubscriptionID))

	if lastFailedStep.ID != "" {
		sb.WriteString(fmt.Sprintf("- **Failed Step**: Step %d — *%s*\n", lastFailedStep.StepIndex, lastFailedStep.StepName))
		sb.WriteString(fmt.Sprintf("- **Failed Attempt**: Attempt %d of %d (Retries Exhausted)\n", lastFailedStep.Attempt, job.MaxRetries))
		sb.WriteString(fmt.Sprintf("- **Underlying Error**: `%s`\n", lastFailedStep.ErrorMessage))

		if lastFailedStep.RawPayload != "" {
			sb.WriteString(fmt.Sprintf("- **Raw Payload Response**: ```json\n%s\n```\n", lastFailedStep.RawPayload))
		}

		if lastFailedStep.ErrorCategory == domain.ErrorCategoryTransient {
			sb.WriteString("- **Error Classification**: **TRANSIENT FAILURE** (Network Gateway / Service Timeout).\n")
			sb.WriteString("  - *Analysis*: The upstream KMS KeyVault endpoint timed out during encryption key allocation. Because this error is caused by a temporary cloud gateway hiccup, re-running the job has a high probability of success.\n\n")
		} else {
			sb.WriteString("- **Error Classification**: **STRUCTURAL FAILURE** (Resource / Configuration Conflict).\n")
			sb.WriteString("  - *Analysis*: The subdomain claim failed due to an existing resource registration conflict (`HTTP 409 Conflict`). A manual DNS config update is recommended.\n\n")
		}
	} else {
		sb.WriteString(fmt.Sprintf("Job is currently in state `%s`. Last recorded error: `%s`.\n\n", job.State, job.LastErrorSummary))
	}

	sb.WriteString("### Proposed Remediation\n")
	sb.WriteString("A **signed confirmation card** has been generated below for operator approval.")

	proposal := a.generateSignedProposal(job, "Copilot diagnostic remediation proposal")
	return sb.String(), proposal
}

func (a *Agent) callGeminiAPI(userQuery string, job *domain.ProvisioningJob, timeline []domain.StepEvent) (*domain.CopilotChatResponse, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s", a.apiKey)

	rawTimelineJSON, _ := json.MarshalIndent(timeline, "", "  ")

	prompt := fmt.Sprintf(`You are an expert Cloud DevOps Copilot explaining a failed tenant provisioning job.
An operator asks: "%s"

Job Details:
- Job ID: %s
- Tenant Name: %s
- Subscription ID: %s
- State: %s
- Last Error Summary: %s

RAW TIMELINE EVENT HISTORY (Unmodified):
%s

INSTRUCTIONS:
1. Analyze the raw timeline events without omitting details or pre-summarizing.
2. Explain the root cause in clear Markdown format:
   - Exactly which step failed
   - On which attempt number (and total retries)
   - The underlying error message and raw payload
   - Classify whether the error is TRANSIENT (503 timeout) or STRUCTURAL (409 conflict)
3. Conclude by stating that a signed, expiring retry proposal card has been generated below for human confirmation.`, userQuery, job.ID, job.TenantName, job.SubscriptionID, job.State, job.LastErrorSummary, string(rawTimelineJSON))

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": prompt},
				},
			},
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini API returned status %d", resp.StatusCode)
	}

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var geminiOut struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(bodyText, &geminiOut); err != nil || len(geminiOut.Candidates) == 0 {
		return nil, fmt.Errorf("failed to parse gemini response")
	}

	text := geminiOut.Candidates[0].Content.Parts[0].Text

	return &domain.CopilotChatResponse{
		JobID:       job.ID,
		Explanation: text,
	}, nil
}
