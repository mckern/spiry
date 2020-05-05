NAME := spiry
BUILDDIR := build

BUILD_DATE := $(shell date '+%s')
GIT_COMMIT := $(shell git rev-parse --short HEAD)
VERSION := $(shell git describe --always --tags --dirty --first-parent)

GIT := $(shell command -v git)
GO := $(shell command -v go)
LINTER := $(shell command -v golangci-lint)
UPX := $(shell command -v upx)

.DEFAULT_TARGET := build
.PHONY: build compress

$(BUILDDIR)/$(NAME): lint
$(BUILDDIR)/$(NAME): export CGO_ENABLED = 0
$(BUILDDIR)/$(NAME):
	set | grep -E '^(CGO_|GOARCH|GOOS|GOPATH|GOROOT)' \
	&& $(GO) build \
		-a \
		-mod=vendor \
		-trimpath \
		-buildmode=pie \
		-ldflags "-s -w -X main.versionNumber=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X 'main.buildDate=$(BUILD_DATE)'" \
		-o $(BUILDDIR)/$(NAME) \
		-trimpath \
		./cmd/...

build: clean $(BUILDDIR)/$(NAME)

compress: $(BUILDDIR)/$(NAME)
ifdef UPX
	$(UPX) -9 --keep --no-progress $(BUILDDIR)/$(NAME) && mv $(BUILDDIR)/$(NAME).~ $(BUILDDIR)/$(NAME).orig
else
	@echo command "upx" not found, cannot compress binary >&2
	@exit 1
endif

lint:
	$(LINTER) run --fast

clean:
	$(RM) -v $(BUILDDIR)/$(NAME) $(BUILDDIR)/$(NAME).orig

cleaner: clean
	$(RM) -rv vendor go.sum
	$(GO) clean -cache -modcache

cleanest: cleaner
	$(GIT) clean -fdx

rebuild: clean build
