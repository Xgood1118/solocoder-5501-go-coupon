package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"coupon-service/internal/constants"
	"coupon-service/internal/model"
	"coupon-service/internal/pkg/logger"
	"coupon-service/internal/pkg/response"
)

var validate = validator.New()

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("request started: method=%s, path=%s, client=%s",
			c.Request.Method, c.Request.URL.Path, c.ClientIP())
		c.Next()
		logger.Info("request finished: method=%s, path=%s, status=%d",
			c.Request.Method, c.Request.URL.Path, c.Writer.Status())
	}
}

func EnumValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req interface{}
		switch c.Request.URL.Path {
		case "/api/v1/templates", "/api/v1/templates/":
			if c.Request.Method == http.MethodPost {
				req = &model.CreateTemplateRequest{}
			}
		case "/api/v1/users/register", "/api/v1/users/register/":
			req = &model.NewUserRegisterRequest{}
		}
		if req != nil {
			if err := c.ShouldBindJSON(req); err == nil {
				if code := validateEnum(req); code != nil {
					response.Error(c, http.StatusUnprocessableEntity, *code)
					c.Abort()
					return
				}
			}
		}
		c.Next()
	}
}

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered: %v", r)
				response.Error(c, http.StatusInternalServerError, constants.CodeServerError)
				c.Abort()
			}
		}()
		c.Next()
	}
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
	case *model.ClaimRequest:
	case *model.UseRequest:
	}
	return nil
}

func ParseIntParam(c *gin.Context, name string) (int64, bool) {
	val := c.Param(name)
	if val == "" {
		response.Error(c, http.StatusBadRequest, constants.CodeParamInvalid, "missing param: "+name)
		return 0, false
	}
	id, err := strconv.ParseInt(val, 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, http.StatusBadRequest, constants.CodeParamInvalid, "invalid param: "+name)
		return 0, false
	}
	return id, true
}

func ParseQueryInt(c *gin.Context, name string, def int) int {
	val := c.Query(name)
	if val == "" {
		return def
	}
	if v, err := strconv.Atoi(val); err == nil {
		return v
	}
	return def
}

func ParseStatusQuery(c *gin.Context) *constants.CouponStatus {
	val := c.Query("status")
	if val == "" {
		return nil
	}
	v, err := strconv.Atoi(val)
	if err != nil {
		return nil
	}
	status := constants.CouponStatus(v)
	if !status.IsValid() {
		return nil
	}
	return &status
}
