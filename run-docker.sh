#!/bin/bash
set -e

IMAGE_NAME="proxmox-mcp-go"
CONTAINER_NAME="proxmox-mcp-go"
CONFIG_FILE="${CONFIG_FILE:-$PWD/config/config.yaml}"
PORT="${PORT:-3002}"

echo "Building ${IMAGE_NAME}..."
docker build -t "${IMAGE_NAME}" .

# Stop existing container if running
if docker ps -q -f name="${CONTAINER_NAME}" | grep -q .; then
    echo "Stopping existing container..."
    docker stop "${CONTAINER_NAME}"
fi

# Remove existing container if exists
if docker ps -aq -f name="${CONTAINER_NAME}" | grep -q .; then
    docker rm "${CONTAINER_NAME}"
fi

echo "Starting ${CONTAINER_NAME} on port ${PORT}..."
docker run -d \
    --name "${CONTAINER_NAME}" \
    -p "${PORT}:${PORT}" \
    -v "${CONFIG_FILE}:/app/config.yaml:ro" \
    "${IMAGE_NAME}" \
    ./proxmox-mcp-go --config=/app/config.yaml -v

echo "Container started. Logs:"
docker logs -f "${CONTAINER_NAME}"
