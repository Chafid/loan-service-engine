package handlers_test

import (
	"bytes"
	"encoding/json"
	"loan-service-engine/config"
	"loan-service-engine/db"
	"loan-service-engine/handlers"
	"loan-service-engine/middleware"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestEnv() {
	config.LoadEnv("../.env")
	db.Connect("../test_db/loan_service.db")

	// Clean slate
	tables := []string{"disbursements", "investments", "approvals", "loans"}
	for _, table := range tables {
		db.DB.Exec("DELETE FROM " + table)
		db.DB.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = '" + table + "'")
	}
}

// helpers

func login(t *testing.T, username, password string) string {
	router := gin.Default()
	router.POST("/login", handlers.Login)

	payload := map[string]string{"username": username, "password": password}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("Login failed: %s", resp.Body.String())
	}
	var result map[string]string
	json.Unmarshal(resp.Body.Bytes(), &result)
	return result["token"]
}

func createDummyFile(path string) {
	os.MkdirAll(filepath.Dir(path), os.ModePerm)
	os.WriteFile(path, []byte("dummy content"), 0644)
}

// main test
func TestLoanLifecycleAndEdgeCases(t *testing.T) {
	setupTestEnv()
	gin.SetMode(gin.TestMode)

	router := gin.Default()
	router.POST("/login", handlers.Login)

	api := router.Group("/api")
	api.Use(middleware.JWTAuthMiddleware())

	api.POST("/requester/create-loan", middleware.RequireRole("requester"), handlers.CreateLoan)
	api.POST("/admin/approve-loan", middleware.RequireRole("admin"), handlers.ApproveLoan)
	api.POST("/investor/invest", middleware.RequireRole("investor"), handlers.InvestInLoan)
	api.POST("/admin/disburse-loan", middleware.RequireRole("admin"), handlers.DisburseLoan)

	// Step 1: Create Loan
	tokenRequester := login(t, "loan_requester1", "loan123")
	loanPayload := map[string]interface{}{
		"borrower_id_number": "1122334455667788",
		"amount":             1000000,
		"rate":               12,
		"roi":                10,
	}
	loanBody, _ := json.Marshal(loanPayload)

	req, _ := http.NewRequest("POST", "/api/requester/create-loan", bytes.NewBuffer(loanBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenRequester)

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("CreateLoan failed: %s", resp.Body.String())
	}

	// Step 2: Approve Loan (happy path)
	tokenAdmin := login(t, "admin", "admin123")
	proofPath := "../test_db/proof.jpg"
	createDummyFile(proofPath)
	//defer os.Remove(proofPath)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("loan_id", "1")
	writer.WriteField("field_validator_employee_id", "EMP001")
	writer.WriteField("approval_date", "2025-06-25")
	fileWriter, _ := writer.CreateFormFile("visit_proof", "proof.jpg")
	fileBytes, err := os.ReadFile(proofPath)
	if err != nil {
		log.Println("error read proof file: ", err.Error())
	}
	fileWriter.Write(fileBytes)
	writer.Close()

	req, _ = http.NewRequest("POST", "/api/admin/approve-loan", body)
	req.Header.Set("Authorization", "Bearer "+tokenAdmin)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("ApproveLoan failed: %s", resp.Body.String())
	}

	// Step 3: Invest (happy path, 2 investors)
	tokenInvestor := login(t, "investor1", "investor123")
	investPayload := map[string]interface{}{
		"loan_id": 1,
		"amount":  500000,
	}
	investBody, _ := json.Marshal(investPayload)

	req, _ = http.NewRequest("POST", "/api/investor/invest", bytes.NewBuffer(investBody))
	req.Header.Set("Authorization", "Bearer "+tokenInvestor)
	req.Header.Set("Content-Type", "application/json")

	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("Investor 1 failed: %s", resp.Body.String())
	}

	tokenInvestor2 := login(t, "investor2", "investor123")
	investPayload2 := map[string]interface{}{
		"loan_id": 1,
		"amount":  500000,
	}
	investBody2, _ := json.Marshal(investPayload2)
	req, _ = http.NewRequest("POST", "/api/investor/invest", bytes.NewBuffer(investBody2))
	req.Header.Set("Authorization", "Bearer "+tokenInvestor2)
	req.Header.Set("Content-Type", "application/json")

	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("Investor 2 failed: %s", resp.Body.String())
	}

	// Step 4: Disbursement
	disbursePath := "test/fixtures/agreement.jpg"
	createDummyFile(disbursePath)
	defer os.Remove(disbursePath)

	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)
	writer.WriteField("loan_id", "1")
	writer.WriteField("field_officer_id", "EMP999")
	writer.WriteField("disbursement_date", "2025-06-26")
	fileWriter, _ = writer.CreateFormFile("signed_agreement", "agreement.jpg")
	fileBytes, _ = os.ReadFile(disbursePath)
	fileWriter.Write(fileBytes)
	writer.Close()

	req, _ = http.NewRequest("POST", "/api/admin/disburse-loan", body)
	req.Header.Set("Authorization", "Bearer "+tokenAdmin)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("Disbursement failed: %s", resp.Body.String())
	}

	// EDGE CASE 1: Approve without proof (should fail)
	setupTestEnv()
	db.DB.Exec(`INSERT INTO loans (id, borrower_id_number, amount, rate, roi, status, requester_id)
			VALUES (1, '8888888888888888', 1000000, 12, 10, 'proposed', 2)`)
	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)
	writer.WriteField("loan_id", "1")
	writer.WriteField("validator_id", "EMP001")
	writer.WriteField("approved_at", "2025-06-25")
	writer.Close()

	req, _ = http.NewRequest("POST", "/api/admin/approve-loan", body)
	req.Header.Set("Authorization", "Bearer "+tokenAdmin)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code == http.StatusOK {
		t.Error("Expected failure when approving without proof")
	}

	// EDGE CASE 2: Leave <10% uninvested (should fail)
	setupTestEnv()
	// Create approved loan for testing
	db.DB.Exec(`INSERT INTO loans (id, borrower_id_number, amount, rate, roi, status, requester_id)
				VALUES (2, '8888888888888888', 1000000, 12, 10, 'approved', 2)`)
	db.DB.Exec(`INSERT INTO approvals (loan_id, validator_id, proof_url, approved_at)
				VALUES (2, 'EMP001', '/dummy.jpg', '2025-06-25')`)
	tokenInvestor3 := login(t, "investor3", "investor123")

	// Invest 999,900 leaving only 100
	investPayload = map[string]interface{}{
		"loan_id": 2,
		"amount":  999900,
	}
	investBody, _ = json.Marshal(investPayload)
	req, _ = http.NewRequest("POST", "/api/investor/invest", bytes.NewBuffer(investBody))
	req.Header.Set("Authorization", "Bearer "+tokenInvestor3)
	req.Header.Set("Content-Type", "application/json")
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code == http.StatusOK {
		t.Error("Expected investment to fail due to low remaining amount, but it succeeded")
	} else {
		var response map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to parse response JSON: %v", err)
		}
		if !strings.Contains(response["error"].(string), "adjust your investment") {
			t.Errorf("Unexpected error message: %v", response["error"])
		}
	}

	// Last investor tries to invest remaining 100 (invalid)
	tokenInvestor4 := login(t, "investor4", "investor123")
	investPayload = map[string]interface{}{
		"loan_id": 2,
		"amount":  100,
	}
	investBody, _ = json.Marshal(investPayload)
	req, _ = http.NewRequest("POST", "/api/investor/invest", bytes.NewBuffer(investBody))
	req.Header.Set("Authorization", "Bearer "+tokenInvestor4)
	req.Header.Set("Content-Type", "application/json")
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code == http.StatusOK {
		t.Error("Expected rejection when last investment leaves <10% unfulfilled")
	}

	os.RemoveAll("test")
	os.RemoveAll("uploads")

}
