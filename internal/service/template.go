package service

import (
	"errors"

	"coupon-service/internal/constants"
	"coupon-service/internal/model"
	"coupon-service/internal/pkg/logger"
	"coupon-service/internal/repository"
)

type TemplateService struct {
	repo *repository.Repository
}

func NewTemplateService(repo *repository.Repository) *TemplateService {
	return &TemplateService{repo: repo}
}

func (s *TemplateService) Create(req *model.CreateTemplateRequest) (*model.CouponTemplate, error) {
	if !req.Type.IsValid() {
		return nil, errors.New("invalid coupon type")
	}
	if !req.ApplicableLevel.IsValid() {
		return nil, errors.New("invalid user level")
	}
	if !req.Category.IsValid() {
		return nil, errors.New("invalid coupon category")
	}
	if req.Value <= 0 {
		return nil, errors.New("coupon value must be positive")
	}
	if req.Threshold < 0 {
		return nil, errors.New("threshold cannot be negative")
	}
	if req.ValidTo.Before(req.ValidFrom) {
		return nil, errors.New("valid_to must be after valid_from")
	}
	tpl := &model.CouponTemplate{
		Name:            req.Name,
		Type:            req.Type,
		Value:           req.Value,
		Threshold:       req.Threshold,
		TotalCount:      req.TotalCount,
		RemainingCount:  req.TotalCount,
		PerUserLimit:    req.PerUserLimit,
		ValidFrom:       req.ValidFrom,
		ValidTo:         req.ValidTo,
		ApplicableLevel: req.ApplicableLevel,
		Category:        req.Category,
	}
	if err := s.repo.CreateTemplate(tpl); err != nil {
		logger.Error("create template failed: %v", err)
		return nil, err
	}
	logger.Info("template created: id=%d, name=%s", tpl.ID, tpl.Name)
	return tpl, nil
}

func (s *TemplateService) Get(id int64) (*model.CouponTemplate, error) {
	return s.repo.GetTemplate(id)
}

func (s *TemplateService) List(page, size int) ([]*model.CouponTemplate, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return s.repo.ListTemplates(page, size)
}

func validateEnum(req interface{}) *constants.ErrorCode {
	switch r := req.(type) {
	case *model.CreateTemplateRequest:
		if !r.Type.IsValid() || !r.ApplicableLevel.IsValid() || !r.Category.IsValid() {
			code := constants.CodeInvalidEnum
			return &code
		}
	case *model.NewUserRegisterRequest:
		if !r.Level.IsValid() {
			code := constants.CodeInvalidEnum
			return &code
		}
	}
	return nil
}
