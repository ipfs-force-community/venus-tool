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

.PHONY: docker
TAG:=test
docker: $(BUILD_DEPS)
ifdef DOCKERFILE
	cp $(DOCKERFILE) ./dockerfile
else
	curl -o dockerfile https://raw.githubusercontent.com/filecoin-project/venus-docs/master/script/docker/dockerfile
endif
	docker build --build-arg https_proxy=$(BUILD_DOCKER_PROXY) --build-arg BUILD_TARGET=venus-tool -t venus-tool  .
	docker tag venus-tool:latest filvenus/venus-tool:$(TAG)

ifdef PRIVATE_REGISTRY
	docker tag venus-tool:latest $(PRIVATE_REGISTRY)/filvenus/venus-tool:$(TAG)
endif


docker-push: docker
	docker push $(PRIVATE_REGISTRY)/filvenus/venus-tool:$(TAG)
