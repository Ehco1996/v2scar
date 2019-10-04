# Start from the latest golang base image
FROM golang:alpine

# Add Maintainer Info
LABEL maintainer="Ehco1996 <zh19960202@gmail.com>"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -v -o v2scar cli/main.go


# Command to run the executable
CMD ["./v2scar"]