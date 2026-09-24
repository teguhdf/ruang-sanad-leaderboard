package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/maulanashalihin/laju-go/app/fingerprint"
	"github.com/maulanashalihin/laju-go/app/models"
	"github.com/maulanashalihin/laju-go/app/queries"
)

var (
	ErrFingerprintCorrectionInvalid = errors.New("fingerprint correction invalid")
	ErrFingerprintPeriodInvalid     = errors.New("fingerprint period invalid")
)

type FingerprintImportService struct {
	querier *queries.Querier
}

func NewFingerprintImportService(querier *queries.Querier) *FingerprintImportService {
	return &FingerprintImportService{querier: querier}
}

func (s *FingerprintImportService) ListImports(ownerUserID int64) ([]models.FingerprintImport, error) {
	return s.querier.ListFingerprintImportsByOwner(context.Background(), ownerUserID, 10)
}

func (s *FingerprintImportService) SaveImport(
	ownerUserID int64,
	fileName string,
	data []byte,
	corrections []models.FingerprintCorrectionInput,
) (*models.FingerprintImportResult, error) {
	ctx := context.Background()

	hashBytes := sha256.Sum256(data)
	fileHash := hex.EncodeToString(hashBytes[:])

	if existing, err := s.querier.FindFingerprintImportByHash(ctx, ownerUserID, fileHash); err == nil {
		return &models.FingerprintImportResult{
			Import:    *existing,
			Duplicate: true,
		}, nil
	} else if !errors.Is(err, queries.ErrFingerprintImportNotFound) {
		return nil, err
	}

	preview, err := fingerprint.ParseXLS(data)
	if err != nil {
		return nil, err
	}
	if preview.PeriodStart == "" || preview.PeriodEnd == "" {
		return nil, ErrFingerprintPeriodInvalid
	}

	members, err := s.querier.ListMembersByOwner(ctx, ownerUserID)
	if err != nil {
		return nil, err
	}

	memberByFingerprint := make(map[string]int64, len(members))
	for _, member := range members {
		key := strings.TrimSpace(member.FingerprintID)
		if key != "" {
			memberByFingerprint[key] = member.ID
		}
	}

	type correctionKey struct {
		FingerprintID string
		Date          string
	}
	correctionByKey := make(map[correctionKey]models.FingerprintCorrectionInput, len(corrections))
	for _, correction := range corrections {
		correction.FingerprintID = strings.TrimSpace(correction.FingerprintID)
		correction.Date = strings.TrimSpace(correction.Date)
		correction.CheckIn = strings.TrimSpace(correction.CheckIn)
		correction.CheckOut = strings.TrimSpace(correction.CheckOut)
		correction.Reason = strings.TrimSpace(correction.Reason)

		if correction.FingerprintID == "" || correction.Date == "" {
			return nil, fmt.Errorf("%w: fingerprint ID dan tanggal wajib diisi", ErrFingerprintCorrectionInvalid)
		}
		if _, err := time.Parse("2006-01-02", correction.Date); err != nil {
			return nil, fmt.Errorf("%w: tanggal koreksi tidak valid", ErrFingerprintCorrectionInvalid)
		}
		if !validOptionalClock(correction.CheckIn) || !validOptionalClock(correction.CheckOut) {
			return nil, fmt.Errorf("%w: format jam harus HH:mm", ErrFingerprintCorrectionInvalid)
		}
		if correction.Reason == "" {
			return nil, fmt.Errorf("%w: alasan koreksi wajib diisi", ErrFingerprintCorrectionInvalid)
		}

		key := correctionKey{FingerprintID: correction.FingerprintID, Date: correction.Date}
		if _, exists := correctionByKey[key]; exists {
			return nil, fmt.Errorf("%w: koreksi duplikat pada ID %s tanggal %s", ErrFingerprintCorrectionInvalid, correction.FingerprintID, correction.Date)
		}
		correctionByKey[key] = correction
	}

	save := models.FingerprintImportSave{
		OwnerUserID: ownerUserID,
		Import: models.FingerprintImport{
			FileName:      fileName,
			FileSHA256:    fileHash,
			PeriodStart:   preview.PeriodStart,
			PeriodEnd:     preview.PeriodEnd,
			FilePeople:    preview.Summary.FilePeople,
			TotalSlots:    preview.Summary.TotalSlots,
			TwoTimes:      preview.Summary.TwoTimes,
			OneTime:       preview.Summary.OneTime,
			Empty:         preview.Summary.Empty,
			MultipleTimes: preview.Summary.Multiple,
		},
		Records: make([]models.FingerprintRecord, 0, preview.Summary.TotalSlots),
	}

	foundCorrection := make(map[correctionKey]bool, len(correctionByKey))

	for _, person := range preview.People {
		fingerprintID := strings.TrimSpace(person.FingerprintID)
		var memberID *int64
		if id, matched := memberByFingerprint[fingerprintID]; matched {
			idCopy := id
			memberID = &idCopy
			save.Import.Matched++
		} else {
			save.Import.Unmatched++
		}

		for _, day := range person.Days {
			if day.Date == "" {
				return nil, ErrFingerprintPeriodInvalid
			}

			rawCheckIn := ""
			rawCheckOut := ""
			if len(day.Times) > 0 {
				rawCheckIn = day.Times[0]
			}
			if len(day.Times) > 1 {
				rawCheckOut = day.Times[1]
			}

			record := models.FingerprintRecord{
				MemberID:       memberID,
				FingerprintID:   fingerprintID,
				FileName:        person.FileName,
				Department:      person.Department,
				AttendanceDate:  day.Date,
				Day:             day.Day,
				RawValue:        day.RawValue,
				RawCheckIn:      rawCheckIn,
				RawCheckOut:     rawCheckOut,
				ReadState:       day.ReadState,
			}

			key := correctionKey{FingerprintID: fingerprintID, Date: day.Date}
			if correction, ok := correctionByKey[key]; ok {
				if correction.CheckIn == rawCheckIn && correction.CheckOut == rawCheckOut {
					return nil, fmt.Errorf("%w: koreksi pada ID %s tanggal %s tidak mengubah data", ErrFingerprintCorrectionInvalid, fingerprintID, day.Date)
				}
				record.CorrectedCheckIn = correction.CheckIn
				record.CorrectedCheckOut = correction.CheckOut
				record.CorrectionReason = correction.Reason
				record.IsCorrected = true
				save.Import.CorrectedCount++
				foundCorrection[key] = true
			}

			save.Records = append(save.Records, record)
		}
	}

	for key := range correctionByKey {
		if !foundCorrection[key] {
			return nil, fmt.Errorf("%w: data ID %s tanggal %s tidak ditemukan pada file", ErrFingerprintCorrectionInvalid, key.FingerprintID, key.Date)
		}
	}

	saved, err := s.querier.SaveFingerprintImport(ctx, save)
	if err != nil {
		// A concurrent/repeated save of the same raw file is idempotent.
		if existing, findErr := s.querier.FindFingerprintImportByHash(ctx, ownerUserID, fileHash); findErr == nil {
			return &models.FingerprintImportResult{
				Import:    *existing,
				Duplicate: true,
			}, nil
		}
		return nil, err
	}

	return &models.FingerprintImportResult{
		Import:    *saved,
		Duplicate: false,
	}, nil
}

func validOptionalClock(value string) bool {
	if value == "" {
		return true
	}
	_, err := time.Parse("15:04", value)
	return err == nil
}
