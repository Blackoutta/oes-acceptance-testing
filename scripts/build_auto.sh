#!/bin/bash
for os in windows linux darwin; do
EXTENSION=-$os
if [ "$os" == "windows" ];then
    EXTENSION=$EXTENSION.exe
fi
GOOS=$os go build -o ./bin/auto/$os/test-datarouter$EXTENSION ./exec/auto/datarouter/datarouter_exec.go
GOOS=$os go build -o ./bin/auto/$os/test-lwm2m$EXTENSION ./exec/auto/lwm2m/lwm2m_exec.go
GOOS=$os go build -o ./bin/auto/$os/test-mqtt$EXTENSION ./exec/auto/mqtt/mqtt_exec.go
GOOS=$os go build -o ./bin/auto/$os/test-mqttps$EXTENSION ./exec/auto/mqttps/mqttps_exec.go
GOOS=$os go build -o ./bin/auto/$os/test-tcp$EXTENSION ./exec/auto/tcp/tcp_exec.go
GOOS=$os go build -o ./bin/auto/$os/test-http$EXTENSION ./exec/auto/http/http_exec.go
GOOS=$os go build -o ./bin/auto/$os/test-all$EXTENSION ./exec/auto/all/all_exec.go
GOOS=$os go build -o ./bin/auto/$os/test-ruleengine$EXTENSION ./exec/auto/ruleengine/ruleengine_exec.go
GOOS=$os go build -o ./bin/auto/$os/test-modbus$EXTENSION ./exec/auto/modbusRTU/modbusRTU_exec.go
done


# go build -o ./bin/auto/test-datarouter ./exec/auto/datarouter/datarouter_exec.go
# go build -o ./bin/auto/test-lwm2m ./exec/auto/lwm2m/lwm2m_exec.go
# go build -o ./bin/auto/test-mqtt ./exec/auto/mqtt/mqtt_exec.go
# go build -o ./bin/auto/test-mqttps ./exec/auto/mqttps/mqttps_exec.go
# go build -o ./bin/auto/test-tcp ./exec/auto/tcp/tcp_exec.go
# go build -o ./bin/auto/test-http ./exec/auto/http/http_exec.go
# go build -o ./bin/auto/test-all ./exec/auto/all/all_exec.go
# go build -o ./bin/auto/test-ruleengine ./exec/auto/ruleengine/ruleengine_exec.go
# go build -o ./bin/auto/test-modbus ./exec/auto/modbusRTU/modbusRTU_exec.go