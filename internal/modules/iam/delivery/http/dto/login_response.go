package dto

// LoginResponse defines the shape of a successful login response.
type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
