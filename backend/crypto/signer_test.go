package crypto

import (
	"testing"

	"tenant-copilot-backend/domain"
)

func TestProposalSigner_ValidSignature(t *testing.T) {
	signer := NewProposalSigner("secret-test-key")

	proposal := &domain.SignedProposal{
		JobID:             "job-101",
		SubscriptionID:    "sub-8921",
		TenantName:        "Acme Financial Corp",
		Action:            "retry_job",
		CurrentState:      "failed",
		TargetState:       "queued",
		RemainingAttempts: 4,
		MaxRetries:        3,
	}

	signer.Sign(proposal, 300) // 5 minutes

	if proposal.Signature == "" {
		t.Fatalf("expected non-empty signature")
	}

	valid, errMsg := signer.Verify(proposal)
	if !valid {
		t.Fatalf("expected proposal signature to be valid, got error: %s", errMsg)
	}
}

func TestProposalSigner_ExpiredToken(t *testing.T) {
	signer := NewProposalSigner("secret-test-key")

	proposal := &domain.SignedProposal{
		JobID:  "job-101",
		Action: "retry_job",
	}

	signer.Sign(proposal, -10) // Expired 10 seconds ago

	valid, errMsg := signer.Verify(proposal)
	if valid {
		t.Fatalf("expected expired proposal to be rejected")
	}
	if errMsg == "" {
		t.Fatalf("expected expiration error message")
	}
}

func TestProposalSigner_TamperedPayload(t *testing.T) {
	signer := NewProposalSigner("secret-test-key")

	proposal := &domain.SignedProposal{
		JobID:        "job-101",
		Action:       "retry_job",
		TargetState:  "queued",
		CurrentState: "failed",
	}

	signer.Sign(proposal, 300)

	// Tamper with target state
	proposal.TargetState = "succeeded"

	valid, _ := signer.Verify(proposal)
	if valid {
		t.Fatalf("expected tampered proposal payload to be rejected")
	}
}
