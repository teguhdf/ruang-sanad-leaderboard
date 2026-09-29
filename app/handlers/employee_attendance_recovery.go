package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maulanashalihin/laju-go/app/models"
)

func (h *EmployeeAttendanceHandler) RequestForgotCheckout(c *fiber.Ctx) error {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var input models.ForgotCheckoutInput
	if err := c.BodyParser(&input); err != nil {
		h.store.Flash(c, "error", "Data lupa checkout tidak dapat dibaca.")
		return h.inertiaService.Redirect(c, "/app/my-attendance")
	}
	if err := h.service.RequestForgotCheckout(userID, input); err != nil {
		setEmployeeAttendanceFlashError(h, c, err)
		return h.inertiaService.Redirect(c, "/app/my-attendance")
	}

	h.store.Flash(c, "success", "Sesi lama ditandai incomplete. Koreksi checkout dikirim ke SPV dan Anda bisa memulai sesi baru.")
	return h.inertiaService.Redirect(c, "/app/my-attendance")
}
