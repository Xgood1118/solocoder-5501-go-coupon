package service

import (
	"coupon-service/internal/constants"
	"coupon-service/internal/model"
	"coupon-service/internal/pkg/logger"
	"coupon-service/internal/repository"
)

type NewUserService struct {
	repo         *repository.Repository
	claimService *ClaimService
}

func NewNewUserService(repo *repository.Repository, claimService *ClaimService) *NewUserService {
	return &NewUserService{repo: repo, claimService: claimService}
}

func (s *NewUserService) Register(userID int64, level constants.UserLevel) (*model.CouponRecord, error) {
	if !level.IsValid() {
		return nil, nil
	}

	_, isNew, err := s.repo.GetOrCreateUser(userID, level)
	if err != nil {
		logger.Error("get or create user failed: %v", err)
		return nil, err
	}
	if !isNew {
		logger.Info("user already exists, skip new user gift, user_id=%d", userID)
		return nil, nil
	}

	sent, err := s.repo.IsNewUserGiftSent(userID)
	if err != nil {
		logger.Error("check new user gift sent failed: %v", err)
		return nil, err
	}
	if sent {
		logger.Info("new user gift already sent, user_id=%d", userID)
		return nil, nil
	}

	tpl, err := s.repo.GetNewUserTemplate()
	if err != nil {
		logger.Error("get new user template failed: %v", err)
		return nil, err
	}
	if tpl == nil {
		logger.Info("no new user template available, user_id=%d", userID)
		return nil, nil
	}

	tx, err := s.repo.DB().Begin()
	if err != nil {
		logger.Error("begin tx for new user failed: %v", err)
		return nil, err
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	rec, err := s.claimService.ClaimInternalTx(tx, userID, tpl, constants.NewUserValidDays)
	if err != nil {
		_ = tx.Rollback()
		logger.Error("claim new user gift failed: %v", err)
		return nil, err
	}

	if err := s.repo.MarkNewUserGiftSentTx(tx, userID); err != nil {
		_ = tx.Rollback()
		logger.Error("mark new user gift sent failed: %v", err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		logger.Error("commit new user tx failed: %v", err)
		return nil, err
	}

	logger.Info("new user gift sent: user_id=%d, record_id=%d, value=%.2f",
		userID, rec.ID, rec.Value)
	return rec, nil
}
