# WebSocket API

通常，一个 WebSocket 消息格式如下：

```json
{
  "id": 1,
  "domain": "",
  "service": "",
  "service_data": {
    "device_id": 1
  }
}
```
* id: 消息ID，必填，服务端会返回对应 ID 的结果
* domain: `plugin`或者插件id

## 设备相关命令

## 插件设备状态变更

```json
{
  "type": "attribute_change",
  "identity": "2762071932",
  "instance_id": 2,
  "attr": {
    "attribute": "power",
    "val": "on",
    "val_type": "string"
  }
}
```

### 发现设备

### req

```json
{
  "id": 1,
  "service": "discover"
}
```

### resp

```json
{
    "id": 1,
    "type": "",
    "result": {
        "device": {
            "name": "zhiting_M1",
            "identity": "hijklmn",
            "model": "M1",
            "manufacturer": "zhiting",
            "plugin_id": "demo"
        }
    },
    "success": true
}
```

### 获取设备属性

#### req

```json
{
  "id": 1,
  "domain": "zhiting",
  "service": "get_attributes",
  "identity": "2762071932"
}
```

#### resp

```json
{
  "id": 1,
  "result": {
    "identity": "2762071932",
    "device": {
      "name": "",
      "identity": "2762071932",
      "instances": [
        {
          "type": "light_bulb",
          "instance_id": 0,
          "attrs": [
            {
              "attribute": "power",
              "val": "on",
              "val_type": "string"
            },
            {
              "attribute": "brightness",
              "val": 55,
              "val_type": "int"
            },
            {
              "attribute": "color_temp",
              "val": 3500,
              "val_type": "int"
            }
          ]
        }
      ]
    }
  },
  "success": true
}
```

### 设置设备属性

#### req

```json
{
  "id": 1,
  "domain": "zhiting",
  "service": "set_attributes",
  "identity": "2762071932",
  "service_data": {
    "attributes": [
      {
        "instance_id": 1,
        "attribute": "power",
        "val": "on"
      }
    ]
  }
}

```

#### resp

```json
{
  "id": 1,
  "type": "response",
  "success": true,
  "error": "error"
}
```