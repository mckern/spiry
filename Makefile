NAME := spiry
BUILDDIR := build

BUILD_DATE := $(shell date '+%s')
GIT_COMMIT := $(shell git rev-parse --short HEAD)
VERSION := $(shell git describe --always --tags --dirty --first-parent)

.DEFAULT_TARGET := build
.PHONY: build compress

$(BUILDDIR)/$(NAME): export CGO_ENABLED = 0
$(BUILDDIR)/$(NAME):
	go build \
	  -o $(BUILDDIR)/$(NAME) \
	  -ldflags "-s -w -X main.versionNumber=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X 'main.buildDate=$(BUILD_DATE)'" \
		./...

build: clean $(BUILDDIR)/$(NAME)

rebuild: clean build

compress: $(BUILDDIR)/$(NAME)
	upx -6 --keep --no-progress -qq $(BUILDDIR)/$(NAME) && mv $(BUILDDIR)/$(NAME).~ $(BUILDDIR)/$(NAME).orig

clean:
	$(RM) $(BUILDDIR)/$(NAME)

cleaner: clean
	$(RM) -rv vendor
	$(RM) -rv go.sum

cleanest: clean
	git clean -ffdx
