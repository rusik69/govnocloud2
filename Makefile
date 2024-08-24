.PHONY: all

get:
	go get -v ./...

build:
	GOARCH=amd64 GOOS=linux go build -o bin/govnocloud2-linux-amd64 cmd/govnocloud2/main.go

buildmac:
	GOARCH=arm64 GOOS=darwin go build -o bin/govnocloud2-darwin-arm64 cmd/govnocloud2/main.go

install:
	bin/govnocloud2-linux-amd64 -m master.govno.cloud -w node0.govno.cloud,node1.govno.cloud,node2.govno.cloud install

uninstall:
	bin/govnocloud2-linux-amd64 -m master.govno.cloud -w node0.govno.cloud,node1.govno.cloud,node2.govno.cloud uninstall

wol:
	bash test/wol.sh

all: get build