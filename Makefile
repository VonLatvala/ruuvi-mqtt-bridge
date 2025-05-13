# Name of the Docker image and container
DOCKER_IMAGE = ruuvi-mqtt-bridge-dev
DOCKER_CONTAINER = ruuvi-dev-container

# Path to the devcontainer Dockerfile
DOCKERFILE_PATH = .devcontainer/Dockerfile

# The default Go version to use in the container
GO_VERSION = go1.24.0

# Build the Docker image for the devcontainer
build:
	docker build -f $(DOCKERFILE_PATH) -t $(DOCKER_IMAGE) .

# Run the devcontainer interactively with the source mounted
run:
	docker run --rm -it \
		-v "$(PWD)":/workspace \
		-w /workspace \
		$(DOCKER_IMAGE) \
		bash

# Build the Go binary inside the devcontainer
build-binary:
	docker run --rm -it \
		-v "$(PWD)":/workspace \
		-w /workspace \
		$(DOCKER_IMAGE) \
		go build -o bin/ruuvi_mqtt_bridge ./src

# Clean up (remove the container image)
clean:
	docker rmi $(DOCKER_IMAGE)
