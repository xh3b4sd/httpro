PROJECT := httpro
BUILD_PATH := $(shell pwd)/.gobuild
PROJECT_PATH := $(BUILD_PATH)/src/github.com/zyndiecate
SOURCE := $(shell find . -name '*.go')
TEMPLATES := $(shell find . -name '*.tmpl')

ifndef GOOS
	GOOS := $(shell go env GOOS)
endif
ifndef GOARCH
	GOARCH := $(shell go env GOARCH)
endif

.PHONY: clean tests test get-deps update-deps fmt deps

all: get-deps $(PROJECT)

clean:
	@rm -rf $(BUILD_PATH) $(PROJECT)

get-deps: .gobuild

deps:
	@${MAKE} -B -s .gobuild

.gobuild:
	@mkdir -p $(PROJECT_PATH)
	@rm -f $(PROJECT_PATH)/$(PROJECT) && cd "$(PROJECT_PATH)" && ln -s ../../../.. $(PROJECT)

	@#
	@# Fetch private packages first (so `go get` skips them later)

	@#
	@# Fetch public dependencies via `go get`
	@GOPATH=$(BUILD_PATH) go get -d -v github.com/zyndiecate/$(PROJECT)

	@#
	@# Build test packages (we only want those two, so we use `-d` in go get)
	@GOPATH=$(BUILD_PATH) go get -d -v github.com/onsi/gomega
	@GOPATH=$(BUILD_PATH) go get -d -v github.com/onsi/ginkgo

$(PROJECT): VERSION $(SOURCE)
	@echo Building for $(GOOS)/$(GOARCH)
	@docker run \
	    --rm \
	    -v $(shell pwd):/usr/code \
	    -e GOPATH=/usr/code/.gobuild \
	    -e GOOS=$(GOOS) \
	    -e GOARCH=$(GOARCH) \
	    -w /usr/code \
	    golang:1.3.1-cross \
	    go build -a

tests:
	@make test test=./...

test:
	@if test "$(test)" = "" ; then \
		echo "missing test parameter, that is, path to test folder e.g. './middleware/v1/'."; \
		exit 1; \
	fi
	@docker run \
	    --rm \
	    -v $(shell pwd):/usr/code \
	    -e GOPATH=/usr/code/.gobuild \
	    -w /usr/code \
	    golang:1.3.1-cross \
	    go test -v $(test)

fmt:
	@gofmt -l -w .
