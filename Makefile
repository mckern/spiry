NAME := spiry
BUILDDIR := build

BUILD_DATE := $(shell date '+%s')
GIT_COMMIT := $(shell git rev-parse --short HEAD)
VERSION := $(shell git describe --always --tags --dirty --first-parent)

.DEFAULT_TARGET := build
.PHONY: build compress

$(BUILDDIR)/$(NAME):
	go build \
	  -o $(BUILDDIR)/$(NAME) \
	  -ldflags "-s -w -X main.versionNumber=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X 'main.buildDate=$(BUILD_DATE)'" \
		./...

build: export CGO_ENABLED = 0
build: clean $(BUILDDIR)/$(NAME)

compress: $(BUILDDIR)/$(NAME)
	upx -9 --keep --no-progress $(BUILDDIR)/$(NAME) && mv $(BUILDDIR)/$(NAME).~ $(BUILDDIR)/$(NAME).orig

clean:
	$(RM) $(BUILDDIR)/$(NAME)

cleanest: clean
	$(RM) -rv vendor
	$(RM) -rv go.sum
	git clean -ffdx
