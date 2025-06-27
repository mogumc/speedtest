# ========== 应用配置 ===========
APP_NAME := speedtest-gd
VERSION := 1.0.0

BUILD_TIME := $(shell date +"%Y-%m-%dT%H:%M:%S")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)

PKG_OUTPUT_DIR := build
LDFLAGS := -ldflags "-X 'main.AppName=$(APP_NAME)' -X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)' -X 'main.GitHash=$(GIT_COMMIT)'"

WAILS ?= wails

# 根据系统决定平台
ifeq ($(OS),Windows_NT)
	GOOS := windows
	ifeq ($(PROCESSOR_ARCHITECTURE),AMD64)
		GOARCH := amd64
	else
		GOARCH := $(PROCESSOR_ARCHITECTURE)
	endif
	EXEC_SUFFIX := .exe
else
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S), Linux)
		GOOS := linux
	endif
	ifeq ($(UNAME_S), Darwin)
		GOOS := darwin
	endif
	UNAME_M := $(shell uname -m)
	ifeq ($(UNAME_M), x86_64)
		GOARCH := amd64
	endif
	ifeq ($(UNAME_M), aarch64)
		GOARCH := arm64
	endif
	ifeq ($(UNAME_M), arm64)
		GOARCH := arm64
	endif
	EXEC_SUFFIX :=
endif

TARGET_NAME := $(APP_NAME)-$(GOOS)-$(GOARCH)$(EXEC_SUFFIX)
OUTPUT_EXEC := $(PKG_OUTPUT_DIR)/$(TARGET_NAME)

# ========== Targets ============
.PHONY: all clean build package release

all: clean build package

clean:
	rm -rf $(PKG_OUTPUT_DIR)
	mkdir -p $(PKG_OUTPUT_DIR)

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) \
	$(WAILS) build $(LDFLAGS) -o $(OUTPUT_EXEC)

package:
ifeq ($(GOOS), windows)
	zip -r $(PKG_OUTPUT_DIR)/$(APP_NAME)-$(VERSION)-$(GOOS)-$(GOARCH).zip $(OUTPUT_EXEC)
else ifeq ($(GOOS), darwin)
	@echo "Packaging macOS app bundle..."
	mkdir -p $(PKG_OUTPUT_DIR)/$(TARGET_NAME).dmgtmp/$(APP_NAME).app/Contents/MacOS/
	cp -R $(OUTPUT_EXEC).app $(PKG_OUTPUT_DIR)/$(TARGET_NAME).app
	cp $(OUTPUT_EXEC) $(PKG_OUTPUT_DIR)/$(TARGET_NAME).app/Contents/MacOS/
	chmod +x $(PKG_OUTPUT_DIR)/$(TARGET_NAME).app/Contents/MacOS/$(APP_NAME)
	hdiutil create -size 500m -fs HFS+ -volname "$(APP_NAME)" -srcfolder $(PKG_OUTPUT_DIR)/$(TARGET_NAME).app $(PKG_OUTPUT_DIR)/$(APP_NAME)-$(VERSION)-$(GOOS)-$(GOARCH).dmg
	rm -rf $(PKG_OUTPUT_DIR)/$(TARGET_NAME).dmgtmp
else # linux
	tar -czvf $(PKG_OUTPUT_DIR)/$(APP_NAME)-$(VERSION)-$(GOOS)-$(GOARCH).tar.gz -C $(PKG_OUTPUT_DIR) $(TARGET_NAME)
endif

release: package
	@echo "Build complete! You can release the following files:"
	ls -l $(PKG_OUTPUT_DIR)/*
