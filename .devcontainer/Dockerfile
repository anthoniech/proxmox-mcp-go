FROM ubuntu:22.04

# Prevent interactive tzdata prompt
ENV DEBIAN_FRONTEND=noninteractive

# Set timezone to Hong Kong
RUN ln -fs /usr/share/zoneinfo/Asia/Hong_Kong /etc/localtime && \
    apt-get update && \
    apt-get install -y \
    curl \
    wget \
    git \
    openssh-client \
    sudo \
    build-essential \
    ca-certificates \
    tzdata \
    unzip \
    gnupg \
    lsb-release \
    software-properties-common \
    && dpkg-reconfigure -f noninteractive tzdata \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

# Create non-root user 'vscode' (standard for DevContainers)
RUN useradd -m vscode && echo "vscode ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Install Go 1.24.0
ENV GOLANG_VERSION=1.24.0
RUN curl -fsSL https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz | tar -C /usr/local -xz \
    && ln -s /usr/local/go/bin/go /usr/local/bin/go \
    && ln -s /usr/local/go/bin/gofmt /usr/local/bin/gofmt

# Set GOPATH and PATH for the vscode user
ENV GOPATH=/home/vscode/go
ENV PATH=$GOPATH/bin:/usr/local/go/bin:$PATH


# Install Go tools (goimports, gopls, golangci-lint, Air) for vscode user
USER vscode
RUN /usr/local/go/bin/go install golang.org/x/tools/cmd/goimports@latest && \
    /usr/local/go/bin/go install golang.org/x/tools/gopls@latest && \
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.62.0 && \
    curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Install Node.js 20 and Yarn
USER root
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y nodejs && \
    npm install -g yarn && \
    npm cache clean --force

# Set working directory
WORKDIR /workspace

# Switch back to vscode user
USER vscode