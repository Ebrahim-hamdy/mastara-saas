// Package httpjson provides hardened helpers for handling JSON in HTTP requests and responses.
package httpjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Ebrahim-hamdy/mastara-saas/pkg/apierror"
)

const defaultMaxBodyBytes = 1_048_576 // 1 MB

// DecodeJSON provides a secure way to decode JSON from an HTTP request body.
// It enforces a request body size limit, checks for the correct Content-Type,
// and prevents unknown fields in the JSON payload.
func DecodeJSON[T any](w http.ResponseWriter, r *http.Request) (T, *apierror.APIError) {
	var dest T

	// Enforce a max body size to prevent DoS attacks.
	r.Body = http.MaxBytesReader(w, r.Body, defaultMaxBodyBytes)

	// Check for the correct Content-Type header.
	if r.Header.Get("Content-Type") != "application/json" {
		return dest, apierror.NewBadRequest("Content-Type header must be 'application/json'", nil)
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // Strict parsing

	if err := dec.Decode(&dest); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at character %d)", syntaxError.Offset)
			return dest, apierror.NewBadRequest(msg, err)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return dest, apierror.NewBadRequest("Request body contains badly-formed JSON", err)

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at character %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return dest, apierror.NewBadRequest(msg, err)

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			return dest, apierror.NewBadRequest(msg, err)

		case errors.As(err, &maxBytesError):
			msg := fmt.Sprintf("Request body must not be larger than %d bytes", maxBytesError.Limit)
			return dest, apierror.NewBadRequest(msg, err)

		case errors.Is(err, io.EOF):
			return dest, apierror.NewBadRequest("Request body must not be empty", err)

		default:
			return dest, apierror.NewInternalServer(err)
		}
	}

	// Check if there is more than one JSON object in the body.
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return dest, apierror.NewBadRequest("Request body must only contain a single JSON object", err)
	}

	return dest, nil
}
