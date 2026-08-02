import { Component, Input, Output, EventEmitter, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ProvisioningJob, CopilotChatResponse, ToolTrace } from '../../models/job.model';
import { JobService } from '../../services/job.service';
import { ConfirmationCardComponent } from '../confirmation-card/confirmation-card.component';

interface ChatMessage {
  sender: 'operator' | 'copilot';
  text: string;
  toolTraces?: ToolTrace[];
  rawTimelineReturned?: boolean;
  proposal?: any;
  timestamp: Date;
}

@Component({
  selector: 'app-copilot',
  standalone: true,
  imports: [CommonModule, FormsModule, ConfirmationCardComponent],
  template: `
    <div class="copilot-panel">
      <!-- Copilot Header -->
      <div class="copilot-header">
        <div class="agent-title">
          <div class="agent-avatar">🤖</div>
          <div>
            <h3>DevOps Copilot Assistant</h3>
            <span class="agent-subtitle">Gemini Powered Incident Diagnostics & Remediation</span>
          </div>
        </div>
        
        <div class="header-actions">
          <span class="read-only-badge">🛡️ Read-Only Path</span>
          <button class="close-btn" (click)="closeCopilot.emit()" title="Minimize Floating Copilot">✕</button>
        </div>
      </div>

      <!-- Messages Thread -->
      <div class="messages-container">
        <div *ngIf="messages.length === 0" class="empty-state">
          <div class="empty-icon">🔍</div>
          <h4>Ask Copilot what went wrong</h4>
          <p>Copilot reads raw event history via read-only tools and interprets root cause in plain language.</p>
          <div class="quick-prompts">
            <button (click)="sendQuery('What went wrong with this provisioning job?')">
              "What went wrong with this provisioning job?"
            </button>
            <button (click)="sendQuery('Explain the failure, last error, and whether it is transient or structural.')">
              "Is this error transient or structural?"
            </button>
          </div>
        </div>

        <div *ngFor="let msg of messages" class="msg-row" [class.user-msg]="msg.sender === 'operator'" [class.copilot-msg]="msg.sender === 'copilot'">
          
          <!-- Avatar -->
          <div class="msg-avatar">
            {{ msg.sender === 'operator' ? '👤' : '🤖' }}
          </div>

          <div class="msg-bubble">
            <div class="msg-meta">
              <span class="sender-name">{{ msg.sender === 'operator' ? 'Operator' : 'DevOps Copilot' }}</span>
              <span class="time">{{ msg.timestamp | date:'HH:mm:ss' }}</span>
            </div>

            <!-- Message Text -->
            <div class="msg-content" [innerHTML]="formatMarkdown(msg.text)"></div>

            <!-- Tool Call Trace Inspector -->
            <div *ngIf="msg.toolTraces && msg.toolTraces.length > 0" class="tool-trace-inspector">
              <div class="trace-header" (click)="toggleTraces(msg)">
                <span class="trace-icon">🔧</span>
                <span>Copilot Read-Only Tool Traces ({{ msg.toolTraces.length }})</span>
                <span class="raw-badge" *ngIf="msg.rawTimelineReturned">RAW TIMELINE FETCHED</span>
              </div>
              
              <div class="trace-body" *ngIf="showTraces">
                <div *ngFor="let t of msg.toolTraces" class="trace-item">
                  <div class="trace-name">Call: <code>{{ t.tool_name }}()</code></div>
                  <pre class="trace-json"><code>{{ t.result | json }}</code></pre>
                </div>
              </div>
            </div>

            <!-- Signed Confirmation Card Widget -->
            <div *ngIf="msg.proposal" class="proposal-wrapper">
              <app-confirmation-card 
                [proposal]="msg.proposal" 
                (confirmed)="onProposalConfirmed()">
              </app-confirmation-card>
            </div>

          </div>

        </div>

        <!-- Thinking Spinner -->
        <div *ngIf="isLoading" class="msg-row copilot-msg">
          <div class="msg-avatar">🤖</div>
          <div class="msg-bubble thinking-bubble">
            <span class="thinking-spinner"></span>
            <span>Copilot is querying read-only tools & analyzing raw timeline events...</span>
          </div>
        </div>
      </div>

      <!-- Input Bar -->
      <div class="input-container">
        <input 
          type="text" 
          [(ngModel)]="userQuery" 
          (keyup.enter)="onSubmit()"
          [disabled]="isLoading"
          placeholder="Ask copilot about job failure, error logs, or remediation..." />
        <button class="send-btn" [disabled]="!userQuery.trim() || isLoading" (click)="onSubmit()">
          Send ➔
        </button>
      </div>
    </div>
  `,
  styles: [`
    .copilot-panel {
      display: flex;
      flex-direction: column;
      height: 100%;
      background: rgba(15, 23, 42, 0.95);
      border: 1.5px solid rgba(99, 102, 241, 0.35);
      border-radius: 16px;
      overflow: hidden;
      box-shadow: 0 20px 50px rgba(0, 0, 0, 0.6), 0 0 30px rgba(99, 102, 241, 0.25);
      backdrop-filter: blur(20px);
    }
    .copilot-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 0.85rem 1.1rem;
      background: rgba(30, 41, 59, 0.9);
      border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    }
    .agent-title {
      display: flex;
      align-items: center;
      gap: 0.65rem;
    }
    .agent-avatar {
      font-size: 1.35rem;
      background: rgba(99, 102, 241, 0.2);
      padding: 0.35rem;
      border-radius: 8px;
    }
    .agent-title h3 {
      margin: 0;
      font-size: 0.95rem;
      color: #f8fafc;
    }
    .agent-subtitle {
      font-size: 0.7rem;
      color: #94a3b8;
    }
    .header-actions {
      display: flex;
      align-items: center;
      gap: 0.6rem;
    }
    .read-only-badge {
      font-size: 0.68rem;
      color: #38bdf8;
      background: rgba(56, 189, 248, 0.12);
      border: 1px solid rgba(56, 189, 248, 0.25);
      padding: 0.2rem 0.5rem;
      border-radius: 20px;
      font-weight: 600;
    }
    .close-btn {
      background: rgba(255, 255, 255, 0.06);
      border: 1px solid rgba(255, 255, 255, 0.1);
      color: #cbd5e1;
      width: 26px; height: 26px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      cursor: pointer;
      font-size: 0.8rem;
      transition: all 0.2s;
    }
    .close-btn:hover {
      background: rgba(239, 68, 68, 0.2);
      color: #f87171;
      border-color: rgba(239, 68, 68, 0.4);
    }
    .messages-container {
      flex: 1;
      overflow-y: auto;
      padding: 1rem;
      display: flex;
      flex-direction: column;
      gap: 1rem;
    }
    .empty-state {
      text-align: center;
      margin: auto;
      max-width: 360px;
      color: #94a3b8;
    }
    .empty-icon { font-size: 2.2rem; margin-bottom: 0.4rem; }
    .empty-state h4 { color: #f1f5f9; margin: 0 0 0.3rem 0; }
    .empty-state p { font-size: 0.8rem; line-height: 1.4; margin: 0 0 1rem 0; }
    .quick-prompts {
      display: flex;
      flex-direction: column;
      gap: 0.45rem;
    }
    .quick-prompts button {
      background: rgba(30, 41, 59, 0.7);
      border: 1px solid rgba(99, 102, 241, 0.3);
      color: #cbd5e1;
      padding: 0.5rem 0.75rem;
      border-radius: 8px;
      font-size: 0.78rem;
      cursor: pointer;
      text-align: left;
      transition: all 0.2s;
    }
    .quick-prompts button:hover {
      background: rgba(99, 102, 241, 0.2);
      border-color: #818cf8;
      color: #ffffff;
    }
    .msg-row {
      display: flex;
      gap: 0.75rem;
      max-width: 92%;
    }
    .user-msg {
      margin-left: auto;
      flex-direction: row-reverse;
    }
    .msg-avatar {
      font-size: 1.1rem;
      width: 28px; height: 28px;
      display: flex;
      align-items: center;
      justify-content: center;
      background: rgba(255, 255, 255, 0.05);
      border-radius: 50%;
      flex-shrink: 0;
    }
    .msg-bubble {
      background: rgba(30, 41, 59, 0.85);
      border: 1px solid rgba(255, 255, 255, 0.08);
      border-radius: 12px;
      padding: 0.75rem 0.85rem;
      color: #e2e8f0;
      font-size: 0.85rem;
      line-height: 1.45;
    }
    .user-msg .msg-bubble {
      background: linear-gradient(135deg, #4f46e5, #4338ca);
      border-color: transparent;
      color: #ffffff;
    }
    .msg-meta {
      display: flex;
      justify-content: space-between;
      gap: 1rem;
      margin-bottom: 0.35rem;
      font-size: 0.7rem;
      opacity: 0.75;
    }
    .tool-trace-inspector {
      margin-top: 0.65rem;
      background: rgba(15, 23, 42, 0.9);
      border: 1px solid rgba(255, 255, 255, 0.1);
      border-radius: 8px;
      overflow: hidden;
    }
    .trace-header {
      padding: 0.45rem 0.65rem;
      font-size: 0.72rem;
      color: #818cf8;
      cursor: pointer;
      display: flex;
      align-items: center;
      gap: 0.4rem;
    }
    .raw-badge {
      margin-left: auto;
      background: rgba(16, 185, 129, 0.2);
      color: #34d399;
      font-size: 0.62rem;
      padding: 0.1rem 0.35rem;
      border-radius: 4px;
    }
    .trace-body {
      padding: 0.65rem;
      border-top: 1px solid rgba(255, 255, 255, 0.06);
      max-height: 180px;
      overflow-y: auto;
    }
    .trace-item { margin-bottom: 0.4rem; }
    .trace-name { font-size: 0.7rem; color: #cbd5e1; }
    .trace-json {
      margin: 0.2rem 0 0 0;
      font-family: monospace;
      font-size: 0.68rem;
      color: #a5b4fc;
      background: #020617;
      padding: 0.35rem;
      border-radius: 4px;
      white-space: pre-wrap;
    }
    .thinking-bubble {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      color: #94a3b8;
      font-size: 0.82rem;
    }
    .thinking-spinner {
      width: 12px; height: 12px;
      border: 2px solid rgba(99, 102, 241, 0.3);
      border-top-color: #818cf8;
      border-radius: 50%;
      animation: spin 0.8s linear infinite;
    }
    .input-container {
      display: flex;
      gap: 0.5rem;
      padding: 0.75rem 1rem;
      background: rgba(30, 41, 59, 0.9);
      border-top: 1px solid rgba(255, 255, 255, 0.08);
    }
    .input-container input {
      flex: 1;
      background: rgba(15, 23, 42, 0.7);
      border: 1px solid rgba(255, 255, 255, 0.1);
      border-radius: 8px;
      padding: 0.55rem 0.75rem;
      color: #f8fafc;
      font-size: 0.85rem;
    }
    .input-container input:focus {
      outline: none;
      border-color: #6366f1;
    }
    .send-btn {
      background: #6366f1;
      color: #ffffff;
      border: none;
      padding: 0.55rem 1rem;
      border-radius: 8px;
      font-weight: 600;
      font-size: 0.82rem;
      cursor: pointer;
    }
    .send-btn:disabled {
      background: #334155;
      color: #64748b;
      cursor: not-allowed;
    }
    @keyframes spin { to { transform: rotate(360deg); } }
  `]
})
export class CopilotComponent {
  @Input() selectedJob?: ProvisioningJob;
  @Output() retryExecuted = new EventEmitter<void>();
  @Output() closeCopilot = new EventEmitter<void>();

