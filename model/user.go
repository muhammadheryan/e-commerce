package model

import "time"

// UserEntity represents the user table entity
type UserEntity struct {
	ID           uint64     `db:"id" json:"id"`
	Name         string     `db:"name" json:"name"`
	Email        string     `db:"email" json:"email"`
	Phone        string     `db:"phone" json:"phone"`
	PasswordHash string     `db:"password_hash" json:"-"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    *time.Time `db:"updated_at" json:"updated_at,omitempty"`
}

// UserFilter for querying users
type UserFilter struct {
	ID    uint64
	Email string
	Phone string
}

// RegisterRequest for user registration
type RegisterRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginRequest for user login (accepts email or phone)
type LoginRequest struct {
	Identifier string `json:"identifier" validate:"required"` // email or phone
	Password   string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Token string `json:"token"`
}

type RegisterResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
