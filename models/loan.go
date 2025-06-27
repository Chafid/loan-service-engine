package models

import "time"

type CreateLoanRequest struct {
	BorrowerIDNumber string  `json:"borrower_id_number" binding:"required"`
	Amount           float64 `json:"amount" binding:"required"`
	Rate             float64 `json:"rate" binding:"required"`
	ROI              float64 `json:"roi" binding:"required"`
}

// Loan represents a full loan record pulled from the DB.
type Loan struct {
	ID                 int     `json:"id"`
	BorrowerIDNumber   string  `json:"borrower_id_number"`
	Amount             float64 `json:"amount"`
	Rate               float64 `json:"rate"`
	ROI                float64 `json:"roi"`
	Status             string  `json:"status"`
	RequesterID        int     `json:"requester_id"`
	AgreementLetterURL string  `json:"agreement_letter_url,omitempty"`
}

// LoanResponse is the version sent back to clients
// (you can remove internal fields like requester_id if needed).
type LoanResponse struct {
	ID               int     `json:"id"`
	BorrowerIDNumber string  `json:"borrower_id_number"`
	Amount           float64 `json:"amount"`
	Rate             float64 `json:"rate"`
	ROI              float64 `json:"roi"`
	Status           string  `json:"status"`
}

type ApprovalRequest struct {
	ValidatorID string    `json:"validator_id" binding:"required"`
	ProofURL    string    `json:"proof_url" binding:"required"`
	ApprovedAt  time.Time `json:"approved_at" binding:"required"`
}

type InvestmentRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type LoanDetails struct {
	ID               int              `json:"id"`
	BorrowerIDNumber string           `json:"borrower_id_number"`
	Amount           float64          `json:"amount"`
	Rate             float64          `json:"rate"`
	ROI              float64          `json:"roi"`
	Status           string           `json:"status"`
	Requester        string           `json:"requester"`
	Approval         ApprovalInfo     `json:"approval,omitempty"`
	Investments      []InvestmentInfo `json:"investments,omitempty"`
	Disbursement     DisbursementInfo `json:"disbursement,omitempty"`
}

type ApprovalInfo struct {
	ValidatorID string `json:"validator_id"`
	ApprovedAt  string `json:"approved_at"`
	ProofURL    string `json:"proof_url"`
}

type InvestmentInfo struct {
	Investor string  `json:"investor"`
	Amount   float64 `json:"amount"`
}

type DisbursementInfo struct {
	OfficerID          string `json:"officer_id"`
	DisbursedAt        string `json:"disbursed_at"`
	SignedAgreementURL string `json:"signed_agreement_url"`
}
