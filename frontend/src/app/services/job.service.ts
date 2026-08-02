import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, Subject } from 'rxjs';
import {
  JobListItem,
  ProvisioningJob,
  StepEvent,
  SignedProposal,
  CopilotChatResponse,
  ConfirmRetryResponse
} from '../models/job.model';

@Injectable({
  providedIn: 'root'
})
export class JobService {
  private http = inject(HttpClient);
  private apiUrl = (typeof window !== 'undefined' && window.location.hostname !== 'localhost') ? '/api' : 'http://localhost:8080/api';

  getJobs(): Observable<JobListItem[]> {
    return this.http.get<JobListItem[]>(`${this.apiUrl}/jobs`);
  }

  getJobDetail(jobId: string): Observable<{ job: ProvisioningJob; subscription: any; timeline: StepEvent[] }> {
    return this.http.get<{ job: ProvisioningJob; subscription: any; timeline: StepEvent[] }>(`${this.apiUrl}/jobs/${jobId}`);
  }

  askCopilot(jobId: string, query: string): Observable<CopilotChatResponse> {
    return this.http.post<CopilotChatResponse>(`${this.apiUrl}/copilot/chat`, {
      job_id: jobId,
      query: query
    });
  }

  confirmRetry(proposal: SignedProposal): Observable<ConfirmRetryResponse> {
    return this.http.post<ConfirmRetryResponse>(`${this.apiUrl}/jobs/retry/confirm`, {
      job_id: proposal.job_id,
      signed_proposal: proposal
    });
  }

  streamJobTimeline(jobId: string): Observable<{ type: 'step_updated' | 'job_updated'; data: any }> {
    return new Observable(observer => {
      const eventSource = new EventSource(`${this.apiUrl}/jobs/stream?job_id=${jobId}`);

      eventSource.addEventListener('step_updated', (event: MessageEvent) => {
        try {
          const data = JSON.parse(event.data);
          observer.next({ type: 'step_updated', data });
        } catch (e) {
          console.error('Failed to parse SSE step_updated event', e);
        }
      });

      eventSource.addEventListener('job_updated', (event: MessageEvent) => {
        try {
          const data = JSON.parse(event.data);
          observer.next({ type: 'job_updated', data });
        } catch (e) {
          console.error('Failed to parse SSE job_updated event', e);
        }
      });

      eventSource.onerror = (error) => {
        observer.error(error);
        eventSource.close();
      };

      return () => {
        eventSource.close();
      };
    });
  }
}
