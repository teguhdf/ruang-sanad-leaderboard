package handlers

import (
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/maulanashalihin/laju-go/app/fingerprint"
)

const maxFingerprintUploadSize = 2 * 1024 * 1024

func (h *AppHandler) FingerprintImport(c *fiber.Ctx) error {
	sess, _ := h.store.Get(c)
	user := sessionUser(sess)

	return h.inertiaService.Render(c, "app/FingerprintImport", fiber.Map{
		"user": user,
	})
}

func (h *AppHandler) PreviewFingerprint(c *fiber.Ctx) error {
	ownerUserID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Pilih file fingerprint .xls terlebih dahulu."})
	}
	if fileHeader.Size <= 0 || fileHeader.Size > maxFingerprintUploadSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Ukuran file fingerprint tidak valid. Maksimal 2 MB."})
	}
	if strings.ToLower(filepath.Ext(fileHeader.Filename)) != ".xls" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format file harus .xls."})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "File fingerprint tidak dapat dibuka."})
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxFingerprintUploadSize+1))
	if err != nil || len(data) == 0 || len(data) > maxFingerprintUploadSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "File fingerprint tidak dapat dibaca."})
	}

	preview, err := fingerprint.ParseXLS(data)
	if err != nil {
		slog.Warn("fingerprint preview parse failed",
			"filename", filepath.Base(fileHeader.Filename),
			"size_bytes", len(data),
			"error", err,
		)
		if errors.Is(err, fingerprint.ErrInvalidWorkbook) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Isi file .xls tidak sesuai format laporan fingerprint yang didukung."})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Preview fingerprint gagal dibuat."})
	}
	preview.FileName = filepath.Base(fileHeader.Filename)
	slog.Info("fingerprint preview parsed",
		"filename", preview.FileName,
		"people", preview.Summary.FilePeople,
		"slots", preview.Summary.TotalSlots,
		"two_times", preview.Summary.TwoTimes,
		"one_time", preview.Summary.OneTime,
		"empty", preview.Summary.Empty,
	)

	members, err := h.memberService.ListMembers(ownerUserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Data anggota gagal dimuat untuk pencocokan."})
	}

	memberByFingerprint := make(map[string]struct {
		ID   int64
		Name string
	}, len(members))
	for _, member := range members {
		key := strings.TrimSpace(member.FingerprintID)
		if key == "" {
			continue
		}
		memberByFingerprint[key] = struct {
			ID   int64
			Name string
		}{ID: member.ID, Name: member.Name}
	}

	for index := range preview.People {
		person := &preview.People[index]
		member, matched := memberByFingerprint[strings.TrimSpace(person.FingerprintID)]
		if matched {
			person.Matched = true
			person.MemberID = member.ID
			person.MemberName = member.Name
			preview.Summary.Matched++
		} else {
			preview.Summary.Unmatched++
		}
	}

	return c.JSON(preview)
}
