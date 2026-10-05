package handlers

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/maulanashalihin/laju-go/app/models"
	"github.com/maulanashalihin/laju-go/app/services"
)

func (h *AppHandler) AttendanceFoundation(c *fiber.Ctx) error {
	ownerUserID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	foundation, err := h.attendanceService.GetFoundationForWeek(ownerUserID, c.Query("week_start"))
	if err != nil {
		var validationErr *services.AttendanceValidationError
		if errors.As(err, &validationErr) {
			h.store.Flash(c, "error", validationErr.Message)
			foundation, err = h.attendanceService.GetFoundation(ownerUserID)
		}
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load attendance foundation"})
		}
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

func (h *AppHandler) SaveRosterDay(c *fiber.Ctx) error {
	ownerUserID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}
	memberID, err := strconv.ParseInt(c.Params("memberID"), 10, 64)
	if err != nil || memberID <= 0 {
		h.store.Flash(c, "error", "Anggota roster tidak valid.")
		return h.inertiaService.Redirect(c, rosterRedirect(c.Query("week_start")))
	}
	workDate := strings.TrimSpace(c.Params("workDate"))
	var input models.RosterDayInput
	if err := c.BodyParser(&input); err != nil {
		h.store.Flash(c, "error", "Data roster tidak dapat dibaca.")
		return h.inertiaService.Redirect(c, rosterRedirect(c.Query("week_start")))
	}
	if err := h.attendanceService.SaveRosterDay(ownerUserID, memberID, workDate, input); err != nil {
		var validationErr *services.AttendanceValidationError
		if errors.As(err, &validationErr) {
			h.store.Flash(c, "error", validationErr.Message)
		} else {
			h.store.Flash(c, "error", "Roster gagal disimpan.")
		}
		return h.inertiaService.Redirect(c, rosterRedirect(c.Query("week_start")))
	}
	h.store.Flash(c, "success", "Roster harian berhasil disimpan.")
	return h.inertiaService.Redirect(c, rosterRedirect(c.Query("week_start")))
}

func (h *AppHandler) ResetRosterDay(c *fiber.Ctx) error {
	ownerUserID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}
	memberID, err := strconv.ParseInt(c.Params("memberID"), 10, 64)
	if err != nil || memberID <= 0 {
		h.store.Flash(c, "error", "Anggota roster tidak valid.")
		return h.inertiaService.Redirect(c, rosterRedirect(c.Query("week_start")))
	}
	if err := h.attendanceService.ResetRosterDay(ownerUserID, memberID, strings.TrimSpace(c.Params("workDate"))); err != nil {
		var validationErr *services.AttendanceValidationError
		if errors.As(err, &validationErr) {
			h.store.Flash(c, "error", validationErr.Message)
		} else {
			h.store.Flash(c, "error", "Override roster gagal dihapus.")
		}
		return h.inertiaService.Redirect(c, rosterRedirect(c.Query("week_start")))
	}
	h.store.Flash(c, "success", "Roster dikembalikan ke jadwal default.")
	return h.inertiaService.Redirect(c, rosterRedirect(c.Query("week_start")))
}

func rosterRedirect(weekStart string) string {
	weekStart = strings.TrimSpace(weekStart)
	if weekStart == "" {
		return "/app/attendance"
	}
	return fmt.Sprintf("/app/attendance?week_start=%s", url.QueryEscape(weekStart))
}
