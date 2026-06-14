package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"coupon-service/internal/constants"
	"coupon-service/internal/middleware"
	"coupon-service/internal/model"
	"coupon-service/internal/pkg/response"
	"coupon-service/internal/service"
)

type Handler struct {
	templateService *service.TemplateService
	claimService    *service.ClaimService
	useService      *service.UseService
	newUserService  *service.NewUserService
}

func NewHandler(
	templateService *service.TemplateService,
	claimService *service.ClaimService,
	useService *service.UseService,
	newUserService *service.NewUserService,
) *Handler {
	return &Handler{
		templateService: templateService,
		claimService:    claimService,
		useService:      useService,
		newUserService:  newUserService,
	}
}

func (h *Handler) CreateTemplate(c *gin.Context) {
	var req model.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, constants.CodeParamInvalid, err.Error())
		return
	}
	if !req.Type.IsValid() || !req.ApplicableLevel.IsValid() || !req.Category.IsValid() {
		response.Error(c, http.StatusUnprocessableEntity, constants.CodeInvalidEnum)
		return
	}
	tpl, err := h.templateService.Create(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, constants.CodeServerError, err.Error())
		return
	}
	response.Success(c, tpl)
}

func (h *Handler) GetTemplate(c *gin.Context) {
	id, ok := middleware.ParseIntParam(c, "id")
	if !ok {
		return
	}
	tpl, err := h.templateService.Get(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, constants.CodeServerError, err.Error())
		return
	}
	if tpl == nil {
		response.Error(c, http.StatusNotFound, constants.CodeCouponNotFound)
		return
	}
	response.Success(c, tpl)
}

func (h *Handler) ListTemplates(c *gin.Context) {
	page := middleware.ParseQueryInt(c, "page", 1)
	size := middleware.ParseQueryInt(c, "size", 20)
	list, total, err := h.templateService.List(page, size)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, constants.CodeServerError, err.Error())
		return
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

func (h *Handler) Claim(c *gin.Context) {
	var req model.ClaimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, constants.CodeParamInvalid, err.Error())
		return
	}
	rec, code, err := h.claimService.Claim(req.UserID, req.TemplateID)
	if code != nil {
		httpStatus := http.StatusConflict
		switch *code {
		case constants.CodeCouponNotFound, constants.CodeCouponStockOut, constants.CodeRecordNotFound:
			httpStatus = http.StatusNotFound
		case constants.CodeCouponExpired:
			httpStatus = http.StatusGone
		case constants.CodeInvalidEnum:
			httpStatus = http.StatusUnprocessableEntity
		}
		response.Error(c, httpStatus, *code, err.Error())
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, constants.CodeServerError, err.Error())
		return
	}
	response.Success(c, rec)
}

func (h *Handler) Use(c *gin.Context) {
	var req model.UseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, constants.CodeParamInvalid, err.Error())
		return
	}
	rec, code, err := h.useService.Use(req.RecordID, req.UserID, req.OrderAmount, req.OrderID)
	if code != nil {
		httpStatus := http.StatusBadRequest
		switch *code {
		case constants.CodeRecordNotFound:
			httpStatus = http.StatusNotFound
		case constants.CodeCouponExpired:
			httpStatus = http.StatusGone
		case constants.CodeOrderNotMeet:
			httpStatus = http.StatusPreconditionFailed
		case constants.CodeNotOwner:
			httpStatus = http.StatusForbidden
		case constants.CodeConflictStock, constants.CodeAlreadyClaimed, constants.CodeCooldownNotMet:
			httpStatus = http.StatusConflict
		}
		response.Error(c, httpStatus, *code, err.Error())
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, constants.CodeServerError, err.Error())
		return
	}
	response.Success(c, rec)
}

func (h *Handler) ListUserRecords(c *gin.Context) {
	userID, ok := middleware.ParseIntParam(c, "user_id")
	if !ok {
		return
	}
	status := middleware.ParseStatusQuery(c)
	list, err := h.claimService.ListUserRecords(userID, status)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, constants.CodeServerError, err.Error())
		return
	}
	response.Success(c, list)
}

func (h *Handler) RegisterUser(c *gin.Context) {
	var req model.NewUserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, constants.CodeParamInvalid, err.Error())
		return
	}
	if !req.Level.IsValid() {
		response.Error(c, http.StatusUnprocessableEntity, constants.CodeInvalidEnum)
		return
	}
	rec, err := h.newUserService.Register(req.UserID, req.Level)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, constants.CodeServerError, err.Error())
		return
	}
	response.Success(c, gin.H{
		"gift": rec,
	})
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		templates := api.Group("/templates")
		{
			templates.POST("", h.CreateTemplate)
			templates.GET("/:id", h.GetTemplate)
			templates.GET("", h.ListTemplates)
		}
		coupons := api.Group("/coupons")
		{
			coupons.POST("/claim", h.Claim)
			coupons.POST("/use", h.Use)
		}
		users := api.Group("/users")
		{
			users.POST("/register", h.RegisterUser)
			users.GET("/:user_id/coupons", h.ListUserRecords)
		}
	}
}
