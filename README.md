# Loan Service Engine

This is a loan processing backend API developed using Go and Gin. It simulates a real-world microfinance workflow, including loan proposal, approval, investment, disbursement, and agreement letter generation.

## System Design
As part of the Amartha interview process, I developed a backend service focused on tracking the full lifecycle of a loan. Several key design decisions were made to balance realism, scalability, and development speed:

### Framework Choice:
The Gin web framework was selected for its performance, simplicity, and my familiarity with it, enabling efficient development within a limited timeframe.

### Database:
SQLite was used as the database engine due to its minimal setup and ease of use. For this specific use case—without complex joins or concurrent access patterns—SQLite provides a lightweight and effective solution.

### Loan Amount Limits:
To maintain realistic constraints, loan amounts are limited to a minimum of 1,000,000 IDR and a maximum of 100,000,000 IDR. This avoids impractical scenarios such as extremely small or excessively large loans.

### Schema Design:
Each stage of the loan process is modeled in its own database table:
- loans
- approvals
- investments
- disbursements
This normalized schema improves data organization and traceability, making it easier to query and reason about each stage independently.

### Authentication Scope:

User registration was intentionally excluded. The project’s focus is on the loan processing flow and role-based transitions, so predefined users are seeded into the database for testing and demonstration purposes.

## Features

- JWT-based authentication with 3 user roles: requester, investor, and admin
- Loan processing state machine:
  - Loan creation (requester)
  - Loan approval (admin with proof upload)
  - Investment by multiple investors (investors)
  - Disbursement after loan is fully funded (admin with signed proof upload)
- Agreement letter generation in PDF format
- Loan list (admin) and individual loan detail (all users) endpoints
- Unit-tested flow and edge cases

## Getting Started

### 1. Clone this repository
```
git clone https://www.github.com/chafid/loan-service-engine
```

### 2. Install dependencies

```bash
cd loan-service-engine
go mod tidy
```

### 3. Initialize the database

```bash
cd db
sqlite3 loan_service.db < schema.sql
```

> This creates the tables and seeds 7 users. 

### 4. .env file

Update the `.env` file with your own secret key:

```
JWT_SECRET=your_super_secret_key
```

### 5. Start the server

go back to project root and run:
```bash
go run main.go
```
The engine will run on port 8080. 
If you run it on local, to call the API would be:
```
http://localhost:8080/login # for login
```
```
http://localhost:8080/api/requester/create-loan # to create loan
```

## Directory structure
```
loan-service-engine/
├── go.mod
├── go.sum
├── main.go
├── /config
│   └── config.go 
├── /db
│   └── db.go
│   └── init-db.sql
│   └── loan_service.db 
├── .env
├── /handlers
│   └── admin.go
│   └── auth.go
│   └── disbursement.go
│   └── investment.go
│   └── loan_flow_test.go   # unit test for the flow of loan process
│   └── loan.go
├── /middleware
│   └── auth.go             # auth process for user roles
├── /models
│   └── loan.go             # structs for loan processes
├── /pdf
│   └── agreement.go        # module to generate agreement pdf to be sent to investors and borrower
├── /test_db
│   └── loan_service.db     # database for unit testing
│   └── proof.jpg           # image needed for approval proof unit test
├── /utils
│   └── email.go            # module to generate email to be sent to investors
├── README.md
```

## Users

Pre-seeded users:

| Role        | Username           | Password     |
|-------------|--------------------|--------------|
| Admin       | admin              | admin123     |
| Requester   | loan_requester1    | loan123      |
| Requester   | loan_requester2    | loan123      |
| Investor    | investor1          | investor123  |
| Investor    | investor2          | investor123  |
| Investor    | investor3          | investor123  |
| Investor    | investor4          | investor123  |

## Authentication

Login to get JWT token:

```
POST /login
{
  "username": "admin",
  "password": "admin123"
}
```

Use the JWT in subsequent requests:

```
Authorization: Bearer <your_token>
```

## Endpoints Overview

| Endpoint                        | Role         | Description                        |
|---------------------------------|--------------|------------------------------------|
| `/login`                        | All          | Login and receive JWT token        |
| `/api/requester/create-loan`    | requester    | Propose a loan                     |
| `/api/admin/approve-loan`       | admin        | Approve a loan with proof upload   |
| `/api/admin/disburse-loan`      | admin        | Disburse a fully invested loan     |
| `/api/admin/loans`              | admin        | List all loans                     |
| `/api/admin/loan/:id`           | admin        | Get details of a single loan       |
| `/api/admin/invest-loan`        | admin        | Invest in a loan                   |

## Testing

Run unit tests:

```bash
go test -v ./handlers
```

## Notes

- This project is designed to demonstrate multi-stage workflow logic and data validation in a finance-related setting.
- Emails are simulated via logs and JSON output only.
- Investment amount must fulfill loan amount exactly; partial remainder below 10% is blocked.
- All uploaded files are stored under `uploads/`.



---

© 2025 Loan Service Engine — built for demonstration and interview evaluation.