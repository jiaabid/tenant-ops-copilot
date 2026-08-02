import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { JobService } from './services/job.service';
import { JobListItem, ProvisioningJob, StepEvent } from './models/job.model';
import { JobListComponent } from './components/job-list/job-list.component';
import { JobDetailComponent } from './components/job-detail/job-detail.component';
import { CopilotComponent } from './components/copilot/copilot.component';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [CommonModule, JobListComponent, JobDetailComponent, CopilotComponent],
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent implements OnInit {
  private jobService = inject(JobService);

  jobs: JobListItem[] = [];
  selectedJob?: ProvisioningJob;
  selectedSubscription?: any;
  selectedTimeline: StepEvent[] = [];
  isLoading: boolean = true;
  apiConnected: boolean = true;
  isCopilotOpen: boolean = true; // Floating Copilot state

  ngOnInit() {
    this.loadJobs();
  }

  toggleCopilot() {
    this.isCopilotOpen = !this.isCopilotOpen;
  }

  loadJobs() {
    this.isLoading = true;
    this.jobService.getJobs().subscribe({
      next: (res) => {
        this.jobs = res;
        this.apiConnected = true;
        this.isLoading = false;

        // Auto select job-101 if not selected
        if (!this.selectedJob && res.length > 0) {
          this.selectJob(res[0].job);
        } else if (this.selectedJob) {
          const found = res.find(i => i.job.id === this.selectedJob?.id);
          if (found) {
            this.selectedJob = found.job;
          }
        }
      },
      error: () => {
        this.apiConnected = false;
        this.isLoading = false;
      }
    });
  }

  selectJob(job: ProvisioningJob) {
    this.selectedJob = job;
    this.jobService.getJobDetail(job.id).subscribe({
      next: (detail) => {
        this.selectedJob = detail.job;
        this.selectedSubscription = detail.subscription;
        this.selectedTimeline = detail.timeline;
      }
    });
  }

  onRetryExecuted() {
    this.loadJobs();
    if (this.selectedJob) {
      this.selectJob(this.selectedJob);
    }
  }
}
