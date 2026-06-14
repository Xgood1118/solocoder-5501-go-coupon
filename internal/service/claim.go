package service

import (
	"database/sql"
	"errors"
	"time"

	"coupon-service/internal/constants"
	"coupon-service/internal/model"
	"coupon-service/internal/pkg/logger"
	"coupon-service/internal/repository"
)

type ClaimService struct {
	repo *repository.Repository
}

func NewClaimService(repo *repository.Repository) *ClaimService {
	return &ClaimService{repo: repo}
}

func (s *ClaimService) Claim(userID, templateID int64) (*model.CouponRecord, *constants.ErrorCode, error) {
	tpl, err := s.repo.GetTemplate(templateID)
	if err != nil {
		logger.Error("get template failed: %v", err)
		return nil, nil, err
	}
	if tpl == nil {
		code := constants.CodeCouponNotFound
		return nil, &code, errors.New("template not found")
	}

	now := logger.Now()
	if now.Before(tpl.ValidFrom) || now.After(tpl.ValidTo) {
		code := constants.CodeCouponExpired
		return nil, &code, errors.New("coupon not in valid period")
	}

	if tpl.RemainingCount <= 0 {
		code := constants.CodeCouponStockOut
		return nil, &code, errors.New("coupon stock out")
	}

	cnt, err := s.repo.CountUserRecords(userID, templateID)
	if err != nil {
		logger.Error("count user records failed: %v", err)
		return nil, nil, err
	}
	if cnt >= tpl.PerUserLimit {
		code := constants.CodeAlreadyClaimed
		return nil, &code, errors.New("already claimed")
	}

	lastClaim, err := s.repo.GetLastClaimTime(userID, templateID)
	if err != nil {
		logger.Error("get last claim time failed: %v", err)
		return nil, nil, err
	}
	if lastClaim != nil {
		cooldown := time.Duration(constants.DefaultCooldownSeconds) * time.Second
		if now.Sub(*lastClaim) < cooldown {
			code := constants.CodeCooldownNotMet
			return nil, &code, errors.New("cooldown not met")
		}
	}

	tx, err := s.repo.DB().Begin()
	if err != nil {
		logger.Error("begin tx failed: %v", err)
		return nil, nil, err
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	tplLocked, err := s.repo.GetTemplateForUpdate(tx, templateID)
	if err != nil {
		_ = tx.Rollback()
		logger.Error("get template for update failed: %v", err)
		return nil, nil, err
	}
	if tplLocked == nil {
		_ = tx.Rollback()
		code := constants.CodeCouponNotFound
		return nil, &code, errors.New("template not found")
	}

	if tplLocked.RemainingCount <= 0 {
		_ = tx.Rollback()
		code := constants.CodeCouponStockOut
		return nil, &code, errors.New("coupon stock out")
	}

	affected, err := s.repo.DecreaseTemplateRemaining(tx, templateID)
	if err != nil {
		_ = tx.Rollback()
		logger.Error("decrease remaining failed: %v", err)
		return nil, nil, err
	}
	if affected == 0 {
		_ = tx.Rollback()
		code := constants.CodeConflictStock
		return nil, &code, errors.New("update affected 0 rows")
	}

	tplCheck, err := s.repo.GetTemplateForUpdate(tx, templateID)
	if err != nil {
		_ = tx.Rollback()
		logger.Error("check template after decrease failed: %v", err)
		return nil, nil, err
	}
	if tplCheck == nil || tplCheck.RemainingCount < 0 {
		_ = tx.Rollback()
		code := constants.CodeConflictStock
		return nil, &code, errors.New("remaining count negative, rollback")
	}

	rec := &model.CouponRecord{
		TemplateID: templateID,
		UserID:     userID,
		Value:      tpl.Value,
		Threshold:  tpl.Threshold,
		ValidFrom:  tpl.ValidFrom,
		ValidTo:    tpl.ValidTo,
	}
	if err := s.repo.CreateRecordTx(tx, rec); err != nil {
		_ = tx.Rollback()
		logger.Error("create record failed: %v", err)
		return nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		logger.Error("commit tx failed: %v", err)
		return nil, nil, err
	}

	logger.Info("coupon claimed: record_id=%d, user_id=%d, template_id=%d", rec.ID, userID, templateID)
	return rec, nil, nil
}

func (s *ClaimService) ClaimInternalTx(tx *sql.Tx, userID int64, tpl *model.CouponTemplate, validDays int) (*model.CouponRecord, error) {
	now := logger.Now()
	validFrom := now
	validTo := now.AddDate(0, 0, validDays)

	tplLocked, err := s.repo.GetTemplateForUpdate(tx, tpl.ID)
	if err != nil {
		return nil, err
	}
	if tplLocked == nil || tplLocked.RemainingCount <= 0 {
		return nil, errors.New("new user coupon stock out")
	}

	affected, err := s.repo.DecreaseTemplateRemaining(tx, tpl.ID)
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, errors.New("decrease remaining failed")
	}

	tplCheck, err := s.repo.GetTemplateForUpdate(tx, tpl.ID)
	if err != nil {
		return nil, err
	}
	if tplCheck == nil || tplCheck.RemainingCount < 0 {
		return nil, errors.New("remaining count negative")
	}

	rec := &model.CouponRecord{
		TemplateID: tpl.ID,
		UserID:     userID,
		Value:      tpl.Value,
		Threshold:  tpl.Threshold,
		ValidFrom:  validFrom,
		ValidTo:    validTo,
	}
	if err := s.repo.CreateRecordTx(tx, rec); err != nil {
		return nil, err
	}

	return rec, nil
}

func (s *ClaimService) ListUserRecords(userID int64, status *constants.CouponStatus) ([]*model.CouponRecord, error) {
	return s.repo.ListUserRecords(userID, status)
}
