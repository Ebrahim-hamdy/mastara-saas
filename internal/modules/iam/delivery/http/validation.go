package http

import (
	"regexp"

	"github.com/Ebrahim-hamdy/mastara-saas/internal/modules/iam/delivery/http/dto"
	z "github.com/Oudwins/zog"
)

// Pre-compile regex patterns for performance.
var e164Regex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

// Defines the schema for the LoginRequest DTO.
var loginRequestSchema = z.Struct(z.Shape{
	"email":        z.String().Email(z.Message("A valid email address is required.")).Optional(),
	"phone_number": z.String().Match(e164Regex, z.Message("A valid E.164 phone number is required.")).Optional(),
	"password":     z.String().Required(z.Message("Password is required.")),
}).TestFunc( // Use TestFunc for cross-field validation on structs.
	func(data any, ctx z.Ctx) bool {
		req, ok := data.(*dto.LoginRequest)
		if !ok {
			return false
		}
		return req.Email != nil || req.Phone != nil
	},
	z.Message("Either email or phone_number must be provided."),
)

// Schema for inviting a new employee.
var inviteEmployeeSchema = z.Struct(z.Shape{
	"full_name":    z.String().Min(4, z.Message("Full name must be at least 4 characters.")),
	"email":        z.String().Email(z.Message("A valid email address is required.")).Optional(),
	"phone_number": z.String().Match(e164Regex, z.Message("A valid E.164 phone number is required.")).Optional(),
	"job_title":    z.String().Optional(),
}).TestFunc(
	func(data any, ctx z.Ctx) bool {
		req, ok := data.(*dto.InviteEmployeeRequest)
		if !ok {
			return false
		}
		return req.Email != nil || req.PhoneNumber != nil
	},
	z.Message("Either email or phone_number must be provided for an invitation."),
)
