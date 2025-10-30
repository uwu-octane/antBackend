# Makefile for go-zero API -> OpenAPI (lives in docs/)
# 约定：当前目录为 docs/，生成物直接落在此目录下（无多余子层级）

SHELL := /bin/bash
API_DIR ?= ../gateway
DOCS_DIR ?= .
SWAGGER_V2 ?= $(DOCS_DIR)/doc.json
SWAGGER_HTML ?= $(DOCS_DIR)/doc.html
OPENAPI_DIR ?= $(DOCS_DIR)/openapi
OPENAPI_V3 ?= $(OPENAPI_DIR)/openapi.json

# 可选：指定某个 .api 文件（用于插件直生 v2 的目标）
API_FILE ?= $(API_DIR)/gateway.api

.PHONY: all openapi v2 v3 convert-v2-to-v3 v3-plugin serve clean check-tools

all: openapi

# 1) 生成 v2（Swagger 2.0）：doc.json + doc.html
v2: check-tools
	@echo "==> 生成 Swagger v2 文档到 $(DOCS_DIR)"
	@mkdir -p $(DOCS_DIR)
	# 生成 Markdown 文档（gateway.md 等），已存在时忽略错误
	- goctl api doc --dir $(API_DIR) --o $(DOCS_DIR) || true
	# 使用 goctl-swagger 插件生成 Swagger v2 JSON（doc.json）到 $(DOCS_DIR)
	@cd $(DOCS_DIR) && goctl api plugin -plugin goctl-swagger="swagger -filename doc.json" --api $(abspath $(API_FILE))
	@test -f $(SWAGGER_V2) || (echo "生成失败：未找到 $(SWAGGER_V2)"; exit 1)
	@echo "✓ $(SWAGGER_V2)"
	@echo "✓ $(SWAGGER_HTML)"

# 2) 从 v2 转 v3（OpenAPI 3.x）
convert-v2-to-v3: check-tools
	@echo "==> 将 Swagger v2 转为 OpenAPI v3 到 $(OPENAPI_V3)"
	@test -f $(SWAGGER_V2) || (echo "缺少 $(SWAGGER_V2)，请先执行 make v2"; exit 1)
	@mkdir -p $(OPENAPI_DIR)
	swagger2openapi $(SWAGGER_V2) -o $(OPENAPI_V3) --patch --resolve
	@test -f $(OPENAPI_V3) || (echo "转换失败：未找到 $(OPENAPI_V3)"; exit 1)
	@echo "✓ $(OPENAPI_V3)"

# 3) 一条龙：doc.json -> openapi.json
openapi: v2 convert-v2-to-v3
	@echo "==> OpenAPI v3 已就绪：$(OPENAPI_V3)"

# 5) 本地预览文档（静态服务）
serve:
	@echo "==> 本地预览 docs 目录，监听 http://127.0.0.1:8000"
	@cd $(DOCS_DIR) && python3 -m http.server 8000

# 6) 清理
clean:
	@echo "==> 清理生成物"
	@rm -rf $(DOCS_DIR)/doc.json $(DOCS_DIR)/doc.html $(OPENAPI_DIR)

# 7) 工具检查
check-tools:
	@command -v goctl >/dev/null 2>&1 || { echo "未找到 goctl，请先安装：go install github.com/zeromicro/go-zero/tools/goctl@latest"; exit 1; }
	@command -v swagger2openapi >/dev/null 2>&1 || { echo "未找到 swagger2openapi，请先安装：npm i -g swagger2openapi"; exit 1; }
	@command -v goctl-swagger >/dev/null 2>&1 || { echo "未找到 goctl-swagger，请先安装：go install github.com/zeromicro/goctl-swagger@latest"; exit 1; }


