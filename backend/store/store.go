package store

import "tenant-copilot/backend/domain"

type Store interface {
	GetAllJobs() []*domain.ProvisioningJob
	GetJob(id string) (*domain.ProvisioningJob, bool)
	GetSubscription(id string) (*domain.Subscription, bool)
	GetTimeline(jobID string) []domain.StepEvent
	AddStepEvent(event domain.StepEvent)
	UpdateJobState(jobID string, state domain.JobState, attempt int, failedStep string, errSummary string, category domain.ErrorCategory)
	ResetJobTimeline(jobID string)
}
