BINARY_NAME=main
MAIN_PATH=./cmd/main.go
BUILD_DIR=bin
CONFIG_FILE=config.yaml

.PHONY: help build run test create-config

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

define CONFIG_TEMPLATE
app:
  debug:

http:
  addr:
  write_timeout:
  read_timeout:
  idle_timeout:
  graceful_shutdown_timeout:
endef
export CONFIG_TEMPLATE

help:
	@echo ""
	@echo "$$LOGO_TEXT"
	@echo ""
	@echo "Доступные команды:"
	@echo ""
	@echo "  build          - Собрать исполняемый файл"
	@echo "  run            - Запустить приложение"
	@echo "  test           - Запустить все тесты"
	@echo "  test-cover     - Вывести покрытие тестами"
	@echo "  create-config  - Создать шаблон конфига"
	@echo ""

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

run: build
	@./$(BUILD_DIR)/$(BINARY_NAME)

test:
	go test -v -race ./internal/...

test-cover:
	go test -cover ./internal/...

create-config:
	@if [ -f $(CONFIG_FILE) ]; then \
		echo "$(CONFIG_FILE) already created"; \
	else \
	    echo "$$CONFIG_TEMPLATE" > $(CONFIG_FILE); \
		echo "$(CONFIG_FILE) created"; \
	fi
