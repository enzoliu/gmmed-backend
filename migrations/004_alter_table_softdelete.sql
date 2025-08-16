ALTER TABLE serials ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP DEFAULT NULL;
CREATE INDEX IF NOT EXISTS idx_serials_deleted_at ON serials (deleted_at);
CREATE TRIGGER update_serials_updated_at 
  BEFORE UPDATE ON serials 
  FOR EACH ROW 
  EXECUTE FUNCTION update_updated_at_column();

ALTER TABLE warranty_registrations ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP DEFAULT NULL;
CREATE INDEX IF NOT EXISTS idx_warranty_registrations_deleted_at ON warranty_registrations (deleted_at);

ALTER TABLE products ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP DEFAULT NULL;
CREATE INDEX IF NOT EXISTS idx_products_deleted_at ON products (deleted_at);

ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP DEFAULT NULL;
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);