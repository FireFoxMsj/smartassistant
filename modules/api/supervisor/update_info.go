package supervisor

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/plugin/docker"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

type updateInfoResp struct {
	Version       string `json:"version"`
	LatestVersion string `json:"latest_version"`
}

// UpdateInfo 查看更新信息
func UpdateInfo(c *gin.Context) {
	var (
		resp updateInfoResp
		err  error
	)
	defer func() {
		response.HandleResponse(c, err, resp)
	}()
	resp.Version = types.Version
	resp.LatestVersion = ""
	tag, err := docker.GetClient().GetImageNewestTag(docker.Image{
		Name:     "smartassistant",
		Tag:      "",
		Registry: types.DockerRegistry,
	})
	if err != nil {
		err = errors.New(status.GetImageVersionErr)
		return
	}
	resp.LatestVersion = tag
}
