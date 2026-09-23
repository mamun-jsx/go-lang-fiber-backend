// Package middleware contains cross-cutting HTTP concerns (error handling, etc.).
package middleware

import (
	"errors"   // errors is used for errors.As
	"log"      // log writes internal errors to stdout
	"net/http" // net/http provides HTTP status codes

	"github.com/gofiber/fiber/v3" // fiber is the web framework

	"github.com/mamun-jsx/go-lang-fiber-backend/internal/apperror" // typed app errors
	"github.com/mamun-jsx/go-lang-fiber-backend/pkg/response"      // JSON envelope helpers
)

// ErrorHandler is the global Fiber error handler: every error returned by a handler ends up here.
// It converts errors into the standard JSON envelope so clients always get a consistent format.
func ErrorHandler(c fiber.Ctx, err error) error {
	// 1) Our own application errors carry their own status + message
	if appErr, ok := apperror.As(err); ok {
		// Log the real cause of server-side errors (never sent to the client)
		if appErr.Code >= http.StatusInternalServerError {
			log.Printf("[ERROR] %s %s: %v", c.Method(), c.Path(), appErr)
		}
		// Send the safe message and optional details
		return response.Error(c, appErr.Code, appErr.Message, appErr.Details)
	}

	// 2) Fiber's built-in errors (e.g. 404 route not found, 405 method not allowed)
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return response.Error(c, fiberErr.Code, fiberErr.Message, nil)
	}

	// 3) Anything else is unexpected: log it and hide the details from the client
	log.Printf("[ERROR] %s %s: %v", c.Method(), c.Path(), err)
	// Respond with a generic 500 message
	return response.Error(c, http.StatusInternalServerError, "internal server error", nil)
}
