# 默认目标
.PHONY: tidy build run air clean format help

# Go modules tidy
tidy:
	go mod tidy

# 构建 Go 二进制文件
build: tidy
	go build -o tmp/main cmd/main.go

# 运行 Go 应用
run: build
	tmp/main

# 使用 Air 启动开发环境（自动重载）
air:
	air -c .air.toml

# 清理构建文件
clean:
	rm -rf tmp/*

# 格式化 Go 代码（使用 goimports-reviser）
format:
	goimports-reviser.exe -format -recursive .

# 显示帮助信息
help:
	@echo "Makefile usage:"
	@echo "  make tidy      - Clean and update Go mod dependencies"
	@echo "  make build     - Build the binary"
	@echo "  make run       - Run the built binary"
	@echo "  make air       - Run with Air for hot reloading"
	@echo "  make clean     - Remove build binaries"
	@echo "  make format    - Format Go code using goimports-reviser"
	@echo "  make help      - Show this message"
