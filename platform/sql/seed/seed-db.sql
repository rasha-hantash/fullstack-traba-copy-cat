DROP FUNCTION IF EXISTS generate_ksuid() ;
CREATE FUNCTION generate_ksuid() RETURNS  VARCHAR(27) AS $$
    DECLARE digits CHAR(62) DEFAULT '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz';

    DECLARE n DECIMAL(49) DEFAULT EXTRACT(EPOCH FROM NOW())::INT - 1400000000; -- 8 byte time portion offset per specification.
    DECLARE f DECIMAL(49) DEFAULT 340282366920938463463374607431768211456;
    DECLARE s VARCHAR(27) ;
    DECLARE i INT DEFAULT 1;
BEGIN

  -- shift and add random bytes
  n:= n * f + (RANDOM() * f)::DECIMAL(49);

  -- base62 encode
  WHILE i <= 27 LOOP
    s:= CONCAT(SUBSTR(digits, ((n % 62) + 1)::INT, 1), s);
    n:= FLOOR(n / 62);
    i:= i + 1;
  END LOOP; 

  RETURN s;
END;
$$  LANGUAGE plpgsql VOLATILE;


-- Seed data for shifts table
INSERT INTO invoices (id, start_date, end_date, shifts_filled, invoice_amount, status, user_id, shift_id, created_by, updated_by, invoice_name)
VALUES
    ('inv_' || generate_ksuid(), '2024-10-01', '2024-10-07', '2024-10-07', 1500.00, 'approved', 'user_' || generate_ksuid(), 'shift_' || generate_ksuid(), 'user_' || generate_ksuid(), 'user_' || generate_ksuid(), 'Summer Music Festival Staff'),
    ('inv_' || generate_ksuid(), '2024-10-08', '2024-10-14', '2024-10-14', 1200.50, 'pending', 'user_' || generate_ksuid(), 'shift_' || generate_ksuid(), 'user_' || generate_ksuid(), 'user_' || generate_ksuid(), 'Corporate Office Relocation'),
    ('inv_' || generate_ksuid(), '2024-10-15', '2024-10-21', '2024-10-21', 1750.25, 'approved', 'user_' || generate_ksuid(), 'shift_' || generate_ksuid(), 'user_' || generate_ksuid(), 'user_' || generate_ksuid(), 'Annual Tech Conference Support'),
    ('inv_' || generate_ksuid(), '2024-10-22', '2024-10-28', '2024-10-28', 1300.75, 'rejected', 'user_' || generate_ksuid(), 'shift_' || generate_ksuid(), 'user_' || generate_ksuid(), 'user_' || generate_ksuid(), 'Holiday Season Retail Assistance'),
    ('inv_' || generate_ksuid(), '2024-10-29', '2024-11-04', '2024-11-04', 1600.00, 'approved', 'user_' || generate_ksuid(), 'shift_' || generate_ksuid(), 'user_' || generate_ksuid(), 'user_' || generate_ksuid(), 'Hospital Emergency Room Coverage'),
    ('inv_' || generate_ksuid(), '2024-11-05', '2024-11-11', '2024-11-11', 1450.50, 'pending', 'user_' || generate_ksuid(), 'shift_' || generate_ksuid(), 'user_' || generate_ksuid(), 'user_' || generate_ksuid(), 'University Orientation Week'),
    ('inv_' || generate_ksuid(), '2024-11-12', '2024-11-18', '2024-11-18', 1800.25, 'approved', 'user_' || generate_ksuid(), 'shift_' || generate_ksuid(), 'user_' || generate_ksuid(), 'user_' || generate_ksuid(), 'Construction Site Safety Team'),
    ('inv_' || generate_ksuid(), '2024-11-19', '2024-11-25', '2024-11-25', 1350.75, 'pending', 'user_' || generate_ksuid(), 'shift_' || generate_ksuid(), 'user_' || generate_ksuid(), 'user_' || generate_ksuid(), 'Film Production Crew'),
    ('inv_' || generate_ksuid(), '2024-11-26', '2024-12-02', '2024-12-02', 1550.00, 'approved', 'user_' || generate_ksuid(), 'shift_' || generate_ksuid(), 'user_' || generate_ksuid(), 'user_' || generate_ksuid(), 'Warehouse Inventory Audit'),
    ('inv_' || generate_ksuid(), '2024-12-03', '2024-12-09', '2024-12-09', 1400.50, 'rejected', 'user_' || generate_ksuid(), 'shift_' || generate_ksuid(), 'user_' || generate_ksuid(), 'user_' || generate_ksuid(), 'Catering Staff for Gala Dinner');