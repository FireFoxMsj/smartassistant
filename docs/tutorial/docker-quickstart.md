# 使用 Docker 运行智汀家庭云

本文档描述如何在 Docker 环境下运行智汀家庭云，并且通过交互式客户端控制虚拟设备。

## 环境准备

只要您的主机安装了 docker 与 docker-compose，都可以运行智汀家庭云。但通常情况下，智汀家庭云更适合运行在 Linux 主机上。

## 运行智汀家庭云

首先确保主机上已安装 docker 与 docker-compose，并且能够正常运行：

``` shell
docker version
docker-compose version
```

创建一个目录作为智汀家庭云运行的根目录，并在该目录内创建文件 docker-compose.yaml，内容如下：

``` yaml
version: "3.9"

services:
  etcd:
    image: bitnami/etcd:3
    environment:
      - ALLOW_NONE_AUTHENTICATION=yes
    ports:
      - 2379:2379
      - 2380:2380
  smartassistant:
    image: zhitingtech/smartassistant:latest
    privileged: true
    ports:
      - "8088:8088"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    depends_on:
      - etcd
  saclient:
    image: zhitingtech/saclient:latest
    network_mode: "host"
    depends_on:
      - smartassistant

```

该文件配置了三个服务：

* smartassistant 为智汀家庭云主服务
* etcd 基础服务，用于插件注册与发现
* saclient 提供一个命令行跟智汀家庭云进行交互，用于 demo 数据的初始化以及设备控制演示

输入以下命令启动服务：

``` shell
docker-compose build
docker-compose up
```

可以通过以下命令查看智汀家庭云状态：

```shell
curl http://localhost:8088/check
```

如果返回以下内容则说明服务已运行起来并且未被绑定

``` json
{"status":0,"reason":"成功","data":{"is_bind":false}}
```

如果返回数据里 is_bind 字段值为 true，则需要删除运行目录的 sadb.db 文件（**注意：** 该操作会删除已有数据），并且重新运行。

## 测试智汀家庭云

下面通过 saclient 命令演示初始化智汀家庭云以及添加设备，控制设备的过程。

运行以下命令进入 saclient 命令行界面：

``` shell
docker-compose exec saclient /app/saclient
```

可以通过 help 子命令查看 saclient 提供的功能

``` shell
saclient » help
```

saclient 主要提供以下子命令：

* bind 初始化智汀家庭云，并且生成管理员帐号
* discover 通过插件扫描设备
* add 将设备添加到智汀家庭云
* info 获取设备信息
* power 向设备发送控制指令

可以按下面的流程进行测试：

``` shell
saclient » bind
add device ok, ID: 1
bind ok, user token: xxxxxx
saclient » discover
{"id":1,"type":"","result":{"device":{"name":"zhiting_M1_abcdefg","identity":"abcdefg","model":"M1","manufacturer":"zhiting","plugin_id":"demo"}},"success":true}
saclient » add -i abcdefg -f zhiting -m M1 -n zhiting_M1_abcdefg -p demo
add device ok, ID: 2
saclient » info -i 2
{"status":0,"reason":"成功","data":{"device_info":{"id":2,"name":"zhiting_M1_abcdefg","logo_url":"device_logo_url","model":"M1","location":{"name":"","id":0},"plugin":{"name":"demo","id":"demo","url":"http://127.0.0.1:8088/api/plugin/demo/html/?device_id=2\u0026identity=abcdefg\u0026model=M1\u0026name=zhiting_M1_abcdefg\u0026plugin_id=demo\u0026sa_id=demo-sa\u0026token=MTYyODA2NDAzNXxOd3dBTkZvM1ZqTkhSRXhYVGtsWVNVTllVbG8yVHpKUVVWTlNWalJhUmtaTFVWVkVWVFF5VkZaQ1RsbEVWMVZJVkVsTlNESktRVUU9fFTRixaAgA1QqIpJVvogB87HvvjTpFI3dk4ifqvupUdE"},"attributes":[{"id":1,"attribute":"power","val":"off","val_type":"string","instance_id":1},{"id":2,"attribute":"color_temp","val":0,"val_type":"int","instance_id":1},{"id":3,"attribute":"brightness","val":0,"val_type":"int","min":1,"max":100,"instance_id":1}],"permissions":{"update_device":true,"delete_device":true}}}}
saclient » power -i abcdefg -p demo -v on 
{"event_type":"attribute_change","data":{"attr":{"id":1,"attribute":"power","val":"on","val_type":"string"},"identity":"abcdefg","instance_id":1}}
saclient » power -i abcdefg -p demo -v off
{"event_type":"attribute_change","data":{"attr":{"id":1,"attribute":"power","val":"off","val_type":"string"},"identity":"abcdefg","instance_id":1}}
```

## 进一步了解

如果您手上有智汀家庭云支持的硬件设备，可以安装第三方插件，然后通过智汀APP接入您的设备。

智汀家庭云是一个开源项目，如果如果您熟悉 go 编程语言，想进一步了解我们的项目，可以访问[开发环境搭建](./development-quickstart.md)

智汀家庭云提供插件系统支持第三方设备接入，如果您的设备不在我们的支持列表，可以参考[开发您的第一个插件](./plugin-quickstart.md)了解插件开发相关内容。

## 常见问题

测试过程中可能会遇到容器中数据不一致的情况，可以通过挂在本地卷的方式解决

使用如下的 docker-compose.yaml 配置：

```yaml
version: "3.9"

services:
  etcd:
    image: bitnami/etcd:3
    environment:
      - ALLOW_NONE_AUTHENTICATION=yes
    ports:
      - 2379:2379
      - 2380:2380
  smartassistant:
    image: zhitingtech/smartassistant:latest
    ports:
      - "8088:8088"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./sadb.db:/app/sadb.db
    depends_on:
      - etcd
  saclient:
    image: zhitingtech/saclient:latest
    network_mode: "host"
    volumes:
      - ./saclient.yaml:/app/app.yaml
    depends_on:
      - smartassistant
```

本地目录创建一个空的 sadb.db 文件；以及具有以下内容的 saclient.yaml：

``` yaml
token: ""
server: http://127.0.0.1:8088
device: ""
```

然后按正常流程启动：

输入以下命令启动服务：

``` shell
docker-compose up
```