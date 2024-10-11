.PHONY: install build

APP_NAME ?= gover
APP_VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")

PACKAGE_NAME ?= github.com/kam9lo/gover

BUILD_LD_FLAGS ?= -ldflags="-X '$(PACKAGE_NAME)/app.Name=$(APP_NAME)' -X '$(PACKAGE_NAME)/app.Version=$(APP_VERSION)'"

GOOS ?= linux
GOARCH ?= amd64

setup:
	@cp ./hooks/prepare-commit-msg .git/hooks/
	@chmod +x .git/hooks/prepare-commit-msg

install:
	@echo "Installing $(APP_NAME) $(APP_VERSION)"
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) \
	go install $(BUILD_LD_FLAGS)


build:
	@echo "Building $(APP_NAME) $(APP_VERSION)"
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) \
	go build $(BUILD_LD_FLAGS) -o ./bin/$(APP_NAME) \

