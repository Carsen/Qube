.PHONY: dev build generate install image release profile bench test clean setup

CGO_ENABLED=0
VERSION=$(shell git describe --abbrev=0 --tags 2>/dev/null || echo "$VERSION")
COMMIT=$(shell git rev-parse --short HEAD || echo "$COMMIT")
BUILD=$(shell git show -s --pretty=format:%cI)
GOCMD=go

DESTDIR=/usr/local/bin

ifeq ($(LOCAL), 1)
IMAGE := r.mills.io/prologic/bitcask
TAG := dev
else
ifeq ($(BRANCH), main)
IMAGE := prologic/bitcask
TAG := latest
else
IMAGE := prologic/bitcask
TAG := dev
endif
endif

all: dev

dev: build
	@./bitcask --version
	@./bitcaskd --version

build: generate
	@$(GOCMD) build \
		-tags "netgo static_build" -installsuffix netgo \
		-ldflags "-w -X $(shell go list)/internal.Version=$(VERSION) -X $(shell go list)/internal.Commit=$(COMMIT) -X $(shell go list)/internal.Build=$(BUILD)" \
		./cmd/bitcask/...
	@$(GOCMD) build \
		-tags "netgo static_build" -installsuffix netgo \
		-ldflags "-w -X $(shell go list)/internal.Version=$(VERSION) -X $(shell go list)/internal.Commit=$(COMMIT) -X $(shell go list)/internal.Build=$(BUILD)" \
		./cmd/bitcaskd/...

generate:
	@$(GOCMD) generate $(shell go list)/...

install: build
	@install -D -m 755 bitcask $(DESTDIR)/bitcask
	@install -D -m 755 bitcaskd $(DESTDIR)/bitcaskd

ifeq ($(PUBLISH), 1)
image: generate ## Build the Docker image
	@docker buildx build \
		--build-arg VERSION="$(VERSION)" \
		--build-arg COMMIT="$(COMMIT)" \
		--build-arg BUILD="$(BUILD)" \
		--platform linux/amd64,linux/arm64 --push -t $(IMAGE):$(TAG) .
else
image: generate
	@docker build  \
		--build-arg VERSION="$(VERSION)" \
		--build-arg COMMIT="$(COMMIT)" \
		--build-arg BUILD="$(BUILD)" \
		-t $(IMAGE):$(TAG) .
endif

release:
	@./tools/release.sh

profile: build
	@$(GOCMD) test -cpuprofile cpu.prof -memprofile mem.prof -v -bench .

bench: build
	@$(GOCMD) test -v -run=XXX -benchmem -bench=. .

test: build
	@$(GOCMD) test -v \
		-cover -coverprofile=coverage.out -covermode=atomic \
		-coverpkg=$(shell go list) \
		-race \
		./...

setup:

clean:
	@git clean -f -d -X
