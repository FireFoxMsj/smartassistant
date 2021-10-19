package plugin

import (
	errors2 "errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

type UploadPluginReq struct {
	File   multipart.File
	header *multipart.FileHeader
}

type UploadPluginResp struct {
	PluginInfo plugin.Plugin `json:"plugin_info"`
}

func UploadPlugin(c *gin.Context) {
	var (
		err  error
		resp UploadPluginResp
		req  UploadPluginReq
	)

	defer func() {
		response.HandleResponse(c, err, &resp)
	}()

	if err = c.Request.ParseMultipartForm(32 << 20); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	req.File, req.header, err = c.Request.FormFile("file")
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	defer req.File.Close()

	if req.File == nil || req.header == nil {
		err = errors.New(status.PluginIsEmpty)
		return
	}

	resp, err = req.uploadPlugin(c)
	if err != nil {
		return
	}
}

func (req *UploadPluginReq) uploadPlugin(c *gin.Context) (resp UploadPluginResp, err error) {

	if err = checkPluginType(req.header.Filename); err != nil {
		return
	}

	if err = os.Mkdir(req.getUploadDir(), 0755); err != nil {
		if !errors2.Is(err, os.ErrExist) {
			err = errors.Wrap(err, errors.InternalServerErr)
			return
		}
		err = nil
	}

	pluginPath := filepath.Join(req.getUploadDir(), req.header.Filename)
	destFile, err := os.Create(pluginPath)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	defer destFile.Close()
	// 复制内容
	_, err = io.Copy(destFile, req.File)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	u := session.Get(c)
	// 从上传目录中将插件build成镜像
	plg, err := plugin.LoadPluginFromZip(pluginPath, u.AreaID)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	// 删除插件压缩包
	os.RemoveAll(pluginPath)
	resp.PluginInfo = plg
	return
}

// getUploadDir 获取上传目录
func (req UploadPluginReq) getUploadDir() string {
	ext := filepath.Ext(req.header.Filename)
	// 以插件包的名字创建目录
	dirName := strings.TrimSuffix(req.header.Filename, ext)

	return fmt.Sprintf("./plugins/%s", dirName)
}

// checkPluginType 插件包类型校验
func checkPluginType(fileName string) (err error) {
	ext := filepath.Ext(fileName)
	// 目前只支持zip格式的压缩包
	if ext != ".zip" {
		err = errors.New(status.PluginTypeNotSupport)
		return
	}
	return
}
