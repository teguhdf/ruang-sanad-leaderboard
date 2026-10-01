package services

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/maulanashalihin/laju-go/app/models"
	"github.com/maulanashalihin/laju-go/app/queries"
)

var (
	ErrMemberNotFound      = errors.New("member not found")
	ErrFingerprintIDExists = errors.New("fingerprint id already exists")
)

type MemberValidationError struct {
	Message string
}

func (e *MemberValidationError) Error() string {
	return e.Message
}

type MemberService struct {
	querier *queries.Querier
}

func NewMemberService(querier *queries.Querier) *MemberService {
	return &MemberService{querier: querier}
}

func (s *MemberService) ListMembers(ownerUserID int64) ([]models.Member, error) {
	return s.querier.ListMembersByOwner(context.Background(), ownerUserID)
}

func (s *MemberService) CreateMember(ownerUserID int64, input models.MemberInput) (*models.Member, error) {
	member, err := normalizeMemberInput(ownerUserID, input)
	if err != nil {
		return nil, err
	}

	created, err := s.querier.CreateMemberRecord(context.Background(), member)
	if err != nil {
		if isFingerprintConflict(err) {
			return nil, ErrFingerprintIDExists
		}
		return nil, err
	}

	return created, nil
}

func (s *MemberService) UpdateMember(ownerUserID, memberID int64, input models.MemberInput) error {
	member, err := normalizeMemberInput(ownerUserID, input)
	if err != nil {
		return err
	}

	if err := s.querier.UpdateMemberRecord(context.Background(), ownerUserID, memberID, member); err != nil {
		if errors.Is(err, queries.ErrMemberNotFound) {
			return ErrMemberNotFound
		}
		if isFingerprintConflict(err) {
			return ErrFingerprintIDExists
		}
		return err
	}

	return nil
}

func (s *MemberService) UpdateMemberPhoto(ownerUserID, memberID int64, photoURL string) error {
	if strings.TrimSpace(photoURL) == "" {
		return &MemberValidationError{Message: "Foto anggota tidak valid."}
	}
	if err := s.querier.UpdateMemberPhotoURL(context.Background(), ownerUserID, memberID, photoURL); err != nil {
		if errors.Is(err, queries.ErrMemberNotFound) {
			return ErrMemberNotFound
		}
		return err
	}
	return nil
}

func normalizeMemberInput(ownerUserID int64, input models.MemberInput) (models.Member, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Division = strings.TrimSpace(input.Division)
	input.StartTime = strings.TrimSpace(input.StartTime)
	input.FingerprintID = strings.TrimSpace(input.FingerprintID)
	input.GoogleSheetURL = strings.TrimSpace(input.GoogleSheetURL)

	if input.Name == "" {
		return models.Member{}, &MemberValidationError{Message: "Nama anggota wajib diisi."}
	}
	if input.Division == "" {
		return models.Member{}, &MemberValidationError{Message: "Divisi wajib diisi."}
	}
	if len(input.WorkDays) == 0 {
		return models.Member{}, &MemberValidationError{Message: "Pilih minimal satu hari kerja."}
	}
	if input.StartTime == "" {
		return models.Member{}, &MemberValidationError{Message: "Jam masuk wajib diisi."}
	}
	if _, err := time.Parse("15:04", input.StartTime); err != nil {
		return models.Member{}, &MemberValidationError{Message: "Format jam masuk tidak valid."}
	}
	if input.FingerprintID == "" {
		return models.Member{}, &MemberValidationError{Message: "ID fingerprint wajib diisi."}
	}
	if input.GoogleSheetURL == "" {
		return models.Member{}, &MemberValidationError{Message: "Link Google Sheets wajib diisi."}
	}

	sheetURL, err := url.Parse(input.GoogleSheetURL)
	if err != nil || sheetURL.Scheme != "https" || !strings.EqualFold(sheetURL.Host, "docs.google.com") || !strings.HasPrefix(sheetURL.Path, "/spreadsheets/") {
		return models.Member{}, &MemberValidationError{Message: "Link Google Sheets tidak valid."}
	}

	allowedDays := map[string]bool{
		"mon": true,
		"tue": true,
		"wed": true,
		"thu": true,
		"fri": true,
		"sat": true,
		"sun": true,
	}
	dayOrder := []string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"}
	selected := make(map[string]bool, len(input.WorkDays))
	for _, day := range input.WorkDays {
		day = strings.ToLower(strings.TrimSpace(day))
		if !allowedDays[day] {
			return models.Member{}, &MemberValidationError{Message: "Hari kerja tidak valid."}
		}
		selected[day] = true
	}
	workDays := make([]string, 0, len(selected))
	for _, day := range dayOrder {
		if selected[day] {
			workDays = append(workDays, day)
		}
	}

	return models.Member{
		OwnerUserID:    ownerUserID,
		Name:           input.Name,
		Division:       input.Division,
		WorkDays:       workDays,
		StartTime:      input.StartTime,
		FingerprintID:  input.FingerprintID,
		GoogleSheetURL: input.GoogleSheetURL,
	}, nil
}

func isFingerprintConflict(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed: members.owner_user_id, members.fingerprint_id")
}
