#!/bin/bash

echo "change work directory to" $GOPATH/src/github.com/SasukeBo/ftpviewer ...
cd $GOPATH/src/github.com/SasukeBo/ftpviewer

echo "start service ..."
go run main.go
