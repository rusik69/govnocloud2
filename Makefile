.PHONY: all get build install uninstall test update-deps run-debug run-debug-mac

get:
	go get -v ./...

build:
	GOARCH=amd64 GOOS=linux go build -o bin/govnocloud2-linux-amd64 cmd/govnocloud2/*.go

buildmac:
	GOARCH=arm64 GOOS=darwin go build -o bin/govnocloud2-darwin-arm64 cmd/govnocloud2/*.go

install:
	DEBUG=true bin/govnocloud2-linux-amd64 --master 10.0.0.1 --ips 10.0.0.2,10.0.0.3 --macs f0:de:f1:67:8c:92,3c:97:0e:71:77:ab install

installmac:
	DEBUG=true bin/govnocloud2-darwin-arm64 --master 192.168.1.64 --ips 10.0.0.2,10.0.0.3 --macs f0:de:f1:67:8c:92,3c:97:0e:71:77:ab install

uninstall:
	bin/govnocloud2-linux-amd64 --master 10.0.0.1 --ips 10.0.0.2,10.0.0.3 uninstall

uninstallmac:
	bin/govnocloud2-darwin-arm64 --master 192.168.1.64 --ips 10.0.0.2,10.0.0.3 uninstall

test:
	go test -v ./...

wol:
	bin/govnocloud2-linux-amd64 tool wol --macs f0:de:f1:67:8c:92,3c:97:0e:71:77:ab --iprange 10.0.0.255 --master 10.0.0.1
	sleep 5

wolmac:
	bin/govnocloud2-darwin-arm64 --macs f0:de:f1:67:8c:92,3c:97:0e:71:77:ab --iprange 10.0.0.255 --master 192.168.1.64 tool wol

suspend:
	bin/govnocloud2-linux-amd64 tool suspend --ips 10.0.0.2,10.0.0.3 --master 10.0.0.1

suspendmac:
	bin/govnocloud2-darwin-arm64 tool suspend --ips 10.0.0.2,10.0.0.3 --master 192.168.1.64

logs:
	sudo journalctl _SYSTEMD_INVOCATION_ID=`systemctl show -p InvocationID --value govnocloud2.service`

deploymac:
	make get buildmac wolmac uninstallmac installmac test
	-make logs
	-make suspend

update-deps:
	go get -u ./...
	go mod tidy
	go mod verify

run-web:
	DEBUG=true bin/govnocloud2-linux-amd64 web 

run-web-mac:
	DEBUG=true bin/govnocloud2-darwin-arm64 web --webpath pkg/web/templates

create-vm-mac:
	./bin/govnocloud2-darwin-arm64 client vms create test-vm ubuntu24 small test --host 192.168.1.83

stop-vm-mac:
	./bin/govnocloud2-darwin-arm64 client vms stop test-vm test --host 192.168.1.83

start-vm-mac:
	./bin/govnocloud2-darwin-arm64 client vms start test-vm test --host 192.168.1.83
exec:
	chmod +x bin/*

all: get build