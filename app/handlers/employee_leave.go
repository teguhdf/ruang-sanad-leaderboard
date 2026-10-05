package handlers

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maulanashalihin/laju-go/app/models"
	"github.com/maulanashalihin/laju-go/app/queries"
	"github.com/maulanashalihin/laju-go/app/services"
)

func (h *EmployeeAttendanceHandler) MyLeave(c *fiber.Ctx) error {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}
	year := time.Now().In(time.FixedZone("WIB", 7*60*60)).Year()
	if value := c.Query("year"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			year = parsed
		}
	}
	balance, requests, err := h.service.GetEmployeeLeave(userID, year)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load leave"})
	}
	sess, _ := h.store.Get(c)
	return h.inertiaService.Render(c, "app/MyLeave", fiber.Map{
		"user":     sessionUser(sess),
		"balance":  balance,
		"requests": requests,
		"year":     year,
	})
}

func (h *EmployeeAttendanceHandler) RequestLeave(c *fiber.Ctx) error {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}
	var input models.LeaveRequestInput
	if err := c.BodyParser(&input); err != nil {
		h.store.Flash(c, "error", "Data pengajuan cuti tidak dapat dibaca.")
		return h.inertiaService.Redirect(c, "/app/my-leave")
	}
	if err := h.service.RequestLeave(userID, input); err != nil {
		setLeaveFlashError(h, c, err)
		return h.inertiaService.Redirect(c, "/app/my-leave")
	}
	h.store.Flash(c, "success", "Pengajuan cuti dikirim ke SPV. Saldo baru berkurang setelah disetujui.")
	return h.inertiaService.Redirect(c, "/app/my-leave")
}

func (h *EmployeeAttendanceHandler) CancelLeave(c *fiber.Ctx) error {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}
	requestID, err := strconv.ParseInt(c.Params("requestID"), 10, 64)
	if err != nil || requestID <= 0 {
		h.store.Flash(c, "error", "Pengajuan cuti tidak valid.")
		return h.inertiaService.Redirect(c, "/app/my-leave")
	}
	if err := h.service.CancelLeave(userID, requestID); err != nil {
		setLeaveFlashError(h, c, err)
		return h.inertiaService.Redirect(c, "/app/my-leave")
	}
	h.store.Flash(c, "success", "Pengajuan cuti dibatalkan.")
	return h.inertiaService.Redirect(c, "/app/my-leave")
}

func (h *EmployeeAttendanceHandler) LeaveAdministration(c *fiber.Ctx) error {
	ownerUserID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}
	year := time.Now().In(time.FixedZone("WIB", 7*60*60)).Year()
	if value := c.Query("year"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			year = parsed
		}
	}
	balances, pending, err := h.service.ListLeaveAdministration(ownerUserID, year)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load leave administration"})
	}
	sess, _ := h.store.Get(c)
	return h.inertiaService.Render(c, "app/LeaveAdministration", fiber.Map{
		"user":     sessionUser(sess),
		"balances": balances,
		"pending":  pending,
		"year":     year,
	})
}

func (h *EmployeeAttendanceHandler) SetLeaveEntitlement(c *fiber.Ctx) error {
	ownerUserID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}
	memberID, err := strconv.ParseInt(c.Params("memberID"), 10, 64)
	if err != nil || memberID <= 0 {
		h.store.Flash(c, "error", "Anggota tidak valid.")
		return h.inertiaService.Redirect(c, "/app/leave")
	}
	var input models.LeaveEntitlementInput
	if err := c.BodyParser(&input); err != nil {
		h.store.Flash(c, "error", "Jatah cuti tidak dapat dibaca.")
		return h.inertiaService.Redirect(c, "/app/leave")
	}
	if err := h.service.SetLeaveEntitlement(ownerUserID, memberID, input); err != nil {
		setLeaveFlashError(h, c, err)
		return h.inertiaService.Redirect(c, "/app/leave")
	}
	h.store.Flash(c, "success", "Jatah cuti tahunan diperbarui.")
	return h.inertiaService.Redirect(c, "/app/leave?year="+strconv.Itoa(input.Year))
}

func (h *EmployeeAttendanceHandler) ReviewLeave(c *fiber.Ctx) error {
	ownerUserID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}
	requestID, err := strconv.ParseInt(c.Params("requestID"), 10, 64)
	if err != nil || requestID <= 0 {
		h.store.Flash(c, "error", "Pengajuan cuti tidak valid.")
		return h.inertiaService.Redirect(c, "/app/leave")
	}
	var input models.ReviewRequestInput
	if err := c.BodyParser(&input); err != nil {
		h.store.Flash(c, "error", "Review cuti tidak dapat dibaca.")
		return h.inertiaService.Redirect(c, "/app/leave")
	}
	if err := h.service.ReviewLeave(ownerUserID, requestID, ownerUserID, input); err != nil {
		setLeaveFlashError(h, c, err)
		return h.inertiaService.Redirect(c, "/app/leave")
	}
	h.store.Flash(c, "success", "Review cuti tersimpan.")
	return h.inertiaService.Redirect(c, "/app/leave")
}

func setLeaveFlashError(h *EmployeeAttendanceHandler, c *fiber.Ctx, err error) {
	var validationErr *services.EmployeeAttendanceValidationError
	switch {
	case errors.As(err, &validationErr):
		h.store.Flash(c, "error", validationErr.Message)
	case errors.Is(err, queries.ErrLeaveRequestConflict):
		h.store.Flash(c, "error", "Tanggal cuti bertabrakan dengan pengajuan lain yang masih pending atau sudah disetujui.")
	case errors.Is(err, queries.ErrLeaveInsufficientBalance):
		h.store.Flash(c, "error", "Sisa cuti tahunan tidak mencukupi.")
	case errors.Is(err, queries.ErrLeaveRequestNotFound):
		h.store.Flash(c, "error", "Pengajuan cuti tidak ditemukan atau sudah diproses.")
	case errors.Is(err, queries.ErrMemberNotFound):
		h.store.Flash(c, "error", "Anggota tidak ditemukan.")
	default:
		h.store.Flash(c, "error", "Operasi cuti gagal. Coba lagi atau cek log server.")
	}
}
