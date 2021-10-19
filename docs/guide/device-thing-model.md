# 设备物模型

## light_bulb灯泡

|    attribute   |description | val_type |required |
| ----------|--- | --- |---|
| power|开关 |  string |true|
| brightness |亮度    |  int |false
| color_temp  |色温     |  int |false

## switch开关

|    attribute   |description | val_type |required |
| ----------|--- | --- |---|
| power|开关 |  string |true|

## outlet插座

|    attribute   |description | val_type |required |
| ----------|--- | --- |---|
| power|开关 |  string |true|

## info设备详情

|    attribute   |description | val_type |required |
| ----------|--- | --- |---|
| identity|id |  string |true|
| model|型号 |  string |true|
| manufacturer|厂商 |  string |true|
| version|版本 |  string |true|
| name|名字 |  string |false|

## curtain窗帘

|    attribute   |description | val_type |required |
| ----------|--- | --- |---|
| current_position|当前位置，0-100 |  int |true|
| target_position|目标位置，0-100 |  int |true|
| state|当前状态，0关1开2暂停 |  enum |true|
| direction|方向，0默认方向1反方向 |  enum |false|
| upper_limit|上限，0删除1设置 |  enum |false|
| lower_limit|下限，0删除1设置 |  enum |false|