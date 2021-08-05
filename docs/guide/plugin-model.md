# 插件模型定义

## 字段解释

- instance 实例
- attribute 属性
- val_type 值类型
- val 值

## val_type

对于int型，属性可设置最小和最大值

## 模型

### light_bulb 灯泡

|    attribute   |description | val_type |required |
| ----------|--- | --- |---|
| power|开关 |  string |true|
| brightness   |亮度    |  int |false
| color_temp  |色温     |  int |false

### switch 开关

|    attribute   |description | val_type |required |
| ----------|--- | --- |---|
| power|开关 |  string |true|

### outlet 插座

|    attribute   |description | val_type |required |
| ----------|--- | --- |---|
| power|开关 |  string |true|

### info 设备详情

|    attribute   |description | val_type |required |
| ----------|--- | --- |---|
| identity|id |  string |true|
| model|型号 |  string |true|
| manufacturer|厂商 |  string |true|
| version|版本 |  string |true|
| name|名字 |  string |false|