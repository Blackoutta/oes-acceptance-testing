# 边缘计算套件验收助手
## 关于本项目
本项目为OneNET智能边缘计算套件的测试验证工作提供辅助，同时包含自动化测试脚本和设备模拟器。自动化测试脚本用于验证流程，设备模拟器用于演示或Debug。

## 目录
- [OneNET智能边缘计算套件验收助手](#onenet智能边缘计算套件验收助手)
  - [关于本项目](#关于本项目)
  - [目录](#目录)
  - [手动测试用设备模拟器使用说明](#手动测试用设备模拟器使用说明)
    - [快速开始](#快速开始)
    - [MQTT设备使用动态注册(即产品级鉴权)](#mqtt设备使用动态注册即产品级鉴权)
  - [自动化验收测试程序使用说明](#自动化验收测试程序使用说明)
    - [Docker执行](#docker执行)
    - [非Docker运行 (以非docker形式运行数据路由场景需安装额外依赖:librdkafka，安装指南见pkg/helpers/notes/how-to-install-librdkafka文件)](#非docker运行-以非docker形式运行数据路由场景需安装额外依赖librdkafka安装指南见pkghelpersnoteshow-to-install-librdkafka文件)
    - [配置文件示例](#配置文件示例)
    - [测试用例参考](#测试用例参考)

## 手动测试用设备模拟器使用说明
### 快速开始
首先直接运行```bin```目录下的可执行文件，查看可用的子命令：
```
oes-sim-windows.exe // Windows运行此文件
./oes-sim-linux     // Linux运行此文件
./oes-sim-mac       // Mac(darwin)运行此文件
```

可以看到打印出来的程序帮助文档：
```
OneNET智能边缘计算套件设备模拟器，支持协议：mqtt, mqtt透传，tcp透传，modbus rtu over tcp, lwm2m, http

Usage:
  oes-sim [flags]
  oes-sim [command]

// 此处显示了oes-sim程序的可用命令
Available Commands:
  help        Help about any command
  http        模拟http设备
  lwm2m       模拟lwm2m设备
  modbus      模拟modbus设备
  mqtt        模拟mqtt设备
  mqttps      模拟mqtt透传设备
  tcp         模拟tcp设备

Flags:
  -h, --help   help for oes-sim

Use "oes-sim [command] --help" for more information about a command.
```

然后通过help命令来查看各个子命令的更详细的帮助信息：
```
oes-sim-windows.exe help mqtt
```
子命令的帮助信息中提供了配置文件的位置和一些注意事项：
```
mqtt设备模拟器，配置文件格式见./configs/manual/mqtt_manual.yaml
支持自定义上行payload，上行payload文件见./data/payload/mqtt.json

Usage:
  oes-sim mqtt [flags]

Flags:
  -h, --help   help for mqtt
```

在修改配置文件和payload文件(如果支持)后，开始运行模拟器：
```
oes-sim-windows.exe mqtt
```

如果配置信息有效，则可以看到模拟器成功上线：
```
2020/07/14 14:33:54 连接成功！等待2秒...
已上行数据: {
    "params": {
        "someBoolean": {
            "value": false
        },
        "someDate": {
            "value": 1594604139000
        },
        "someDouble": {
            "value": 999999.999999999
        },
        "someEnum": {
            "value": 20
        },
       ......
```

只要更换子命令，即可在不同协议的模拟器之间切换，切记启动前修改配置文件和调整上行payload文件(如果支持)
```
oes-sim-windows.exe mqtt
oes-sim-windows.exe mqttps
oes-sim-windows.exe lwm2m
oes-sim-windows.exe tcp
oes-sim-windows.exe modbus
oes-sim-windows.exe http
```

### MQTT设备使用动态注册(即产品级鉴权)
使用```-d```flag来让模拟器使用动态注册鉴权。注意，要使用动态注册服务，需首先在MQTT产品详情页面上打开"动态注册"开关，如在未打开该开关的情况下进行动态注册，会返回错误："产品 {产品id} 未开启动态注册。另外需要在configs/manual/mqtt_manual.yaml配置文件中配置好pid, prodSecret, devName和dynamicRegisterServer。
```
oes-sim-windows.exe mqtt -d
```


## 自动化验收测试程序使用说明
### Docker执行
由于数据路由测试场景需要使用到confluent的kafka消费者，而它依赖librdkafka-1.3，客户环境难以安装，所以建议按下方步骤使用Docker来运行测试脚本

构建测试镜像
```
docker build -t oes-acc-test:5.0 -f scripts/Dockerfile .
```

运行不同场景测试
```
docker run \
-e CONFIGFILE={指定json配置文件, 支持从容器外挂载, 见下方示例} \
--add-host={域名}:{IP} \
-e SUITE={测试场景} \
oes-acceptance-test
```

通过挂载容器外部的配置文件来执行测试：
```
docker run \
-v {外部配置文件的绝对路径}:/cf/config.json
-e CONFIGFILE=/cf/config.json \
--add-host={域名}:{IP} \
-e SUITE={测试场景} \
oes-acceptance-test
```

由于每个环境的IP不同，所以容器内的host需通过--add-host这个flag来定义

以下是现可以使用的测试场景，请赋值给环境变量$SUITE
```
test-datarouter //测试MQTT设备接入并通过数据路由将上行的数据转发至mqtt broker, mysql, kafka和rest服务器
test-ruleengine //测试MQTT设备接入后通过规则引擎进行场景联动
test-lwm2m      //测试lwm2m设备接入+数据上下行+数据点查询
test-mqtt       //测试mqtt设备接入+数据上下行+数据点查询
test-mqttps     //测试mqtt透传设备接入+数据上下行+数据点查询
test-tcp        //测试tcp设备接入+数据上下行+数据点查询
test-http       //测试http设备数据上行+数据点查询
test-modbus     //测试modbus设备上下行+数据点查询

test-all        //上述所有测试场景一起并行执行。不指定测试场景时默认运行该场景
```

例子 - 执行Mqtt测试场景
```
docker run \
-e CONFIGFILE='configs/auto/cf.test.json' \
-e SUITE=test-mqtt \
--add-host={hostname}:{host ip} \
oes-acc-test:5.0
```

### 非Docker运行 (以非docker形式运行数据路由场景需安装额外依赖:librdkafka，安装指南见pkg/helpers/notes/how-to-install-librdkafka文件)

首先设置环境变量：
```
export CONFIGFILE=./configs/auto/cf.test.json
```
如有安装Go，则直接使用Go运行：
```
go run exec/auto/mqtt/mqtt_exec.go   
```
如未安装Go，可直接通过二进制文件执行(只有linux的二进制文件包含test-all和test-datarouter套件，因为Golang的kafka消费者拥有linux系统依赖)：
```
./bin/auto/linux/test-mqtt-linux
```
正常情况下测试完成后，脚本会自动清理所有创建的资源，**如想跳过清理阶段且让套件中的设备一直持续运行，可使用**```-h```flag（注意，如果套件失败，则该操作无效）：
```
./bin/auto/linux/test-mqtt-linux -h
```


### 配置文件示例
```
//注意，请勿修改字段名
{
    "baseURL": "xxxxxxxxx",   //openapi的url
    "accessKeyId": "xxxxxxxxx",                  //可通过页面上的顶部导航栏-accessKey-添加accessKey获取
    "secret": "xxxxxxxxxxxxxxxxx",   //可通过页面上的顶部导航栏-accessKey-添加accessKey获取
    "tcpServer": "xxx.xxx.xxx.xxx",                   //tcp接入机地址
    "mqttBroker": "xxx.xxx.xxx.xxx",                  //mqtt接入机地址
    "mqttsBroker": "xxx.xxx.xxx.xxx",                 //mqtts接入机地址，用于边缘网关接入
    "lwm2mServer": "xxx.xxx.xxx.xxx",                 //lwm2m接入机地址
    "modbusRTUServer": "xxx.xxx.xxx.xxx",             //modbus-rtu接入机地址
    "httpServer": "xxx.xxx.xxx.xxx"                   //http接入机地址
    "dataRouterDestination": "xxx.xxx.xxx.xxx"            //测试数据转发时的目的地服务器
}
```

[返回目录](#目录)