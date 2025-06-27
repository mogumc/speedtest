# ========== 应用配置 ===========
APP_NAME := speedtest-gd
VERSION := 1.0.0

BUILD_TIME := $(shell date +"%Y-%m-%dT%H:%M:%S")
GIT_COMMIT := $(shell git rev-parse --short=7 HEAD 2>/dev/null || echo dev)

PKG_OUTPUT_DIR := build
LDFLAGS := -ldflags "-X 'main.AppName=$(APP_NAME)' -X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)' -X 'main.GitHash=$(GIT_COMMIT)'"

WAILS ?= wails

# ========== 构建输出名称 =========
WINDOWS_AMD64 := $(PKG_OUTPUT_DIR)/$(APP_NAME)-windows-amd64.exe
WINDOWS_ARM64 := $(PKG_OUTPUT_DIR)/$(APP_NAME)-windows-arm64.exe
LINUX_AMD64 := $(PKG_OUTPUT_DIR)/$(APP_NAME)-linux-amd64
LINUX_ARM64 := $(PKG_OUTPUT_DIR)/$(APP_NAME)-linux-arm64
MAC_AMD64 := $(PKG_OUTPUT_DIR)/$(APP_NAME)-macos-amd64
MAC_ARM64 := $(PKG_OUTPUT_DIR)/$(APP_NAME)-macos-arm64

# ========== Targets ============
.PHONY: all clean build package package_all release

all: clean build package_all

clean:
	rm -rf $(PKG_OUTPUT_DIR)
	mkdir -p $(PKG_OUTPUT_DIR)

build: windows_amd64 windows_arm64 linux_amd64 linux_arm64 mac_amd64 mac_arm64

# ========== 各平台构建 ==========
windows_amd64:
	GOOS=windows GOARCH=amd64 $(WAILS) build -platform=windows/amd64 $(LDFLAGS) -o $(WINDOWS_AMD64)

windows_arm64:
	GOOS=windows GOARCH=arm64 $(WAILS) build -platform=windows/arm64 $(LDFLAGS) -o $(WINDOWS_ARM64)

linux_amd64:
	GOOS=linux GOARCH=amd64 $(WAILS) build -platform=linux/amd64 $(LDFLAGS) -o $(LINUX_AMD64)

linux_arm64:
	GOOS=linux GOARCH=arm64 $(WAILS) build -platform=linux/arm64 $(LDFLAGS) -o $(LINUX_ARM64)

mac_amd64:
	GOOS=darwin GOARCH=amd64 $(WAILS) build -platform=darwin/amd64 $(LDFLAGS) -o $(MAC_AMD64)

mac_arm64:
	GOOS=darwin GOARCH=arm64 $(WAILS) build -platform=darwin/arm64 $(LDFLAGS) -o $(MAC_ARM64)

# ========== 打包任务 =============

package_all: package_windows package_linux package_macos

package_windows: windows_amd64 windows_arm64
	zip -r $(PKG_OUTPUT_DIR)/$(APP_NAME)-$(VERSION)-windows-amd64.zip $(WINDOWS_AMD64)
	zip -r $(PKG_OUTPUT_DIR)/$(APP_NAME)-$(VERSION)-windows-arm64.zip $(WINDOWS_ARM64)

package_linux: linux_amd64 linux_arm64
	tar -czvf $(PKG_OUTPUT_DIR)/$(APP_NAME)-$(VERSION)-linux-amd64.tar.gz -C $(PKG_OUTPUT_DIR) $(notdir $(LINUX_AMD64))
	tar -czvf $(PKG_OUTPUT_DIR)/$(APP_NAME)-$(VERSION)-linux-arm64.tar.gz -C $(PKG_OUTPUT_DIR) $(notdir $(LINUX_ARM64))

package_macos: mac_amd64 mac_arm64
	hdiutil verify $(MAC_AMD64)
	hdiutil mount $(MAC_AMD64)
	hdiutil create -volname "$(APP_NAME)" -srcfolder $(MAC_AMD64) -ov -format UDZO $(PKG_OUTPUT_DIR)/$(APP_NAME)-$(VERSION)-macos-universal.dmg

# ========== 发布用途 =============
release: package_all
	@echo "Build complete! You can release the following files:"
	ls -l $(PKG_OUTPUT_DIR)/*
