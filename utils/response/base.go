package response

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	error2 "gitlab.yctc.tech/root/smartassistent.git/utils/errors"
)

type BaseResponse struct {
	error2.Code
	Data interface{} `json:"data,omitempty"`
}

func getResponse(err error, resp interface{}) *BaseResponse {
	baseResult := BaseResponse{error2.OK, resp}
	if err != nil {
		switch v := err.(type) {
		case error2.Error:
			log.Printf("%+v\n", v.Err)
			baseResult.Code = v.Code
		default:
			log.Printf("%+v\n", err)
			baseResult.Code = error2.InternalServerErr
		}
	}
	return &baseResult
}

func HandleResponse(ctx *gin.Context, err error, response interface{}) {
	HandleResponseWithStatus(ctx, http.StatusOK, err, response)
}

func HandleResponseWithStatus(ctx *gin.Context, status int, err error, response interface{}) {
	baseResult := getResponse(err, response)
	ctx.JSON(status, baseResult)
}
