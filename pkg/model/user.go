package model

type Role string

const (
	Admin     Role = "ADMIN"
	Reporter  Role = "REPORTER"
	Undefined Role = "UNDEFINED"
)

func RoleFromString(s string) Role {
	switch s {
	case string(Admin):
		return Admin
	case string(Reporter):
		return Reporter
	default:
		return Undefined
	}
}

// User represents an application user
type User struct {
	ID      string
	Name    string
	EMail   string
	PwdHash string
	Role    Role
}
