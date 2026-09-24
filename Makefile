PREFIX ?= /usr/local
GIT_VERSION = $(shell git describe --abbrev=0)

lure:
	CGO_ENABLED=0 go build -ldflags="-X 'github.com/sintan1729/lure/internal/config.Version=$(GIT_VERSION)'" -o "target/lure"

build-%:
	GOOS=linux GOARCH=$* CGO_ENABLED=0 go build -ldflags="-X 'github.com/sintan1729/lure/internal/config.Version=$(GIT_VERSION)'" -o "target/$*/lure"

GH_TOKEN := $(shell cat ~/.config/github_token)
release: build-amd64 build-arm64 build-arm build-386 build-riscv64 gencmp
	for d in amd64 arm64 arm 386 riscv64; do \
		rm -rf target/$$d/misc; \
		mkdir -p target/$$d/misc/completion; \
		mkdir -p target/$$d/misc/man; \
		cp target/{lure.bash,lure.fish,_lure} target/$$d/misc/completion/; \
		cp docs/lure.1 target/$$d/misc/man/; \
		t="lure-v$(GIT_VERSION)-linux-$$d"; \
		rm -rf target/"$$t"; \
		mv target/{$$d,"$$t"}; \
		tar -czf "target/$$t.tar.gz" -C target "$$t"; \
	done;
	gh release create "${GIT_VERSION}" --notes "$$(git-cliff --latest --github-token ${GH_TOKEN})" "target/*.tar.gz"
	$(MAKE) clean

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
