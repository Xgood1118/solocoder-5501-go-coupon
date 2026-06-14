package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"coupon-service/internal/constants"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: int(constants.CodeSuccess),
		Msg:  constants.ErrorMsg[constants.CodeSuccess],
		Data: data,
	})
}

func Error(c *gin.Context, httpStatus int, code constants.ErrorCode, msg ...string) {
	message := constants.ErrorMsg[code]
	if len(msg) > 0 && msg[0] != "" {
		message = msg[0]
	}
	c.JSON(httpStatus, Response{
		Code: int(code),
		Msg:  message,
	})
}

func ErrorWithData(c *gin.Context, httpStatus int, code constants.ErrorCode, data interface{}, msg ...string) {
	message := constants.ErrorMsg[code]
	if len(msg) > 0 && msg[0] != "" {
		message = msg[0]
	}
	c.JSON(httpStatus, Response{
		Code: int(code),
		Msg:  message,
		Data: data,
	})
}
