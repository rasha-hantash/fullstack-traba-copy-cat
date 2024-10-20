-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Seed data for shifts table
INSERT INTO shifts (start_date, end_date, shifts_filled, invoice_id, invoice_amount, status, created_by, updated_by)
VALUES
    ('2024-10-01', '2024-10-07', '2024-10-07', uuid_generate_v4(), 1500.00, 'approved', uuid_generate_v4(), uuid_generate_v4()),
    ('2024-10-08', '2024-10-14', '2024-10-14', uuid_generate_v4(), 1200.50, 'pending', uuid_generate_v4(), uuid_generate_v4()),
    ('2024-10-15', '2024-10-21', '2024-10-21', uuid_generate_v4(), 1750.25, 'approved', uuid_generate_v4(), uuid_generate_v4()),
    ('2024-10-22', '2024-10-28', '2024-10-28', uuid_generate_v4(), 1300.75, 'rejected', uuid_generate_v4(), uuid_generate_v4()),
    ('2024-10-29', '2024-11-04', '2024-11-04', uuid_generate_v4(), 1600.00, 'approved', uuid_generate_v4(), uuid_generate_v4()),
    ('2024-11-05', '2024-11-11', '2024-11-11', uuid_generate_v4(), 1450.50, 'pending', uuid_generate_v4(), uuid_generate_v4()),
    ('2024-11-12', '2024-11-18', '2024-11-18', uuid_generate_v4(), 1800.25, 'approved', uuid_generate_v4(), uuid_generate_v4()),
    ('2024-11-19', '2024-11-25', '2024-11-25', uuid_generate_v4(), 1350.75, 'pending', uuid_generate_v4(), uuid_generate_v4()),
    ('2024-11-26', '2024-12-02', '2024-12-02', uuid_generate_v4(), 1550.00, 'approved', uuid_generate_v4(), uuid_generate_v4()),
    ('2024-12-03', '2024-12-09', '2024-12-09', uuid_generate_v4(), 1400.50, 'rejected', uuid_generate_v4(), uuid_generate_v4());