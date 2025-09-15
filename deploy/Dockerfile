# Build stage
FROM golang:1.25-alpine AS build

WORKDIR /app

# Copy go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy application source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Copy other necessary files
COPY sqlc.yaml .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /myapp ./cmd/api

# Final stage
FROM alpine:3.18

# Create a non-root user
RUN addgroup -S myapp && adduser -S myapp -G myapp

WORKDIR /app

# Copy the built application from the build stage
COPY --from=build /myapp .

# Expose the application port
EXPOSE 8080

# Switch to the non-root user
USER myapp

# Run the application
CMD ["./myapp"]
