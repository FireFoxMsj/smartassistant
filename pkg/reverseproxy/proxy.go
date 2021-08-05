// Package reverseproxy 反向代理
package reverseproxy

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	response2 "github.com/zhiting-tech/smartassistant/internal/api/utils/response"
	session2 "github.com/zhiting-tech/smartassistant/internal/utils/session"
)

var (
	ErrorUpstreamNotValid = errors.New("upstream not valid")
	ErrorUpstreamExists   = errors.New("upstream exists")
	ErrorUpstreamNotFound = errors.New("upstream not found")
)

var servers = make(map[string]upstream)
var mu sync.RWMutex

type upstream struct {
	proxy *httputil.ReverseProxy
}

func (u *upstream) handleRequest(ctx *gin.Context) {
	req := ctx.Request.Clone(context.Background())
	user := session2.Get(ctx)
	if user != nil {
		req.Header.Add("scope-user-id", strconv.Itoa(user.UserID))
	}

	u.proxy.ServeHTTP(ctx.Writer, req)
}

func RegisterUpstream(path string, serverUrl string) (err error) {
	return registerUpstream(path, serverUrl)
}

func UnregisterUpstream(path string) (err error) {
	return unregisterUpstream(path)
}

// registerUpstream 注册一个反向代理后端
// serverUrl 支持 :8090, 127.0.0.1:8090，http://127.0.0.1:8090 格式
func registerUpstream(path string, serverUrl string) (err error) {
	if len(path) == 0 || len(serverUrl) == 0 {
		err = ErrorUpstreamNotValid
		return
	}
	if strings.HasPrefix(serverUrl, ":") {
		// 只提供端口
		serverUrl = fmt.Sprintf("http://127.0.0.1%s", serverUrl)
	} else if !strings.HasPrefix(serverUrl, "http") {
		serverUrl = fmt.Sprintf("http://%s", serverUrl)
	}
	su, err := url.Parse(serverUrl)
	if err != nil {
		return err
	}
	mu.Lock()
	defer mu.Unlock()
	_, ok := servers[path]
	if ok {
		err = ErrorUpstreamExists
		return
	}
	up := upstream{
		proxy: httputil.NewSingleHostReverseProxy(su),
	}
	servers[path] = up
	return
}

func unregisterUpstream(path string) (err error) {
	if len(path) == 0 {
		err = ErrorUpstreamNotValid
		return
	}
	mu.Lock()
	delete(servers, path)
	mu.Unlock()
	return
}

// ProxyToPlugin 根据路径转发到后端插件
func ProxyToPlugin(ctx *gin.Context) {
	path := ctx.Param("plugin")
	mu.RLock()
	up, ok := servers[path]
	mu.RUnlock()
	if !ok {
		response2.HandleResponseWithStatus(ctx, http.StatusBadGateway,
			ErrorUpstreamNotFound, nil)
		return
	}
	logrus.Debugf("plugin %s http server found,proxy...", path)
	up.handleRequest(ctx)
}
