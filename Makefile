.PHONY: setup build fmt

setup:
	export GONOSUMDB=v2ray.com/core,github.com/v2ray/v2ray-core
	go mod download

build:
	go build -o v2scar cmd/main.go

fmt:
	gofmt -d -s -w .
