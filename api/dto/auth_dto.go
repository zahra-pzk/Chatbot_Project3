package dto

import (
	"github.com/google/uuid"
	"github.com/zahra-pzk/Chatbot_Project3/util"
)

type CreateGuestUserRequest struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
}

type CreateGuestUserResponse struct {
	UserExternalID string          `json:"user_external_id"`
	FirstName      string          `json:"first_name"`
	LastName       string          `json:"last_name"`
	Email          string          `json:"email"`
	Role           string          `json:"role"`
	IsOnline       bool            `json:"is_online"`
	LastSeen       util.JalaliTime `json:"last_seen,omitempty"`
	CreatedAt      util.JalaliTime `json:"created_at"`
	UpdatedAt      util.JalaliTime `json:"updated_at"`
}

type CreateUserRequest struct {
	FirstName  string          `json:"first_name" validate:"required"`
	LastName   string          `json:"last_name" validate:"required"`
	Username   string          `json:"username" validate:"required"`
	Password   string          `json:"password" validate:"required,min=8"`
	Email      string          `json:"email" validate:"required,email"`
	Phone      string          `json:"phone_number" validate:"required"`
	AvatarURLs []string        `json:"photos"`
	BirthDate  util.JalaliTime `json:"birth_date" validate:"required"`
}

type UserResponse struct {
	UserExternalID string          `json:"user_external_id"`
	FirstName      string          `json:"first_name"`
	LastName       string          `json:"last_name"`
	Username       string          `json:"username,omitempty"`
	Email          string          `json:"email"`
	Phone          string          `json:"phone_number,omitempty"`
	Role           string          `json:"role"`
	HashedPassword string          `json:"hashed_password"`
	Status         string          `json:"status"`
	AvatarURLs     []string        `json:"photos"`
	BirthDate      util.JalaliTime `json:"birth_date,omitempty"`
	IsOnline       bool            `json:"is_online"`
	CreatedAt      util.JalaliTime `json:"created_at"`
	UpdatedAt      util.JalaliTime `json:"updated_at"`
	LastSeen       util.JalaliTime `json:"last_seen,omitempty"`
}

type LoginUserWithUsernameRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginUserWithEmailRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginUserWithPhoneRequest struct {
	Phone    string `json:"phone_number" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginUserResponse struct {
	SessionExternalID     uuid.UUID       `json:"session_external_id"`
	AccessToken           string          `json:"access_token"`
	AccessTokenExpiresAt  util.JalaliTime `json:"access_token_expires_at"`
	RefreshToken          string          `json:"refresh_token"`
	RefreshTokenExpiresAt util.JalaliTime `json:"refresh_token_expires_at"`
	User                  UserResponse    `json:"user"`
}

