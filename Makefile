# Audio Loss Checker Makefile
# 支持 Windows, macOS, Linux 跨平台编译

# 项目信息
BINARY_NAME=audio-loss-checker
VERSION=1.1.0
BUILD_TIME=$(shell date '+%Y-%m-%d %H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go 构建参数
LDFLAGS=-ldflags "-s -w -X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)' -X 'main.GitCommit=$(GIT_COMMIT)'"

# 检测操作系统
ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
    EXE_EXT := .exe
    RM_CMD := del /q
    MKDIR_CMD := mkdir
    ECHO_CMD := echo
else
    DETECTED_OS := $(shell uname -s)
    EXE_EXT := 
    RM_CMD := rm -f
    MKDIR_CMD := mkdir -p
    ECHO_CMD := echo
endif

# 默认目标
.PHONY: all
all: build

# 显示帮助信息
.PHONY: help
help:
	@echo "Audio Loss Checker Build Tool"
	@echo "Available Commands:"
	@echo "  build          - Build current platform version"
	@echo "  build-all      - Build all platform versions"
	@echo "  build-windows  - Build Windows version"
	@echo "  build-linux    - Build Linux version"
	@echo "  build-darwin   - Build macOS version"
	@echo "  clean          - Clean build files"
	@echo "  test           - Run tests"
	@echo "  deps           - Download dependencies"
	@echo "  fmt            - Format code"
	@echo "  install        - Install to system"
	@echo "  version        - Show version info"

# 下载依赖
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# 编译当前平台版本
.PHONY: build
build: deps
	@echo "Building $(DETECTED_OS) version..."
	go build $(LDFLAGS) -o $(BINARY_NAME)$(EXE_EXT) .
	@echo "Build completed: $(BINARY_NAME)$(EXE_EXT)"

# 编译所有平台版本
.PHONY: build-all
build-all: build-windows build-linux build-darwin
	@echo "All platforms build completed!"

# 编译 Windows 版本
.PHONY: build-windows
build-windows: deps
	@echo "Building Windows 64-bit version..."
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe .
	@echo "Build completed: dist/$(BINARY_NAME)-windows-amd64.exe"

# 编译 Linux 版本  
.PHONY: build-linux
build-linux: deps
	@echo "Building Linux 64-bit version..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 .
	@echo "Build completed: dist/$(BINARY_NAME)-linux-amd64"

# 编译 macOS 版本
.PHONY: build-darwin
build-darwin: deps
	@echo "Building macOS 64-bit version..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 .
	@echo "Build completed: dist/$(BINARY_NAME)-darwin-amd64"

# 编译 ARM 版本（适用于新的 Mac 和 ARM Linux）
.PHONY: build-arm
build-arm: deps
	@echo "正在编译 ARM64 版本..."
	$(MKDIR_CMD) dist 2>/dev/null || true
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 .
	@echo "ARM64 版本编译完成"

# 创建发布包
.PHONY: release
release: clean build-all build-arm
	@echo "正在创建发布包..."
	$(MKDIR_CMD) release 2>/dev/null || true
ifeq ($(DETECTED_OS),Windows)
	powershell Compress-Archive -Path "dist/*" -DestinationPath "release/$(BINARY_NAME)-$(VERSION)-all-platforms.zip" -Force
	@echo "发布包已创建: release/$(BINARY_NAME)-$(VERSION)-all-platforms.zip"
else
	cd dist && tar -czf ../release/$(BINARY_NAME)-$(VERSION)-all-platforms.tar.gz *
	@echo "发布包已创建: release/$(BINARY_NAME)-$(VERSION)-all-platforms.tar.gz"
endif

# 运行测试
.PHONY: test
test:
	@echo "正在运行测试..."
	go test -v ./...

# 运行基准测试
.PHONY: bench
bench:
	@echo "正在运行基准测试..."
	go test -bench=. -benchmem ./...

# 格式化代码
.PHONY: fmt
fmt:
	@echo "正在格式化代码..."
	go fmt ./...
	gofmt -s -w .

# 代码检查
.PHONY: lint
lint:
	@echo "正在进行代码检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint 未安装，使用 go vet 替代"; \
		go vet ./...; \
	fi

# 安装到系统
.PHONY: install
install: build
	@echo "正在安装到系统..."
ifeq ($(DETECTED_OS),Windows)
	@echo "Windows 用户请手动将 $(BINARY_NAME).exe 复制到 PATH 中的目录"
else
	sudo cp $(BINARY_NAME) /usr/local/bin/
	@echo "已安装到 /usr/local/bin/$(BINARY_NAME)"
endif

# 清理构建文件
.PHONY: clean
clean:
	@echo "正在清理构建文件..."
ifeq ($(DETECTED_OS),Windows)
	-$(RM_CMD) $(BINARY_NAME).exe 2>nul
	-rmdir /s /q dist 2>nul
	-rmdir /s /q release 2>nul
else
	-$(RM_CMD) $(BINARY_NAME)
	-rm -rf dist/
	-rm -rf release/
endif
	@echo "清理完成"

# 显示版本信息
.PHONY: version
version:
	@echo "Audio Loss Checker"
	@echo "版本: $(VERSION)"
	@echo "构建时间: $(BUILD_TIME)"
	@echo "Git提交: $(GIT_COMMIT)"
	@echo "目标平台: $(DETECTED_OS)"

# 快速测试编译结果
.PHONY: test-build
test-build: build
	@echo "正在测试编译结果..."
	./$(BINARY_NAME)$(EXE_EXT) --version
	@echo "测试完成"

# 开发模式（监听文件变化并重新编译）
.PHONY: dev
dev:
	@echo "开发模式需要安装 air 工具: go install github.com/cosmtrek/air@latest"
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "请先安装 air: go install github.com/cosmtrek/air@latest"; \
	fi

# 创建目录结构
dist:
	$(MKDIR_CMD) dist

release:
	$(MKDIR_CMD) release

# 确保目录存在
build-windows build-linux build-darwin: | dist