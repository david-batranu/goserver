.PHONY: build-linux-static

build-linux-static:
	@echo "Building static Linux AMD64 binary with musl..."
	CC=/usr/bin/musl-gcc CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags '-linkmode external -extldflags "-static"' .
