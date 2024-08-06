# Use the official Golang image with the version specified in go.mod
FROM golang:1.22.5 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy Go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code of the project
COPY . .

# Compile the application into a statically linked executable
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gophermart ./cmd/gophermart

# Start a new build stage using the scratch image for minimal size
FROM scratch

# Copy the executable from the previous build stage
COPY --from=builder /app/gophermart .

# Define the port that the application will listen on
EXPOSE 8080

# Run the binary
CMD ["./gophermart"]