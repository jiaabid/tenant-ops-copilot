package store

import (
	"sync"
	"time"

	"tenant-copilot/backend/domain"
)

type MemoryStore struct {
	mu            sync.RWMutex
	jobs          map[string]*domain.ProvisioningJob
	subscriptions map[string]*domain.Subscription
	timelines     map[string][]domain.StepEvent
}

func NewMemoryStore() *MemoryStore {
	s := &MemoryStore{
		jobs:          make(map[string]*domain.ProvisioningJob),
		subscriptions: make(map[string]*domain.Subscription),
		timelines:     make(map[string][]domain.StepEvent),
	}
	s.seedData()
	return s
}

func (s *MemoryStore) seedData() {
	now := time.Now()

	// Incident Scenario 1: KeyVault API Timeout (Transient Error)
	job1ID := "job-101"
	sub1ID := "sub-8921"
	s.subscriptions[sub1ID] = &domain.Subscription{
		ID:         sub1ID,
		TenantName: "Acme Financial Corp",
		Plan:       "Enterprise-Tier",
		Status:     "failed",
		JobID:      job1ID,
		UpdatedAt:  now.Add(-15 * time.Minute),
	}

	s.jobs[job1ID] = &domain.ProvisioningJob{
		ID:               job1ID,
		SubscriptionID:   sub1ID,
		TenantName:       "Acme Financial Corp",
		Region:           "us-east-1",
		State:            domain.JobStateFailed,
		CurrentAttempt:   3,
		MaxRetries:       3,
		FailedStepName:   "Provision Key Vault & Store Encryption Keys",
		LastErrorSummary: "HTTP 503 Service Unavailable: Key Vault KMS endpoint timeout after 30000ms",
		ErrorCategory:    domain.ErrorCategoryTransient,
		CreatedAt:        now.Add(-45 * time.Minute),
		UpdatedAt:        now.Add(-15 * time.Minute),
	}

	s.timelines[job1ID] = []domain.StepEvent{
		// Step 1 - Attempt 1
		{
			ID:          "evt-101-1",
			JobID:       job1ID,
			StepIndex:   1,
			StepName:    "Validate Tenant Domain & Identity Metadata",
			Attempt:     1,
			MaxAttempts: 3,
			Status:      domain.StepStatusSucceeded,
			Timestamp:   now.Add(-44 * time.Minute),
			DurationMs:  450,
		},
		// Step 2 - Attempt 1
		{
			ID:          "evt-101-2",
			JobID:       job1ID,
			StepIndex:   2,
			StepName:    "Allocate Isolated Tenant Database Schema",
			Attempt:     1,
			MaxAttempts: 3,
			Status:      domain.StepStatusSucceeded,
			Timestamp:   now.Add(-43 * time.Minute),
			DurationMs:  1200,
		},
		// Step 3 - Attempt 1 (Failed - Transient Timeout)
		{
			ID:            "evt-101-3",
			JobID:         job1ID,
			StepIndex:     3,
			StepName:      "Provision Key Vault & Store Encryption Keys",
			Attempt:       1,
			MaxAttempts:   3,
			Status:        domain.StepStatusFailed,
			ErrorMessage:  "HTTP 503 Service Unavailable: Gateway Timeout reaching kms-us-east-1.internal.cloud",
			ErrorCategory: domain.ErrorCategoryTransient,
			RawPayload:    `{"error": "KMS_GATEWAY_TIMEOUT", "statusCode": 503, "endpoint": "kms-us-east-1.internal.cloud", "retryAfterSeconds": 5}`,
			Timestamp:     now.Add(-40 * time.Minute),
			DurationMs:    30000,
		},
		// Step 3 - Attempt 2 (Failed - Transient Timeout)
		{
			ID:            "evt-101-4",
			JobID:         job1ID,
			StepIndex:     3,
			StepName:      "Provision Key Vault & Store Encryption Keys",
			Attempt:       2,
			MaxAttempts:   3,
			Status:        domain.StepStatusFailed,
			ErrorMessage:  "HTTP 503 Service Unavailable: Gateway Timeout reaching kms-us-east-1.internal.cloud",
			ErrorCategory: domain.ErrorCategoryTransient,
			RawPayload:    `{"error": "KMS_GATEWAY_TIMEOUT", "statusCode": 503, "endpoint": "kms-us-east-1.internal.cloud", "retryAfterSeconds": 15}`,
			Timestamp:     now.Add(-30 * time.Minute),
			DurationMs:    30000,
		},
		// Step 3 - Attempt 3 (Failed - Retries Exhausted)
		{
			ID:            "evt-101-5",
			JobID:         job1ID,
			StepIndex:     3,
			StepName:      "Provision Key Vault & Store Encryption Keys",
			Attempt:       3,
			MaxAttempts:   3,
			Status:        domain.StepStatusFailed,
			ErrorMessage:  "HTTP 503 Service Unavailable: Max retry limit reached (3/3). Provisioning pipeline aborted.",
			ErrorCategory: domain.ErrorCategoryTransient,
			RawPayload:    `{"error": "RETRY_LIMIT_EXHAUSTED", "statusCode": 503, "totalAttempts": 3, "lastAttemptError": "KMS_GATEWAY_TIMEOUT"}`,
			Timestamp:     now.Add(-15 * time.Minute),
			DurationMs:    30000,
		},
	}

	// Incident Scenario 2: Structural Domain Conflict Error
	job2ID := "job-102"
	sub2ID := "sub-4419"
	s.subscriptions[sub2ID] = &domain.Subscription{
		ID:         sub2ID,
		TenantName: "Global Logistics Inc",
		Plan:       "Business-Tier",
		Status:     "failed",
		JobID:      job2ID,
		UpdatedAt:  now.Add(-2 * time.Hour),
	}

	s.jobs[job2ID] = &domain.ProvisioningJob{
		ID:               job2ID,
		SubscriptionID:   sub2ID,
		TenantName:       "Global Logistics Inc",
		Region:           "eu-west-1",
		State:            domain.JobStateFailed,
		CurrentAttempt:   3,
		MaxRetries:       3,
		FailedStepName:   "Configure Custom Subdomain DNS & TLS Certificate",
		LastErrorSummary: "HTTP 409 Conflict: Subdomain 'globallogistics.cloudplatform.com' is already registered to another active tenant",
		ErrorCategory:    domain.ErrorCategoryStructural,
		CreatedAt:        now.Add(-3 * time.Hour),
		UpdatedAt:        now.Add(-2 * time.Hour),
	}

	s.timelines[job2ID] = []domain.StepEvent{
		{
			ID:          "evt-102-1",
			JobID:       job2ID,
			StepIndex:   1,
			StepName:    "Validate Tenant Domain & Identity Metadata",
			Attempt:     1,
			MaxAttempts: 3,
			Status:      domain.StepStatusSucceeded,
			Timestamp:   now.Add(-3 * time.Hour),
			DurationMs:  300,
		},
		{
			ID:            "evt-102-2",
			JobID:         job2ID,
			StepIndex:     2,
			StepName:      "Configure Custom Subdomain DNS & TLS Certificate",
			Attempt:       1,
			MaxAttempts:   3,
			Status:        domain.StepStatusFailed,
			ErrorMessage:  "HTTP 409 Conflict: Domain claim rejected by routing registry",
			ErrorCategory: domain.ErrorCategoryStructural,
			RawPayload:    `{"error": "SUBDOMAIN_ALREADY_EXISTS", "statusCode": 409, "subdomain": "globallogistics.cloudplatform.com"}`,
			Timestamp:     now.Add(-2 * time.Hour),
			DurationMs:    150,
		},
		{
			ID:            "evt-102-3",
			JobID:         job2ID,
			StepIndex:     2,
			StepName:      "Configure Custom Subdomain DNS & TLS Certificate",
			Attempt:       2,
			MaxAttempts:   3,
			Status:        domain.StepStatusFailed,
			ErrorMessage:  "HTTP 409 Conflict: Subdomain conflict unresolved",
			ErrorCategory: domain.ErrorCategoryStructural,
			RawPayload:    `{"error": "SUBDOMAIN_ALREADY_EXISTS", "statusCode": 409, "subdomain": "globallogistics.cloudplatform.com"}`,
			Timestamp:     now.Add(-2 * time.Hour),
			DurationMs:    140,
		},
		{
			ID:            "evt-102-4",
			JobID:         job2ID,
			StepIndex:     2,
			StepName:      "Configure Custom Subdomain DNS & TLS Certificate",
			Attempt:       3,
			MaxAttempts:   3,
			Status:        domain.StepStatusFailed,
			ErrorMessage:  "HTTP 409 Conflict: Retries exhausted. Structural DNS configuration change required.",
			ErrorCategory: domain.ErrorCategoryStructural,
			RawPayload:    `{"error": "RETRY_LIMIT_EXHAUSTED", "statusCode": 409, "structuralFixRequired": true}`,
			Timestamp:     now.Add(-2 * time.Hour),
			DurationMs:    160,
		},
	}
}

