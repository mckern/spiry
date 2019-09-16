NAME := spiry
BUILDDIR := build

BUILD_DATE := $(shell date '+%s')
GIT_COMMIT := $(shell git rev-parse --short HEAD)
VERSION := $(shell git describe --always --tags --dirty --first-parent)

.DEFAULT_TARGET := build
.PHONY: build

build: export CGO_ENABLED = 0
build: clean
	go build \
	  -o $(BUILDDIR)/$(NAME) \
	  -ldflags "-X main.versionNumber=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X 'main.buildDate=$(BUILD_DATE)'" \
		./...

clean:
	$(RM) $(BUILDDIR)/$(NAME)

cleanest: clean
	$(RM) -rv vendor
	$(RM) -rv go.sum
