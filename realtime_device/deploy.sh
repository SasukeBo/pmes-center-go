#!/bin/bash

echo "change work directory to" $GOPATH/src/github.com/SasukeBo/pmes-data-center/realtime_device/ ...
cd $GOPATH/src/github.com/SasukeBo/pmes-data-center/realtime_device/

echo "start service ..."
go run server.go
