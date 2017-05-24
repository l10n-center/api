package model

// Role is a global constant permissions
type Role int32

const (
	// RoleAdmin may edit applications
	RoleAdmin Role = 1 << iota
	// RoleManager binding to group and may edit resources and versions of applications in this group
	RoleManager
)
