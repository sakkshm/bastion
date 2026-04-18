APP_NAME := bastion
CMD_PATH := ./cmd/bastion
CONFIG_PATH := config/config.toml
BUILD_DIR := bin

.PHONY: build run fmt clean generate_client_python

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(CMD_PATH)

run:
	go run $(CMD_PATH) run --config="$(CONFIG_PATH)"

fmt:
	go fmt ./...

clean:
	rm -rf $(BUILD_DIR)

generate_client_python:
	rm -rf sdk/python/bastion/generated_client
	openapi-generator-cli generate \
		-i openapi/openapi.yaml \
		-g python \
		-o sdk/python/bastion \
		--package-name generated_client \
		--library urllib3 \
		--global-property=apis,models,supportingFiles \
		--skip-validate-spec \
		--additional-properties=generateApiTests=false,generateModelTests=false,generateSourceCodeOnly=true
	rm -rf sdk/python/bastion/.openapi-generator 
	rm -rf sdk/python/bastion/.openapi-generator-ignore
	rm -rf sdk/python/bastion/generated_client_README.md
