.PHONY: all get build install uninstall test

get:
	go get -v ./...

build:
	GOARCH=amd64 GOOS=linux go build -o bin/govnocloud2-linux-amd64 cmd/govnocloud2/*.go
	GOARCH=arm64 GOOS=darwin go build -o bin/govnocloud2-darwin-arm64 cmd/govnocloud2/*.go

buildmac:
	GOARCH=arm64 GOOS=darwin go build -o bin/govnocloud2-darwin-arm64 cmd/govnocloud2/*.go

install:
	sudo bin/govnocloud2-linux-amd64 --master 10.0.0.1 --ips 10.0.0.2,10.0.0.3 --macs f0:de:f1:67:8c:92,3c:97:0e:71:77:ab install

uninstall:
	sudo bin/govnocloud2-linux-amd64 --master 10.0.0.1 --ips 10.0.0.2,10.0.0.3 uninstall

test:
	go test -v ./...

wol:
	bin/govnocloud2-linux-amd64 tool wol --macs f0:de:f1:67:8c:92,3c:97:0e:71:77:ab --iprange 10.0.0.255
	sleep 5

wolmac:
	bin/govnocloud2-darwin-arm64 tool wol --macs f0:de:f1:67:8c:92,3c:97:0e:71:77:ab --iprange 10.0.0.255	

suspend:
	bin/govnocloud2-linux-amd64 tool suspend --ips 10.0.0.2,10.0.0.3

logs:
	journalctl _SYSTEMD_INVOCATION_ID=`systemctl show -p InvocationID --value govnocloud2.service`

deploy:
	make get buildmac
	make wolmac
	make uninstall
	make install
	make test
	-make logs
	-make suspend

all: get build