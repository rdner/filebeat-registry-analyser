BINARY_NAME=bin/regan
REPO=github.com/rdner/filebeat-registry-analyser

bin/regan:
	go build -o ${BINARY_NAME} ${REPO}
