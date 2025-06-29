package pdf

import (
	"database/sql"
	"fmt"
	"loan-service-engine/db"
	"os"
	"path/filepath"
	"time"

	"github.com/jung-kurt/gofpdf"
)

func GenerateAgreementPDF(loanID int, investorName string, amount float64) (string, error) {
	filename := fmt.Sprintf("agreement_loan%d_%s.pdf", loanID, investorName)
	outputDir := "uploads"
	outputPath := filepath.Join(outputDir, filename)

	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create output dir: %v", err)
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)

	pdf.Cell(40, 10, "Loan Agreement")
	pdf.Ln(12)

	date := time.Now().Format("02 Jan 2006")
	content := fmt.Sprintf(`
Date: %s

This document serves as an agreement that %s has invested an amount of %.2f into Loan #%d.

The agreement becomes effective once the loan reaches its funding goal.

This document is automatically generated by the Loan Service System.
`, date, investorName, amount, loanID)

	for _, line := range splitLines(content) {
		pdf.Cell(0, 10, line)
		pdf.Ln(8)
	}

	err = pdf.OutputFileAndClose(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to generate pdf: %v", err)
	}

	return "/" + outputPath, nil
}

func splitLines(text string) []string {
	var lines []string
	for _, line := range []byte(text) {
		lines = append(lines, string(line))
	}
	return splitByNewline(text)
}

func splitByNewline(text string) []string {
	var lines []string
	current := ""
	for _, ch := range text {
		if ch == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func GenerateBorrowerAgreementPDF(loanID int, path string) error {
	var (
		borrowerName, borrowerEmail, nik string
		amount, rate, roi                float64
	)

	err := db.DB.QueryRow(`
		SELECT u.username, u.email, l.borrower_id_number, l.amount, l.rate, l.roi
		FROM loans l
		JOIN users u ON u.id = l.requester_id
		WHERE l.id = ?
	`, loanID).Scan(&borrowerName, &borrowerEmail, &nik, &amount, &rate, &roi)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("loan or borrower not found")
		}
		return fmt.Errorf("DB error: %v", err)
	}

	err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create output dir: %v", err)
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)

	pdf.Cell(40, 10, "Loan Agreement - Borrower Copy")
	pdf.Ln(12)

	date := time.Now().Format("02 Jan 2006")

	content := fmt.Sprintf(`
Date: %s

Loan ID: %d

Borrower Details:
Name  : %s
Email : %s
NIK   : %s

Loan Terms:
Amount      : Rp %.2f
Interest    : %.2f%%
Expected ROI: %.2f%%

This agreement certifies that the borrower agrees to the above loan terms,
including repayment of principal and interest.

Signed by Borrower: _________________________

Date: _______________

`, date, loanID, borrowerName, borrowerEmail, nik, amount, rate, roi)

	for _, line := range splitByNewline(content) {
		pdf.Cell(0, 10, line)
		pdf.Ln(8)
	}

	return pdf.OutputFileAndClose(path)
}
