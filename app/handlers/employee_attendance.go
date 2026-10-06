package handlers

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/maulanashalihin/laju-go/app/models"
	"github.com/maulanashalihin/laju-go/app/queries"
	"github.com/maulanashalihin/laju-go/app/services"
	"github.com/maulanashalihin/laju-go/app/session"
)

type EmployeeAttendanceHandler struct {
	service *services.EmployeeAttendanceService
	store *session.Store
	inertiaService *services.InertiaService
}

func NewEmployeeAttendanceHandler(service *services.EmployeeAttendanceService, store *session.Store, inertiaService *services.InertiaService) *EmployeeAttendanceHandler {
	return &EmployeeAttendanceHandler{service: service, store: store, inertiaService: inertiaService}
}

func (h *EmployeeAttendanceHandler) MyAttendance(c *fiber.Ctx) error {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}
	summary, err := h.service.GetMyAttendance(userID)
	if err != nil {
		if errors.Is(err, queries.ErrEmployeeProfileNotFound) {
			sess, _ := h.store.Get(c)
			return h.inertiaService.Render(c, "app/EmployeeIdentityRequired", fiber.Map{
				"user":        sessionUser(sess),
				"group":       "my-attendance",
				"title":       "Absensi Saya",
				"description": "Akun ini belum terhubung ke profil anggota untuk absensi.",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load attendance"})
	}
	sess, _ := h.store.Get(c)
	return h.inertiaService.Render(c, "app/MyAttendance", fiber.Map{
		"user":       sessionUser(sess),
		"attendance": summary,
	})
}

func (h *EmployeeAttendanceHandler) StartWork(c *fiber.Ctx) error {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}
	var input models.StartWorkInput
	if err := c.BodyParser(&input); err != nil {
		h.store.Flash(c, "error", "Lokasi kerja tidak dapat dibaca.")
		return h.inertiaService.Redirect(c, "/app/my-attendance")
	}
	if err := h.service.StartWork(userID, input); err != nil {
		setEmployeeAttendanceFlashError(h, c, err)
		return h.inertiaService.Redirect(c, "/app/my-attendance")
	}
	h.store.Flash(c, "success", "Sesi kerja dimulai menggunakan waktu server.")
	return h.inertiaService.Redirect(c, "/app/my-attendance")
}

func (h *EmployeeAttendanceHandler) StopWork(c *fiber.Ctx) error {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}
	if err := h.service.StopWork(userID); err != nil {
		setEmployeeAttendanceFlashError(h, c, err)
		return h.inertiaService.Redirect(c, "/app/my-attendance")
	}
	h.store.Flash(c, "success", "Sesi kerja selesai dan durasi tersimpan.")
	return h.inertiaService.Redirect(c, "/app/my-attendance")
}

func (h *EmployeeAttendanceHandler) RequestAdjustment(c *fiber.Ctx) error {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}
	var input models.AttendanceAdjustmentInput
	if err := c.BodyParser(&input); err != nil {
		h.store.Flash(c, "error", "Data koreksi tidak dapat dibaca.")
		return h.inertiaService.Redirect(c, "/app/my-attendance")
	}
	if err := h.service.RequestAdjustment(userID, input); err != nil {
		setEmployeeAttendanceFlashError(h, c, err)
		return h.inertiaService.Redirect(c, "/app/my-attendance")
	}
	h.store.Flash(c, "success", "Permintaan koreksi dikirim ke SPV. Jam belum dihitung sampai disetujui.")
	return h.inertiaService.Redirect(c, "/app/my-attendance")
}

func (h *EmployeeAttendanceHandler) RequestOvertime(c *fiber.Ctx) error {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}
	var input models.OvertimeRequestInput
	if err := c.BodyParser(&input); err != nil {
		h.store.Flash(c, "error", "Data lembur tidak dapat dibaca.")
		return h.inertiaService.Redirect(c, "/app/my-attendance")
	}
	if err := h.service.RequestOvertime(userID, input); err != nil {
		setEmployeeAttendanceFlashError(h, c, err)
		return h.inertiaService.Redirect(c, "/app/my-attendance")
	}
	h.store.Flash(c, "success", "Permintaan lembur dikirim. Lembur hanya sah setelah approval SPV.")
	return h.inertiaService.Redirect(c, "/app/my-attendance")
}

