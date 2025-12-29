// Package dto contains the Data Transfer Objects for the IAM module's API contract.
package dto

// RegisterRequest defines the shape of the request body for user registration.
// Binding tags are used for validation by Gin.
type RegisterRequest struct {
	FullName    string  `json:"full_name" binding:"required,min=2"`
	Email       *string `json:"email" binding:"omitempty,email"`
	PhoneNumber *string `json:"phone_number" binding:"omitempty,e164"` // e.g., +201234567890
	Password    string  `json:"password" binding:"required,min=8"`
	JobTitle    *string `json:"job_title"`
}
