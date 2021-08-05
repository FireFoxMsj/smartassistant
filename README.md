# 智汀家庭云

智汀家庭云（SmartAssistant），立项于2021年，结合国内智能家居各厂商软件特点
，研发“智汀家庭云”，并对该生态系统全面开源，为国内首个采用智能家居系统全生态开源协议（Apache License, Version 2.0）的软件。

## 核心功能

* 局域网内智能设备的发现，管理与场景互动
* 开放插件接口，并且提供插件开发SDK，方便第三方设备接入
* 智汀家庭云提供PC版、IOS版、安卓版的终端
* 通过绑定到智汀云帐号，提供外网控制的功能

## 快速入门

如果您机器上安装有 docker 与 docker-compose 环境，可按照 [使用 Docker 运行智汀家庭云](docs/tutorial/docker-quickstart.md) 的步骤体验智汀家庭云的基本功能。

智汀家庭云是一个开源项目，如果如果您熟悉 go 编程语言，想参与到项目的开发中，可以访问 [开发环境搭建](docs/tutorial/development-quickstart.md) 。

智汀家庭云提供插件系统支持第三方设备接入，如果您的设备不在我们的支持列表，可以参考 [开发您的第一个插件](docs/tutorial/plugin-quickstart.md)了解插件开发相关内容。

## 开发指南

* [架构概述](docs/guide/architecture.md)
* [用户模块](docs/guide/user-module.md)
* [用户认证与第三方授权](docs/guide/authenticate.md)
* [设备模块](docs/guide/device-module.md)
* [设备控制场景](docs/guide/device-scene.md)
* [插件模块](docs/guide/plugin-module.md)
* [HTTP API 接口规范](docs/guide/http-api.md)
* [WebSocket API 消息定义](docs/guide/web-socket-api.md)

## 参与项目

您可以通过给我们提交问题反馈、提交合并请求（pull request）或者开发第三方插件的方式参与到项目中。关于参与项目的详细指引请阅读 [如何参与项目](docs/guide/contributing.md) 文档。

## 开源协议

智汀家庭云项目源码基于 [APACHE LICENSE, VERSION 2.0](https://www.apache.org/licenses/LICENSE-2.0) 发布。
