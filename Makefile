BINARY_NAME=main
MAIN_PATH=./cmd/main.go
BUILD_DIR=bin
DOCS_PKGS=./cmd,./internal/api,./internal/auth/models,./internal/auth/handler,./internal/board/handler/dto,./internal/board/handler,./internal/health/handler,./internal/profile/handler/dto,./internal/profile/handler,./internal/section/handler/dto,./internal/section/handler,./internal/card/handler/dto,./internal/card/handler

.PHONY: help build run proto

define LOGO_TEXT
  /$$$$$$$$$$$$  /$$$$                            /$$$$$$$$$$$$  /$$$$
 /$$$$__  $$$$| $$$$                           /$$$$__  $$$$| $$$$
| $$$$  \\__/| $$$$  /$$$$$$$$$$$$   /$$$$$$$$$$$$$$      | $$$$  \\__/| $$$$  /$$$$$$$$$$$$   /$$$$$$$$$$$$$$
| $$$$      | $$$$ |____  $$$$ /$$$$_____/      | $$$$      | $$$$ |____  $$$$ /$$$$_____/
| $$$$      | $$$$  /$$$$$$$$$$$$$$| $$$$            | $$$$      | $$$$  /$$$$$$$$$$$$$$| $$$$
| $$$$    $$$$| $$$$ /$$$$__  $$$$| $$$$            | $$$$    $$$$| $$$$ /$$$$__  $$$$| $$$$
|  $$$$$$$$$$$$/| $$$$|  $$$$$$$$$$$$$$|  $$$$$$$$$$$$$$      |  $$$$$$$$$$$$/| $$$$|  $$$$$$$$$$$$$$|  $$$$$$$$$$$$$$
 \\______/ |__/ \\_______/ \\_______/       \\______/ |__/ \\_______/ \\_______/

 --- The best team ever
endef
export LOGO_TEXT

help:
	@echo ""
	@echo "$$LOGO_TEXT"
	@echo ""
	@echo "Доступные команды:"
	@echo ""
	@echo "  build          - Собрать исполняемый файл"
	@echo "  run            - Запустить приложение"
	@echo "  docs           - Сгенерировать документацию"
	@echo "  proto          - Сгенерировать proto"
	@echo ""

build:
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

run: build
	@./$(BUILD_DIR)/$(BINARY_NAME)

docs:
	swag init -g main.go -o internal/docs -d $(DOCS_PKGS)

proto:
	protoc --proto_path=. --go_out=. --go_opt=module=github.com/go-park-mail-ru/2026_1_Clac_Clac \
	--go-grpc_out=. --go-grpc_opt=module=github.com/go-park-mail-ru/2026_1_Clac_Clac \
	proto/board/v1/board.proto proto/section/v1/section.proto proto/card/v1/card.proto
