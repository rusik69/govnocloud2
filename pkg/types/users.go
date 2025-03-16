package types

// User represents a user in the database
type User struct {
	Name       string   `json:"name"`
	Password   string   `json:"password"`
	Namespaces []string `json:"namespaces"`
}

// UserList is a list of users
type UserList []User
