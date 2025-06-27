package handlers

import (
	"database/sql"
	"fmt"
	"loan-service-engine/db"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func DisburseLoan(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can disburse loans"})
		return
	}
	adminID := c.GetInt("userID")

	// Parse form fields
	loanIDStr := c.PostForm("loan_id")
	fieldOfficerID := c.PostForm("field_officer_id")
	disbursementDate := c.PostForm("disbursement_date")
	file, err := c.FormFile("signed_agreement")

	if loanIDStr == "" || fieldOfficerID == "" || disbursementDate == "" || file == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
		return
	}

	loanID, err := strconv.Atoi(loanIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid loan ID"})
		return
	}

	// Check loan exists and status
	var currentStatus string
	err = db.DB.QueryRow(`SELECT status FROM loans WHERE id = ?`, loanID).Scan(&currentStatus)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Loan not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}
	if currentStatus != "invested" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only 'invested' loans can be disbursed"})
		return
	}

	// Save uploaded file
	filename := fmt.Sprintf("signed_agreement_loan%d_%d%s", loanID, time.Now().Unix(), filepath.Ext(file.Filename))
	savePath := filepath.Join("uploads", filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	fileURL := "/uploads/" + filename

	// Insert disbursement record
	_, err = db.DB.Exec(`
		INSERT INTO disbursements (loan_id, disbursed_at, field_officer_id, agreement_url, admin_id)
		VALUES (?, ?, ?, ?, ?)
	`, loanID, disbursementDate, fieldOfficerID, fileURL, adminID)
	if err != nil {
		log.Println("Disbursement insert failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record disbursement"})
		return
	}

	// Update loan status
	_, err = db.DB.Exec(`UPDATE loans SET status = 'disbursed' WHERE id = ?`, loanID)
	if err != nil {
		log.Println("Loan status update failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update loan status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "Loan disbursed",
		"agreement_url":    fileURL,
		"disbursed_by":     adminID,
		"field_officer_id": fieldOfficerID,
		"disbursed_at":     disbursementDate,
	})
}
