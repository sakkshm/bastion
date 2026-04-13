APP_NAME := bastion
CMD_PATH := ./cmd/bastion
CONFIG_PATH := config/config.toml
BUILD_DIR := bin

.PHONY: build run fmt clean

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(CMD_PATH)

run:
	go run $(CMD_PATH) run --config="$(CONFIG_PATH)"

fmt:
	go fmt ./...

clean:
	rm -rf $(BUILD_DIR)