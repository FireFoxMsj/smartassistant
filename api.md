## REST API

- GET /api/users/:user_id/devices
- POST /api/users/:user_id/devices
- GET /api/users/:user_id/devices/:device_id

## WebSocket API

id为必要字段

#### 安装插件

```json
{
  "id": 1,
  "domain": "plugin",
  "service": "install",
  "service_data": {
    "plugin_id": "plugin_id"
  }
}
```

```json
{
  "id": 1,
  "success": true
}
```

#### 更新插件

```json
{
  "id": 1,
  "domain": "plugin",
  "service": "update",
  "service_data": {
    "plugin_id": "plugin_id"
  }
}
```

```json
{
  "id": 1,
  "success": true
}
```

#### 删除插件

```json
{
  "id": 1,
  "domain": "plugin",
  "service": "remove",
  "service_data": {
    "plugin_id": "plugin_id"
  }
}
```

#### 设备状态变更

```json
{
  "event_type": "state_changed",
  "data": {
    "device_id": 1,
    "state": {
      "is_online": true,
      "power": "on/off",
      "brightness": 69,
      "color_temp": 5500
    }
  },
  "origin": ""
}
```

#### 设备功能

```json
{
  "domain": "plugin",
  "id": 1,
  "service": "get_actions",
  "service_data": {
    "device_id": 1
  }
}
```

```json
{
  "id": 0,
  "result": {
    "actions": {
      "set_bright": {
        "cmd": "set_bright",
        "name": "调节亮度",
        "is_permit": false
      },
      "set_color_temp": {
        "cmd": "set_color_temp",
        "name": "调节色温",
        "is_permit": false
      },
      "switch": {
        "cmd": "switch",
        "name": "开关",
        "is_permit": true
      }
    }
  },
  "success": true
}
```

### YeeLight

#### 发现设备

```json
{
  "domain": "yeelight",
  "id": 1,
  "service": "discover"
}
```

```json
{
  "id": 1,
  "success": true,
  "result": {
    "device": {
      "id": "21394770a79648e6a3416239e1ebecb9",
      "address": "192.168.0.195:55443",
      "identity": "0x0000000012ed37c8",
      "name": "yeelight_ceiling17_0x0000000012ed37c8",
      "model": "ceiling17",
      "sw_version": "5",
      "manufacturer": "yeelight",
      "power": "on",
      "bright": 80,
      "color_mode": 2,
      "ct": 3017,
      "rgb": 0,
      "hue": 0,
      "sat": 0
    }
  }
}
```

#### 设备信息

```json
{
  "domain": "yeelight",
  "id": 1,
  "service": "state",
  "service_data": {
    "device_id": "device_id"
  }
}
```

```json
{
  "id": 1,
  "result": {
    "state": {
      "power": "on/off",
      "brightness": 55,
      "color_temp": 4000
    }
  },
  "success": true
}
```

#### 开关

```json
{
  "domain": "yeelight",
  "id": 1,
  "service": "switch",
  "service_data": {
    "device_id": "device_id",
    "power": "on/off/toggle"
  }
}
```

#### 设置亮度

```json
{
  "domain": "yeelight",
  "id": 1,
  "service": "set_bright",
  "service_data": {
    "device_id": "device_id",
    "brightness": 100
  }
}
```

#### 设置色温

```json
{
  "domain": "yeelight",
  "id": 1,
  "service": "set_color_temp",
  "service_data": {
    "device_id": "device_id",
    "color_temp": 5000
  }
}
```

