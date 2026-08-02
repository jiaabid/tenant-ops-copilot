package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"tenant-copilot-backend/api"
	"tenant-copilot-backend/copilot"
	"tenant-copilot-backend/crypto"
	"tenant-copilot-backend/store"
	"tenant-copilot-backend/worker"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	secretKey := os.Getenv("HMAC_SECRET_KEY")
	if secretKey == "" {
		secretKey = "copilot-secure-gated-remediation-secret-key"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "tenant_copilot.db"
	}

	dbStore, err := store.NewDBStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize SQLite database (%s): %v", dbPath, err)
	}
	defer dbStore.Close()

	proposalSigner := crypto.NewProposalSigner(secretKey)
	copilotAgent := copilot.NewAgent(dbStore, proposalSigner)
	workerSim := worker.NewWorker(dbStore)

	handler := api.NewAPIHandler(dbStore, proposalSigner, copilotAgent, workerSim)

	mux := http.NewServeMux()

	// REST Endpoints
	mux.HandleFunc("/api/jobs", handler.EnableCORS(handler.GetJobs))
	mux.HandleFunc("/api/jobs/", handler.EnableCORS(handler.GetJobDetail))
	mux.HandleFunc("/api/copilot/chat", handler.EnableCORS(handler.CopilotChat))
	mux.HandleFunc("/api/jobs/retry/confirm", handler.EnableCORS(handler.ConfirmRetry))
	mux.HandleFunc("/api/jobs/stream", handler.EnableCORS(handler.StreamJobTimeline))

	log.Printf("==========================================================================")
	log.Printf("🚀 Tenant Copilot Backend Server running on port :%s", port)
	log.Printf("   SQLite Storage: Active (%s)", dbPath)
	log.Printf("   HMAC Signer Secret: Active (5-Minute Expiring Proposal Cards)")
	log.Printf("   Copilot Read-Only Tools: get_job_timeline, get_job_error")
	log.Printf("   Copilot Remediation Proposal Tool: propose_job_retry (Signed Card)")
	log.Printf("==========================================================================")

	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
