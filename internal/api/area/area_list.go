package area

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/internal/api/utils/response"

	"github.com/zhiting-tech/smartassistant/internal/entity"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// areaListResp 家庭列表接口返回数据
type areaListResp struct {
	Areas []Area `json:"areas"` // 家庭列表
}

// Area 家庭信息
type Area struct {
	Name string `json:"name"` // 家庭名称
	ID   int    `json:"id"`   // 家庭ID
}

// ListArea 用于处理家庭列表接口的请求
func ListArea(c *gin.Context) {
	var (
		err   error
		resp  areaListResp
		areas []entity.Area
	)
	defer func() {
		if resp.Areas == nil {
			resp.Areas = make([]Area, 0)
		}
		response.HandleResponse(c, err, resp)
	}()
	areas, err = entity.GetAreas()
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	resp.Areas = WrapAreas(areas)
	return

}

func WrapAreas(areas []entity.Area) (result []Area) {
	for _, s := range areas {
		area := Area{
			ID:   s.ID,
			Name: s.Name,
		}
		result = append(result, area)

	}

	return result
}
