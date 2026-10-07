PREFIX ?= /usr/local
GIT_VERSION = $(shell git describe --abbrev=0)

lure:
	CGO_ENABLED=0 go build -ldflags="-X 'github.com/sintan1729/lure/internal/config.Version=$(GIT_VERSION)'" -o "target/lure"

build-all:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-X 'github.com/sintan1729/lure/internal/config.Version=$(GIT_VERSION)'" -o "target/x86_64/lure"
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-X 'github.com/sintan1729/lure/internal/config.Version=$(GIT_VERSION)'" -o "target/arm64/lure"
	GOOS=linux GOARM=5 GOARCH=arm CGO_ENABLED=0 go build -ldflags="-X 'github.com/sintan1729/lure/internal/config.Version=$(GIT_VERSION)'" -o "target/armv5/lure"
	GOOS=linux GOARM=6 GOARCH=arm CGO_ENABLED=0 go build -ldflags="-X 'github.com/sintan1729/lure/internal/config.Version=$(GIT_VERSION)'" -o "target/armv6/lure"
	GOOS=linux GOARM=7 GOARCH=arm CGO_ENABLED=0 go build -ldflags="-X 'github.com/sintan1729/lure/internal/config.Version=$(GIT_VERSION)'" -o "target/armv7/lure"
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -ldflags="-X 'github.com/sintan1729/lure/internal/config.Version=$(GIT_VERSION)'" -o "target/i686/lure"
	GOOS=linux GOARCH=riscv64 CGO_ENABLED=0 go build -ldflags="-X 'github.com/sintan1729/lure/internal/config.Version=$(GIT_VERSION)'" -o "target/riscv64/lure"

GH_TOKEN := $(shell cat ~/.config/github_token)
release: build-all gencmp
	for d in x86_64 arm64 armv5 armv6 armv7 i686 riscv64; do \
		rm -rf "target/$$d/misc"; \
		mkdir -p "target/$$d/misc/completion"; \
		mkdir -p "target/$$d/misc/man"; \
		cp target/lure.bash target/lure.fish target/_lure "target/$$d/misc/completion/"; \
		cp docs/lure.1 "target/$$d/misc/man/"; \
		t="lure-v$(GIT_VERSION)-linux-$$d"; \
		rm -rf "target/$$t"; \
		mv "target/$$d" "target/$$t"; \
		tar -czf "target/$$t.tar.gz" -C target "$$t"; \
	done
# gh release create "${GIT_VERSION}" --notes "$$(git-cliff --latest --github-token ${GH_TOKEN})" "target/*.tar.gz"
# $(MAKE) clean

gencmp: lure
	target/lure completion bash > target/lure.bash
	target/lure completion fish > target/lure.fish
	target/lure completion zsh > target/_lure

clean:
	rm -rf target

install: lure installmisc
	install -Dm755 lure $(DESTDIR)$(PREFIX)/bin/lure

uninstall:
	rm -f /usr/local/bin/lure

.PHONY: install clean uninstall installmisc lure build release
