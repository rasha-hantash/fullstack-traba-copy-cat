-- First drop tables in reverse order to handle foreign key dependencies
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS shifts;
DROP TABLE IF EXISTS users;

-- Drop the extension last
DROP EXTENSION IF EXISTS "uuid-ossp";