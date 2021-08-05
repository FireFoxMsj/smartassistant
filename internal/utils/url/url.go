package url

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/zhiting-tech/smartassistant/internal/config"
)

// BuildQuery 将map中数据转换成 url.Values
func BuildQuery(query map[string]interface{}) url.Values {
	values := url.Values{}
	for k, v := range query {
		values.Add(k, fmt.Sprintf("%v", v))
	}
	return values
}

// BuildURL 根据请求和相对路径转换成地址
func BuildURL(path string, query map[string]interface{}, req *http.Request) string {
	values := BuildQuery(query)
	u := url.URL{
		Host:     req.Host,
		Path:     path,
		Scheme:   req.URL.Scheme,
		RawQuery: values.Encode(),
	}

	if u.Scheme == "" { // 使用nginx代理则需要配置才能获取scheme
		u.Scheme = req.Header.Get("X-Scheme")
		if u.Scheme == "" {
			u.Scheme = "http"
		}
	}
	return u.String()
}

func StaticPath() string {
	return ConcatPath("api", "static", config.GetConf().SmartAssistant.ID)
}

// ConcatPath 拼接路径
func ConcatPath(paths ...string) string {
	return strings.Join(paths, "/")
}

// SAImageUrl SA的Logo地址
func SAImageUrl(req *http.Request) string {
	path := ConcatPath(StaticPath(), "img", "智慧中心.png")
	return BuildURL(path, nil, req)
}
