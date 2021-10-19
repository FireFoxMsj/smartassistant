package cloud

import (
	"bytes"
	"context"
	"encoding/json"
	errors2 "errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/supervisor"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

const (
	HttpRequestTimeout = (time.Duration(30) * time.Second)
)

type AreaMigrationReq struct {
	MigrationUrl string `json:"migration_url"`
	MD5          string `json:"md5"`
}

func (req *AreaMigrationReq) SendAreaMigrationToSAC() (file string, err error) {
	var (
		content    []byte
		httpReq    *http.Request
		ofile      *os.File
		fileLength int64
	)
	body := map[string]interface{}{
		"md5":  req.MD5,
		"said": config.GetConf().SmartAssistant.ID,
		"key":  config.GetConf().SmartAssistant.Key,
	}
	content, err = json.Marshal(body)
	if err != nil {
		return
	}
	httpReq, err = http.NewRequest(http.MethodPost, req.MigrationUrl, bytes.NewBuffer(content))
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), HttpRequestTimeout)
	httpReq.WithContext(ctx)
	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return
	}
	if httpResp.StatusCode != http.StatusOK {
		text := fmt.Sprintf("Status Not OK, Status Code %d", httpResp.StatusCode)
		err = errors2.New(text)
		return
	}
	defer httpResp.Body.Close()

	ofile, err = ioutil.TempFile(supervisor.GetManager().BackupPath, "temp")
	if err != nil {
		return
	}
	defer ofile.Close()

	fileLength, err = io.Copy(ofile, httpResp.Body)
	if err != nil {
		return
	} else if fileLength != httpResp.ContentLength {
		text := fmt.Sprintf("write %d bytes, file content length %d", fileLength, httpResp.ContentLength)
		err = errors2.New(text)
		return
	}
	file = filepath.Base(ofile.Name())

	return
}

func AreaMigration(c *gin.Context) {
	var (
		req  AreaMigrationReq
		err  error
		file string
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	err = c.BindJSON(&req)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	file, err = req.SendAreaMigrationToSAC()
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	err = supervisor.GetManager().StartRestoreCloudJob(file)
	if err != nil {
		if os.IsNotExist(err) {
			err = errors.Wrap(err, status.FileNotExistErr)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
	}
}
