BINARY := bin/musiccli
GO_CMD := ./cmd/musiccli

.PHONY: all build web dev clean

all: build

web:
	cd web && npm install && npm run build

build: web
	go build -o $(BINARY) $(GO_CMD)

dev-go:
	go run $(GO_CMD) serve

dev-web:
	cd web && npm run dev

clean:
	rm -rf $(BINARY) internal/web/dist web/node_modules

test:
	go test ./...
