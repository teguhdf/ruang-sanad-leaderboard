package handlers

import (
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/maulanashalihin/laju-go/app/services"
	"github.com/maulanashalihin/laju-go/app/session"
)

type PasswordResetHandler struct {
	mailerService  *services.MailerService
	userService    *services.UserService
	store          *session.Store
	inertiaService *services.InertiaService
}

func NewPasswordResetHandler(
	mailerService *services.MailerService,
	userService *services.UserService,
	store *session.Store,
	inertiaService *services.InertiaService,
) *PasswordResetHandler {
	return &PasswordResetHandler{
		mailerService:  mailerService,
		userService:    userService,
		store:          store,
		inertiaService: inertiaService,
	}
}

func (h *PasswordResetHandler) ShowForgotPasswordForm(c *fiber.Ctx) error {
	return h.inertiaService.Render(c, "auth/ForgotPassword", fiber.Map{
		"Title": "Forgot Password",
	})
}

func (h *PasswordResetHandler) SendResetLink(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}

	if err := c.BodyParser(&req); err != nil {
		h.store.Flash(c, "error", "Invalid request body")
		return h.inertiaService.Redirect(c, "/forgot-password")
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		h.store.Flash(c, "error", "Please provide a valid email address")
		return h.inertiaService.Redirect(c, "/forgot-password")
	}

	// Keep the same generic response whether or not the account exists.
	successMsg := "If an account exists with that email, we've sent a password reset link."
	user, err := h.userService.GetProfileByEmail(req.Email)
	if err == nil {
		if err := h.mailerService.SendPasswordResetEmail(c.Context(), user.Email, user.ID); err != nil {
			slog.Error("failed to send password reset email", "user_id", user.ID, "error", err)
		}
	}

	h.store.Flash(c, "success", successMsg)
	return h.inertiaService.Redirect(c, "/forgot-password")
}

func (h *PasswordResetHandler) ShowResetPasswordForm(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		h.store.Flash(c, "error", "Invalid reset link")
		return h.inertiaService.Redirect(c, "/forgot-password")
	}

	if _, err := h.mailerService.ValidateResetToken(c.Context(), token); err != nil {
		h.store.Flash(c, "error", "Invalid or expired reset link")
		return h.inertiaService.Redirect(c, "/forgot-password")
	}

	return h.inertiaService.Render(c, "auth/ResetPassword", fiber.Map{
		"Title": "Reset Password",
		"token": token,
	})
}

func (h *PasswordResetHandler) ResetPassword(c *fiber.Ctx) error {
	token := c.Params("token")
	resetPath := "/reset-password/" + token

	var req struct {
		Password          string `json:"password"`
		PasswordConfirmed string `json:"password_confirmation"`
	}

	if err := c.BodyParser(&req); err != nil {
		h.store.Flash(c, "error", "Invalid request body")
		return h.inertiaService.Redirect(c, resetPath)
	}

	if req.Password == "" || len(req.Password) < 8 {
		h.store.Flash(c, "error", "Password must be at least 8 characters")
		return h.inertiaService.Redirect(c, resetPath)
	}

	if err := validatePasswordStrength(req.Password); err != nil {
		h.store.Flash(c, "error", err.Error())
		return h.inertiaService.Redirect(c, resetPath)
	}

	if req.Password != req.PasswordConfirmed {
		h.store.Flash(c, "error", "Passwords do not match")
		return h.inertiaService.Redirect(c, resetPath)
	}

	tokenEntry, err := h.mailerService.ValidateResetToken(c.Context(), token)
	if err != nil {
		h.store.Flash(c, "error", "Invalid or expired reset link")
		return h.inertiaService.Redirect(c, "/forgot-password")
	}

	user, err := h.userService.GetProfile(tokenEntry.UserID)
	if err != nil {
		h.store.Flash(c, "error", "User not found")
		return h.inertiaService.Redirect(c, "/forgot-password")
	}

	hashedPassword, err := services.HashPassword(req.Password)
	if err != nil {
		slog.Error("failed to hash reset password", "user_id", user.ID, "error", err)
		h.store.Flash(c, "error", "Failed to reset password")
		return h.inertiaService.Redirect(c, resetPath)
	}

	if err := h.userService.UpdatePassword(user.ID, hashedPassword); err != nil {
		slog.Error("failed to update reset password", "user_id", user.ID, "error", err)
		h.store.Flash(c, "error", "Failed to reset password")
		return h.inertiaService.Redirect(c, resetPath)
	}

	h.mailerService.InvalidateResetToken(c.Context(), token)
	if err := h.userService.InvalidateAllSessions(user.ID); err != nil {
		slog.Error("failed to invalidate sessions after password reset", "user_id", user.ID, "error", err)
	}

	h.store.Flash(c, "success", "Password reset successfully. You can now login with your new password.")
	return h.inertiaService.Redirect(c, "/login")
}
