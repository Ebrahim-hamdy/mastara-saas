package model

// Permission represents an atomic capability in the system.
type Permission struct {
	ID            int16  `db:"id"`
	PermissionKey string `db:"permission_key"`
}
