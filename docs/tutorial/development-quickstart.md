# 开发环境搭建

此文档描述如何搭建智汀家庭云开发环境，下载，编译与运行。如果你只是想体验智汀家庭云的功能，可以先阅读[使用 Docker 运行智汀家庭云](./docker-quickstart.md)；如果你是想进行插件开发，可参考[开发您的第一个插件](./plugin-quickstart.md)。

## 环境准备

* go 版本为 1.15.0 或以上
* 确保能生成 gRPC 代码，请参考 [gRPC Quick start](https://grpc.io/docs/languages/go/quickstart/)
* docker 与 docker-compose, 如果需要 smartassistant 与插件进行交互，则需要安装此依赖

## 步骤

获取代码

``` shell
git clone https://github.com/zhiting-tech/smartassistant.git
```

同步依赖

``` shell
go mod tidy
```

复制 app.yaml.example 到 app.yaml 并配置

``` yaml
debug: true
# 智汀云
smartcloud:
    # 留空则不连接智汀云
    domain: ""  
    tls: false
# 智汀家庭云
smartassistant:
    # 由智汀云分配的ID
    id: "demo-sa"
    # 由智汀云分配的设备密钥
    key: "aGVsbG93b3JsZA"
    db: "sadb.db"
    host: 0.0.0.0
    port: 8088
    grpc_port: 9234

docker:
    server: ""
    username: ""
    password: ""
   
```

编译运行

``` shell
go run ./cmd/smartassistant/main.go 
```

如果已安装 docker 与 docker-compose，则可以通过以下命令进行打包与运行

``` shell
make build
make run
```

然后可以访问以下地址确认服务是否正常运行

``` shell
curl http://localhost:8088/check
```

正常会返回

``` json
{"status":0,"reason":"成功","data":{"is_bind":false}}
```
