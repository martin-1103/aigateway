// models/role.enum.go
package models

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleUser     Role = "user"
	RoleProvider Role = "provider"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleUser, RoleProvider:
		return true
	}
	return false
}

func (r Role) CanAccessAI() bool {
	return r == RoleAdmin || r == RoleUser
}

func (r Role) CanManageUsers() bool {
	return r == RoleAdmin
}

func (r Role) CanManageAccounts() bool {
	return r == RoleAdmin || r == RoleProvider
}

func (r Role) CanManageAPIKeys() bool {
	return r == RoleAdmin || r == RoleUser
}
