FROM golang:1.24-bullseye

# Install common dependencies
RUN apt-get update && apt-get install -y \
    git

ARG USER_ID=1000
ARG GROUP_ID=1000

RUN groupadd -g $GROUP_ID devuser && \
    useradd -m -u $USER_ID -g devuser devuser

USER devuser
WORKDIR /workspace
