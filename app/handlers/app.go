package handlers

import (
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/maulanashalihin/laju-go/app/models"
	"github.com/maulanashalihin/laju-go/app/services"
	"github.com/maulanashalihin/laju-go/app/session"
)

type AppHandler struct {
	userService    *services.UserService
	store          *session.Store
	inertiaService *services.InertiaService
}

func NewAppHandler(userService *services.UserService, store *session.Store, inertiaService *services.InertiaService) *AppHandler {
	return &AppHandler{
		userService:    userService,
		store:          store,
		inertiaService: inertiaService,
	}
}

// sessionUser builds a UserResponse from session values.
func sessionUser(sess *session.Session) *models.UserResponse {
	return &models.UserResponse{
		ID:            sess.Get("user_id").(int64),
		Name:          toStr(sess.Get("name")),
		Email:         toStr(sess.Get("email")),
		Avatar:        toStr(sess.Get("avatar")),
		Role:          models.UserRole(toStr(sess.Get("role"))),
		EmailVerified: toBool(sess.Get("email_verified")),
	}
}

// Dashboard renders the main app dashboard using Inertia
func (h *AppHandler) Dashboard(c *fiber.Ctx) error {
	sess, _ := h.store.Get(c)
	user := sessionUser(sess)

	return h.inertiaService.Render(c, "app/Dashboard", fiber.Map{
		"user": user,
	})
}

// Profile returns user profile (Inertia)
func (h *AppHandler) Profile(c *fiber.Ctx) error {
	sess, _ := h.store.Get(c)
	user := sessionUser(sess)

	return h.inertiaService.Render(c, "app/Profile", fiber.Map{
		"user": user,
	})
}

// UpdateProfile updates user profile (Inertia)
func (h *AppHandler) UpdateProfile(c *fiber.Ctx) error {
	// Get user info from locals (set by AuthRequired middleware)
	userID := c.Locals("user_id")

	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	var req models.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	user, err := h.userService.UpdateProfile(userID.(int64), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update profile",
		})
	}

	// Sync session with updated name/avatar
	sess, _ := h.store.Get(c)
	if req.Name != "" {
		sess.Set("name", user.Name)
	}
	if req.Avatar != "" {
		sess.Set("avatar", user.Avatar)
	}
	sess.Save()

	return h.inertiaService.Render(c, "app/Profile", fiber.Map{
		"user":    user,
		"success": "Profile updated successfully",
	})
}

// UploadTest renders the upload test page
func (h *AppHandler) UploadTest(c *fiber.Ctx) error {
	sess, _ := h.store.Get(c)
	user := sessionUser(sess)

	return h.inertiaService.Render(c, "app/UploadTest", fiber.Map{
		"user": user,
	})
}

// UpdatePassword updates user password (Inertia)
func (h *AppHandler) UpdatePassword(c *fiber.Ctx) error {
	// Get user info from locals (set by AuthRequired middleware)
	userID := c.Locals("user_id")

	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
		ConfirmPassword string `json:"confirm_password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate passwords
	if req.NewPassword != req.ConfirmPassword {
		return h.inertiaService.Render(c, "app/Profile", fiber.Map{
			"error": "Passwords do not match",
		})
	}

	if err := validatePasswordStrength(req.NewPassword); err != nil {
		return h.inertiaService.Render(c, "app/Profile", fiber.Map{
			"error": err.Error(),
		})
	}

	// Change password
	if err := h.userService.ChangePassword(userID.(int64), req.CurrentPassword, req.NewPassword); err != nil {
		return h.inertiaService.Render(c, "app/Profile", fiber.Map{
			"error": err.Error(),
		})
	}

	// Invalidate all sessions for this user — forces re-login on all devices.
	// The current session is also invalidated; user stays logged in via the
	// session cookie which will be re-seeded on the next request.
	uid := userID.(int64)
	if err := h.userService.InvalidateAllSessions(uid); err != nil {
		slog.Error("failed to invalidate sessions after password change", "user_id", uid, "error", err)
	}

	// Re-create the current session so the user doesn't get logged out immediately
	sess, _ := h.store.Get(c)
	sess.Regenerate()
	user := sessionUser(sess)

	return h.inertiaService.Render(c, "app/Profile", fiber.Map{
		"user":    user,
		"success": "Password changed successfully",
	})
}

// toStr safely extracts a string from an interface{}, defaulting to empty string.
func toStr(v interface{}) string {
	if v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

// toBool safely extracts a bool from an interface{}, defaulting to false.
func toBool(v interface{}) bool {
	if v == nil {
		return false
	}
	b, ok := v.(bool)
	if !ok {
		return false
	}
	return b
}

// validatePasswordStrength enforces minimum password complexity:
// at least 8 characters, one uppercase, one lowercase, one digit.
func validatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	var hasUpper, hasLower, hasDigit bool
	for _, r := range password {
		switch {
		case r >= 'A' && r <= 'Z':
			hasUpper = true
		case r >= 'a' && r <= 'z':
			hasLower = true
		case r >= '0' && r <= '9':
			hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		return fmt.Errorf("password must contain at least one uppercase letter, one lowercase letter, and one digit")
	}
	return nil
}
