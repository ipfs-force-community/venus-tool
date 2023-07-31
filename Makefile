git=$(subst -,.,$(shell git describe --always --match=NeVeRmAtCh --dirty 2>/dev/null || git rev-parse --short HEAD 2>/dev/null))

ldflags=-X=github.com/ipfs-force-community/venus-tool/version.CurrentCommit=+git.$(git)
ifneq ($(strip $(LDFLAGS)),)
	ldflags+=-extldflags=$(LDFLAGS)
endif

GOFLAGS+=-ldflags="$(ldflags)"

all: dashboard/build venus-tool

venus-tool: deps
	rm -rf venus-tool
	go build $(GOFLAGS) -o venus-tool ./cmd


gen:
	@go generate ./...

lint: deps
	@golangci-lint run

test:
	@go test -race ./...

dev-init:
	ln -s ../../.githooks/pre-commit .git/hooks/pre-commit
	ln -s ../../.githooks/pre-push .git/hooks/pre-push

dashboard/build: dashboard/src
	cd dashboard && yarn install && yarn build

.PHONY: docker
TAG:=test
docker: $(BUILD_DEPS) all
	docker build --build-arg https_proxy=$(BUILD_DOCKER_PROXY) --build-arg BUILD_TARGET=venus-tool -t venus-tool  .
	docker tag venus-tool:latest filvenus/venus-tool:$(TAG)

docker-push: docker
	docker push filvenus/venus-tool:$(TAG)


build_deps:
	mkdir $@

build_deps/.update-modules: build_deps
	git submodule update --init --recursive
	touch $@

FFI_PATH:=extern/filecoin-ffi/
build_deps/.filecoin-install: build_deps $(FFI_PATH)
	make -C $(FFI_PATH)
	@touch $@


deps: build_deps/.update-modules build_deps/.filecoin-install
