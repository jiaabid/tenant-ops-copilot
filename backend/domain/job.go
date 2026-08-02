package domain

import "time"

type JobState string

const (
	JobStateQueued    JobState = "queued"
	JobStateRunning   JobState = "running"
	JobStateFailed    JobState = "failed"
	JobStateSucceeded JobState = "succeeded"
)

type StepStatus string

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusRunning   StepStatus = "running"
	StepStatusSucceeded StepStatus = "succeeded"
	StepStatusFailed    StepStatus = "failed"
)

type ErrorCategory string

const (
	ErrorCategoryTransient  ErrorCategory = "transient"
	ErrorCategoryStructural ErrorCategory = "structural"
	ErrorCategoryNone       ErrorCategory = "none"
)

// ProvisioningJob represents a tenant provisioning workflow job
type ProvisioningJob struct {
	ID               string        `json:"id"`
	SubscriptionID   string        `json:"subscription_id"`
	TenantName       string        `json:"tenant_name"`
	Region           string        `json:"region"`
	State            JobState      `json:"state"`
	CurrentAttempt   int           `json:"current_attempt"`
	MaxRetries       int           `json:"max_retries"`
	FailedStepName   string        `json:"failed_step_name,omitempty"`
	LastErrorSummary string        `json:"last_error_summary,omitempty"`
	ErrorCategory    ErrorCategory `json:"error_category,omitempty"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}

// StepEvent represents an individual step execution in a provisioning job
type StepEvent struct {
	ID            string        `json:"id"`
	JobID         string        `json:"job_id"`
	StepIndex     int           `json:"step_index"`
	StepName      string        `json:"step_name"`
	Attempt       int           `json:"attempt"`
	MaxAttempts   int           `json:"max_attempts"`
	Status        StepStatus    `json:"status"`
	ErrorMessage  string        `json:"error_message,omitempty"`
	ErrorCategory ErrorCategory `json:"error_category,omitempty"`
	RawPayload    string        `json:"raw_payload,omitempty"`
	Timestamp     time.Time     `json:"timestamp"`
	DurationMs    int64         `json:"duration_ms"`
}

// Subscription represents tenant subscription status
type Subscription struct {
	ID         string    `json:"id"`
	TenantName string    `json:"tenant_name"`
	Plan       string    `json:"plan"`
	Status     string    `json:"status"`
	JobID      string    `json:"job_id"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// SignedProposal represents a cryptographic confirmation card for job remediation
type SignedProposal struct {
	JobID             string `json:"job_id"`
	SubscriptionID    string `json:"subscription_id"`
	TenantName        string `json:"tenant_name"`
	Action            string `json:"action"`
	CurrentState      string `json:"current_state"`
	TargetState       string `json:"target_state"`
	RemainingAttempts int    `json:"remaining_attempts"`
	MaxRetries        int    `json:"max_retries"`
	WhatReRunningDoes string `json:"what_rerunning_does"`
	IssuedAt          int64  `json:"issued_at"`
	ExpiresAt         int64  `json:"expires_at"`
	Nonce             string `json:"nonce"`
	Signature         string `json:"signature"`
}

type ConfirmRetryRequest struct {
	JobID          string         `json:"job_id"`
	SignedProposal SignedProposal `json:"signed_proposal"`
}

type CopilotChatRequest struct {
	JobID string `json:"job_id"`
	Query string `json:"query"`
}

type ToolTrace struct {
	ToolName string      `json:"tool_name"`
	Args     interface{} `json:"args"`
	Result   interface{} `json:"result"`
}

type CopilotChatResponse struct {
	JobID               string          `json:"job_id"`
	Explanation         string          `json:"explanation"`
	ToolTraces          []ToolTrace     `json:"tool_traces"`
	RawTimelineReturned bool            `json:"raw_timeline_returned"`
	Proposal            *SignedProposal `json:"proposal,omitempty"`
}

type JobListItem struct {
	Job          ProvisioningJob `json:"job"`
	Subscription *Subscription   `json:"subscription"`
}

type JobDetailResponse struct {
	Job          ProvisioningJob `json:"job"`
	Subscription *Subscription   `json:"subscription"`
	Timeline     []StepEvent     `json:"timeline"`
}

type SSEEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
