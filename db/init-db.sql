-- USERS TABLE
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    role TEXT NOT NULL
);

-- LOANS TABLE
CREATE TABLE IF NOT EXISTS loans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    borrower_id_number TEXT NOT NULL,
    amount REAL NOT NULL,
    rate REAL NOT NULL,
    roi REAL NOT NULL,
    status TEXT NOT NULL DEFAULT 'proposed',
    requester_id INTEGER NOT NULL,
    agreement_letter_url TEXT,
    FOREIGN KEY (requester_id) REFERENCES users(id)
);

-- APPROVALS TABLE
CREATE TABLE IF NOT EXISTS approvals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    loan_id INTEGER NOT NULL UNIQUE,
    validator_id TEXT NOT NULL,
    proof_url TEXT NOT NULL,
    approved_at DATETIME NOT NULL,
    FOREIGN KEY (loan_id) REFERENCES loans(id)
);

-- INVESTMENTS TABLE
CREATE TABLE IF NOT EXISTS investments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    loan_id INTEGER NOT NULL,
    investor_id INTEGER NOT NULL,
    amount REAL NOT NULL,
    investment_date TEXT NOT NULL,
    FOREIGN KEY (loan_id) REFERENCES loans(id),
    FOREIGN KEY (investor_id) REFERENCES users(id)
);

-- DISBURSEMENTS TABLE
CREATE TABLE IF NOT EXISTS disbursements (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    loan_id INTEGER NOT NULL UNIQUE,
    disbursed_at TEXT NOT NULL,
    field_officer_id TEXT NOT NULL,
    agreement_url TEXT NOT NULL,
    admin_id INTEGER NOT NULL,
    FOREIGN KEY (loan_id) REFERENCES loans(id),
    FOREIGN KEY (admin_id) REFERENCES users(id)
);


-- Seed Users
INSERT OR IGNORE INTO users (username, email, password, role) VALUES
('admin', 'admin@email.com','$2a$12$j.rFEx1xe/Bu8E6K9n5qce.CvmB6CWFncUHPAFwRpZLPp2KefKas6', 'admin'),
('loan_requester1', 'loan1@email.com', '$2a$12$NFOq.C20AIixRD9v8Sms4.8kFUT.cHZg0p2Vt8NBGdqGInx7V2E7K', 'requester'),
('loan_requester2', 'loan2@email.com', '$2a$12$NFOq.C20AIixRD9v8Sms4.8kFUT.cHZg0p2Vt8NBGdqGInx7V2E7K', 'requester'),
('investor1', 'investor1@email.com', '$2a$12$NYTvY3idcI42xAOGzZllA.8iSDxjhTifhJ0QVRJCsGYQKwURpPpM.', 'investor'),
('investor2', 'investor2@email.com', '$2a$12$NYTvY3idcI42xAOGzZllA.8iSDxjhTifhJ0QVRJCsGYQKwURpPpM.', 'investor'),
('investor3', 'investor3@email.com', '$2a$12$NYTvY3idcI42xAOGzZllA.8iSDxjhTifhJ0QVRJCsGYQKwURpPpM.', 'investor'),
('investor4', 'investor4@email.com', '$2a$12$NYTvY3idcI42xAOGzZllA.8iSDxjhTifhJ0QVRJCsGYQKwURpPpM.', 'investor');

-- Explanation:
-- Passwords are pre-hashed using bcrypt:
-- admin:         admin123
-- loan_requesters: loan123
-- investors:     investor123
