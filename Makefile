SHELL := /bin/bash

NAME = spiry
BUILDDIR := build

GIT := $(shell command -v git)
GO := $(shell command -v go)

BUILD_DATE := $(shell date '+%s')
GIT_COMMIT := $(shell $(GIT) rev-parse --short HEAD)
VERSION := $(shell $(GIT) describe --always --tags --dirty --first-parent)

LDFLAGS := -s -w -X main.versionNumber=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X 'main.buildDate=$(BUILD_DATE)'

GOVER := 1.21
CGO_ENABLED := 0

DOCKER := $(shell command -v docker)
LINTER := $(shell command -v golangci-lint)
UPX := $(shell command -v upx)

.DEFAULT_TARGET := $(BUILDDIR)/$(NAME)
.PHONY: build compress test lint vendor

$(BUILDDIR)/$(NAME):
	$(GO) build \
		-a \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILDDIR)/$(NAME) \
		-trimpath \
		./cmd/spiry

build: $(BUILDDIR)/$(NAME)

compress: $(BUILDDIR)/$(NAME)
ifdef UPX
	$(UPX) -9 --keep --no-progress $(BUILDDIR)/$(NAME) && mv $(BUILDDIR)/$(NAME).~ $(BUILDDIR)/$(NAME).orig
else
	@echo command "upx" not found, cannot compress binary >&2
	@exit 1
endif

package: compress
	tar cfv $(BUILDDIR)/$(NAME)-$(VERSION).tar $(BUILDDIR)/$(NAME)
	ls -hl $(BUILDDIR)

test:
	$(GO) test -v ./internal/...

containerized-tests: clean vendor
ifdef DOCKER
	$(DOCKER) run \
		--mount "type=bind,source="${PWD}",target=/app" \
		--env="CGO_ENABLED=0" \
		--workdir="/app" \
		--rm \
		golang:$(GOVER)-alpine \
		/bin/sh -c 'go test -v ./internal/...'
else
	@echo command "docker" not found, cannot run isolated privileged tests inside Docker container
	@exit 1
endif

lint:
ifdef LINTER
	@$(LINTER) run ./...
else
	@echo command "golangci-lint" not found, cannot lint codebase
	@exit 1
endif

tidy:
	@$(GO) mod tidy

vendor:
	@$(GO) mod vendor

clean:
	@$(RM) -v $(BUILDDIR)/$(NAME) $(BUILDDIR)/$(NAME).orig

cleaner: clean
	@$(RM) -rv vendor go.sum
	@$(GO) clean -cache -modcache

cleanest: cleaner
	@$(GIT) clean -fdx

rebuild: clean build