func (h *EmployeeAttendanceHandler) InviteEmployee(c *fiber.Ctx) error {
	ownerUserID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}
	memberID, err := strconv.ParseInt(c.Params("memberID"), 10, 64)
	if err != nil || memberID <= 0 {
		h.store.Flash(c, "error", "Anggota tidak valid.")
		return h.inertiaService.Redirect(c, "/app/attendance")
	}
	var input struct {
		Email string `json:"email" form:"email"`
	}
	if err := c.BodyParser(&input); err != nil {
		h.store.Flash(c, "error", "Email employee tidak dapat dibaca.")
		return h.inertiaService.Redirect(c, "/app/attendance")
	}
	if err := h.service.InviteEmployee(ownerUserID, memberID, input.Email); err != nil {
		setEmployeeAttendanceFlashError(h, c, err)
		return h.inertiaService.Redirect(c, "/app/attendance")
	}
	h.store.Flash(c, "success", "Akun employee terhubung. Link aktivasi/set password telah dikirim.")
	return h.inertiaService.Redirect(c, "/app/attendance")
}

func (h *EmployeeAttendanceHandler) ReviewAdjustment(c *fiber.Ctx) error {
	ownerUserID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}
	requestID, err := strconv.ParseInt(c.Params("requestID"), 10, 64)
	if err != nil || requestID <= 0 {
		h.store.Flash(c, "error", "Permintaan koreksi tidak valid.")
		return h.inertiaService.Redirect(c, "/app/attendance")
	}
	var input models.ReviewRequestInput
	if err := c.BodyParser(&input); err != nil {
		h.store.Flash(c, "error", "Review koreksi tidak dapat dibaca.")
		return h.inertiaService.Redirect(c, "/app/attendance")
	}
	if err := h.service.ReviewAdjustment(ownerUserID, requestID, ownerUserID, input); err != nil {
		setEmployeeAttendanceFlashError(h, c, err)
		return h.inertiaService.Redirect(c, "/app/attendance")
	}
	h.store.Flash(c, "success", "Review koreksi absensi tersimpan.")
	return h.inertiaService.Redirect(c, "/app/attendance")
}

func (h *EmployeeAttendanceHandler) ReviewOvertime(c *fiber.Ctx) error {
	ownerUserID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}
	requestID, err := strconv.ParseInt(c.Params("requestID"), 10, 64)
	if err != nil || requestID <= 0 {
		h.store.Flash(c, "error", "Permintaan lembur tidak valid.")
		return h.inertiaService.Redirect(c, "/app/attendance")
	}
	var input models.ReviewRequestInput
	if err := c.BodyParser(&input); err != nil {
		h.store.Flash(c, "error", "Review lembur tidak dapat dibaca.")
		return h.inertiaService.Redirect(c, "/app/attendance")
	}
	if err := h.service.ReviewOvertime(ownerUserID, requestID, ownerUserID, input); err != nil {
		setEmployeeAttendanceFlashError(h, c, err)
		return h.inertiaService.Redirect(c, "/app/attendance")
	}
	h.store.Flash(c, "success", "Review lembur tersimpan.")
	return h.inertiaService.Redirect(c, "/app/attendance")
}

func setEmployeeAttendanceFlashError(h *EmployeeAttendanceHandler, c *fiber.Ctx, err error) {
	var validationErr *services.EmployeeAttendanceValidationError
	switch {
	case errors.As(err, &validationErr):
		h.store.Flash(c, "error", validationErr.Message)
	case errors.Is(err, services.ErrEmployeeAlreadyLinked):
		h.store.Flash(c, "error", "Anggota sudah terhubung dengan akun employee lain.")
	case errors.Is(err, services.ErrEmployeeInviteConflict):
		h.store.Flash(c, "error", "Email sudah digunakan akun non-employee. Gunakan email lain atau audit akun tersebut.")
	case errors.Is(err, services.ErrAttendanceSessionActive):
		h.store.Flash(c, "error", "Masih ada sesi kerja aktif. Selesaikan sesi tersebut lebih dulu.")
	case errors.Is(err, services.ErrNoActiveAttendanceSession):
		h.store.Flash(c, "error", "Tidak ada sesi aktif untuk diselesaikan.")
	case errors.Is(err, queries.ErrAdjustmentRequestNotFound):
		h.store.Flash(c, "error", "Permintaan koreksi tidak ditemukan atau sudah direview.")
	case errors.Is(err, queries.ErrOvertimeRequestNotFound):
		h.store.Flash(c, "error", "Permintaan lembur tidak ditemukan atau sudah direview.")
	default:
		h.store.Flash(c, "error", "Operasi absensi gagal. Coba lagi atau cek log server.")
	}
}
