# Use the official Golang image as the base image
FROM golang:1.23.2-alpine

# Install SQLite dependencies
RUN apk add --no-cache gcc musl-dev

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Copy static files to correct location
COPY src/static /app/src/static

# Enable CGO for SQLite
ENV CGO_ENABLED=1
RUN go build -o main .

# Command to run the executable
CMD ["./main"]

# Expose port 8080 to the outside world
EXPOSE 8080
