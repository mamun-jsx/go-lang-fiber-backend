// Package response provides a single, consistent JSON envelope for all API responses.
package response

import "github.com/gofiber/fiber/v3" // fiber provides the request context

// Envelope is the standard JSON structure returned by every endpoint.
type Envelope struct {
	Success bool   `json:"success"`          // Success tells whether the request succeeded
	Message string `json:"message"`          // Message is a human readable message
	Data    any    `json:"data,omitempty"`   // Data holds the payload on success
	Meta    any    `json:"meta,omitempty"`   // Meta holds extra info such as pagination
	Errors  any    `json:"errors,omitempty"` // Errors holds error details on failure
}

// Success writes a successful JSON response with the given status code.
func Success(c fiber.Ctx, status int, message string, data any) error {
	// Set the status and serialize the envelope as JSON
	return c.Status(status).JSON(Envelope{Success: true, Message: message, Data: data})
}

// SuccessWithMeta writes a successful JSON response that also includes metadata.
func SuccessWithMeta(c fiber.Ctx, status int, message string, data, meta any) error {
	// Same as Success but with the Meta field filled in
	return c.Status(status).JSON(Envelope{Success: true, Message: message, Data: data, Meta: meta})
}

// Error writes a failed JSON response with the given status code.
func Error(c fiber.Ctx, status int, message string, errs any) error {
	// Set the status and serialize the error envelope as JSON
	return c.Status(status).JSON(Envelope{Success: false, Message: message, Errors: errs})
}
