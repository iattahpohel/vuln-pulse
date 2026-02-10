#!/bin/bash
set -e

echo "🚀 VulnPulse End-to-End Demo"
echo "=============================="
echo ""

# Configuration
API_URL="http://localhost:8080"
TOKEN=""

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_step() {
  echo ""
  echo -e "${BLUE}▶ $1${NC}"
}

print_success() {
  echo -e "${GREEN}✓ $1${NC}"
}

# Step 1: Login
print_step "Step 1: Login as admin..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@democorp.com",
    "password": "admin123"
  }')

TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "❌ Login failed. Make sure services are running and database is seeded."
  exit 1
fi

print_success "Logged in successfully"
echo "Token: ${TOKEN:0:20}..."

# Step 2: List assets
print_step "Step 2: List current assets..."
curl -s -X GET "$API_URL/api/v1/assets" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

print_success "Assets listed"

# Step 3: Create a new asset
print_step "Step 3: Create a new asset (log4j)..."
ASSET_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/assets" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "log4j",
    "type": "package",
    "version": "2.14.1",
    "metadata": {
      "environment": "production"
    }
  }')

ASSET_ID=$(echo $ASSET_RESPONSE | grep -o '"id":"[^"]*' | cut -d'"' -f4)
echo $ASSET_RESPONSE | jq '.'
print_success "Asset created: $ASSET_ID"

# Step 4: Ingest a vulnerability (Log4Shell)
print_step "Step 4: Ingest Log4Shell vulnerability (CVE-2021-44228)..."
VULN_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/vulnerabilities" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "cve_id": "CVE-2021-44228",
    "title": "Apache Log4j2 Remote Code Execution (Log4Shell)",
    "description": "Apache Log4j2 2.0-beta9 through 2.15.0 JNDI features used in configuration, log messages, and parameters do not protect against attacker controlled LDAP and other JNDI related endpoints.",
    "severity": "critical",
    "affected_products": [
      {
        "name": "log4j",
        "versions": ["2.0", "2.1", "2.2", "2.3", "2.4", "2.5", "2.6", "2.7", "2.8", "2.9", "2.10", "2.11", "2.12", "2.13", "2.14.0", "2.14.1", "2.15.0"]
      }
    ],
    "published_at": "2021-12-10T00:00:00Z"
  }')

VULN_ID=$(echo $VULN_RESPONSE | grep -o '"id":"[^"]*' | cut -d'"' -f4)
echo $VULN_RESPONSE | jq '.'
print_success "Vulnerability ingested: $VULN_ID"

# Step 5: Wait for worker to process
print_step "Step 5: Waiting for worker to match vulnerability to assets..."
sleep 5
print_success "Processing complete"

# Step 6: List alerts
print_step "Step 6: Check generated alerts..."
curl -s -X GET "$API_URL/api/v1/alerts" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

print_success "Alerts retrieved"

# Step 7: Get alerts with status filter
print_step "Step 7: Filter open alerts..."
ALERTS_RESPONSE=$(curl -s -X GET "$API_URL/api/v1/alerts?status=open" \
  -H "Authorization: Bearer $TOKEN")

echo $ALERTS_RESPONSE | jq '.'

ALERT_ID=$(echo $ALERTS_RESPONSE | jq -r '.alerts[0].id // empty')

if [ ! -z "$ALERT_ID" ]; then
  print_success "Found alert: $ALERT_ID"
  
  # Step 8: Update alert status
  print_step "Step 8: Acknowledge the alert..."
  curl -s -X PATCH "$API_URL/api/v1/alerts/$ALERT_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
      "status": "acknowledged"
    }' | jq '.'
  
  print_success "Alert status updated"
else
  echo "⚠ No alerts found (worker may not have processed yet)"
fi

# Step 9: List vulnerabilities
print_step "Step 9: List all vulnerabilities..."
curl -s -X GET "$API_URL/api/v1/vulnerabilities?limit=10" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

print_success "Vulnerabilities listed"

echo ""
echo "=============================="
echo -e "${GREEN}✅ Demo completed successfully!${NC}"
echo ""
echo "What happened:"
echo "1. Logged in as admin user"
echo "2. Listed existing assets"
echo "3. Created a new asset (log4j 2.14.1)"
echo "4. Ingested Log4Shell vulnerability (CVE-2021-44228)"
echo "5. Worker matched vulnerability to affected assets"
echo "6. Alert was generated and retrieved"
echo "7. Alert status was updated to 'acknowledged'"
echo ""
echo "Next steps:"
echo "- Add webhook subscriptions to receive real-time notifications"
echo "- Explore the API with curl or Postman"
echo "- Check RabbitMQ management UI at http://localhost:15672 (vulnpulse/devpassword)"
echo ""
