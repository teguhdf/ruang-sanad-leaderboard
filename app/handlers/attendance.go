package handlers

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/maulanashalihin/laju-go/app/models"
	"github.com/maulanashalihin/laju-go/app/services"
)

func (h *AppHandler) AttendanceFoundation(c *fiber.Ctx) error {
	ownerUserID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	foundation, err := h.attendanceService.GetFoundation(ownerUserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load attendance foundation"})
	}

	sess, _ := h.store.Get(c)
	user := sessionUser(sess)

	return h.inertiaService.Render(c, "app/Attendance", fiber.Map{
		"user":       user,
		"foundation": foundation,
	})
}

func (h *AppHandler) UpdateAttendanceProfile(c *fiber.Ctx) error {
	ownerUserID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	memberID, err := strconv.ParseInt(c.Params("memberID"), 10, 64)
	if err != nil || memberID <= 0 {
		h.store.Flash(c, "error", "Anggota tidak valid.")
		return h.inertiaService.Redirect(c, "/app/attendance")
	}

	var input models.MemberAttendanceProfileInput
	if err := c.BodyParser(&input); err != nil {
		h.store.Flash(c, "error", "Pengaturan absensi tidak dapat dibaca.")
		return h.inertiaService.Redirect(c, "/app/attendance")
	}

	if err := h.attendanceService.UpdateProfile(ownerUserID, memberID, input); err != nil {
		var validationErr *services.AttendanceValidationError
		switch {
		case errors.As(err, &validationErr):
			h.store.Flash(c, "error", validationErr.Message)
		case errors.Is(err, services.ErrAttendanceProfileNotFound):
			h.store.Flash(c, "error", "Profil absensi anggota tidak ditemukan.")
		default:
			h.store.Flash(c, "error", "Pengaturan absensi gagal disimpan.")
		}
		return h.inertiaService.Redirect(c, "/app/attendance")
	}

	h.store.Flash(c, "success", "Pengaturan absensi anggota berhasil diperbarui.")
	return h.inertiaService.Redirect(c, "/app/attendance")
}
