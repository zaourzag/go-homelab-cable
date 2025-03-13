# Use the official Golang image as the base image
FROM golang:1.20 as builder

# Install VLC and LibVLC
RUN apt-get update && apt-get install -y vlc libvlc-dev

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Go Modules manifests
COPY go.mod go.sum ./

# Download Go modules
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o main ./cmds/cli

# Start a new stage from scratch
FROM debian:bullseye-slim

# Install VLC and LibVLC
RUN apt-get update && apt-get install -y vlc libvlc-dev

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main /app/main

# Expose port 3004 to the outside world
EXPOSE 3004

# Command to run the executable
ENTRYPOINT ["/app/main", "server", "--path", "/app/media"]
