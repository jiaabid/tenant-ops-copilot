import { Component, Input, Output, EventEmitter, OnInit, OnDestroy, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SignedProposal } from '../../models/job.model';
import { JobService } from '../../services/job.service';

@Component({
  selector: 'app-confirmation-card',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="confirmation-card" [class.expired]="isExpired" [class.executed]="isExecuted">
      <!-- Top Security Header -->
      <div class="card-header">
        <div class="header-title">
          <span class="shield-icon">🛡️</span>
          <div>
            <h4>Human Approval Required</h4>
            <span class="signature-tag">HMAC-SHA256 Signed Proposal</span>
          </div>
        </div>

        <div class="timer-badge" [class.danger]="secondsLeft <= 60" *ngIf="!isExecuted && !isExpired">
          <span class="timer-icon">⏳</span>
          <span>Expires in {{ formattedTime }}</span>
        </div>

        <div class="timer-badge expired-badge" *ngIf="isExpired">
          <span>⚠️ PROPOSAL EXPIRED</span>
        </div>

        <div class="timer-badge success-badge" *ngIf="isExecuted">
          <span>✅ APPROVED & QUEUED</span>
        </div>
      </div>

      <!-- Action Details -->
      <div class="card-body">
        <div class="proposal-details">
          <div class="detail-row">
            <span class="label">Target Tenant:</span>
            <span class="val font-highlight">{{ proposal.tenant_name }} ({{ proposal.subscription_id }})</span>
          </div>
          <div class="detail-row">
            <span class="label">Proposed Action:</span>
            <span class="val action-code">{{ proposal.action }} ➔ {{ proposal.target_state }}</span>
          </div>
          <div class="detail-row">
            <span class="label">Effect:</span>
            <span class="val description">{{ proposal.what_rerunning_does }}</span>
          </div>
        </div>

        <!-- HMAC Security Metadata Token -->
        <div class="security-meta">
          <div class="meta-item">
            <span class="meta-label">Issued At:</span>
            <span class="meta-val mono">{{ proposal.issued_at | date:'HH:mm:ss' }}</span>
          </div>
          <div class="meta-item">
            <span class="meta-label">Nonce:</span>
            <span class="meta-val mono">{{ proposal.nonce }}</span>
          </div>
          <div class="meta-item signature-item">
            <span class="meta-label">HMAC Signature:</span>
            <span class="meta-val mono signature-hash">{{ proposal.signature }}</span>
          </div>
        </div>

        <!-- Error feedback if confirmation fails -->
        <div class="error-banner" *ngIf="errorMessage">
          🚨 {{ errorMessage }}
        </div>
      </div>

      <!-- Footer Action Buttons -->
      <div class="card-footer" *ngIf="!isExecuted && !isExpired">
        <button 
          class="approve-btn" 
          [disabled]="isSubmitting" 
          (click)="confirmProposal()">
          <span class="btn-spinner" *ngIf="isSubmitting"></span>
          <span>{{ isSubmitting ? 'Verifying Signature...' : '🔒 Approve & Re-run Job' }}</span>
        </button>
      </div>
    </div>
  `,
  styles: [`
    .confirmation-card {
      background: linear-gradient(135deg, rgba(30, 41, 59, 0.95), rgba(15, 23, 42, 0.95));
      border: 1.5px solid rgba(99, 102, 241, 0.4);
      border-radius: 12px;
      overflow: hidden;
      box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
      margin-top: 0.85rem;
    }
    .confirmation-card.expired {
      border-color: rgba(239, 68, 68, 0.4);
      opacity: 0.8;
    }
    .confirmation-card.executed {
      border-color: rgba(16, 185, 129, 0.4);
    }
    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 0.75rem 1rem;
      background: rgba(30, 41, 59, 0.9);
      border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    }
    .header-title {
      display: flex;
      align-items: center;
      gap: 0.5rem;
    }
    .shield-icon { font-size: 1.2rem; }
    .header-title h4 {
      margin: 0;
      font-size: 0.88rem;
      color: #f8fafc;
    }
    .signature-tag {
      font-size: 0.68rem;
      color: #818cf8;
    }
    .timer-badge {
      display: inline-flex;
      align-items: center;
      gap: 0.3rem;
      background: rgba(99, 102, 241, 0.15);
      border: 1px solid rgba(99, 102, 241, 0.3);
      color: #818cf8;
      padding: 0.25rem 0.6rem;
      border-radius: 20px;
      font-size: 0.72rem;
      font-weight: 700;
    }
    .timer-badge.danger {
      background: rgba(239, 68, 68, 0.15);
      border-color: rgba(239, 68, 68, 0.3);
      color: #f87171;
    }
    .expired-badge {
      background: rgba(239, 68, 68, 0.2);
      color: #f87171;
    }
    .success-badge {
      background: rgba(16, 185, 129, 0.2);
      color: #34d399;
    }
    .card-body {
      padding: 0.85rem 1rem;
    }
    .proposal-details {
      display: flex;
      flex-direction: column;
      gap: 0.35rem;
      margin-bottom: 0.75rem;
      font-size: 0.82rem;
    }
    .detail-row {
      display: flex;
      gap: 0.5rem;
    }
    .label {
      color: #64748b;
      width: 110px;
      flex-shrink: 0;
    }
    .val { color: #e2e8f0; }
    .font-highlight { font-weight: 600; color: #f1f5f9; }
    .action-code {
      font-family: monospace;
      color: #38bdf8;
      background: rgba(56, 189, 248, 0.1);
      padding: 0.1rem 0.35rem;
      border-radius: 4px;
    }
    .description {
      font-size: 0.78rem;
      color: #cbd5e1;
    }
    .security-meta {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 0.5rem;
      background: #020617;
      border: 1px solid rgba(255, 255, 255, 0.08);
      border-radius: 6px;
      padding: 0.5rem;
      font-size: 0.7rem;
    }
    .meta-item {
      display: flex;
      flex-direction: column;
    }
    .meta-label { color: #64748b; }
    .meta-val { color: #94a3b8; }
    .mono { font-family: monospace; }
    .signature-hash {
      color: #a78bfa;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
    .error-banner {
      margin-top: 0.6rem;
      background: rgba(239, 68, 68, 0.2);
      border: 1px solid rgba(239, 68, 68, 0.4);
      color: #fca5a5;
      padding: 0.4rem 0.6rem;
      border-radius: 6px;
      font-size: 0.78rem;
    }
    .card-footer {
      padding: 0.6rem 1rem;
      background: rgba(15, 23, 42, 0.8);
      border-top: 1px solid rgba(255, 255, 255, 0.06);
    }
    .approve-btn {
      width: 100%;
      background: linear-gradient(135deg, #10b981, #059669);
      color: #ffffff;
      border: none;
      padding: 0.6rem;
      border-radius: 8px;
      font-size: 0.85rem;
      font-weight: 700;
      cursor: pointer;
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 0.4rem;
      box-shadow: 0 0 15px rgba(16, 185, 129, 0.3);
      transition: all 0.2s;
    }
    .approve-btn:hover {
      background: linear-gradient(135deg, #059669, #047857);
      box-shadow: 0 0 20px rgba(16, 185, 129, 0.5);
    }
    .approve-btn:disabled {
      opacity: 0.6;
      cursor: not-allowed;
    }
    .btn-spinner {
      width: 12px; height: 12px;
      border: 2px solid rgba(255, 255, 255, 0.3);
      border-top-color: #ffffff;
      border-radius: 50%;
      animation: spin 0.8s linear infinite;
    }
    @keyframes spin { to { transform: rotate(360deg); } }
  `]
})
export class ConfirmationCardComponent implements OnInit, OnDestroy {
  @Input() proposal!: SignedProposal;
  @Output() confirmed = new EventEmitter<void>();

  private jobService = inject(JobService);

  secondsLeft: number = 0;
  isExpired: boolean = false;
  isExecuted: boolean = false;
  isSubmitting: boolean = false;
  errorMessage: string = '';
  private timerInterval?: any;

  ngOnInit() {
    this.calculateTimeLeft();
    this.timerInterval = setInterval(() => {
      this.calculateTimeLeft();
    }, 1000);
  }

  ngOnDestroy() {
    if (this.timerInterval) {
      clearInterval(this.timerInterval);
    }
  }

  get formattedTime(): string {
    const mins = Math.floor(this.secondsLeft / 60);
    const secs = this.secondsLeft % 60;
    return `${mins}:${secs < 10 ? '0' : ''}${secs}`;
  }

  private calculateTimeLeft() {
    if (!this.proposal || !this.proposal.expires_at) return;
    const now = Math.floor(Date.now() / 1000);
    const diff = this.proposal.expires_at - now;

    if (diff <= 0) {
      this.secondsLeft = 0;
      this.isExpired = true;
      if (this.timerInterval) {
        clearInterval(this.timerInterval);
      }
    } else {
      this.secondsLeft = diff;
    }
  }

  confirmProposal() {
    if (this.isExpired || this.isSubmitting || this.isExecuted) return;

    this.isSubmitting = true;
    this.errorMessage = '';

    this.jobService.confirmRetry(this.proposal).subscribe({
      next: () => {
        this.isSubmitting = false;
        this.isExecuted = true;
        this.confirmed.emit();
      },
      error: (err) => {
        this.isSubmitting = false;
        this.errorMessage = err.error?.error || 'HMAC verification failed or card expired.';
      }
    });
  }
}
