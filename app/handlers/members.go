package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maulanashalihin/laju-go/app/models"
	"github.com/maulanashalihin/laju-go/app/services"
)

const maxMemberPhotoBytes = 3 * 1024 * 1024

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

func (h *AppHandler) UploadMemberPhoto(c *fiber.Ctx) error {
	ownerUserID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	memberID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || memberID <= 0 {
		h.store.Flash(c, "error", "Anggota tidak valid.")
		return h.inertiaService.Redirect(c, "/app/members")
	}

	fileHeader, err := c.FormFile("file")
	if err != nil || fileHeader == nil {
		h.store.Flash(c, "error", "Pilih foto anggota terlebih dahulu.")
		return h.inertiaService.Redirect(c, "/app/members")
	}
	if fileHeader.Size <= 0 || fileHeader.Size > maxMemberPhotoBytes {
		h.store.Flash(c, "error", "Ukuran foto maksimal 3 MB.")
		return h.inertiaService.Redirect(c, "/app/members")
	}

	file, err := fileHeader.Open()
	if err != nil {
		h.store.Flash(c, "error", "Foto tidak dapat dibuka.")
		return h.inertiaService.Redirect(c, "/app/members")
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxMemberPhotoBytes+1))
	if err != nil || len(data) == 0 || len(data) > maxMemberPhotoBytes {
		h.store.Flash(c, "error", "Foto tidak dapat dibaca atau terlalu besar.")
		return h.inertiaService.Redirect(c, "/app/members")
	}

	contentType := http.DetectContentType(data)
	extensions := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	}
	ext, allowed := extensions[contentType]
	if !allowed {
		h.store.Flash(c, "error", "Format foto harus JPG, PNG, atau WEBP.")
		return h.inertiaService.Redirect(c, "/app/members")
	}

	storageDir := filepath.Join("storage", "member-photos")
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		h.store.Flash(c, "error", "Penyimpanan foto belum tersedia.")
		return h.inertiaService.Redirect(c, "/app/members")
	}

	filename := fmt.Sprintf("member_%d_%d%s", memberID, time.Now().UnixNano(), ext)
	storagePath := filepath.Join(storageDir, filename)
	if err := os.WriteFile(storagePath, data, 0644); err != nil {
		h.store.Flash(c, "error", "Foto gagal disimpan.")
		return h.inertiaService.Redirect(c, "/app/members")
	}

	photoURL := "/storage/member-photos/" + filename
	if err := h.memberService.UpdateMemberPhoto(ownerUserID, memberID, photoURL); err != nil {
		_ = os.Remove(storagePath)
		setMemberFlashError(h, c, err)
		return h.inertiaService.Redirect(c, "/app/members")
	}

	h.store.Flash(c, "success", "Foto anggota berhasil diperbarui.")
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
