# Use the official Go image as the base image
FROM golang:1.23.1-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sailor ./cmd/sailor

# Use a minimal alpine image for the final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create a non-root user
RUN addgroup -g 1001 -S sailor && \
    adduser -u 1001 -S sailor -G sailor

# Set the working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/sailor .

# Create the configs directory
RUN mkdir -p /app/configs && \
    chown -R sailor:sailor /app

# Switch to the non-root user
USER sailor

# Create a volume for the configs directory
VOLUME ["/app/configs"]

# Expose port 7766
EXPOSE 7766

# Run the application
CMD ["./sailor"]
