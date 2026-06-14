package constants

type CouponStatus int
type UserLevel int
type CouponCategory int
type CouponType int

const (
	CouponStatusUnused  CouponStatus = 0
	CouponStatusUsed    CouponStatus = 1
	CouponStatusExpired CouponStatus = 2
	CouponStatusFrozen  CouponStatus = 3
)

const (
	UserLevelNormal   UserLevel = 0
	UserLevelSilver   UserLevel = 1
	UserLevelGold     UserLevel = 2
	UserLevelPlatinum UserLevel = 3
)

const (
	CouponCategoryAll       CouponCategory = 0
	CouponCategoryFood      CouponCategory = 1
	CouponCategoryElectronics CouponCategory = 2
	CouponCategoryClothing  CouponCategory = 3
	CouponCategoryHome      CouponCategory = 4
)

const (
	CouponTypeGeneral  CouponType = 0
	CouponTypeNewUser  CouponType = 1
)

const (
	DefaultCooldownSeconds = 3600
	DefaultTickerMinutes   = 1
	DefaultPort            = "8080"
	DefaultDBPath          = "./coupon.db"
	NewUserValidDays       = 7
)

type ErrorCode int

const (
	CodeSuccess            ErrorCode = 0
	CodeParamInvalid       ErrorCode = 40001
	CodeCouponNotFound     ErrorCode = 40401
	CodeCouponStockOut     ErrorCode = 40402
	CodeRecordNotFound     ErrorCode = 40403
	CodeCouponExpired      ErrorCode = 41001
	CodeOrderNotMeet       ErrorCode = 41201
	CodeNotOwner           ErrorCode = 40301
	CodeConflictStock      ErrorCode = 40901
	CodeAlreadyClaimed     ErrorCode = 40902
	CodeCooldownNotMet     ErrorCode = 40903
	CodeInvalidEnum        ErrorCode = 42201
	CodeServerError        ErrorCode = 50001
)

var ErrorMsg = map[ErrorCode]string{
	CodeSuccess:        "success",
	CodeParamInvalid:   "参数错误",
	CodeCouponNotFound: "优惠券模板不存在",
	CodeCouponStockOut: "优惠券已领完",
	CodeRecordNotFound: "优惠券记录不存在",
	CodeCouponExpired:  "优惠券已过期",
	CodeOrderNotMeet:   "订单金额未达使用门槛",
	CodeNotOwner:       "无权使用该优惠券",
	CodeConflictStock:  "库存冲突，请重试",
	CodeAlreadyClaimed: "已领取过该优惠券",
	CodeCooldownNotMet: "领取冷却时间未到",
	CodeInvalidEnum:    "枚举值非法",
	CodeServerError:    "服务器内部错误",
}

func (s CouponStatus) String() string {
	switch s {
	case CouponStatusUnused:
		return "UNUSED"
	case CouponStatusUsed:
		return "USED"
	case CouponStatusExpired:
		return "EXPIRED"
	case CouponStatusFrozen:
		return "FROZEN"
	default:
		return "UNKNOWN"
	}
}

func (s CouponStatus) IsValid() bool {
	return s >= CouponStatusUnused && s <= CouponStatusFrozen
}

func (l UserLevel) IsValid() bool {
	return l >= UserLevelNormal && l <= UserLevelPlatinum
}

func (c CouponCategory) IsValid() bool {
	return c >= CouponCategoryAll && c <= CouponCategoryHome
}

func (t CouponType) IsValid() bool {
	return t == CouponTypeGeneral || t == CouponTypeNewUser
}
