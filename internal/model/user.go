package model

import "time"

type User struct {
	ID         uint      `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`
	Username   *string   `json:"username" db:"username"`
	Email      string    `json:"email" db:"email"`
	Password   string    `json:"-" db:"password"`
	AvatarPath *string   `json:"avatar_path" db:"avatar_path"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
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
