.DEFAULT_GOAL := build

BINARY_NAME = svc
BUILD_PATH = cmd/build

doc_build:
	swagger generate spec -o ./doc/doc.yaml

build:
	mkdir -p $(BUILD_PATH)
	cp .conf.yml $(BUILD_PATH)/
	cp -R doc $(BUILD_PATH)/
	CGO_ENABLED=0 go build -o $(BUILD_PATH)/$(BINARY_NAME) main.go

clean:
	rm -rf $(BUILD_PATH)
