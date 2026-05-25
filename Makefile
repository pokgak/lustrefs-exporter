BINARY  := lustrefs-exporter
DIST    := dist

.PHONY: all build test snapshot clean

all: build

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -ldflags="-s -w" -o $(DIST)/$(BINARY) .

test:
	go test -mod=vendor ./...

snapshot:
	goreleaser release --snapshot --clean

clean:
	rm -rf $(DIST)
