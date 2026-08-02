export type JobState = 'queued' | 'running' | 'failed' | 'succeeded';
export type StepStatus = 'pending' | 'running' | 'succeeded' | 'failed';
export type ErrorCategory = 'transient' | 'structural' | 'none';

export interface ProvisioningJob {
  id: string;
  subscription_id: string;
  tenant_name: string;
  region: string;
  state: JobState;
  current_attempt: number;
  max_retries: number;
  failed_step_name?: string;
  last_error_summary?: string;
  error_category?: ErrorCategory;
  created_at: string;
  updated_at: string;
}

export interface StepEvent {
  id: string;
  job_id: string;
  step_index: number;
  step_name: string;
  attempt: number;
  max_attempts: number;
  status: StepStatus;
  error_message?: string;
  error_category?: ErrorCategory;
  raw_payload?: string;
  timestamp: string;
  duration_ms: number;
}

export interface Subscription {
  id: string;
  tenant_name: string;
  plan: string;
  status: string;
  job_id: string;
  updated_at: string;
}

export interface JobListItem {
  job: ProvisioningJob;
  subscription: Subscription;
}

export interface SignedProposal {
  job_id: string;
  subscription_id: string;
  tenant_name: string;
  action: string;
  current_state: string;
  target_state: string;
  remaining_attempts: number;
  max_retries: number;
  what_rerunning_does: string;
  issued_at: number;
  expires_at: number;
  nonce: string;
  signature: string;
}

export interface ToolTrace {
  tool_name: string;
  args: any;
  result: any;
}

export interface CopilotChatResponse {
  job_id: string;
  explanation: string;
  tool_traces: ToolTrace[];
  raw_timeline_returned: boolean;
  proposal?: SignedProposal;
}

export interface ConfirmRetryResponse {
  success: boolean;
  message?: string;
  error?: string;
  job?: ProvisioningJob;
}
