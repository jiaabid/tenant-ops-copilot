package worker

import (
	"fmt"
	"log"
	"sync"
	"time"

	"tenant-copilot/backend/domain"
	"tenant-copilot/backend/store"
)

type Worker struct {
	store       store.Store
	mu          sync.RWMutex
	subscribers map[string][]chan domain.SSEEvent
}

func NewWorker(st store.Store) *Worker {
	return &Worker{
		store:       st,
		subscribers: make(map[string][]chan domain.SSEEvent),
	}
}

func (w *Worker) Subscribe(jobID string) chan domain.SSEEvent {
	w.mu.Lock()
	defer w.mu.Unlock()

	ch := make(chan domain.SSEEvent, 20)
	w.subscribers[jobID] = append(w.subscribers[jobID], ch)
	return ch
}

func (w *Worker) Unsubscribe(jobID string, ch chan domain.SSEEvent) {
	w.mu.Lock()
	defer w.mu.Unlock()

	subs := w.subscribers[jobID]
	for i, sub := range subs {
		if sub == ch {
			w.subscribers[jobID] = append(subs[:i], subs[i+1:]...)
			close(ch)
			break
		}
	}
}

func (w *Worker) broadcast(jobID string, eventType string, data interface{}) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	event := domain.SSEEvent{
		Type: eventType,
		Data: data,
	}

	for _, ch := range w.subscribers[jobID] {
		select {
		case ch <- event:
		default:
		}
	}
}

func (w *Worker) StartRetryJob(jobID string) {
	go func() {
		job, exists := w.store.GetJob(jobID)
		if !exists {
			return
		}

		log.Printf("Worker picking up job %s (Tenant: %s) for re-execution...", jobID, job.TenantName)

		// Set job state to queued
		w.store.UpdateJobState(jobID, domain.JobStateQueued, job.CurrentAttempt, "", "", "")
		if updated, ok := w.store.GetJob(jobID); ok {
			w.broadcast(jobID, "job_updated", updated)
		}

		time.Sleep(1 * time.Second)

		// Set job state to running
		w.store.UpdateJobState(jobID, domain.JobStateRunning, job.CurrentAttempt, "", "", "")
		if updated, ok := w.store.GetJob(jobID); ok {
			w.broadcast(jobID, "job_updated", updated)
		}

		// Reset timeline events
		w.store.ResetJobTimeline(jobID)

		// Run 5 provisioning pipeline steps dynamically
		steps := []struct {
			index int
			name  string
		}{
			{1, "Validate Tenant Domain & Identity Metadata"},
			{2, "Allocate Isolated Tenant Database Schema"},
			{3, "Provision Key Vault & Store Encryption Keys"},
			{4, "Configure Custom Subdomain DNS & TLS Certificate"},
			{5, "Initialize Core Platform Services & User Access"},
		}

		newAttempt := job.CurrentAttempt + 1

		for _, step := range steps {
			// Running event
			evtID := fmt.Sprintf("evt-retry-%d-%d", jobID, step.index)
			runningEvt := domain.StepEvent{
				ID:          evtID,
				JobID:       jobID,
				StepIndex:   step.index,
				StepName:    step.name,
				Attempt:     newAttempt,
				MaxAttempts: job.MaxRetries,
				Status:      domain.StepStatusRunning,
				Timestamp:   time.Now(),
			}
			w.store.AddStepEvent(runningEvt)
			w.broadcast(jobID, "step_updated", runningEvt)

			time.Sleep(1200 * time.Millisecond)

			// Completed event
			succeededEvt := domain.StepEvent{
				ID:          evtID,
				JobID:       jobID,
				StepIndex:   step.index,
				StepName:    step.name,
				Attempt:     newAttempt,
				MaxAttempts: job.MaxRetries,
				Status:      domain.StepStatusSucceeded,
				Timestamp:   time.Now(),
				DurationMs:  800,
			}
			w.store.AddStepEvent(succeededEvt)
			w.broadcast(jobID, "step_updated", succeededEvt)
		}

		// Update job as Succeeded!
		w.store.UpdateJobState(jobID, domain.JobStateSucceeded, newAttempt, "", "", "")
		if updated, ok := w.store.GetJob(jobID); ok {
			w.broadcast(jobID, "job_updated", updated)
		}

		log.Printf("Job %s successfully completed onboarding pipeline!", jobID)
	}()
}
