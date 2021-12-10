package model

import "fmt"

type Role string

const (
	Admin     Role = "ADMIN"
	Reporter  Role = "REPORTER"
	Undefined Role = "UNDEFINED"
)

// String implements Stringer interface
func (r Role) String() string {
	return string(r)
}

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
	ID    string
	Name  string
	EMail string
	Pwd   []byte
	Role  Role
}

// String implements Stringer interface
func (u *User) String() string {
	return fmt.Sprintf("RequestedUser { id = %q, role = %q, email = ***, name = ***, pwd = *** }", u.ID, u.Role)
}

// RequestedUser represent a user that
// need to be created
// it contains no id - it will be generated while persisting
// and no role - it can be assigned by the admin user later on
type RequestedUser struct {
	EMail, Name string
	Pwd         []byte
}

// String implements Stringer interface
func (u *RequestedUser) String() string {
	return "RequestedUser { email = ***, name = ***, pwd = *** }"
}
