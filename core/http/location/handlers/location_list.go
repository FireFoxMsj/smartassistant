package handlers

import (
	"github.com/gin-gonic/gin"
	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
)

var LocationDefaultChosen = [...]string{
	"客厅",
	"餐厅",
	"主人房",
	"书房",
	"卫生间",
}

var LocationDefaultNotChosen = [...]string{
	"行政部",
	"市场部",
	"研发部",
	"总裁办",
}

type Location struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Sort int    `json:"sort"`
}

type locationListResp struct {
	Locations []Location `json:"locations"`
}

type locationsDefaultResp struct {
	LocationsDefault []locationDefault `json:"locations"`
}
type locationDefault struct {
	Name   string `json:"name"`
	Chosen bool   `json:"chosen"`
}

func ListLocation(c *gin.Context) {
	var (
		err       error
		resp      locationListResp
		locations []orm.Location
	)
	defer func() {
		if resp.Locations == nil {
			resp.Locations = make([]Location, 0)
		}
		response.HandleResponse(c, err, resp)
	}()

	if locations, err = orm.GetLocations(); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	resp.Locations = WrapLocations(locations)
	return

}

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

func WrapLocations(locations []orm.Location) (result []Location) {
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
