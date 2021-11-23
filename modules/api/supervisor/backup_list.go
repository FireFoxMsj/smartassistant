package supervisor

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/supervisor"
)

// backupListResp 备份列表返回
type backupListResp struct {
	Backups []Backup `json:"backups"`
}

type Backup struct {
	FileName  string `json:"file_name"`
	Note      string `json:"note"`
	CreatedAt int64  `json:"created_at"`
}

// ListBackup 备份列表
func ListBackup(c *gin.Context) {
	var (
		resp backupListResp
		err  error
	)
	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	resp.Backups = wrapResponse(supervisor.GetManager().ListBackups())
}

func wrapResponse(backups []supervisor.Backup) []Backup {
	baks := make([]Backup, 0, len(backups))
	for _, b := range backups {
		baks = append(baks, Backup{
			FileName:  b.FileName,
			Note:      b.Note,
			CreatedAt: b.CreatedAt.Unix(),
		})
	}
	return baks
}
