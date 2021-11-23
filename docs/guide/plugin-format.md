# 插件包格式

插件包，是一个zip压缩文件，可以通过上传插件包的形式将您的插件上传到智汀家庭云运行；插件包也可以上传到智汀官方开发者中心，在审核后开放给其他用户使用。

通常一个插件包解压后的目录如下（假设是以Go语言开发）：

``` txt
├── Dockerfile          插件包上传 docker build 输入
├── go.mod              后端 go 包管理
├── html                前端资源，需要 Dockerfile 显式 COPY 到镜像里
│   ├── index.html      
│   └── static
│       ├── css
│       ├── img
│       └── js
├── main.go             后端代码入口
└── config.json         插件描述文件

```

其中 config.json 与 Dockerfile 为必须

## config.json 定义

每个插件均需包含一个 config.json 文件，描述的插件功能，包括以下方面：

* 插件的 ID，名称，版本，描述等基础信息
* 支持的设备信息，安装插件后，用户可以在列表中选择插件支持的设备进行置网与绑定
* config.json 中使用到的 logo 图片资源需要包含在压缩包中

``` json
{
  "id": "zhiting", 
  "name": "zhiting",
  "version": "0.0.1",
  "info": "智汀设备插件",
  "support_devices": [
    {
      "name": "吸顶灯",
      "model": "ceiling17",
      "type_name": "light",
      "logo": "html/static/img/led.0bd29fdd.png",
      "provisioning": "esp32_softap", 智汀官方提供默认实现
      "control": "light" 智汀官方提供默认实现
    },
    {
      "name": "台灯",
      "model": "lamp9",
      "type_name": "light",
      "logo": "html/static/img/lamp.da1b67cc.png",
      "provisioning": "html/my_custom_provisioning.html", 用户自定义置网页面
      "control": "html/my_custom_light_control.html" 用户自定义控制页面
    }
  ]
}
```

字段含义如下:

|字段名称|含义|必填|
|---|----------|---|
|id|插件ID，全局唯一|是|
|name|插件名称|是|
|version|版本，遵循 SemVer|是|
|info|插件描述|否|
|image|插件镜像信息，参考下面 image 字段的介绍|否|
|support_devices||是|

support_devices 字段为数组，其各个Item字段含义如下：

|字段名称|含义|必填|
|---|----------|---|
|name||否|
|model|产品型号|是|
|type_name|产品所属类型，选项由智汀官方提供|是|
|logo|设备logo图片地| 否   |
|provisioning|置网页在前端资源中的相对路径|否|
|control|设备详情（控制）页在前端资源中的相对路径|否|

## Dockerfile 与其他文件

每个插件均需包含一个 Dockerfile 文件，用于对插件进行打包；为了保障安全，所有插件均需要通过智汀云进行打包，在进行安全审核后再发布；为了保障您的插件能顺利通过审核，请尽量基于官方可信镜像构建您的插件。

您的插件在上传后会使用 docker build 对插件进行镜像打包，并且 docker image tag 到对应镜像版本。

如果插件上传环境是本地智汀家庭云，则镜像在 build 与 tag 后会用于本地插件安装。

如果插件上传到智汀开发中心，镜像将上传到智汀 docker registry。

需要注意的是，插件需要用到的文件，如 html 目录，需要在 Dockerfile 中用 COPY 命令拷贝到镜像里，插件系统只会根据 config.json 的 image 信息运行插件，插件包的其他文件只在 build 阶段保留。

## 插件上传流程

* 插件以zip包上传
* 后台接收到插件包后，解析插件包中的 config.json文件，校验字段是否齐全，图片资源是否存在
* 图片资源上传到 oss，把 config.json 信息解析入库（包括插件信息、支持的设备信息）
* 执行 docker build；以及tag到 config.json 中配置的tag，上传到docker registry。，进入待审核状态
* 待后台审核后发布；如果build失败，或者审核不通过，相关信息需要则管理后台反馈给开发者
