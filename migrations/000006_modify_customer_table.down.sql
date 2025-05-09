ALTER TABLE customers DROP COLUMN first_name;
ALTER TABLE customers DROP COLUMN last_name;
ALTER TABLE customers ADD COLUMN name VARCHAR(255);
ALTER TABLE customers RENAME COLUMN phone_number TO phone;
ALTER TABLE customers DROP COLUMN active;