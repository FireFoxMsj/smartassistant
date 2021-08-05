package location

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/internal/api/utils/response"
	"github.com/zhiting-tech/smartassistant/internal/entity"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// 房间默认勾选列表中默认已勾选的房间
var LocationDefaultChosen = [...]string{
	"客厅",
	"餐厅",
	"主人房",
	"书房",
	"卫生间",
}

// 房间默认勾选列表中默认未勾选的房间
var LocationDefaultNotChosen = [...]string{
	"行政部",
	"市场部",
	"研发部",
	"总裁办",
}

// Location 房间信息
type Location struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Sort int    `json:"sort"`
}

// locationListResp 房间列表接口返回数据
type locationListResp struct {
	Locations []Location `json:"locations"`
}

// locationDefaulResp 房间默认勾选列表接口返回数据
type locationsDefaultResp struct {
	LocationsDefault []locationDefault `json:"locations"`
}

// locationDefault 模板列表中的房间信息
type locationDefault struct {
	Name   string `json:"name"`
	Chosen bool   `json:"chosen"`
}

// ListLocation 用于处理房间列表接口的请求
func ListLocation(c *gin.Context) {
	var (
		err       error
		resp      locationListResp
		locations []entity.Location
	)
	defer func() {
		if resp.Locations == nil {
			resp.Locations = make([]Location, 0)
		}
		response.HandleResponse(c, err, resp)
	}()

	if locations, err = entity.GetLocations(); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	resp.Locations = WrapLocations(locations)
	return

}

// ListDefaultLocation 用于处理房间默认勾选列表接口的请求
func ListDefaultLocation(c *gin.Context) {
	var (
		locationsDefault []locationDefault
		locationsResp    locationsDefaultResp
		err              error
	)
	defer func() {
		response.HandleResponse(c, err, locationsResp)
	}()
	locationsDefault = []locationDefault{}
	for _, name := range LocationDefaultChosen {
		locationsDefault = append(locationsDefault, locationDefault{
			Name:   name,
			Chosen: true,
		})
	}
	for _, name := range LocationDefaultNotChosen {
		locationsDefault = append(locationsDefault, locationDefault{
			Name:   name,
			Chosen: false,
		})
	}
	locationsResp.LocationsDefault = locationsDefault
	return

}

func WrapLocations(locations []entity.Location) (result []Location) {
	for _, a := range locations {
		location := Location{
			ID:   a.ID,
			Name: a.Name,
			Sort: a.Sort,
		}
		result = append(result, location)

	}

	return result
}
