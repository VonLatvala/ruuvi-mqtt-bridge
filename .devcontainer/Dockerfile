FROM golang:1.24-bullseye

# Install common dependencies
RUN apt-get update && apt-get install -y \
    git curl vim build-essential

# Set up work directory
WORKDIR /workspace

# Add Go tools (optional)
ENV GO111MODULE=on
