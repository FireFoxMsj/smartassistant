package page

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func CreateSceneGuideRedirect(c *gin.Context) {
	c.Redirect(http.StatusFound, Page.CreateSceneGuideLink)
}
