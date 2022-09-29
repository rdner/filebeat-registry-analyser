BINARY_NAME?=bin/regan
REPO=github.com/rdner/filebeat-registry-analyser

bin/regan: build

.PHONY: build
build:
	@echo "Building ${BINARY_NAME}..."
	@go build -o ${BINARY_NAME} ${REPO}
