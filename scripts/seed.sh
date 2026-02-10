#!/bin/bash
set -e

echo "🌱 Seeding VulnPulse database..."

# Wait for services to be ready
echo "Waiting for services to be ready..."
sleep 5

# Database connection URL (use from env or default)
DATABASE_URL="${DATABASE_URL:-postgres://vulnpulse:devpassword@localhost:5432/vulnpulse?sslmode=disable}"

# Seed tenant
TENANT_ID="11111111-1111-1111-1111-111111111111"
echo "Creating tenant..."
psql "$DATABASE_URL" <<SQL
INSERT INTO tenants (id, name, created_at, updated_at)
VALUES ('$TENANT_ID', 'Demo Corp', NOW(), NOW())
ON CONFLICT DO NOTHING;
SQL

# Seed admin user (password: admin123)
# bcrypt hash for "admin123": \$2a\$10\$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi
echo "Creating admin user..."
psql "$DATABASE_URL" <<SQL
INSERT INTO users (id, tenant_id, email, password_hash, role, created_at, updated_at)
VALUES (
  gen_random_uuid(),
  '$TENANT_ID',
  'admin@democorp.com',
  '\$2a\$10\$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
  'admin',
  NOW(),
  NOW()
)
ON CONFLICT (email) DO NOTHING;
SQL

# Seed some demo assets
echo "Creating demo assets..."
psql "$DATABASE_URL" <<SQL
INSERT INTO assets (id, tenant_id, name, type, version, metadata, created_at, updated_at)
VALUES
  (gen_random_uuid(), '$TENANT_ID', 'nodejs', 'runtime', '14.17.0', '{}', NOW(), NOW()),
  (gen_random_uuid(), '$TENANT_ID', 'express', 'package', '4.17.1', '{}', NOW(), NOW()),
  (gen_random_uuid(), '$TENANT_ID', 'lodash', 'package', '4.17.20', '{}', NOW(), NOW()),
  (gen_random_uuid(), '$TENANT_ID', 'apache-httpd', 'server', '2.4.48', '{}', NOW(), NOW()),
  (gen_random_uuid(), '$TENANT_ID', 'nginx', 'server', '1.19.6', '{}', NOW(), NOW())
ON CONFLICT DO NOTHING;
SQL

echo "✅ Seeding complete!"
echo ""
echo "Demo credentials:"
echo "  Email: admin@democorp.com"
echo "  Password: admin123"
echo ""
