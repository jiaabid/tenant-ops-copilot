package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"tenant-copilot/backend/domain"
)

type ProposalSigner struct {
	secretKey []byte
}

func NewProposalSigner(secret string) *ProposalSigner {
	return &ProposalSigner{
		secretKey: []byte(secret),
	}
}

// ComputeSignature generates HMAC-SHA256 for a given proposal
func (s *ProposalSigner) ComputeSignature(proposal *domain.SignedProposal) string {
	payload := fmt.Sprintf("%s|%s|%s|%s|%d|%d|%s",
		proposal.JobID,
		proposal.SubscriptionID,
		proposal.Action,
		proposal.TargetState,
		proposal.IssuedAt,
		proposal.ExpiresAt,
		proposal.Nonce,
	)

	h := hmac.New(sha256.New, s.secretKey)
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

// Sign sets IssuedAt, ExpiresAt, Nonce, and Signature on the proposal card
func (s *ProposalSigner) Sign(proposal *domain.SignedProposal, ttlSeconds int64) {
	now := time.Now().Unix()
	proposal.IssuedAt = now
	proposal.ExpiresAt = now + ttlSeconds
	if proposal.Nonce == "" {
		proposal.Nonce = fmt.Sprintf("nonce-%d", time.Now().UnixNano())
	}
	proposal.Signature = s.ComputeSignature(proposal)
}

// Verify returns true if the proposal signature is authentic and not expired
func (s *ProposalSigner) Verify(proposal *domain.SignedProposal) (bool, string) {
	now := time.Now().Unix()
	if now > proposal.ExpiresAt {
		return false, "Proposal confirmation card has expired. Please ask Copilot for a new remediation proposal."
	}

	expectedSig := s.ComputeSignature(proposal)
	if !hmac.Equal([]byte(proposal.Signature), []byte(expectedSig)) {
		return false, "Invalid HMAC signature on proposal card. Proposal may have been tampered with."
	}

	return true, ""
}
