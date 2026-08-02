import { Component, Input, OnInit, OnChanges, SimpleChanges, OnDestroy, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Subscription as RxSubscription } from 'rxjs';
import { ProvisioningJob, StepEvent, Subscription } from '../../models/job.model';
import { JobService } from '../../services/job.service';

@Component({
  selector: 'app-job-detail',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="timeline-container" *ngIf="job">
      <!-- Header Summary -->
      <div class="header-card">
        <div class="header-main">
          <div class="tenant-meta">
            <h2>{{ job.tenant_name }}</h2>
            <div class="meta-pills">
              <span class="pill mono">Job ID: {{ job.id }}</span>
              <span class="pill sub-pill">Account Sub: {{ job.subscription_id }}</span>
              <span class="pill region-pill">Cloud Region: {{ job.region }}</span>
            </div>
          </div>

          <div class="job-status-badge" [ngClass]="'status-' + job.state">
            <span class="status-pulse" *ngIf="job.state === 'running'"></span>
            <span class="status-icon">
              <ng-container [ngSwitch]="job.state">
                <span *ngSwitchCase="'failed'">🔴</span>
                <span *ngSwitchCase="'running'">⚡</span>
                <span *ngSwitchCase="'queued'">⏳</span>
                <span *ngSwitchCase="'succeeded'">✅</span>
              </ng-container>
            </span>
            <span class="status-text">{{ job.state === 'failed' ? 'SETUP FAILED' : job.state === 'running' ? 'RUNNING NOW' : job.state === 'queued' ? 'QUEUED TO START' : 'SUCCESSFULLY ONBOARDED' }}</span>
          </div>
        </div>

        <div class="job-metrics">
          <div class="metric">
            <span class="metric-label">Retry Attempts</span>
            <span class="metric-val">{{ job.current_attempt }} of {{ job.max_retries }} max retries used</span>
          </div>
          <div class="metric">
            <span class="metric-label">Error Diagnosis Type</span>
            <span class="metric-val" [class.transient]="job.error_category === 'transient'" [class.structural]="job.error_category === 'structural'">
              {{ job.error_category === 'transient' ? '⚡ Temporary Server Glitch' : job.error_category === 'structural' ? '⚠️ Configuration Conflict' : 'None' }}
            </span>
          </div>
          <div class="metric">
            <span class="metric-label">Customer Subscription</span>
            <span class="metric-val sub-status" [class.failed-status]="sub?.status === 'failed'">
              {{ (sub?.status || 'unknown') | uppercase }}
            </span>
          </div>
        </div>
      </div>

      <!-- Steps Timeline Tree -->
      <div class="timeline-tree">
        <div class="tree-header">
          <div>
            <h3>Step-by-Step Setup Pipeline History</h3>
            <span class="sub-tree-text">Shows the 5 automated steps required to set up this customer</span>
          </div>
          
          <span class="stream-indicator" *ngIf="isLiveStreaming">
            <span class="live-dot"></span> STREAMING LIVE PROGRESS
          </span>
        </div>

        <div class="steps-list">
          <div 
            *ngFor="let step of displaySteps; let i = index" 
            class="step-card"
            [ngClass]="'step-status-' + step.status">
            
            <div class="step-connector" *ngIf="i < displaySteps.length - 1"></div>

            <div class="step-header">
              <div class="step-num-badge">Step {{ step.step_index }} of 5</div>
              
              <div class="step-title-group">
                <h4 class="step-name">{{ step.step_name }}</h4>
                <div class="step-meta">
                  <span class="attempt-tag">Attempt {{ step.attempt }} of {{ step.max_attempts }}</span>
                  <span class="duration-tag" *ngIf="step.duration_ms > 0">Took {{ step.duration_ms }}ms</span>
                  <span class="timestamp-tag">{{ step.timestamp | date:'HH:mm:ss' }}</span>
                </div>
              </div>

              <div class="step-state-pill" [ngClass]="'pill-' + step.status">
                <span *ngIf="step.status === 'running'" class="small-spinner"></span>
                {{ step.status === 'succeeded' ? '✅ COMPLETED' : step.status === 'failed' ? '❌ FAILED' : step.status === 'running' ? '⚡ RUNNING' : '⏳ PENDING' }}
              </div>
            </div>

            <!-- Layman Error Box -->
            <div class="error-box" *ngIf="step.status === 'failed' && step.error_message">
              <div class="error-header">
                <span class="error-icon">🚨</span>
                <span class="error-title">Step {{ step.step_index }} Error Details</span>
                <span class="err-cat-badge" [class.transient]="step.error_category === 'transient'">
                  {{ step.error_category === 'transient' ? '⚡ Temporary Glitch' : '⚠️ Setting Conflict' }}
                </span>
              </div>
              
              <div class="error-msg"><strong>Error Message:</strong> {{ step.error_message }}</div>
              
              <div class="layman-translation">
                <strong>💡 Layman Summary:</strong> 
                <span *ngIf="step.error_category === 'transient'">
                  The cloud server timed out while creating security keys. Re-running the setup will likely fix this automatically.
                </span>
                <span *ngIf="step.error_category === 'structural'">
                  The custom domain address requested by this customer is already claimed by another company. The address setting needs to be updated.
                </span>
              </div>

              <div class="raw-payload" *ngIf="step.raw_payload">
                <span class="payload-label">Raw Developer Technical Log (Unmodified):</span>
                <pre><code>{{ step.raw_payload }}</code></pre>
              </div>
            </div>

          </div>
        </div>
      </div>

    </div>
  `,
  styles: [`
    .timeline-container {
      display: flex;
      flex-direction: column;
      gap: 1.5rem;
    }
    .header-card {
      background: rgba(30, 41, 59, 0.6);
      border: 1px solid rgba(255, 255, 255, 0.08);
      border-radius: 14px;
      padding: 1.25rem;
      backdrop-filter: blur(12px);
    }
    .header-main {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      margin-bottom: 1rem;
    }
    .tenant-meta h2 {
      margin: 0 0 0.4rem 0;
      font-size: 1.35rem;
      color: #f8fafc;
      font-weight: 700;
    }
    .meta-pills {
      display: flex;
      gap: 0.5rem;
    }
    .pill {
      font-size: 0.75rem;
      background: rgba(255, 255, 255, 0.05);
      border: 1px solid rgba(255, 255, 255, 0.1);
      padding: 0.2rem 0.5rem;
      border-radius: 6px;
      color: #94a3b8;
    }
    .mono { font-family: monospace; }
    .job-status-badge {
      display: inline-flex;
      align-items: center;
      gap: 0.4rem;
      padding: 0.4rem 0.85rem;
      border-radius: 30px;
      font-size: 0.85rem;
      font-weight: 700;
      letter-spacing: 0.04em;
    }
    .status-failed {
      background: rgba(239, 68, 68, 0.15);
      color: #f87171;
      border: 1px solid rgba(239, 68, 68, 0.3);
    }
    .status-running {
      background: rgba(245, 158, 11, 0.15);
      color: #fbbf24;
      border: 1px solid rgba(245, 158, 11, 0.3);
    }
    .status-queued {
      background: rgba(56, 189, 248, 0.15);
      color: #38bdf8;
      border: 1px solid rgba(56, 189, 248, 0.3);
    }
    .status-succeeded {
      background: rgba(16, 185, 129, 0.15);
      color: #34d399;
      border: 1px solid rgba(16, 185, 129, 0.3);
    }
    .status-pulse {
      width: 8px; height: 8px;
      border-radius: 50%;
      background: #fbbf24;
      box-shadow: 0 0 10px #fbbf24;
      animation: pulse 1s infinite;
    }
    .job-metrics {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 1rem;
      border-top: 1px solid rgba(255, 255, 255, 0.06);
      padding-top: 0.85rem;
    }
    .metric {
      display: flex;
      flex-direction: column;
      gap: 0.2rem;
    }
    .metric-label {
      font-size: 0.7rem;
      color: #64748b;
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }
    .metric-val {
      font-size: 0.88rem;
      font-weight: 600;
      color: #e2e8f0;
    }
    .metric-val.transient { color: #38bdf8; }
    .metric-val.structural { color: #f43f5e; }
    .failed-status { color: #f87171; }

    .timeline-tree {
      background: rgba(15, 23, 42, 0.4);
      border: 1px solid rgba(255, 255, 255, 0.06);
      border-radius: 14px;
      padding: 1.25rem;
    }
    .tree-header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      margin-bottom: 1.25rem;
    }
    .tree-header h3 {
      margin: 0 0 0.2rem 0;
      font-size: 1.05rem;
      color: #cbd5e1;
    }
    .sub-tree-text { font-size: 0.75rem; color: #94a3b8; }
    .stream-indicator {
      display: inline-flex;
      align-items: center;
      gap: 0.4rem;
      background: rgba(56, 189, 248, 0.15);
      color: #38bdf8;
      padding: 0.25rem 0.6rem;
      border-radius: 12px;
      font-size: 0.72rem;
      font-weight: 700;
    }
    .live-dot {
      width: 6px; height: 6px;
      border-radius: 50%;
      background: #38bdf8;
      box-shadow: 0 0 8px #38bdf8;
      animation: pulse 1s infinite;
    }
    .steps-list {
      display: flex;
      flex-direction: column;
      gap: 1rem;
      position: relative;
    }
    .step-card {
      position: relative;
      background: rgba(30, 41, 59, 0.5);
      border: 1px solid rgba(255, 255, 255, 0.06);
      border-radius: 10px;
      padding: 1rem;
      transition: all 0.2s ease;
    }
    .step-status-failed {
      border-color: rgba(239, 68, 68, 0.35);
      background: rgba(239, 68, 68, 0.04);
    }
    .step-status-succeeded {
      border-color: rgba(16, 185, 129, 0.25);
    }
    .step-status-running {
      border-color: rgba(245, 158, 11, 0.4);
      background: rgba(245, 158, 11, 0.05);
    }
    .step-connector {
      position: absolute;
      left: 1.65rem;
      bottom: -1rem;
      width: 2px;
      height: 1rem;
      background: rgba(255, 255, 255, 0.1);
      z-index: 1;
    }
    .step-header {
      display: flex;
      align-items: center;
      gap: 0.85rem;
    }
    .step-num-badge {
      background: rgba(255, 255, 255, 0.06);
      color: #94a3b8;
      border: 1px solid rgba(255, 255, 255, 0.1);
      padding: 0.3rem 0.6rem;
      border-radius: 8px;
      font-size: 0.78rem;
      font-weight: 700;
    }
    .step-title-group {
      flex: 1;
    }
    .step-name {
      margin: 0 0 0.2rem 0;
      font-size: 0.95rem;
      color: #f1f5f9;
      font-weight: 600;
    }
    .step-meta {
      display: flex;
      gap: 0.75rem;
      font-size: 0.75rem;
      color: #64748b;
    }
    .attempt-tag { color: #a78bfa; }
    .duration-tag { color: #94a3b8; }
    .step-state-pill {
      font-size: 0.72rem;
      font-weight: 700;
      padding: 0.25rem 0.6rem;
      border-radius: 6px;
      display: flex;
      align-items: center;
      gap: 0.35rem;
    }
    .pill-succeeded { background: rgba(16, 185, 129, 0.15); color: #34d399; }
    .pill-failed { background: rgba(239, 68, 68, 0.15); color: #f87171; }
    .pill-running { background: rgba(245, 158, 11, 0.15); color: #fbbf24; }
    .pill-pending { background: rgba(255, 255, 255, 0.05); color: #64748b; }
    
    .small-spinner {
      width: 10px; height: 10px;
      border: 2px solid rgba(251, 191, 36, 0.3);
      border-top-color: #fbbf24;
      border-radius: 50%;
      animation: spin 0.8s linear infinite;
    }

    .error-box {
      margin-top: 0.85rem;
      background: rgba(15, 23, 42, 0.8);
      border: 1px solid rgba(239, 68, 68, 0.3);
      border-radius: 8px;
      padding: 0.85rem;
    }
    .error-header {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      margin-bottom: 0.4rem;
    }
    .error-title {
      font-weight: 700;
      color: #f87171;
      font-size: 0.85rem;
    }
    .err-cat-badge {
      margin-left: auto;
      font-size: 0.68rem;
      padding: 0.15rem 0.4rem;
      border-radius: 4px;
      background: rgba(239, 68, 68, 0.2);
      color: #fca5a5;
    }
    .err-cat-badge.transient {
      background: rgba(56, 189, 248, 0.2);
      color: #7dd3fc;
    }
    .error-msg {
      font-size: 0.85rem;
      color: #e2e8f0;
      margin-bottom: 0.5rem;
    }
    .layman-translation {
      background: rgba(99, 102, 241, 0.12);
      border: 1px solid rgba(99, 102, 241, 0.25);
      color: #cbd5e1;
      padding: 0.5rem 0.75rem;
      border-radius: 6px;
      font-size: 0.82rem;
      margin-bottom: 0.5rem;
    }
    .raw-payload {
      margin-top: 0.4rem;
    }
    .payload-label {
      font-size: 0.7rem;
      color: #64748b;
      display: block;
      margin-bottom: 0.2rem;
    }
    .raw-payload pre {
      margin: 0;
      background: #020617;
      padding: 0.5rem;
      border-radius: 6px;
      font-family: 'JetBrains Mono', monospace;
      font-size: 0.75rem;
      color: #a5b4fc;
      overflow-x: auto;
    }
    @keyframes spin { to { transform: rotate(360deg); } }
  `]
})
export class JobDetailComponent implements OnInit, OnChanges, OnDestroy {
  @Input() job?: ProvisioningJob;
  @Input() sub?: Subscription;
  @Input() timeline: StepEvent[] = [];

  private jobService = inject(JobService);
  private sseSub?: RxSubscription;
  isLiveStreaming: boolean = false;
  displaySteps: StepEvent[] = [];

  ngOnInit() {
    this.updateDisplaySteps();
    this.checkStreamingState();
  }

  ngOnChanges(changes: SimpleChanges) {
    if (changes['timeline'] || changes['job']) {
      this.updateDisplaySteps();
      this.checkStreamingState();
    }
  }

  ngOnDestroy() {
    this.stopStreaming();
  }

  private updateDisplaySteps() {
    if (!this.timeline) {
      this.displaySteps = [];
      return;
    }
    this.displaySteps = [...this.timeline];
  }

  private checkStreamingState() {
    if (this.job && (this.job.state === 'queued' || this.job.state === 'running')) {
      this.startStreaming();
    } else {
      this.stopStreaming();
    }
  }

  private startStreaming() {
    if (this.isLiveStreaming || !this.job) return;
    this.isLiveStreaming = true;

    this.sseSub = this.jobService.streamJobTimeline(this.job.id).subscribe({
      next: (event) => {
        if (event.type === 'step_updated') {
          const step: StepEvent = event.data;
          const idx = this.displaySteps.findIndex(s => s.id === step.id);
          if (idx >= 0) {
            this.displaySteps[idx] = step;
          } else {
            this.displaySteps.push(step);
          }
        } else if (event.type === 'job_updated') {
          this.job = event.data;
          if (this.job?.state === 'succeeded' || this.job?.state === 'failed') {
            this.stopStreaming();
          }
        }
      },
      error: () => {
        this.stopStreaming();
      }
    });
  }

  private stopStreaming() {
    this.isLiveStreaming = false;
    if (this.sseSub) {
      this.sseSub.unsubscribe();
      this.sseSub = undefined;
    }
  }
}
