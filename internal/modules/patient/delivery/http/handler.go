package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Ebrahim-hamdy/mastara-saas/internal/middleware"
	"github.com/Ebrahim-hamdy/mastara-saas/internal/modules/patient"
	"github.com/Ebrahim-hamdy/mastara-saas/internal/modules/patient/delivery/http/dto"
	"github.com/Ebrahim-hamdy/mastara-saas/internal/modules/patient/model"
	"github.com/Ebrahim-hamdy/mastara-saas/pkg/apierror"
	z "github.com/Oudwins/zog"
	"github.com/Oudwins/zog/zhttp"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service patient.Service
}

func NewHandler(service patient.Service) *Handler {
	return &Handler{service: service}
}

// RegisterPatient handles the creation of a new, fully registered patient by a staff member.
func (h *Handler) RegisterPatient(c *gin.Context) *apierror.APIError {
	payload, err := middleware.GetAuthPayload(c.Request.Context())
	if err != nil {
		return apierror.NewInternalServer(err)
	}

	var req dto.RegisterPatientRequest
	if issues := registerPatientSchema.Parse(zhttp.Request(c.Request), &req); issues != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"validation_errors": z.Issues.Flatten(issues)})
		return nil
	}

	serviceReq := patient.RegisterPatientRequest{
		ClinicID:    payload.ClinicID,
		FullName:    req.FullName,
		PhoneNumber: req.PhoneNumber,
		Email:       req.Email,
		NationalID:  req.NationalID,
		DateOfBirth: req.DateOfBirth,
	}

	profile, err := h.service.RegisterNewPatient(c.Request.Context(), payload.ClinicID, serviceReq)
	if err != nil {
		var apiErr *apierror.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return apierror.NewInternalServer(err)
	}

	c.JSON(http.StatusCreated, toProfileResponse(profile))
	return nil
}

// CompleteGuestProfile handles updating a guest profile to a registered one.
func (h *Handler) CompleteGuestProfile(c *gin.Context) *apierror.APIError {
	payload, err := middleware.GetAuthPayload(c.Request.Context())
	if err != nil {
		return apierror.NewInternalServer(err)
	}

	profileID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apierror.NewBadRequest("Invalid profile ID format.", err)
	}

	var req dto.CompleteGuestRequest
	if issues := CompleteGuestProfile.Parse(zhttp.Request(c.Request), &req); issues != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"validation_errors": z.Issues.Flatten(issues)})
		return nil
	}

	serviceReq := patient.CompleteGuestRequest{
		ClinicID:    payload.ClinicID,
		ProfileID:   profileID,
		FullName:    req.FullName,
		Email:       req.Email,
		NationalID:  req.NationalID,
		DateOfBirth: req.DateOfBirth,
	}

	profile, err := h.service.CompleteGuestRegistration(c.Request.Context(), payload.ClinicID, serviceReq)
	if err != nil {
		var apiErr *apierror.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return apierror.NewInternalServer(err)
	}

	c.JSON(http.StatusOK, toProfileResponse(profile))
	return nil
}

// GetPatient retrieves a single patient profile by staff.
func (h *Handler) GetPatient(c *gin.Context) *apierror.APIError {
	payload, err := middleware.GetAuthPayload(c.Request.Context())
	if err != nil {
		return apierror.NewInternalServer(err)
	}

	profileID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apierror.NewBadRequest("Invalid profile ID format.", err)
	}

	profile, err := h.service.GetProfileByID(c.Request.Context(), payload.ClinicID, profileID)
	if err != nil {
		var apiErr *apierror.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return apierror.NewInternalServer(err)
	}

	c.JSON(http.StatusOK, toProfileResponse(profile))
	return nil
}

// ListPatients retrieves a paginated list of patients for a clinic.
func (h *Handler) ListPatients(c *gin.Context) *apierror.APIError {
	payload, err := middleware.GetAuthPayload(c.Request.Context())
	if err != nil {
		return apierror.NewInternalServer(err)
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "25"))

	profiles, err := h.service.ListProfiles(c.Request.Context(), payload.ClinicID, page, pageSize)
	if err != nil {
		var apiErr *apierror.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return apierror.NewInternalServer(err)
	}

	response := make([]dto.ProfileResponse, len(profiles))
	for i, p := range profiles {
		response[i] = toProfileResponse(&p)
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
	return nil
}

// toProfileResponse maps the internal profile model to the public DTO.
func toProfileResponse(profile *model.Profile) dto.ProfileResponse {
	return dto.ProfileResponse{
		ID:            profile.ID,
		ClinicID:      profile.ClinicID,
		FullName:      profile.FullName,
		PhoneNumber:   profile.PhoneNumber,
		Email:         profile.Email,
		NationalID:    profile.NationalID,
		DateOfBirth:   profile.DateOfBirth,
		ProfileStatus: string(profile.ProfileStatus),
		CreatedAt:     profile.CreatedAt,
		UpdatedAt:     profile.UpdatedAt,
	}
}
