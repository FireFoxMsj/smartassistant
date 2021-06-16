package handlers

import (
	"github.com/gin-gonic/gin"

	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
)

type areaListResp struct {
	Areas []Area `json:"areas"`
}

type Area struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

func ListArea(c *gin.Context) {
	var (
		err   error
		resp  areaListResp
		areas []orm.Area
	)
	defer func() {
		if resp.Areas == nil {
			resp.Areas = make([]Area, 0)
		}
		response.HandleResponse(c, err, resp)
	}()
	areas, err = orm.GetAreas()
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	resp.Areas = WrapAreas(areas)
	return

}

func WrapAreas(areas []orm.Area) (result []Area) {
	for _, s := range areas {
		area := Area{
			ID:   s.ID,
			Name: s.Name,
		}
		result = append(result, area)

	}

	return result
}
