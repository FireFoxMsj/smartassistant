# 物模型

## 什么是物模型

智汀为设备定义了一套物模型，用于描述设备的功能。

物模型是物理世界的实体东西的一个抽象，是进行数字化描述后，用于数字世界的数字模型。

以智能灯为例，不同的灯，尽管规格不同，但它们的属性是相似的，比如都有开关状态的属性，功能逻辑也相仿。我们可以将这些特征标准化，形成智能灯的物模型。

第三方开发者通过使用智汀定义的物模型自由组合就可以描述产品/硬件，并将其能力接入到智汀家庭云。

## 主要概念解释

|    定义     | 说明 |
|------------|-----|
|  device    | 设备，具体的产品/设备，包含一个或多个instance  |
|  instance  | 实例，基于智汀物模型生成的实例  |
|  attribute | 属性，物模型对应的属性或者功能  |

一个设备/产品等同于多个物模型的集合，相当于使用物模型组合并描述这个设备

举例：

智汀定义了开关的物模型[switch](device-thing-model.md#switch)，拥有一个属性-开关；

现有一个设备是一个三键开关，这个三键开关就是device，拥有三个switch instance，每个instance分别拥有开关的属性。

这样子，开发者通过使用物模型描述设备并将设备的能力集成到智汀家庭云中。