func (s *MemoryStore) GetAllJobs() []*domain.ProvisioningJob {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*domain.ProvisioningJob, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobCopy := *job
		result = append(result, &jobCopy)
	}
	return result
}

func (s *MemoryStore) GetJob(id string) (*domain.ProvisioningJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.jobs[id]
	if !exists {
		return nil, false
	}
	jobCopy := *job
	return &jobCopy, true
}

func (s *MemoryStore) GetSubscription(id string) (*domain.Subscription, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sub, exists := s.subscriptions[id]
	if !exists {
		return nil, false
	}
	subCopy := *sub
	return &subCopy, true
}

func (s *MemoryStore) GetTimeline(jobID string) []domain.StepEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events, exists := s.timelines[jobID]
	if !exists {
		return []domain.StepEvent{}
	}
	result := make([]domain.StepEvent, len(events))
	copy(result, events)
	return result
}

func (s *MemoryStore) AddStepEvent(event domain.StepEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.timelines[event.JobID] = append(s.timelines[event.JobID], event)
}

func (s *MemoryStore) UpdateJobState(jobID string, state domain.JobState, attempt int, failedStep string, errSummary string, category domain.ErrorCategory) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[jobID]
	if !exists {
		return
	}

	job.State = state
	job.CurrentAttempt = attempt
	job.FailedStepName = failedStep
	job.LastErrorSummary = errSummary
	job.ErrorCategory = category
	job.UpdatedAt = time.Now()

	// Sync subscription state
	if sub, ok := s.subscriptions[job.SubscriptionID]; ok {
		if state == domain.JobStateFailed {
			sub.Status = "failed"
		} else if state == domain.JobStateSucceeded {
			sub.Status = "active"
		} else if state == domain.JobStateQueued || state == domain.JobStateRunning {
			sub.Status = "provisioning"
		}
		sub.UpdatedAt = time.Now()
	}
}

func (s *MemoryStore) ResetJobTimeline(jobID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Keep history or clear timeline for fresh run
	s.timelines[jobID] = []domain.StepEvent{}
}
