#!/bin/bash
GOOS=linux go build \
-o ./bin/manual/oes-sim-linux  \
-ldflags '-s -w' \
./exec/manual/main.go

GOOS=windows go build \
-o ./bin/manual/oes-sim-windows.exe  \
-ldflags '-s -w' \
./exec/manual/main.go

GOOS=darwin go build \
-o ./bin/manual/oes-sim-mac  \
-ldflags '-s -w' \
./exec/manual/main.go