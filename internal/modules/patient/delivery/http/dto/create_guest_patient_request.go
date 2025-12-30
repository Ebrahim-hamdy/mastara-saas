package dto

// CreateGuestPatientRequest is used for the "3-Tap Booking" flow.
type CreateGuestPatientRequest struct {
	FullName    string `json:"full_name" binding:"required,min=2"`
	PhoneNumber string `json:"phone_number" binding:"required,e164"`
}
