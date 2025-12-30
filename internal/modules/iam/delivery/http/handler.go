package http

import (
	"errors"
	"net/http"

	"github.com/Ebrahim-hamdy/mastara-saas/internal/middleware"
	"github.com/Ebrahim-hamdy/mastara-saas/internal/modules/iam"
	"github.com/Ebrahim-hamdy/mastara-saas/internal/modules/iam/delivery/http/dto"
	"github.com/Ebrahim-hamdy/mastara-saas/internal/modules/iam/model"
	"github.com/Ebrahim-hamdy/mastara-saas/pkg/apierror"
	z "github.com/Oudwins/zog"
	"github.com/Oudwins/zog/zhttp"
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

// InviteEmployee handles the HTTP request for inviting a new staff member.
func (h *Handler) InviteEmployee(c *gin.Context) *apierror.APIError {
	inviterPayload, err := middleware.GetAuthPayload(c.Request.Context())
	if err != nil {
		// This indicates a server configuration error, as this handler should only be
		// reached if the authenticator middleware has already run successfully.
		return apierror.NewInternalServer(err)
	}
	var req dto.InviteEmployeeRequest
	if issues := inviteEmployeeSchema.Parse(zhttp.Request(c.Request), &req); issues != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"validation_errors": z.Issues.Flatten(issues)})
		return nil
	}

	serviceReq := iam.InviteEmployeeRequest{
		FullName:    req.FullName,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		JobTitle:    req.JobTitle,
	}

	employee, err := h.service.InviteEmployee(c.Request.Context(), inviterPayload.ClinicID, inviterPayload.UserID, serviceReq)
	if err != nil {
		var apiErr *apierror.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return apierror.NewInternalServer(err)
	}

	c.JSON(http.StatusCreated, toEmployeeResponse(employee))
	return nil
}

// LoginEmployee handles the HTTP request for staff authentication.
func (h *Handler) LoginEmployee(c *gin.Context) *apierror.APIError {
	var req dto.LoginRequest
	if issues := loginRequestSchema.Parse(zhttp.Request(c.Request), &req); issues != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"validation_errors": z.Issues.Flatten(issues)})
		return nil
	}

	serviceReq := iam.LoginEmployeeRequest{
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
	}

	token, employee, err := h.service.LoginEmployee(c.Request.Context(), serviceReq)
	if err != nil {
		var apiErr *apierror.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return apierror.NewInternalServer(err)
	}

	response := dto.LoginResponse{
		Token:    token,
		Employee: toEmployeeResponse(employee),
	}

	c.JSON(http.StatusOK, response)
	return nil
}

// toEmployeeResponse maps the internal employee and its nested profile to the public DTO.
func toEmployeeResponse(employee *model.Employee) dto.EmployeeResponse {
	return dto.EmployeeResponse{
		ID:          employee.ProfileID,
		ClinicID:    employee.ClinicID,
		Email:       employee.Profile.Email,
		PhoneNumber: employee.Profile.PhoneNumber,
		FullName:    employee.Profile.FullName,
		JobTitle:    employee.JobTitle,
		Status:      string(employee.Status),
	}
}
