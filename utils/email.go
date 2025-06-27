package utils

import (
	"fmt"
	"log"
)

type EmailPreview struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// Simulates sending an agreement email to one investor
func ComposeAgreementEmail(to string, loanID int, agreementURL string) EmailPreview {
	subject := fmt.Sprintf("Loan Agreement for Loan #%d", loanID)
	body := fmt.Sprintf(`Dear Investor,

Thank you for investing in Loan #%d.
Please review the loan agreement at the link below:

https://localhost:8000%s

Sincerely,
Loan Service Team`, loanID, agreementURL)

	log.Printf("[Email composed]\nTo: %s\nSubject: %s\n\n%s\n", to, subject, body)

	return EmailPreview{
		To:      to,
		Subject: subject,
		Body:    body,
	}
}
