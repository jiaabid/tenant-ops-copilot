package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"tenant-copilot-backend/copilot"
	"tenant-copilot-backend/crypto"
	"tenant-copilot-backend/domain"
	"tenant-copilot-backend/store"
	"tenant-copilot-backend/worker"
)

type APIHandler struct {
	store        store.Store
	signer       *crypto.ProposalSigner
	copilotAgent *copilot.Agent
	worker       *worker.Worker
}

func NewAPIHandler(st store.Store, signer *crypto.ProposalSigner, copilotAgent *copilot.Agent, w *worker.Worker) *APIHandler {
	return &APIHandler{
		store:        st,
		signer:       signer,
		copilotAgent: copilotAgent,
		worker:       w,
	}
}

func (h *APIHandler) EnableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func (h *APIHandler) GetJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobs := h.store.GetAllJobs()
	var list []domain.JobListItem

	for _, job := range jobs {
		sub, _ := h.store.GetSubscription(job.SubscriptionID)
		list = append(list, domain.JobListItem{
			Job:          *job,
			Subscription: sub,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *APIHandler) GetJobDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	jobID := parts[3]

	job, exists := h.store.GetJob(jobID)
	if !exists {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	sub, _ := h.store.GetSubscription(job.SubscriptionID)
	timeline := h.store.GetTimeline(jobID)

	resp := domain.JobDetailResponse{
		Job:          *job,
		Subscription: sub,
		Timeline:     timeline,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *APIHandler) CopilotChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.CopilotChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	resp, err := h.copilotAgent.ProcessQuery(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Copilot processing error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *APIHandler) ConfirmRetry(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var proposal domain.SignedProposal
	if err := json.NewDecoder(r.Body).Decode(&proposal); err != nil {
		http.Error(w, "Invalid JSON proposal", http.StatusBadRequest)
		return
	}

	// HMAC Signature & Expiration Verification
	if err := h.signer.Verify(&proposal); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Security Verification Failed: %v", err),
		})
		return
	}

	// Trigger async background retry worker
	h.worker.StartRetryJob(proposal.JobID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Retry confirmed & queued for job %s. Live timeline streaming initiated.", proposal.JobID),
	})
}

func (h *APIHandler) StreamJobTimeline(w http.ResponseWriter, r *http.Request) {
	jobID := r.URL.Query().Get("job_id")
	if jobID == "" {
		http.Error(w, "Missing job_id parameter", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := h.worker.Subscribe(jobID)
	defer h.worker.Unsubscribe(jobID, ch)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	notify := r.Context().Done()

	for {
		select {
		case <-notify:
			return
		case evt, ok := <-ch:
			if !ok {
				return
			}
			jsonData, _ := json.Marshal(evt.Data)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evt.Type, jsonData)
			flusher.Flush()
		}
	}
}
