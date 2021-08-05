package response

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

type BaseResponse struct {
	errors.Code
	Data interface{} `json:"data,omitempty"`
}

func getResponse(err error, resp interface{}) *BaseResponse {
	baseResult := BaseResponse{errors.GetCode(errors.OK), resp}
	if err != nil {
		switch v := err.(type) {
		case errors.Error:
			//log.Printf("%+v\n", v.Err)
			baseResult.Code = v.Code
		default:
			log.Printf("%+v\n", err)
			baseResult.Code = errors.GetCode(errors.InternalServerErr)
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
