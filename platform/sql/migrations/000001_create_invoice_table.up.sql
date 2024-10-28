CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION citext;

CREATE TABLE users (
    id VARCHAR(255) PRIMARY KEY,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    email citext NOT NULL UNIQUE, -- todo make this unique
    phone_number VARCHAR(255) NOT NULL,
    company_name VARCHAR(255) NOT NULL,
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP
);

CREATE TABLE shifts (
    id VARCHAR(255) PRIMARY KEY,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    location VARCHAR(255) NOT NULL,
    shift_name VARCHAR(255) NOT NULL,
    shifts_filled INTEGER NOT NULL,
    worker_id VARCHAR(255) NOT NULL,
    shift_description TEXT,
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP,
    FOREIGN KEY (worker_id) REFERENCES users(id)
);

CREATE TABLE invoices (
    id VARCHAR(255) PRIMARY KEY,
    invoice_amount INTEGER NOT NULL,
    status VARCHAR(255) NOT NULL,
    shift_id VARCHAR(255) NOT NULL,
    invoice_name VARCHAR(255),
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP,
    FOREIGN KEY (shift_id) REFERENCES shifts(id)
);

-- Indexes (todo: look into the indexes that i want to create)
-- CREATE INDEX idx_shifts_start_date ON shifts(start_date);
-- CREATE INDEX idx_shifts_end_date ON shifts(end_date);
-- CREATE INDEX idx_invoices_status ON invoices(status);
-- CREATE INDEX idx_users_email ON users(email);
-- CREATE INDEX idx_users_role ON users(role);


-- todo: look into creating view for balance sheet (balance of a ledger account)