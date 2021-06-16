# smartassistent

## Light
* Yeelight，[文档](https://www.yeelight.com/download/Yeelight_Inter-Operation_Spec.pdf)
* 绿米，[文档](http://docs.opencloud.aqara.com/development/gateway-LAN-communication/)
* 百度，[文档](https://dueros.baidu.com/didp/doc/dueros-bot-platform/dbp-smart-home/protocol/intro-protocol_markdown)

### TODO
* [ ] 接口定义方法必须实现限制
* [ ] 设备状态广播：不同实体数据更新
* [ ] 设备状态广播：不同长连接推送
* [ ] 异步任务执行

### 插件定义
  采用grpc双向绑定数据流(stream)来实现
  
  ```sh
    cd core/grpc/
    // ruby:
    grpc_tools_ruby_protoc -I ./proto/ --ruby_out=./proto/ --grpc_out=./proto/ proto/plugin.proto
    // golang:
    protoc --proto_path=:. *.proto  --go_out=plugins=grpc:.
  ```
 
```yaml
  id: "插件ID"
  name: "插件名字，只允许英文数字"
  download_url: "下载插件url"
  version: "版本号"
  info: "描述"
  logo_url: "logo地址" 
  brand: "品牌"

  support_devices: # 支持的设备列表
    - yeelight_celling7:
      logo_url: "www.baidu.com/device_logo.jpg"
      name: "yeelight_ceiling17_0x0000000012ed37c8"
```  