  private jobService = inject(JobService);
  userQuery: string = '';
  isLoading: boolean = false;
  messages: ChatMessage[] = [];
  showTraces: boolean = true;

  sendQuery(text: string) {
    this.userQuery = text;
    this.onSubmit();
  }

  toggleTraces(msg: ChatMessage) {
    this.showTraces = !this.showTraces;
  }

  onSubmit() {
    if (!this.userQuery.trim() || this.isLoading) return;

    const q = this.userQuery.trim();
    this.userQuery = '';

    const jobId = this.selectedJob?.id || 'job-101';

    this.messages.push({
      sender: 'operator',
      text: q,
      timestamp: new Date()
    });

    this.isLoading = true;

    this.jobService.askCopilot(jobId, q).subscribe({
      next: (res) => {
        this.isLoading = false;
        this.messages.push({
          sender: 'copilot',
          text: res.explanation,
          toolTraces: res.tool_traces,
          rawTimelineReturned: res.raw_timeline_returned,
          proposal: res.proposal,
          timestamp: new Date()
        });
      },
      error: () => {
        this.isLoading = false;
        this.messages.push({
          sender: 'copilot',
          text: '⚠️ Unable to contact DevOps Copilot agent. Please check backend connection.',
          timestamp: new Date()
        });
      }
    });
  }

  onProposalConfirmed() {
    this.retryExecuted.emit();
  }

  formatMarkdown(text: string): string {
    if (!text) return '';
    return text
      .replace(/### (.*)/g, '<h4 style="margin:0.4rem 0;color:#818cf8;">$1</h4>')
      .replace(/ - \*\*(.*)\*\*: (.*)/g, '<li><strong>$1</strong>: $2</li>')
      .replace(/\*\*(.*)\*\*/g, '<strong>$1</strong>')
      .replace(/`([^`]+)`/g, '<code style="background:rgba(0,0,0,0.4);padding:0.15rem 0.35rem;border-radius:4px;color:#a78bfa;">$1</code>')
      .replace(/\n/g, '<br/>');
  }
}
