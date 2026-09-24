package model

// Role adalah baris pada tabel roles.
type Role struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Permission adalah baris pada tabel permissions.
type Permission struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
