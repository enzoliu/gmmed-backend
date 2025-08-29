ALTER TABLE warranty_registrations ADD COLUMN IF NOT EXISTS step INT DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_warranty_registrations_step ON warranty_registrations (step);

ALTER TABLE products DROP CONSTRAINT IF EXISTS chk_products_warranty_years_non_negative;