FROM blackouta/librdkafka-golang1141-alpine
WORKDIR /oes-acceptance-test
COPY . .
RUN go build -o test-datarouter ./exec/auto/datarouter/datarouter_exec.go
RUN go build -o test-lwm2m ./exec/auto/lwm2m/lwm2m_exec.go
RUN go build -o test-mqtt ./exec/auto/mqtt/mqtt_exec.go
RUN go build -o test-mqttps ./exec/auto/mqttps/mqttps_exec.go
RUN go build -o test-tcp ./exec/auto/tcp/tcp_exec.go
RUN go build -o test-http ./exec/auto/http/http_exec.go
RUN go build -o test-all ./exec/auto/all/all_exec.go
RUN go build -o test-ruleengine ./exec/auto/ruleengine/ruleengine_exec.go
RUN go build -o test-modbus ./exec/auto/modbusRTU/modbusRTU_exec.go

RUN chmod +x ./scripts/entrypoint.sh
CMD [ "./scripts/entrypoint.sh" ] 