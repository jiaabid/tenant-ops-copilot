import { Component, Input, Output, EventEmitter } from '@angular/core';
import { CommonModule } from '@angular/common';
import { JobListItem, ProvisioningJob } from '../../models/job.model';

@Component({
  selector: 'app-job-list',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="job-list-card">
      <div class="list-header">
        <div>
          <h3>Onboarding Incidents Dashboard</h3>
          <span class="sub-header-text">Select an account setup failure to inspect with AI Copilot</span>
        </div>
        <span class="count-badge">{{ jobs.length }} Active Incidents</span>
      </div>

      <div class="scenario-buttons">
        <span class="scenario-label">Try Failure Scenarios:</span>
        <button 
          *ngFor="let item of jobs" 
          class="scenario-btn" 
          [class.active]="selectedJob?.id === item.job.id"
          (click)="onSelect(item.job)">
          <span class="scen-icon" [class.transient]="item.job.error_category === 'transient'" [class.structural]="item.job.error_category === 'structural'">
            {{ item.job.error_category === 'transient' ? '⚡ Scenario A: Temporary Server Glitch' : '⚠️ Scenario B: Setting Conflict' }}
          </span>
          <span class="scen-name">— {{ item.job.tenant_name }}</span>
        </button>
      </div>

      <div class="job-cards-grid">
        <div 
          *ngFor="let item of jobs" 
          class="job-card-item"
          [class.selected]="selectedJob?.id === item.job.id"
          (click)="onSelect(item.job)">
          
          <div class="item-top">
            <span class="sub-id">Sub ID: {{ item.subscription.id }}</span>
            <span class="state-tag" [ngClass]="'state-' + item.job.state">
              {{ item.job.state === 'failed' ? '🔴 FAILED' : item.job.state === 'running' ? '⚡ RUNNING' : '🔵 QUEUED' }}
            </span>
          </div>

          <h4 class="tenant-title">{{ item.job.tenant_name }}</h4>

          <div class="item-details">
            <span class="detail-pill">Failed at: {{ item.job.failed_step_name || 'None' }}</span>
            <span class="detail-pill">Attempts: {{ item.job.current_attempt }}/{{ item.job.max_retries }}</span>
          </div>

          <!-- Layman Error Tag -->
          <div class="layman-badge" [class.transient]="item.job.error_category === 'transient'" [class.structural]="item.job.error_category === 'structural'">
            <span *ngIf="item.job.error_category === 'transient'">
              💡 <strong>Temporary Glitch:</strong> Server timed out. Safe to re-run.
            </span>
            <span *ngIf="item.job.error_category === 'structural'">
              ⚠️ <strong>Setting Conflict:</strong> Website address already in use.
            </span>
          </div>

        </div>
      </div>
    </div>
  `,
  styles: [`
    .job-list-card {
      background: rgba(30, 41, 59, 0.6);
      border: 1px solid rgba(255, 255, 255, 0.08);
      border-radius: 14px;
      padding: 1.25rem;
      backdrop-filter: blur(12px);
    }
    .list-header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      margin-bottom: 1rem;
    }
    .list-header h3 {
      margin: 0 0 0.2rem 0;
      font-size: 1.1rem;
      color: #f8fafc;
    }
    .sub-header-text {
      font-size: 0.75rem;
      color: #94a3b8;
    }
    .count-badge {
      font-size: 0.72rem;
      background: rgba(255, 255, 255, 0.06);
      border: 1px solid rgba(255, 255, 255, 0.1);
      color: #94a3b8;
      padding: 0.2rem 0.6rem;
      border-radius: 12px;
    }
    .scenario-buttons {
      display: flex;
      align-items: center;
      gap: 0.6rem;
      margin-bottom: 1.25rem;
      background: rgba(15, 23, 42, 0.6);
      padding: 0.65rem 0.85rem;
      border-radius: 10px;
      flex-wrap: wrap;
    }
    .scenario-label {
      font-size: 0.75rem;
      color: #94a3b8;
      font-weight: 600;
    }
    .scenario-btn {
      background: rgba(30, 41, 59, 0.8);
      border: 1px solid rgba(255, 255, 255, 0.1);
      color: #cbd5e1;
      padding: 0.4rem 0.85rem;
      border-radius: 8px;
      font-size: 0.78rem;
      cursor: pointer;
      display: flex;
      align-items: center;
      gap: 0.4rem;
      transition: all 0.2s;
    }
    .scenario-btn.active {
      border-color: #6366f1;
      background: rgba(99, 102, 241, 0.25);
      color: #ffffff;
    }
    .scen-icon.transient { color: #38bdf8; }
    .scen-icon.structural { color: #f43f5e; }
    .job-cards-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
      gap: 1rem;
    }
    .job-card-item {
      background: rgba(15, 23, 42, 0.6);
      border: 1px solid rgba(255, 255, 255, 0.08);
      border-radius: 10px;
      padding: 1rem;
      cursor: pointer;
      transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
    }
    .job-card-item:hover {
      transform: translateY(-2px);
      border-color: rgba(99, 102, 241, 0.4);
    }
    .job-card-item.selected {
      border-color: #6366f1;
      box-shadow: 0 0 15px rgba(99, 102, 241, 0.2);
    }
    .item-top {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 0.4rem;
    }
    .sub-id {
      font-family: monospace;
      font-size: 0.72rem;
      color: #94a3b8;
    }
    .state-tag {
      font-size: 0.68rem;
      font-weight: 700;
      padding: 0.15rem 0.45rem;
      border-radius: 4px;
    }
    .state-failed { background: rgba(239, 68, 68, 0.2); color: #f87171; }
    .state-queued { background: rgba(56, 189, 248, 0.2); color: #38bdf8; }
    .state-running { background: rgba(245, 158, 11, 0.2); color: #fbbf24; }
    .tenant-title {
      margin: 0 0 0.5rem 0;
      font-size: 1rem;
      color: #f1f5f9;
    }
    .item-details {
      display: flex;
      flex-wrap: wrap;
      gap: 0.4rem;
      margin-bottom: 0.65rem;
    }
    .detail-pill {
      font-size: 0.7rem;
      background: rgba(255, 255, 255, 0.04);
      color: #94a3b8;
      padding: 0.15rem 0.4rem;
      border-radius: 4px;
    }
    .layman-badge {
      font-size: 0.75rem;
      padding: 0.45rem 0.6rem;
      border-radius: 6px;
      line-height: 1.35;
    }
    .layman-badge.transient {
      background: rgba(56, 189, 248, 0.12);
      color: #7dd3fc;
      border: 1px solid rgba(56, 189, 248, 0.25);
    }
    .layman-badge.structural {
      background: rgba(244, 63, 94, 0.12);
      color: #fda4af;
      border: 1px solid rgba(244, 63, 94, 0.25);
    }
  `]
})
export class JobListComponent {
  @Input() jobs: JobListItem[] = [];
  @Input() selectedJob?: ProvisioningJob;
  @Output() jobSelected = new EventEmitter<ProvisioningJob>();

  onSelect(job: ProvisioningJob) {
    this.jobSelected.emit(job);
  }
}
