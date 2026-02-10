package domain

import (
	"time"

	"github.com/google/uuid"
)

// Tenant represents a customer organization
type Tenant struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// User represents a user within a tenant
type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	TenantID     uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         string    `json:"role" db:"role"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Role constants
const (
	RoleAdmin   = "admin"
	RoleAnalyst = "analyst"
	RoleViewer  = "viewer"
)

// Asset represents a software component, server, or infrastructure
type Asset struct {
	ID        uuid.UUID              `json:"id" db:"id"`
	TenantID  uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	Name      string                 `json:"name" db:"name"`
	Type      string                 `json:"type" db:"type"`
	Version   *string                `json:"version,omitempty" db:"version"`
	Metadata  map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
}

// Vulnerability represents a CVE or security advisory
type Vulnerability struct {
	ID               uuid.UUID              `json:"id" db:"id"`
	CveID            string                 `json:"cve_id" db:"cve_id"`
	Title            string                 `json:"title" db:"title"`
	Description      string                 `json:"description,omitempty" db:"description"`
	Severity         string                 `json:"severity" db:"severity"`
	AffectedProducts []AffectedProduct      `json:"affected_products,omitempty" db:"affected_products"`
	PublishedAt      *time.Time             `json:"published_at,omitempty" db:"published_at"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

// AffectedProduct represents a product affected by a vulnerability
type AffectedProduct struct {
	Name     string   `json:"name"`
	Versions []string `json:"versions"`
}

// Severity constants
const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
)

// Alert represents a matched vulnerability for a specific asset
type Alert struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	TenantID        uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	AssetID         uuid.UUID  `json:"asset_id" db:"asset_id"`
	VulnerabilityID uuid.UUID  `json:"vulnerability_id" db:"vulnerability_id"`
	Severity        string     `json:"severity" db:"severity"`
	Status          string     `json:"status" db:"status"`
	MatchedAt       time.Time  `json:"matched_at" db:"matched_at"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
}

// Alert status constants
const (
	AlertStatusOpen          = "open"
	AlertStatusAcknowledged  = "acknowledged"
	AlertStatusResolved      = "resolved"
	AlertStatusFalsePositive = "false_positive"
)

// WebhookSubscription represents a webhook endpoint for event notifications
type WebhookSubscription struct {
	ID        uuid.UUID `json:"id" db:"id"`
	TenantID  uuid.UUID `json:"tenant_id" db:"tenant_id"`
	URL       string    `json:"url" db:"url"`
	Secret    string    `json:"-" db:"secret"`
	Events    []string  `json:"events" db:"events"`
	Active    bool      `json:"active" db:"active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID             uuid.UUID              `json:"id" db:"id"`
	SubscriptionID uuid.UUID              `json:"subscription_id" db:"subscription_id"`
	EventType      string                 `json:"event_type" db:"event_type"`
	Payload        map[string]interface{} `json:"payload" db:"payload"`
	Status         string                 `json:"status" db:"status"`
	Attempts       int                    `json:"attempts" db:"attempts"`
	LastAttemptAt  *time.Time             `json:"last_attempt_at,omitempty" db:"last_attempt_at"`
	NextRetryAt    *time.Time             `json:"next_retry_at,omitempty" db:"next_retry_at"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
}

// Webhook delivery status constants
const (
	DeliveryStatusPending = "pending"
	DeliveryStatusSuccess = "success"
	DeliveryStatusFailed  = "failed"
)

// Event type constants
const (
	EventAlertCreated  = "alert.created"
	EventAlertResolved = "alert.resolved"
)
