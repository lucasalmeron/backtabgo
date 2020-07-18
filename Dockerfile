# Start from the latest golang base image
FROM golang:alpine

# Add Maintainer Info
LABEL maintainer="Lucas Salmerón <luko.ar@gmail.com>"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

ENV MONGODB_URI="mongodb://mongo"
ENV PORT="3500"

# Build the Go app
RUN go build cmd/main.go

# This container exposes port 8080 to the outside world
EXPOSE 3500

# Run the binary program produced by `go install`
CMD ["./main"]