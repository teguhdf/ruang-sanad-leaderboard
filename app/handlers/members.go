package handlers

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/maulanashalihin/laju-go/app/models"
	"github.com/maulanashalihin/laju-go/app/services"
)

func (h *AppHandler) Members(c *fiber.Ctx) error {
	ownerUserID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	members, err := h.memberService.ListMembers(ownerUserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load members"})
	}

	sess, _ := h.store.Get(c)
	user := sessionUser(sess)

	return h.inertiaService.Render(c, "app/Members", fiber.Map{
		"user":    user,
		"members": members,
	})
}

func (h *AppHandler) CreateMember(c *fiber.Ctx) error {
	ownerUserID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var input models.MemberInput
	if err := c.BodyParser(&input); err != nil {
		h.store.Flash(c, "error", "Data anggota tidak dapat dibaca.")
		return h.inertiaService.Redirect(c, "/app/members")
	}

	if _, err := h.memberService.CreateMember(ownerUserID, input); err != nil {
		setMemberFlashError(h, c, err)
		return h.inertiaService.Redirect(c, "/app/members")
	}

	h.store.Flash(c, "success", "Anggota berhasil ditambahkan.")
	return h.inertiaService.Redirect(c, "/app/members")
}

func (h *AppHandler) UpdateMember(c *fiber.Ctx) error {
	ownerUserID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	memberID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || memberID <= 0 {
		h.store.Flash(c, "error", "Anggota tidak valid.")
		return h.inertiaService.Redirect(c, "/app/members")
	}

	var input models.MemberInput
	if err := c.BodyParser(&input); err != nil {
		h.store.Flash(c, "error", "Data anggota tidak dapat dibaca.")
		return h.inertiaService.Redirect(c, "/app/members")
	}

	if err := h.memberService.UpdateMember(ownerUserID, memberID, input); err != nil {
		setMemberFlashError(h, c, err)
		return h.inertiaService.Redirect(c, "/app/members")
	}

	h.store.Flash(c, "success", "Data anggota berhasil diperbarui.")
	return h.inertiaService.Redirect(c, "/app/members")
}

func authenticatedUserID(c *fiber.Ctx) (int64, bool) {
	switch value := c.Locals("user_id").(type) {
	case int64:
		return value, value > 0
	case int:
		return int64(value), value > 0
	default:
		return 0, false
	}
}

func setMemberFlashError(h *AppHandler, c *fiber.Ctx, err error) {
	var validationErr *services.MemberValidationError

	switch {
	case errors.As(err, &validationErr):
		h.store.Flash(c, "error", validationErr.Message)
	case errors.Is(err, services.ErrFingerprintIDExists):
		h.store.Flash(c, "error", "ID fingerprint sudah digunakan anggota lain.")
	case errors.Is(err, services.ErrMemberNotFound):
		h.store.Flash(c, "error", "Anggota tidak ditemukan.")
	default:
		h.store.Flash(c, "error", "Data anggota gagal disimpan.")
	}
}
