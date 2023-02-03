git=$(subst -,.,$(shell git describe --always --match=NeVeRmAtCh --dirty 2>/dev/null || git rev-parse --short HEAD 2>/dev/null))

ldflags=-X=github.com/ipfs-force-community/venus-tool/version.CurrentCommit=+git.$(git)
ifneq ($(strip $(LDFLAGS)),)
	ldflags+=-extldflags=$(LDFLAGS)
endif

GOFLAGS+=-ldflags="$(ldflags)"

build:
	rm -rf venus-tool
	go build $(GOFLAGS) -o venus-tool ./cmd


gen:
	@go generate ./...

lint:
	@golangci-lint run

test:
	@go test -race ./...

dev-init:
	ln -s ../../.githooks/pre-commit .git/hooks/pre-commit
	ln -s ../../.githooks/pre-push .git/hooks/pre-push
