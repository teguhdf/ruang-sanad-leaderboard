package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func (h *EmployeeAttendanceHandler) AttendanceReviews(c *fiber.Ctx) error {
	ownerUserID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}
	adjustments, overtime, err := h.service.ListPendingReviews(ownerUserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load attendance reviews"})
	}
	sess, _ := h.store.Get(c)
	return h.inertiaService.Render(c, "app/AttendanceReviews", fiber.Map{
		"user": sessionUser(sess),
		"adjustments": adjustments,
		"overtime": overtime,
	})
}
