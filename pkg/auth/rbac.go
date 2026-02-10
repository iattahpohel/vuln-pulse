package auth

import (
	"fmt"
)

// Permission represents an action that can be performed
type Permission string

const (
	// Tenant permissions
	PermissionCreateTenant Permission = "tenant:create"
	PermissionListTenants  Permission = "tenant:list"

	// Asset permissions
	PermissionCreateAsset Permission = "asset:create"
	PermissionReadAsset   Permission = "asset:read"
	PermissionUpdateAsset Permission = "asset:update"
	PermissionDeleteAsset Permission = "asset:delete"

	// Vulnerability permissions
	PermissionIngestVuln Permission = "vulnerability:ingest"
	PermissionReadVuln   Permission = "vulnerability:read"

	// Alert permissions
	PermissionReadAlert   Permission = "alert:read"
	PermissionUpdateAlert Permission = "alert:update"

	// Webhook permissions
	PermissionCreateWebhook Permission = "webhook:create"
	PermissionReadWebhook   Permission = "webhook:read"
	PermissionUpdateWebhook Permission = "webhook:update"
	PermissionDeleteWebhook Permission = "webhook:delete"
)

// rolePermissions maps roles to their permissions
var rolePermissions = map[string][]Permission{
	"admin": {
		PermissionCreateTenant,
		PermissionListTenants,
		PermissionCreateAsset,
		PermissionReadAsset,
		PermissionUpdateAsset,
		PermissionDeleteAsset,
		PermissionIngestVuln,
		PermissionReadVuln,
		PermissionReadAlert,
		PermissionUpdateAlert,
		PermissionCreateWebhook,
		PermissionReadWebhook,
		PermissionUpdateWebhook,
		PermissionDeleteWebhook,
	},
	"analyst": {
		PermissionCreateAsset,
		PermissionReadAsset,
		PermissionUpdateAsset,
		PermissionDeleteAsset,
		PermissionIngestVuln,
		PermissionReadVuln,
		PermissionReadAlert,
		PermissionUpdateAlert,
		PermissionCreateWebhook,
		PermissionReadWebhook,
		PermissionUpdateWebhook,
		PermissionDeleteWebhook,
	},
	"viewer": {
		PermissionReadAsset,
		PermissionReadVuln,
		PermissionReadAlert,
		PermissionReadWebhook,
	},
}

// HasPermission checks if a role has a specific permission
func HasPermission(role string, permission Permission) error {
	permissions, ok := rolePermissions[role]
	if !ok {
		return fmt.Errorf("unknown role: %s", role)
	}

	for _, p := range permissions {
		if p == permission {
			return nil
		}
	}

	return fmt.Errorf("permission denied: %s requires %s", role, permission)
}
