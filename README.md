# Tenant Copilot — Incident Diagnostics & Human-Gated Remediation Workbench

> **An AI DevOps Copilot for Cloud Multi-Tenant Provisioning & Recovery**

**Tenant Copilot** is a full-stack enterprise platform built for cloud onboarding pipelines. When tenant provisioning jobs fail after exhausting retries, an **AI Copilot** reads raw event history via read-only tools and explains the root cause in plain English. 

To safeguard production systems, the AI **does not apply mutations directly**. Instead, it generates a **cryptographically signed (HMAC-SHA256), 5-minute expiring action proposal card**. Upon human operator confirmation, the Go backend re-queues the job and streams live step progress to the dashboard via Server-Sent Events (SSE).

---

## ✨ Key Features

- **🤖 AI Incident Diagnostics (Gemini Powered)**: Reads unredacted raw step logs via read-only tools (`get_job_timeline`, `get_job_error`) and synthesizes root cause explanations.
- **🛡️ Human-Gated Remediation**: Action proposals (`propose_job_retry`) produce HMAC-SHA256 signed cards with 300-second expiration windows to prevent unauthorized or automated mutations.
- **⚡ Real-Time SSE Streaming**: Live progress stream as the worker re-executes all 5 provisioning steps.
- **🗄️ SQLite Persistent Storage**: Built with pure Go SQLite (`modernc.org/sqlite`) for persistent job, subscription, and step event storage (`tenant_copilot.db`).
- **🎨 Master-Detail Dashboard**: Master list of failure scenarios on the left; instant step timeline visualization on the right.
- **💬 Floating AI Drawer**: Minimizable floating DevOps Copilot tab in the bottom-right corner.

---

## 🛠️ Technology Stack

| Layer | Technology |
| :--- | :--- |
| **Backend** | Go (Golang) — REST API, Async Worker, SSE Engine, HMAC Signer |
| **Frontend** | Angular 19 (Standalone Components, RxJS, EventSource) |
| **AI Model** | Google Gemini API (`gemini-2.5-flash` / `gemini-1.5-flash`) |
| **Database** | SQLite (`modernc.org/sqlite` pure Go driver) |
| **Security** | HMAC-SHA256 Token Verification with 5-minute TTL |

---

## 🚀 Quick Start Guide

### Prerequisites
- [Go 1.22+](https://go.dev/dl/)
- [Node.js 18+](https://nodejs.org/) & npm

---

### 1. Run Backend Server (Go)

```bash
cd backend
go run main.go
```
*Backend server runs at `http://localhost:8080` and initializes SQLite database `tenant_copilot.db`.*

> **Optional**: Set your Gemini API key:
> ```bash
> export GEMINI_API_KEY="your-google-gemini-api-key"
> ```

---

### 2. Run Frontend Dashboard (Angular 19)

```bash
cd frontend
npm install
npx ng serve --port 4200
```
*Frontend opens at `http://localhost:4200`.*

---

## 📂 Project Structure

```
tenant-copilot/
├── backend/
│   ├── api/          # REST & SSE HTTP Handlers
│   ├── copilot/      # Gemini AI Agent & Tool Traces
│   ├── crypto/       # HMAC-SHA256 Proposal Signer & Verifier
│   ├── domain/       # Core Domain Structs & Event Models
│   ├── store/        # SQLite Persistent Storage (`db.go`)
│   ├── worker/       # Async Job Simulator & SSE Broadcaster
│   ├── main.go       # Go Main Server Entrypoint
│   └── go.mod
│
└── frontend/
    ├── src/app/
    │   ├── components/
    │   │   ├── confirmation-card/  # Signed Card Widget with Countdown
    │   │   ├── copilot/            # Floating AI Copilot Drawer
    │   │   ├── job-detail/         # Step Timeline Tree Visualizer
    │   │   └── job-list/           # Incident Scenarios Grid
    │   ├── models/                 # TypeScript Interfaces
    │   └── services/               # REST HTTP Client & SSE Subscriber
    └── package.json
```

---

## 🔐 Security Governance Architecture

```
[ Failed Job ] ➔ [ AI Copilot (Read-Only Path) ]
                          │
                Generates Signed Proposal
                          ▼
            [ 🔒 HMAC-SHA256 Card (5-Min TTL) ]
                          │
                   Human Approval
                          ▼
            [ Go Worker Re-queues Pipeline ] ➔ [ SSE Live Stream ]
```

---

## 📄 License

MIT License. Built for Enterprise Tenant Provisioning & AI Governance.
