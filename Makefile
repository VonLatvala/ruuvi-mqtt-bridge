# Name of the Docker image and container
DOCKER_IMAGE = ruuvi-mqtt-bridge-dev
DOCKER_CONTAINER = ruuvi-dev-container

# Path to the devcontainer Dockerfile
DOCKERFILE_PATH = .devcontainer/Dockerfile

BINARY_BASENAME = ruuvi-mqtt-bridge
INSTALL_PATH = /opt/ruuvi-mqtt-bridge/current
SYMLINK_PATH = /usr/local/bin/

# The default Go version to use in the container
GO_VERSION = go1.24.0

# Build the Docker image for the devcontainer
build:
	docker build \
		-f $(DOCKERFILE_PATH) \
		-t $(DOCKER_IMAGE) \
		--build-arg USER_ID=$(shell id -u) \
		--build-arg GROUP_ID=$(shell id -g) \
		.

# Run the devcontainer interactively with the source mounted
run:
	docker run --rm -it \
		-v "$(PWD)":/workspace \
		-w /workspace \
		--user $(shell id -u):$(shell id -g) \
		$(DOCKER_IMAGE) \
		bash

install:
	mkdir -p "$(INSTALL_PATH)"
	cp bin/$(BINARY_BASENAME)-linux-amd64 $(INSTALL_PATH)/$(BINARY_BASENAME)
	ln -sf $(INSTALL_PATH)/$(BINARY_BASENAME) $(SYMLINK_PATH)

# Build the Go binary inside the devcontainer
build-binary:
	docker run --rm -it \
		-v "$(PWD)":/workspace \
		-w /workspace \
		--user $(shell id -u):$(shell id -g) \
		$(DOCKER_IMAGE) \
		go build -o bin/$(BINARY_BASENAME)-linux-amd64 ./src
	chmod +x bin/$(BINARY_BASENAME)-linux-amd64

# Clean up (remove the container image)
clean:
	docker rmi $(DOCKER_IMAGE); rm bin/$(BINARY_BASENAME)-linux-amd64
