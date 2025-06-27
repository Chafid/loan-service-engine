package handlers

import (
	"fmt"
	"loan-service-engine/db"
	"loan-service-engine/models"
	"loan-service-engine/pdf"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateLoan(c *gin.Context) {
	role := c.GetString("role")
	if role != "requester" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only requesters can create loans"})
		return
	}
	userID := c.GetInt("userID")

	var req models.CreateLoanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Validate loan amount range
	if req.Amount < 1000000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Minimum loan amount is Rp 1,000,000"})
		return
	}
	if req.Amount > 100000000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum loan amount is Rp 100,000,000"})
		return
	}

	// Validate rate > ROI
	if req.Rate <= req.ROI {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Interest rate must be higher than ROI",
		})
		return
	}

	_, err := db.DB.Exec(`
		INSERT INTO loans (borrower_id_number, amount, rate, roi, status, requester_id)
		VALUES (?, ?, ?, ?, ?, ?)
	`, req.BorrowerIDNumber, req.Amount, req.Rate, req.ROI, "proposed", userID)

	if err != nil {
		log.Println("Failed to insert loan:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create loan"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Loan created and in proposed state"})
}

func DownloadLoanAgreement(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can download agreements"})
		return
	}

	loanID, err := strconv.Atoi(c.Param("loan_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid loan ID"})
		return
	}

	// Simulate generating agreement on-the-fly
	filename := fmt.Sprintf("agreement_loan%d_borrower.pdf", loanID)
	path := filepath.Join("uploads", filename)

	// If file doesn't exist, generate it (optional logic)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Println("Agreement PDF not found. Generating...")

		err := pdf.GenerateBorrowerAgreementPDF(loanID, path)
		if err != nil {
			log.Printf("PDF generation failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate agreement"})
			return
		}
	}

	c.FileAttachment(path, filename)
}

func GetLoanDetails(c *gin.Context) {
	loanID := c.Param("id")

	// Base loan info
	var loan models.LoanDetails
	err := db.DB.QueryRow(`
		SELECT l.id, l.borrower_id_number, l.amount, l.rate, l.roi, l.status, u.username
		FROM loans l
		JOIN users u ON u.id = l.requester_id
		WHERE l.id = ?
	`, loanID).Scan(
		&loan.ID,
		&loan.BorrowerIDNumber,
		&loan.Amount,
		&loan.Rate,
		&loan.ROI,
		&loan.Status,
		&loan.Requester,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Loan not found"})
		return
	}

	// Approval info (optional)
	_ = db.DB.QueryRow(`
		SELECT validator_id, approved_at, proof_url
		FROM approvals
		WHERE loan_id = ?
	`, loanID).Scan(&loan.Approval.ValidatorID, &loan.Approval.ApprovedAt, &loan.Approval.ProofURL)

	// Disbursement info (optional)
	_ = db.DB.QueryRow(`
		SELECT field_officer_id, disbursed_at, agreement_url
		FROM disbursements
		WHERE loan_id = ?
	`, loanID).Scan(&loan.Disbursement.OfficerID, &loan.Disbursement.DisbursedAt, &loan.Disbursement.SignedAgreementURL)

	// Investment info
	rows, _ := db.DB.Query(`
		SELECT u.username, i.amount
		FROM investments i
		JOIN users u ON u.id = i.investor_id
		WHERE i.loan_id = ?
	`, loanID)

	for rows.Next() {
		var inv models.InvestmentInfo
		_ = rows.Scan(&inv.Investor, &inv.Amount)
		loan.Investments = append(loan.Investments, inv)
	}

	c.JSON(http.StatusOK, loan)
}
