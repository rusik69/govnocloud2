.PHONY: all get build

get:
	go get -v ./...

build:
	GOARCH=amd64 GOOS=linux go build -o bin/govnocloud2-linux-amd64 cmd/govnocloud2/main.go
	chmod +x bin/*

install:
	bin/govnocloud2-linux-amd64 install --master master.govno.cloud --nodes node0.govno.cloud,node1.govno.cloud,node2.govno.cloud

uninstall:
	bin/govnocloud2-linux-amd64 uninstall --master master.govno.cloud --nodes node0.govno.cloud,node1.govno.cloud,node2.govno.cloud

all: get build