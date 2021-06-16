package handlers

import (
	"github.com/gin-gonic/gin"
	"gitlab.yctc.tech/root/smartassistent.git/core/http/page/config"
	"net/http"
)

func CreateSceneGuideRedirect(c *gin.Context) {
	c.Redirect(http.StatusFound, config.Page.CreateSceneGuideLink)
}
