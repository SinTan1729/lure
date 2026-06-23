PREFIX ?= /usr/local
GIT_VERSION = $(shell git describe --abbrev=0)

lure:
	CGO_ENABLED=0 go build -ldflags="-X 'github.com/sintan1729/lure/internal/config.Version=$(GIT_VERSION)'" -o "target/lure"

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-X 'github.com/sintan1729/lure/internal/config.Version=$(GIT_VERSION)'" -o "target/lure-v$(GIT_VERSION)-linux-x86_64"
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-X 'github.com/sintan1729/lure/internal/config.Version=$(GIT_VERSION)'" -o "target/lure-v$(GIT_VERSION)-linux-aarch64"
	GOOS=linux GOARCH=arm CGO_ENABLED=0 go build -ldflags="-X 'github.com/sintan1729/lure/internal/config.Version=$(GIT_VERSION)'" -o "target/lure-v$(GIT_VERSION)-linux-arm"
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -ldflags="-X 'github.com/sintan1729/lure/internal/config.Version=$(GIT_VERSION)'" -o "target/lure-v$(GIT_VERSION)-linux-i386"
	GOOS=linux GOARCH=riscv64 CGO_ENABLED=0 go build -ldflags="-X 'github.com/sintan1729/lure/internal/config.Version=$(GIT_VERSION)'" -o "target/lure-v$(GIT_VERSION)-linux-riscv64"

release: build
	for f in target/lure-*; do \
		if [ "$${f##*.}" = "gz" ]; then \
			continue; \
		fi; \
		tar -czf "$$f.tar.gz" \
			-C target "$$(basename "$$f")" \
			-C ../docs lure.1; \
	done

clean:
	rm -rf target/*

install: lure installmisc
	install -Dm755 lure $(DESTDIR)$(PREFIX)/bin/lure

uninstall:
	rm -f /usr/local/bin/lure

.PHONY: install clean uninstall installmisc lure build release
