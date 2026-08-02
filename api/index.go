package handler

import (
	"net/http"
	"os"
	"strings"
	"sync"

	apiHandler "tenant-copilot/backend/api"
	"tenant-copilot/backend/copilot"
	"tenant-copilot/backend/crypto"
	"tenant-copilot/backend/store"
	"tenant-copilot/backend/worker"
)

var (
	once    sync.Once
	handler *apiHandler.APIHandler
)

func initHandler() {
	secretKey := os.Getenv("HMAC_SECRET_KEY")
	if secretKey == "" {
		secretKey = "copilot-secure-gated-remediation-secret-key"
	}

	memStore := store.NewMemoryStore()
	proposalSigner := crypto.NewProposalSigner(secretKey)
	copilotAgent := copilot.NewAgent(memStore, proposalSigner)
	workerSim := worker.NewWorker(memStore)

	handler = apiHandler.NewAPIHandler(memStore, proposalSigner, copilotAgent, workerSim)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(initHandler)

	path := r.URL.Path

	switch {
	case path == "/api/jobs" || path == "/api/jobs/":
		handler.EnableCORS(handler.GetJobs)(w, r)
	case strings.HasPrefix(path, "/api/jobs/retry/confirm"):
		handler.EnableCORS(handler.ConfirmRetry)(w, r)
	case strings.HasPrefix(path, "/api/jobs/stream"):
		handler.EnableCORS(handler.StreamJobTimeline)(w, r)
	case strings.HasPrefix(path, "/api/jobs/"):
		handler.EnableCORS(handler.GetJobDetail)(w, r)
	case path == "/api/copilot/chat" || path == "/api/copilot/chat/":
		handler.EnableCORS(handler.CopilotChat)(w, r)
	default:
		http.Error(w, "Endpoint not found", http.StatusNotFound)
	}
}
