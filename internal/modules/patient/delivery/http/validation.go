package http

import (
	"regexp"
	"time"

	z "github.com/Oudwins/zog"
)

var e164Regex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

// Schema for creating a new, fully registered patient by staff.
var registerPatientSchema = z.Struct(z.Shape{
	"full_name":     z.String().Min(2, z.Message("Full name must be at least 2 characters.")),
	"phone_number":  z.String().Match(e164Regex, z.Message("A valid E.164 phone number is required.")),
	"email":         z.String().Email(z.Message("A valid email address is required.")).Optional(),
	"national_id":   z.String().Optional(),
	"date_of_birth": z.Time(z.Time.Format(time.DateOnly)).Optional(), // Expects "YYYY-MM-DD"
})

// Schema for updating a patient's details (including completing a guest profile).
var CompleteGuestProfile = z.Struct(z.Shape{
	"full_name":   z.String().Min(2, z.Message("Full name must be at least 2 characters.")),
	"email":       z.String().Email(z.Message("A valid email address is required.")).Optional(),
	"national_id": z.String().Optional(),
	// "date_of_birth": z.Time(z.TimeOpts{Layout: "2006-01-02"}).Optional(),
	"date_of_birth": z.Time(z.Time.Format(time.DateOnly)).Optional(),
})
