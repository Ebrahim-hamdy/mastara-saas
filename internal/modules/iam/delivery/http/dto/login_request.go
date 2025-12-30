// Package dto contains the Data Transfer Objects for the IAM module's API contract.
package dto

// LoginRequest defines the shape of the request body for user login.
type LoginRequest struct {
	ClinicID string  `json:"clinic_id" binding:"required,uuid"`
	Email    *string `json:"email" binding:"omitempty,email"`
	Phone    *string `json:"phone_number" binding:"omitempty,e164"`
	Password string  `json:"password" binding:"required"`
}
