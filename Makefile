.PHONY: all get build install uninstall test

get:
	go get -v ./...

build:
	GOARCH=amd64 GOOS=linux go build -o bin/govnocloud2-linux-amd64 cmd/govnocloud2/*.go

buildmac:
	GOARCH=arm64 GOOS=darwin go build -o bin/govnocloud2-darwin-arm64 cmd/govnocloud2/*.go

install:
	bin/govnocloud2-linux-amd64 -m master.govno.cloud -w node0.govno.cloud,node2.govno.cloud,node1.govno.cloud install

uninstall:
	bin/govnocloud2-linux-amd64 -m master.govno.cloud -w node0.govno.cloud,node1.govno.cloud,node2.govno.cloud uninstall

wol:
	bash test/wol.sh

test:
	go test -v ./...

logs:
	ssh root@master.govno.cloud "journalctl -u govnocloud2.service"

all: get build