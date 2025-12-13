package dto

import "time"

type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Phone     string `json:"phone_number"`
	BirthDate string `json:"birth_date"`
}

type UploadAvatarResponse struct {
	Url string `json:"url"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type UserDocumentResponse struct {
	ID         string    `json:"id"`
	Filename   string    `json:"filename"`
	MimeType   string    `json:"mime_type"`
	Size       int64     `json:"size"`
	UploadedAt time.Time `json:"uploaded_at"`
	Status     string    `json:"status"`
}