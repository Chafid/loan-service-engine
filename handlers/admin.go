package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"loan-service-engine/db"
	"loan-service-engine/models"

	"github.com/gin-gonic/gin"
)

func ApproveLoan(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can approve loans"})
		return
	}

	// Parse multipart form fields
	loanIDStr := c.PostForm("loan_id")
	validatorID := c.PostForm("field_validator_employee_id")
	approvedAt := c.PostForm("approval_date") // expected in YYYY-MM-DD
	file, err := c.FormFile("visit_proof")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid proof image"})
		return
	}
	log.Println("Proof image: ", file.Filename)

	log.Println("loan id ", loanIDStr)
	log.Println("validator id ", validatorID)
	log.Println("approved at ", approvedAt)
	//log.Println("file ", file.Filename)

	// Validate fields
	if loanIDStr == "" || validatorID == "" || approvedAt == "" || file == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required including image proof"})
		return
	}

	loanID, err := strconv.Atoi(loanIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid loan_id"})
		return
	}

	// Check if loan exists and in proposed state
	var currentStatus string
	err = db.DB.QueryRow(`SELECT status FROM loans WHERE id = ?`, loanID).Scan(&currentStatus)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Loan not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if currentStatus != "proposed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Loan must be in 'proposed' state to approve"})
		return
	}

	// Save proof image
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("proof_%d_%s", timestamp, filepath.Base(file.Filename))
	savePath := filepath.Join("uploads", filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}
	proofURL := "/uploads/" + filename

	// Insert into approvals table
	_, err = db.DB.Exec(`
		INSERT INTO approvals (loan_id, validator_id, proof_url, approved_at)
		VALUES (?, ?, ?, ?)
	`, loanID, validatorID, proofURL, approvedAt)

	if err != nil {
		log.Println("Error inserting approval:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record approval"})
		return
	}

	// Update loan status to 'approved'
	_, err = db.DB.Exec(`
		UPDATE loans SET status = 'approved' WHERE id = ?
	`, loanID)

	if err != nil {
		log.Println("Error updating loan status:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update loan status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Loan approved",
		"proof_url":   proofURL,
		"approved_at": approvedAt,
	})
}

func ListLoans(c *gin.Context) {
	rows, err := db.DB.Query(`
		SELECT 
			id, borrower_id_number, amount, rate, roi, status
		FROM loans
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve loans"})
		return
	}
	defer rows.Close()

	var loans []models.LoanResponse
	for rows.Next() {
		var loan models.LoanResponse
		if err := rows.Scan(
			&loan.ID,
			&loan.BorrowerIDNumber,
			&loan.Amount,
			&loan.Rate,
			&loan.ROI,
			&loan.Status,
		); err != nil {
			log.Println("Scan error:", err)
			continue
		}
		loans = append(loans, loan)
	}

	c.JSON(http.StatusOK, gin.H{"loans": loans})
}
