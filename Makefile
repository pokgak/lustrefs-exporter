VERSION ?= 0.1.0
BINARY  := lustrefs-exporter
DIST    := dist

.PHONY: all build test deb clean

all: build

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -ldflags="-s -w" -o $(DIST)/$(BINARY) .

test:
	go test -mod=vendor ./...

deb: build
	VERSION=$(VERSION) nfpm pkg --config nfpm.yaml --packager deb --target $(DIST)/

clean:
	rm -rf $(DIST)
