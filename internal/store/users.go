package store

import (
	"context"

	"github.com/imfeelingtheagi/netctl/internal/tenancy"
)

// Users is the tenant-scoped user repository (per-tenant identity; SSO in S18).
type Users struct{}

const userCols = `id::text, tenant_id::text, email, display_name, status, created_at, updated_at`

func scanUser(row interface{ Scan(...any) error }, u *User) error {
	return row.Scan(&u.ID, &u.TenantID, &u.Email, &u.DisplayName, &u.Status, &u.CreatedAt, &u.UpdatedAt)
}

// Create inserts a user in the caller's tenant.
func (Users) Create(ctx context.Context, s tenancy.Scope, email, displayName string) (*User, error) {
	var u User
	err := scanUser(s.Q.QueryRow(ctx,
		`INSERT INTO users (tenant_id, email, display_name) VALUES ($1, $2, $3) RETURNING `+userCols,
		s.Tenant.String(), email, displayName), &u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Get returns a user by id.
func (Users) Get(ctx context.Context, s tenancy.Scope, id string) (*User, error) {
	var u User
	if err := scanUser(s.Q.QueryRow(ctx, `SELECT `+userCols+` FROM users WHERE id = $1`, id), &u); err != nil {
		return nil, notFound("user", err)
	}
	return &u, nil
}

// GetByEmail returns a user by email within the tenant.
func (Users) GetByEmail(ctx context.Context, s tenancy.Scope, email string) (*User, error) {
	var u User
	if err := scanUser(s.Q.QueryRow(ctx, `SELECT `+userCols+` FROM users WHERE email = $1`, email), &u); err != nil {
		return nil, notFound("user", err)
	}
	return &u, nil
}

// List returns the tenant's users.
func (Users) List(ctx context.Context, s tenancy.Scope) ([]User, error) {
	rows, err := s.Q.Query(ctx, `SELECT `+userCols+` FROM users ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []User
	for rows.Next() {
		var u User
		if err := scanUser(rows, &u); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// UpdateStatus changes a user's status (active/suspended/disabled).
func (Users) UpdateStatus(ctx context.Context, s tenancy.Scope, id, status string) (*User, error) {
	var u User
	if err := scanUser(s.Q.QueryRow(ctx,
		`UPDATE users SET status = $2, updated_at = now() WHERE id = $1 RETURNING `+userCols,
		id, status), &u); err != nil {
		return nil, notFound("user", err)
	}
	return &u, nil
}
