package handlers

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maulanashalihin/laju-go/app/services"
)

const (
	maxAmaliyahWorkspaceBytes = 10 * 1024 * 1024
	maxAmaliyahWorkspaceFiles = 40
)

func (h *AppHandler) AmaliyahLeaderboard(c *fiber.Ctx) error {
	ownerUserID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	periodMonth := strings.TrimSpace(c.Query("month"))
	if !isValidPeriodMonth(periodMonth) {
		periodMonth = currentJakartaPeriod()
	}

	members, err := h.memberService.ListMembers(ownerUserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load members"})
	}
	leaderboard, err := h.memberService.GetAmaliyahLeaderboard(ownerUserID, periodMonth)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load amaliyah leaderboard"})
	}

	sess, _ := h.store.Get(c)
	return h.inertiaService.Render(c, "app/AmaliyahLeaderboard", fiber.Map{
		"user":         sessionUser(sess),
		"members":      members,
		"period_month": periodMonth,
		"leaderboard":  leaderboard,
	})
}

func (h *AppHandler) ImportAmaliyahWorkspaces(c *fiber.Ctx) error {
	ownerUserID, ok := authenticatedUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	form, err := c.MultipartForm()
	if err != nil {
		h.store.Flash(c, "error", "Upload tidak dapat dibaca.")
		return h.inertiaService.Redirect(c, "/app/amaliyah")
	}

	periodMonth := firstFormValue(form.Value["period_month"])
	if !isValidPeriodMonth(periodMonth) {
		h.store.Flash(c, "error", "Pilih bulan leaderboard yang valid.")
		return h.inertiaService.Redirect(c, "/app/amaliyah")
	}
	redirectURL := "/app/amaliyah?month=" + periodMonth

	files := form.File["files"]
	memberIDs := form.Value["member_ids"]
	if len(files) == 0 {
		h.store.Flash(c, "error", "Pilih minimal satu workspace .xlsx.")
		return h.inertiaService.Redirect(c, redirectURL)
	}
	if len(files) > maxAmaliyahWorkspaceFiles {
		h.store.Flash(c, "error", fmt.Sprintf("Maksimal %d workspace per upload.", maxAmaliyahWorkspaceFiles))
		return h.inertiaService.Redirect(c, redirectURL)
	}
	if len(memberIDs) != len(files) {
		h.store.Flash(c, "error", "Setiap workspace harus dipasangkan ke satu anggota.")
		return h.inertiaService.Redirect(c, redirectURL)
	}

	imported := 0
	for i, fileHeader := range files {
		memberID, err := strconv.ParseInt(strings.TrimSpace(memberIDs[i]), 10, 64)
		if err != nil || memberID <= 0 {
			h.store.Flash(c, "error", fmt.Sprintf("Anggota untuk file %s belum dipilih.", fileHeader.Filename))
			return h.inertiaService.Redirect(c, redirectURL)
		}
		if !strings.EqualFold(filepath.Ext(fileHeader.Filename), ".xlsx") {
			h.store.Flash(c, "error", fmt.Sprintf("%s bukan file .xlsx.", fileHeader.Filename))
			return h.inertiaService.Redirect(c, redirectURL)
		}
		if fileHeader.Size <= 0 || fileHeader.Size > maxAmaliyahWorkspaceBytes {
			h.store.Flash(c, "error", fmt.Sprintf("%s terlalu besar. Maksimal 10 MB.", fileHeader.Filename))
			return h.inertiaService.Redirect(c, redirectURL)
		}

		file, err := fileHeader.Open()
		if err != nil {
			h.store.Flash(c, "error", fmt.Sprintf("%s tidak dapat dibuka.", fileHeader.Filename))
			return h.inertiaService.Redirect(c, redirectURL)
		}
		data, readErr := io.ReadAll(io.LimitReader(file, maxAmaliyahWorkspaceBytes+1))
		_ = file.Close()
		if readErr != nil || len(data) == 0 || len(data) > maxAmaliyahWorkspaceBytes {
			h.store.Flash(c, "error", fmt.Sprintf("%s tidak dapat dibaca.", fileHeader.Filename))
			return h.inertiaService.Redirect(c, redirectURL)
		}
		if len(data) < 2 || data[0] != 'P' || data[1] != 'K' {
			h.store.Flash(c, "error", fmt.Sprintf("%s bukan workbook .xlsx yang valid.", fileHeader.Filename))
			return h.inertiaService.Redirect(c, redirectURL)
		}

		if err := h.memberService.ImportAmaliyahWorkspace(ownerUserID, memberID, periodMonth, filepath.Base(fileHeader.Filename), data); err != nil {
			var validationErr *services.AmaliyahValidationError
			switch {
			case errors.As(err, &validationErr):
				h.store.Flash(c, "error", fmt.Sprintf("%s: %s", fileHeader.Filename, validationErr.Message))
			case errors.Is(err, services.ErrAmaliyahInvalidPeriod):
				h.store.Flash(c, "error", "Periode leaderboard tidak valid.")
			default:
				h.store.Flash(c, "error", fmt.Sprintf("%s gagal diproses.", fileHeader.Filename))
			}
			return h.inertiaService.Redirect(c, redirectURL)
		}
		imported++
	}

	h.store.Flash(c, "success", fmt.Sprintf("%d workspace berhasil dihitung dan leaderboard diperbarui.", imported))
	return h.inertiaService.Redirect(c, redirectURL)
}

func firstFormValue(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return strings.TrimSpace(values[0])
}

func isValidPeriodMonth(value string) bool {
	if len(value) != 7 {
		return false
	}
	_, err := time.Parse("2006-01", value)
	return err == nil
}

func currentJakartaPeriod() string {
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.Now().Format("2006-01")
	}
	return time.Now().In(location).Format("2006-01")
}
