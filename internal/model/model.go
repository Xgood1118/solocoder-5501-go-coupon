package model

import (
	"time"

	"coupon-service/internal/constants"
)

type CouponTemplate struct {
	ID             int64                      `json:"id" db:"id"`
	Name           string                     `json:"name" db:"name"`
	Type           constants.CouponType       `json:"type" db:"type"`
	Value          float64                    `json:"value" db:"value"`
	Threshold      float64                    `json:"threshold" db:"threshold"`
	TotalCount     int64                      `json:"total_count" db:"total_count"`
	RemainingCount int64                      `json:"remaining_count" db:"remaining_count"`
	PerUserLimit   int64                      `json:"per_user_limit" db:"per_user_limit"`
	ValidFrom      time.Time                  `json:"valid_from" db:"valid_from"`
	ValidTo        time.Time                  `json:"valid_to" db:"valid_to"`
	ApplicableLevel constants.UserLevel       `json:"applicable_level" db:"applicable_level"`
	Category       constants.CouponCategory   `json:"category" db:"category"`
	CreatedAt      time.Time                  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time                  `json:"updated_at" db:"updated_at"`
}

type CouponRecord struct {
	ID         int64                 `json:"id" db:"id"`
	TemplateID int64                 `json:"template_id" db:"template_id"`
	UserID     int64                 `json:"user_id" db:"user_id"`
	Status     constants.CouponStatus `json:"status" db:"status"`
	Value      float64               `json:"value" db:"value"`
	Threshold  float64               `json:"threshold" db:"threshold"`
	ValidFrom  time.Time             `json:"valid_from" db:"valid_from"`
	ValidTo    time.Time             `json:"valid_to" db:"valid_to"`
	UsedAt     *time.Time            `json:"used_at" db:"used_at"`
	OrderID    *string               `json:"order_id" db:"order_id"`
	CreatedAt  time.Time             `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time             `json:"updated_at" db:"updated_at"`
}

type User struct {
	ID              int64              `json:"id" db:"id"`
	Level           constants.UserLevel `json:"level" db:"level"`
	NewUserGiftSent bool               `json:"new_user_gift_sent" db:"new_user_gift_sent"`
	CreatedAt       time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" db:"updated_at"`
}

type CreateTemplateRequest struct {
	Name           string                     `json:"name" binding:"required"`
	Type           constants.CouponType       `json:"type" binding:"required"`
	Value          float64                    `json:"value" binding:"required,min=0"`
	Threshold      float64                    `json:"threshold" binding:"required,min=0"`
	TotalCount     int64                      `json:"total_count" binding:"required,min=1"`
	PerUserLimit   int64                      `json:"per_user_limit" binding:"required,min=1"`
	ValidFrom      time.Time                  `json:"valid_from" binding:"required"`
	ValidTo        time.Time                  `json:"valid_to" binding:"required,gtfield=ValidFrom"`
	ApplicableLevel constants.UserLevel       `json:"applicable_level" binding:"required"`
	Category       constants.CouponCategory   `json:"category" binding:"required"`
}

type ClaimRequest struct {
	TemplateID int64 `json:"template_id" binding:"required,min=1"`
	UserID     int64 `json:"user_id" binding:"required,min=1"`
}

type UseRequest struct {
	RecordID    int64   `json:"record_id" binding:"required,min=1"`
	UserID      int64   `json:"user_id" binding:"required,min=1"`
	OrderAmount float64 `json:"order_amount" binding:"required,min=0"`
	OrderID     string  `json:"order_id" binding:"required"`
}

type NewUserRegisterRequest struct {
	UserID int64                `json:"user_id" binding:"required,min=1"`
	Level  constants.UserLevel  `json:"level" binding:"required"`
}
