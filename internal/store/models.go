package store

import (
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/imfeelingtheagi/netctl/internal/apierror"
)

// Domain models. UUIDs are carried as canonical strings so they map cleanly onto
// OTel resource attributes (S6/S22) and across the API (S9).

// Tenant is the outermost entity and security boundary (F50).
type Tenant struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Organization is the top of the in-tenant hierarchy (Tenant → Org → Team → Project).
type Organization struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Team belongs to an organization.
type Team struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	OrgID     string    `json:"org_id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Project belongs to a team.
type Project struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	TeamID    string    `json:"team_id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// User is a tenant member (per-tenant identity; SSO lands in S18).
type User struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Role is a tenant-scoped RBAC role (enforcement lands in S18).
type Role struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProviderOperator is a provider/MSP operator — NOT a tenant user (F51).
type ProviderOperator struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BreakGlassGrant is a time-bounded, consented, audited grant for an operator to
// access one tenant's data (F51).
type BreakGlassGrant struct {
	ID         string     `json:"id"`
	OperatorID string     `json:"operator_id"`
	TenantID   string     `json:"tenant_id"`
	Reason     string     `json:"reason"`
	Scope      string     `json:"scope"`
	GrantedBy  string     `json:"granted_by"`
	GrantedAt  time.Time  `json:"granted_at"`
	ExpiresAt  time.Time  `json:"expires_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	RevokedBy  *string    `json:"revoked_by,omitempty"`
}

// notFound maps a pgx no-rows error to a domain NotFound; other errors pass
// through unchanged.
func notFound(entity string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return apierror.NotFound(entity + " not found")
	}
	return err
}
