#!/usr/bin/env bash

#cd example
#go run main.go -config-dir=./ -venture=ru -env=dev -addr=127.0.0.1:8080 -adm-addr=127.0.0.1:9080 -etcd-endpoints=http://127.0.0.1:4001/

go run example/main.go -config=example/app.ini -venture=ru -env=dev -addr=127.0.0.1:8080 -adm-addr=127.0.0.1:9080 -etcd-endpoints=http://127.0.0.1:4001/