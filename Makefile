NAME := spiry
BUILDDIR := build

BUILD_DATE := $(shell date '+%s')
GIT_COMMIT := $(shell git rev-parse --short HEAD)
VERSION := $(shell git describe --always --tags --dirty --first-parent)

UPX := $(shell command -v upx)

.DEFAULT_TARGET := build
.PHONY: build compress

$(BUILDDIR)/$(NAME): export CGO_ENABLED = 0
$(BUILDDIR)/$(NAME):
	set | grep -E '^(CGO_|GOARCH|GOOS|GOPATH|GOROOT)' \
	&& go build \
		-a \
		-buildmode=pie \
		-ldflags "-s -w -X main.versionNumber=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X 'main.buildDate=$(BUILD_DATE)'" \
		-o $(BUILDDIR)/$(NAME) \
		-trimpath \
		./cmd/...

build: clean $(BUILDDIR)/$(NAME)

compress: $(BUILDDIR)/$(NAME)
ifdef UPX
	upx -9 --keep --no-progress $(BUILDDIR)/$(NAME) && mv $(BUILDDIR)/$(NAME).~ $(BUILDDIR)/$(NAME).orig
else
	@echo command "upx" not found, cannot compress binary >&2
	@exit 1
endif

clean:
	$(RM) -v $(BUILDDIR)/$(NAME) $(BUILDDIR)/$(NAME).orig

cleaner: clean
	$(RM) -rv vendor go.sum

cleanest: cleaner
	git clean -fdx

rebuild: clean build
