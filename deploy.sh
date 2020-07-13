#!/bin/bash

echo "change work directory to" $GOPATH/src/github.com/SasukeBo/pmes-data-center ...
cd $GOPATH/src/github.com/SasukeBo/pmes-data-center

echo "start service ..."
go run main.go
