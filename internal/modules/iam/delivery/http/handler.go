package http

import (
	"errors"
	"net/http"

	"github.com/Ebrahim-hamdy/mastara-saas/internal/modules/iam"
	"github.com/Ebrahim-hamdy/mastara-saas/internal/modules/iam/delivery/http/dto"
	"github.com/Ebrahim-hamdy/mastara-saas/internal/modules/iam/model"
	"github.com/Ebrahim-hamdy/mastara-saas/pkg/apierror"
	"github.com/Ebrahim-hamdy/mastara-saas/pkg/httpjson"
	"github.com/gin-gonic/gin"
)

// Handler holds the dependencies for the IAM HTTP handlers.
type Handler struct {
	service iam.Service
}

// NewHandler creates a new IAM handler with the given service.
func NewHandler(service iam.Service) *Handler {
	return &Handler{service: service}
}

// RegisterUser handles the HTTP request for creating a new user.
func (h *Handler) RegisterUser(c *gin.Context) *apierror.APIError {
	// 1. Decode and validate the request body into our DTO.
	req, apiErr := httpjson.DecodeJSON[dto.RegisterRequest](c.Writer, c.Request)
	if apiErr != nil {
		return apiErr
	}

	// Basic semantic validation
	if req.Email == nil && req.PhoneNumber == nil {
		return apierror.NewBadRequest("Either email or phone_number must be provided.", nil)
	}

	// 2. Map the DTO to the service layer's request struct.
	serviceReq := iam.RegisterUserRequest{
		FullName:    req.FullName,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		Password:    req.Password,
		JobTitle:    req.JobTitle,
	}

	// 3. Call the business logic service.
	user, err := h.service.RegisterUser(c.Request.Context(), serviceReq)
	if err != nil {
		// The service layer returns an APIError, so we can just pass it up.
		var apiErr *apierror.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return apierror.NewInternalServer(err)
	}

	// 4. Map the domain model to our public response DTO and send.
	c.JSON(http.StatusCreated, toUserResponse(user))
	return nil
}

// LoginUser handles the HTTP request for user authentication.
func (h *Handler) LoginUser(c *gin.Context) *apierror.APIError {
	req, apiErr := httpjson.DecodeJSON[dto.LoginRequest](c.Writer, c.Request)
	if apiErr != nil {
		return apiErr
	}

	if req.Email == nil && req.Phone == nil {
		return apierror.NewBadRequest("Either email or phone_number must be provided.", nil)
	}

	serviceReq := iam.LoginUserRequest{
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
	}

	token, user, err := h.service.LoginUser(c.Request.Context(), serviceReq)
	if err != nil {
		var apiErr *apierror.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return apierror.NewInternalServer(err)
	}

	response := dto.LoginResponse{
		Token: token,
		User:  toUserResponse(user),
	}

	c.JSON(http.StatusOK, response)
	return nil
}

// toUserResponse is a helper function to map the internal user model to the public DTO.
func toUserResponse(user *model.User) dto.UserResponse {
	return dto.UserResponse{
		ID:       user.ID,
		ClinicID: user.ClinicID,
		Email:    user.Email,
		Phone:    user.PhoneNumber,
		FullName: user.FullName,
		JobTitle: user.JobTitle,
		Status:   user.Status,
	}
}
