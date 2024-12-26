# Use the official Go image as the base image
FROM golang:1.21-alpine

# Set the Current Working Directory inside the container
WORKDIR /todo-app

# Install dependencies required to run 'gin'
RUN apk add --no-cache git

# Install gin globally
RUN go install github.com/codegangsta/gin@latest

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all the dependencies (with vendoring)
RUN go mod vendor

# Copy the entire project
COPY . .

# Add Go's bin directory to PATH (in case it's needed for local binaries)
ENV PATH=$PATH:/go/bin

WORKDIR /todo-app/app

# Expose port 8080
EXPOSE 8080

# Run the app using 'gin'
CMD ["go", "run", "."]
