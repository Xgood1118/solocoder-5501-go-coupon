package service

import (
	"errors"

	"coupon-service/internal/constants"
	"coupon-service/internal/model"
	"coupon-service/internal/pkg/logger"
	"coupon-service/internal/repository"
)

type UseService struct {
	repo *repository.Repository
}

func NewUseService(repo *repository.Repository) *UseService {
	return &UseService{repo: repo}
}

func (s *UseService) Use(recordID, userID int64, orderAmount float64, orderID string) (*model.CouponRecord, *constants.ErrorCode, error) {
	rec, err := s.repo.GetRecord(recordID)
	if err != nil {
		logger.Error("get record failed: %v", err)
		return nil, nil, err
	}
	if rec == nil {
		code := constants.CodeRecordNotFound
		return nil, &code, errors.New("record not found")
	}

	if rec.UserID != userID {
		code := constants.CodeNotOwner
		return nil, &code, errors.New("not the owner")
	}

	now := logger.Now()
	if now.After(rec.ValidTo) {
		code := constants.CodeCouponExpired
		return nil, &code, errors.New("coupon expired")
	}
	if now.Before(rec.ValidFrom) {
		code := constants.CodeCouponExpired
		return nil, &code, errors.New("coupon not active yet")
	}

	if rec.Status != constants.CouponStatusUnused {
		code := constants.CodeCouponExpired
		if rec.Status == constants.CouponStatusUsed {
			code = constants.CodeCouponExpired
		}
		return nil, &code, errors.New("coupon not available, status=" + rec.Status.String())
	}

	if orderAmount < rec.Threshold {
		code := constants.CodeOrderNotMeet
		return nil, &code, errors.New("order amount below threshold")
	}

	if err := s.repo.UseRecord(recordID, orderID); err != nil {
		logger.Error("use record failed: %v", err)
		code := constants.CodeConflictStock
		return nil, &code, err
	}

	updatedRec, err := s.repo.GetRecord(recordID)
	if err != nil {
		logger.Error("get updated record failed: %v", err)
		return nil, nil, err
	}

	logger.Info("coupon used: record_id=%d, user_id=%d, order_id=%s, value=%.2f",
		recordID, userID, orderID, rec.Value)
	return updatedRec, nil, nil
}
