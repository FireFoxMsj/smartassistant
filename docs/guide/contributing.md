# 如何参与项目

智汀家庭云是基于 Apache License, Version 2.0 发布的开源软件，您可以通过给我们提交问题反馈、
提交合并请求（pull request）或者开发第三方插件的方式参与到项目中。

## 提交问题反馈

本项目使用 git issue 跟踪反馈问题，在提交问题反馈前，请先进行以下操作：

* 确保您的应用版本已经是最新，因为您遇到的问题可能在最新版本中已经修复
* 可以尝试切换到旧版本，测试问题是否依然存在，这有助于我们可以快速定位问题
* 查看项目的 issue 列表中是否已存在该问题

### 问题反馈需要包含的信息

为了让项目开发者能快速复现问题，建议提交的问题反馈中至少包含以下信息：

* 应用程序版本，可以是某个 release 版本号，或者对应代码提交的 commit id
* 本地运行环境，譬如 Linux 发行版，Windows 或者 MacOS 版本，64位还是32位系统，越详细越好
* 使用的 Golang 版本号，如 Golang 1.16，Golang 1.15
* 开发者如何能复现 bug？可以包括一系列的操作，或者是一段代码，也可以是任何相关的上下文信息；当然，也是越详细越好

## 提交合并请求

在开始编码前建议先阅读项目的快速入门文档以及开发文档，如
[使用 Docker 运行智汀家庭云](../tutorial/docker-quickstart.md),
[开发环境搭建](../tutorial/development-quickstart.md),
[架构概述](./architecture.md) 等。

需要注意的是智汀家庭云基于 Apache License, Version 2.0 开源协议发布，请确保您的代码与该协议兼容。

### 编码规范

编码规范主要参考 [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md) （ 
 [Uber Go 语言编码规范](https://github.com/xxjwxc/uber_go_guide_cn) ）

### 开发流程

项目主要包含以下分支：

* master 预发布的分支，我们会基于 master 分支来打新版本的标签，如 1.1.0，1.2.0
* dev 主开发分支，当新功能开发、测试完成后会合并到 master 分支上，**建议基于此分支提交合并请求**
* production 分支针对最新版本的修复，合并后会打 1.1.1，1.1.2 等标签进行发布

开发流程
* Fork 项目源码到您的帐号，然后从 master, dev 或者 production 分支进行开发
* 编写代码，并且同步更新文档
* 确保您的代码符合我们的编码规范以及开源协议
* 测试您的代码
* 提交合并请求到 dev 或者 production 分支

## 开发第三方插件

您也可以通过开发插件的形式参与到项目中，完善智汀家庭云对第三方硬件的支持，让更多用户受惠。

可以先阅读[开发您的第一个插件](../tutorial/plugin-quickstart.md)来快速入门插件开发，
然后阅读[插件系统设计技术概要](./plugin-module.md),[设备插件开发](./device-module.md)
等文档进一步了解插件的实现机制。
