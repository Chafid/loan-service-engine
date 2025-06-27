package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"loan-service-engine/db"
	"loan-service-engine/pdf"
	"loan-service-engine/utils"

	"github.com/gin-gonic/gin"
)

type InvestRequest struct {
	LoanID int     `json:"loan_id" binding:"required"`
	Amount float64 `json:"amount" binding:"required"`
}

func InvestInLoan(c *gin.Context) {
	role := c.GetString("role")
	if role != "investor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only investors can invest in loans"})
		return
	}

	userID := c.GetInt("userID")

	var req InvestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	// 1. Check loan status and amount
	var status string
	var loanAmount float64
	err := db.DB.QueryRow(`SELECT status, amount FROM loans WHERE id = ?`, req.LoanID).Scan(&status, &loanAmount)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Loan not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if status != "approved" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can only invest in loans that are approved"})
		return
	}

	// Calculate 10% minimum
	minInvestment := loanAmount * 0.10
	if req.Amount < minInvestment {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Minimum investment is 10%% of loan amount (Rp %.0f)", minInvestment),
		})
		return
	}

	// 2. Check current total investment
	var totalInvested float64
	err = db.DB.QueryRow(`SELECT COALESCE(SUM(amount), 0) FROM investments WHERE loan_id = ?`, req.LoanID).Scan(&totalInvested)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check total investment"})
		return
	}

	if totalInvested+req.Amount > loanAmount {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":           "Investment would exceed loan principal",
			"loan_principal":  loanAmount,
			"already_raised":  totalInvested,
			"requested_extra": req.Amount,
		})
		return
	}

	remainingAmount := loanAmount - totalInvested
	futureRemaining := remainingAmount - req.Amount
	if futureRemaining < minInvestment && futureRemaining > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf(
				"This investment would leave only Rp %.0f remaining, which is below the minimum allowed (Rp %.0f). Please adjust your investment to fully fund the loan.",
				futureRemaining, minInvestment,
			),
		})
		return
	}

	// 3. Insert investment
	_, err = db.DB.Exec(`
		INSERT INTO investments (loan_id, investor_id, amount, investment_date)
		VALUES (?, ?, ?, ?)
	`, req.LoanID, userID, req.Amount, time.Now().Format("2006-01-02"))

	if err != nil {
		log.Println("Insert investment failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record investment"})
		return
	}

	// 4. Recalculate total — did we fully fund the loan?
	totalInvested += req.Amount
	if totalInvested == loanAmount {
		// Update loan status to 'invested'
		_, err = db.DB.Exec(`UPDATE loans SET status = 'invested' WHERE id = ?`, req.LoanID)
		if err != nil {
			log.Println("Failed to update loan status to 'invested':", err)
		} else {
			// Simulate sending email to all investors
			go NotifyInvestorsOfAgreement(req.LoanID, c)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "Investment recorded",
		"total_invested":    totalInvested,
		"loan_fully_funded": totalInvested == loanAmount,
	})
}

func NotifyInvestorsOfAgreement(loanID int, c *gin.Context) {
	rows, err := db.DB.Query(`
		SELECT u.username, u.id, i.amount, u.email
		FROM investments i
		JOIN users u ON i.investor_id = u.id
		WHERE i.loan_id = ?
	`, loanID)
	if err != nil {
		log.Println("Error fetching investors for agreement notification:", err)
		return
	}
	defer rows.Close()

	var previews []utils.EmailPreview
	for rows.Next() {
		var username string
		var userID int
		var amount float64
		var email string
		if err := rows.Scan(&username, &userID, &amount, &email); err == nil {
			pdfURL, err := pdf.GenerateAgreementPDF(loanID, username, amount)
			if err != nil {
				log.Printf("Failed to generate PDF for %s: %v", username, err)
				continue
			}
			emailPreview := utils.ComposeAgreementEmail(email, loanID, pdfURL)
			previews = append(previews, emailPreview)
		}
	}

	// At this point we successfully composed emails for each of the investors.
	// In the actual system, we could add another function here to send the composed emails
	// using SMTP or service like SendGrim/Mailgun.
	// for example:
	// utils.SendAgreementEmails(previews)

	c.JSON(http.StatusOK, gin.H{
		"message":        "Investment successful and loan fully funded",
		"email_previews": previews,
	})
}
