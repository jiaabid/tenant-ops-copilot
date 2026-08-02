package store

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "modernc.org/sqlite"
	"tenant-copilot/backend/domain"
)

type DBStore struct {
	mu sync.RWMutex
	db *sql.DB
}

func NewDBStore(dbPath string) (*DBStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db at %s: %w", dbPath, err)
	}

	store := &DBStore{db: db}
	if err := store.initTables(); err != nil {
		return nil, fmt.Errorf("failed to init db tables: %w", err)
	}

	if err := store.seedIfEmpty(); err != nil {
		log.Printf("Warning: error seeding DB: %v", err)
	}

	return store, nil
}

func (s *DBStore) initTables() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	queries := []string{
		`CREATE TABLE IF NOT EXISTS subscriptions (
			id TEXT PRIMARY KEY,
			tenant_name TEXT NOT NULL,
			plan TEXT NOT NULL,
			status TEXT NOT NULL,
			job_id TEXT NOT NULL,
			updated_at DATETIME NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS jobs (
			id TEXT PRIMARY KEY,
			subscription_id TEXT NOT NULL,
			tenant_name TEXT NOT NULL,
			region TEXT NOT NULL,
			state TEXT NOT NULL,
			current_attempt INTEGER NOT NULL,
			max_retries INTEGER NOT NULL,
			failed_step_name TEXT,
			last_error_summary TEXT,
			error_category TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS step_events (
			id TEXT PRIMARY KEY,
			job_id TEXT NOT NULL,
			step_index INTEGER NOT NULL,
			step_name TEXT NOT NULL,
			attempt INTEGER NOT NULL,
			max_attempts INTEGER NOT NULL,
			status TEXT NOT NULL,
			error_message TEXT,
			error_category TEXT,
			raw_payload TEXT,
			timestamp DATETIME NOT NULL,
			duration_ms INTEGER NOT NULL
		);`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func (s *DBStore) seedIfEmpty() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM jobs").Scan(&count)
	if err != nil || count > 0 {
		return nil // Already seeded or error
	}

	now := time.Now()

	// Seed Job 101 (Acme Financial Corp - KeyVault 503 Transient)
	job1ID := "job-101"
	sub1ID := "sub-8921"

	_, _ = s.db.Exec(`INSERT INTO subscriptions (id, tenant_name, plan, status, job_id, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		sub1ID, "Acme Financial Corp", "Enterprise-Tier", "failed", job1ID, now.Add(-15*time.Minute))

	_, _ = s.db.Exec(`INSERT INTO jobs (id, subscription_id, tenant_name, region, state, current_attempt, max_retries, failed_step_name, last_error_summary, error_category, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		job1ID, sub1ID, "Acme Financial Corp", "us-east-1", string(domain.JobStateFailed), 3, 3,
		"Provision Key Vault & Store Encryption Keys", "HTTP 503 Service Unavailable: Key Vault KMS endpoint timeout after 30000ms",
		string(domain.ErrorCategoryTransient), now.Add(-45*time.Minute), now.Add(-15*time.Minute))

	// Events for Job 101
	events1 := []domain.StepEvent{
		{ID: "evt-101-1", JobID: job1ID, StepIndex: 1, StepName: "Validate Tenant Domain & Identity Metadata", Attempt: 1, MaxAttempts: 3, Status: domain.StepStatusSucceeded, Timestamp: now.Add(-44 * time.Minute), DurationMs: 450},
		{ID: "evt-101-2", JobID: job1ID, StepIndex: 2, StepName: "Allocate Isolated Tenant Database Schema", Attempt: 1, MaxAttempts: 3, Status: domain.StepStatusSucceeded, Timestamp: now.Add(-43 * time.Minute), DurationMs: 1200},
		{ID: "evt-101-3", JobID: job1ID, StepIndex: 3, StepName: "Provision Key Vault & Store Encryption Keys", Attempt: 1, MaxAttempts: 3, Status: domain.StepStatusFailed, ErrorMessage: "HTTP 503 Service Unavailable: Gateway Timeout reaching kms-us-east-1.internal.cloud", ErrorCategory: domain.ErrorCategoryTransient, RawPayload: `{"error": "KMS_GATEWAY_TIMEOUT", "statusCode": 503, "endpoint": "kms-us-east-1.internal.cloud", "retryAfterSeconds": 5}`, Timestamp: now.Add(-40 * time.Minute), DurationMs: 30000},
		{ID: "evt-101-4", JobID: job1ID, StepIndex: 3, StepName: "Provision Key Vault & Store Encryption Keys", Attempt: 2, MaxAttempts: 3, Status: domain.StepStatusFailed, ErrorMessage: "HTTP 503 Service Unavailable: Gateway Timeout reaching kms-us-east-1.internal.cloud", ErrorCategory: domain.ErrorCategoryTransient, RawPayload: `{"error": "KMS_GATEWAY_TIMEOUT", "statusCode": 503, "endpoint": "kms-us-east-1.internal.cloud", "retryAfterSeconds": 15}`, Timestamp: now.Add(-30 * time.Minute), DurationMs: 30000},
		{ID: "evt-101-5", JobID: job1ID, StepIndex: 3, StepName: "Provision Key Vault & Store Encryption Keys", Attempt: 3, MaxAttempts: 3, Status: domain.StepStatusFailed, ErrorMessage: "HTTP 503 Service Unavailable: Max retry limit reached (3/3). Provisioning pipeline aborted.", ErrorCategory: domain.ErrorCategoryTransient, RawPayload: `{"error": "RETRY_LIMIT_EXHAUSTED", "statusCode": 503, "totalAttempts": 3, "lastAttemptError": "KMS_GATEWAY_TIMEOUT"}`, Timestamp: now.Add(-15 * time.Minute), DurationMs: 30000},
	}

	for _, e := range events1 {
		_, _ = s.db.Exec(`INSERT INTO step_events (id, job_id, step_index, step_name, attempt, max_attempts, status, error_message, error_category, raw_payload, timestamp, duration_ms) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			e.ID, e.JobID, e.StepIndex, e.StepName, e.Attempt, e.MaxAttempts, string(e.Status), e.ErrorMessage, string(e.ErrorCategory), e.RawPayload, e.Timestamp, e.DurationMs)
	}

	// Seed Job 102 (Global Logistics Inc - Subdomain 409 Structural)
	job2ID := "job-102"
	sub2ID := "sub-4419"

	_, _ = s.db.Exec(`INSERT INTO subscriptions (id, tenant_name, plan, status, job_id, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		sub2ID, "Global Logistics Inc", "Business-Tier", "failed", job2ID, now.Add(-2*time.Hour))

	_, _ = s.db.Exec(`INSERT INTO jobs (id, subscription_id, tenant_name, region, state, current_attempt, max_retries, failed_step_name, last_error_summary, error_category, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		job2ID, sub2ID, "Global Logistics Inc", "eu-west-1", string(domain.JobStateFailed), 3, 3,
		"Configure Custom Subdomain DNS & TLS Certificate", "HTTP 409 Conflict: Subdomain 'globallogistics.cloudplatform.com' is already registered to another active tenant",
		string(domain.ErrorCategoryStructural), now.Add(-3*time.Hour), now.Add(-2*time.Hour))

	events2 := []domain.StepEvent{
		{ID: "evt-102-1", JobID: job2ID, StepIndex: 1, StepName: "Validate Tenant Domain & Identity Metadata", Attempt: 1, MaxAttempts: 3, Status: domain.StepStatusSucceeded, Timestamp: now.Add(-3 * time.Hour), DurationMs: 300},
		{ID: "evt-102-2", JobID: job2ID, StepIndex: 2, StepName: "Configure Custom Subdomain DNS & TLS Certificate", Attempt: 1, MaxAttempts: 3, Status: domain.StepStatusFailed, ErrorMessage: "HTTP 409 Conflict: Domain claim rejected by routing registry", ErrorCategory: domain.ErrorCategoryStructural, RawPayload: `{"error": "SUBDOMAIN_ALREADY_EXISTS", "statusCode": 409, "subdomain": "globallogistics.cloudplatform.com"}`, Timestamp: now.Add(-2 * time.Hour), DurationMs: 150},
		{ID: "evt-102-3", JobID: job2ID, StepIndex: 2, StepName: "Configure Custom Subdomain DNS & TLS Certificate", Attempt: 2, MaxAttempts: 3, Status: domain.StepStatusFailed, ErrorMessage: "HTTP 409 Conflict: Subdomain conflict unresolved", ErrorCategory: domain.ErrorCategoryStructural, RawPayload: `{"error": "SUBDOMAIN_ALREADY_EXISTS", "statusCode": 409, "subdomain": "globallogistics.cloudplatform.com"}`, Timestamp: now.Add(-2 * time.Hour), DurationMs: 140},
		{ID: "evt-102-4", JobID: job2ID, StepIndex: 2, StepName: "Configure Custom Subdomain DNS & TLS Certificate", Attempt: 3, MaxAttempts: 3, Status: domain.StepStatusFailed, ErrorMessage: "HTTP 409 Conflict: Retries exhausted. Structural DNS configuration change required.", ErrorCategory: domain.ErrorCategoryStructural, RawPayload: `{"error": "RETRY_LIMIT_EXHAUSTED", "statusCode": 409, "structuralFixRequired": true}`, Timestamp: now.Add(-2 * time.Hour), DurationMs: 160},
	}

	for _, e := range events2 {
		_, _ = s.db.Exec(`INSERT INTO step_events (id, job_id, step_index, step_name, attempt, max_attempts, status, error_message, error_category, raw_payload, timestamp, duration_ms) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			e.ID, e.JobID, e.StepIndex, e.StepName, e.Attempt, e.MaxAttempts, string(e.Status), e.ErrorMessage, string(e.ErrorCategory), e.RawPayload, e.Timestamp, e.DurationMs)
	}

	log.Printf("Successfully seeded SQLite database with initial incident scenarios.")
	return nil
}

func (s *DBStore) GetAllJobs() []*domain.ProvisioningJob {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query("SELECT id, subscription_id, tenant_name, region, state, current_attempt, max_retries, failed_step_name, last_error_summary, error_category, created_at, updated_at FROM jobs")
	if err != nil {
		return []*domain.ProvisioningJob{}
	}
	defer rows.Close()

	var result []*domain.ProvisioningJob
	for rows.Next() {
		var j domain.ProvisioningJob
		var stateStr, catStr string
		var failedStep, errSummary sql.NullString

		if err := rows.Scan(&j.ID, &j.SubscriptionID, &j.TenantName, &j.Region, &stateStr, &j.CurrentAttempt, &j.MaxRetries, &failedStep, &errSummary, &catStr, &j.CreatedAt, &j.UpdatedAt); err == nil {
			j.State = domain.JobState(stateStr)
			j.ErrorCategory = domain.ErrorCategory(catStr)
			j.FailedStepName = failedStep.String
			j.LastErrorSummary = errSummary.String
			result = append(result, &j)
		}
	}
	return result
}

func (s *DBStore) GetJob(id string) (*domain.ProvisioningJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow("SELECT id, subscription_id, tenant_name, region, state, current_attempt, max_retries, failed_step_name, last_error_summary, error_category, created_at, updated_at FROM jobs WHERE id = ?", id)

	var j domain.ProvisioningJob
	var stateStr, catStr string
	var failedStep, errSummary sql.NullString

	err := row.Scan(&j.ID, &j.SubscriptionID, &j.TenantName, &j.Region, &stateStr, &j.CurrentAttempt, &j.MaxRetries, &failedStep, &errSummary, &catStr, &j.CreatedAt, &j.UpdatedAt)
	if err != nil {
		return nil, false
	}

	j.State = domain.JobState(stateStr)
	j.ErrorCategory = domain.ErrorCategory(catStr)
	j.FailedStepName = failedStep.String
	j.LastErrorSummary = errSummary.String

	return &j, true
}

func (s *DBStore) GetSubscription(id string) (*domain.Subscription, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow("SELECT id, tenant_name, plan, status, job_id, updated_at FROM subscriptions WHERE id = ?", id)

	var sub domain.Subscription
	err := row.Scan(&sub.ID, &sub.TenantName, &sub.Plan, &sub.Status, &sub.JobID, &sub.UpdatedAt)
	if err != nil {
		return nil, false
	}

	return &sub, true
}

func (s *DBStore) GetTimeline(jobID string) []domain.StepEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query("SELECT id, job_id, step_index, step_name, attempt, max_attempts, status, error_message, error_category, raw_payload, timestamp, duration_ms FROM step_events WHERE job_id = ? ORDER BY timestamp ASC", jobID)
	if err != nil {
		return []domain.StepEvent{}
	}
	defer rows.Close()

	var events []domain.StepEvent
	for rows.Next() {
		var e domain.StepEvent
		var statusStr, catStr sql.NullString
		var errMsg, rawPayload sql.NullString

		if err := rows.Scan(&e.ID, &e.JobID, &e.StepIndex, &e.StepName, &e.Attempt, &e.MaxAttempts, &statusStr, &errMsg, &catStr, &rawPayload, &e.Timestamp, &e.DurationMs); err == nil {
			e.Status = domain.StepStatus(statusStr.String)
			e.ErrorCategory = domain.ErrorCategory(catStr.String)
			e.ErrorMessage = errMsg.String
			e.RawPayload = rawPayload.String
			events = append(events, e)
		}
	}
	return events
}

func (s *DBStore) AddStepEvent(event domain.StepEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, _ = s.db.Exec(`INSERT INTO step_events (id, job_id, step_index, step_name, attempt, max_attempts, status, error_message, error_category, raw_payload, timestamp, duration_ms) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		event.ID, event.JobID, event.StepIndex, event.StepName, event.Attempt, event.MaxAttempts, string(event.Status), event.ErrorMessage, string(event.ErrorCategory), event.RawPayload, event.Timestamp, event.DurationMs)
}

func (s *DBStore) UpdateJobState(jobID string, state domain.JobState, attempt int, failedStep string, errSummary string, category domain.ErrorCategory) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	_, _ = s.db.Exec(`UPDATE jobs SET state = ?, current_attempt = ?, failed_step_name = ?, last_error_summary = ?, error_category = ?, updated_at = ? WHERE id = ?`,
		string(state), attempt, failedStep, errSummary, string(category), now, jobID)

	var subID string
	_ = s.db.QueryRow("SELECT subscription_id FROM jobs WHERE id = ?", jobID).Scan(&subID)
	if subID != "" {
		subStatus := "provisioning"
		if state == domain.JobStateFailed {
			subStatus = "failed"
		} else if state == domain.JobStateSucceeded {
			subStatus = "active"
		}
		_, _ = s.db.Exec(`UPDATE subscriptions SET status = ?, updated_at = ? WHERE id = ?`, subStatus, now, subID)
	}
}

func (s *DBStore) ResetJobTimeline(jobID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, _ = s.db.Exec(`DELETE FROM step_events WHERE job_id = ?`, jobID)
}

func (s *DBStore) Close() error {
	return s.db.Close()
}
