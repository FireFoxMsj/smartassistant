module github.com/zhiting-tech/smartassistant

go 1.15

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

require (
	github.com/containerd/containerd v1.5.2 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/docker v20.10.7+incompatible
	github.com/gin-contrib/sessions v0.0.3
	github.com/gin-gonic/gin v1.7.2
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.2.0
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/websocket v1.4.2
	github.com/inlets/inlets v0.0.0-20210509192755-9df7d77ced40
	github.com/jinzhu/now v1.1.2
	github.com/json-iterator/go v1.1.11
	github.com/mattn/go-sqlite3 v1.14.6 // indirect
	github.com/micro/go-micro/v2 v2.9.1
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/rancher/remotedialer v0.2.6-0.20201012155453-8b1b7bb7d05f
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.8.1
	github.com/twinj/uuid v1.0.0
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110
	google.golang.org/grpc v1.33.2
	gopkg.in/yaml.v2 v2.4.0
	gorm.io/datatypes v1.0.1
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.21.12
)
