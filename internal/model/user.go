package model

import "time"

type User struct {
	ID         uint    `json:"id" db:"id"`
	Name       string  `json:"name" db:"name"`
	Username   *string `json:"username" db:"username"`
	Email      string  `json:"email" db:"email"`
	Password   string  `json:"-" db:"password"`
	AvatarPath *string `json:"avatar_path" db:"avatar_path"`
	// Role is "admin" for app-wide administrators, "user" otherwise. It's
	// separate from workspace/board roles, which only apply inside one of those.
	Role string `json:"role" db:"role"`
	// AuthProvider records how the account was created: "password" (manual
	// register) or "google" (first sign-in through Google).
	AuthProvider string    `json:"auth_provider" db:"auth_provider"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// AdminUser is a user as listed on the admin users page, along with the
// workspaces they already belong to.
type AdminUser struct {
	User
	Workspaces []AdminUserWorkspace `json:"workspaces"`
}

type AdminUserWorkspace struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type GoogleAuthRequest struct {
	Credential string `json:"credential" binding:"required"`
}

type UpdateProfileRequest struct {
	Name     string  `json:"name" binding:"required"`
	Username *string `json:"username"`
}
