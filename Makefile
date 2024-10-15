.PHONY: all get build install uninstall test

get:
	go get -v ./...

build:
	GOARCH=amd64 GOOS=linux go build -o bin/govnocloud2-linux-amd64 cmd/govnocloud2/*.go

buildmac:
	GOARCH=arm64 GOOS=darwin go build -o bin/govnocloud2-darwin-arm64 cmd/govnocloud2/*.go

install:
	bin/govnocloud2-linux-amd64 --master master.govno.cloud --workersips 10.0.0.1,10.0.0.2,10.0.0.3 --workersmacs f0:de:f1:67:8c:92,3c:97:0e:71:77:ab,28:d2:44:ed:85:f9 install

uninstall:
	bin/govnocloud2-linux-amd64 --master master.govno.cloud --workersips 10.0.0.1,10.0.0.2,10.0.0.3 uninstall

wol:
	bash test/wol.sh

test:
	go test -v ./...

logs:
	journalctl _SYSTEMD_INVOCATION_ID=`systemctl show -p InvocationID --value govnocloud2.service`

all: get build