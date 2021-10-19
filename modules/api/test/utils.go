package test

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"github.com/zhiting-tech/smartassistant/modules/api/middleware"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/utils"
	"github.com/zhiting-tech/smartassistant/modules/utils/hash"

	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

type ApiTestCase struct {
	Method  string
	Path    string
	Body    string
	Status  int64
	Reason  string
	IsArray []string
	IsID    []string
}

type RegisterRouterFunc func(r gin.IRouter)

type options struct {
	isLogin bool
	roles   []string
	areaID  uint64
}

type Option interface {
	apply(*options)
}
type optionFunc func(*options)

func (f optionFunc) apply(o *options) { f(o) }

func WithRoles(roles ...string) Option {
	return optionFunc(func(o *options) {
		o.isLogin = true
		o.roles = roles
	})
}

func WithAreas(areaID uint64) Option {
	return optionFunc(func(o *options) {
		o.areaID = areaID
	})
}

var areaID uint64 = utils.SAAreaID()

// CreateRecord 用来在测试过程中往数据库插入一条记录
func CreateRecord(s interface{}) {
	entity.GetDB().Create(s)
}

// RunApiTest 根据配置运行API测试
func RunApiTest(t *testing.T, rFunc RegisterRouterFunc, cases []ApiTestCase, opts ...Option) {
	var user entity.User
	options := options{
		isLogin: false,
		areaID: areaID,
	}
	for _, o := range opts {
		o.apply(&options)
	}
	if options.isLogin {
		user = initUser(options.roles, options.areaID)
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.DefaultMiddleware())
	rFunc(r)
	for _, c := range cases {
		var reader io.Reader
		if len(c.Body) > 0 {
			reader = strings.NewReader(c.Body)
		}
		req, _ := http.NewRequest(c.Method, c.Path, reader)
		if user.ID > 0 {
			req.Header.Add(types.SATokenKey, user.Token)
		}
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)
		assert.Equal(t, resp.Code, 200)
		body, _ := ioutil.ReadAll(resp.Body)
		data := string(body)
		assert.Equal(t, c.Status, gjson.Get(data, "status").Int())
		if len(c.Reason) > 0 {
			reason := gjson.Get(data, "reason").String()
			if !bytes.Equal([]byte(c.Reason), []byte(reason)) {
				t.Errorf("reason not match: expect %s, get %s ", c.Reason, reason)
			}
		}
		if len(c.IsArray) > 0 {
			for _, item := range c.IsArray {
				assert.True(t, gjson.Get(data, item).IsArray(), item)
			}
		}
		if len(c.IsID) > 0 {
			for _, item := range c.IsID {
				assert.True(t, gjson.Get(data, item).Int() > 0, item)
			}
		}
	}
}

func initUser(roles []string, areaID uint64) entity.User {
	user := entity.User{
		Nickname:  "test_user",
		Token:     hash.GetSaToken(),
		CreatedAt: time.Now(),
		AreaID:    areaID,
	}
	entity.CreateUser(&user, entity.GetDB())
	var userRoles []entity.UserRole
	for _, r := range roles {
		var role entity.Role
		entity.GetDB().Where("name=?", r).First(&role)
		if role.ID > 0 {
			userRoles = append(userRoles, entity.UserRole{
				UserID: user.ID,
				RoleID: role.ID,
			})
		}
	}
	if len(userRoles) > 0 {
		entity.CreateUserRole(userRoles)
	}
	return user
}

func InitApiTest(m *testing.M) {
	config.TestSetup()
	_ = entity.InitRole(entity.GetDB(), areaID)
	code := m.Run()
	config.TestTeardown()
	os.Exit(code)
}

func GetAreas() (areas []entity.Area) {
	area,_ := entity.GetAreas()
	return area
}