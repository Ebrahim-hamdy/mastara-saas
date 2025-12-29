# -------------------------------------
# Stage 1: Builder
# -------------------------------------
# Use the official Golang Alpine image that matches our go.mod file's requirement.
FROM golang:1.24-alpine AS builder

# Set necessary environment variables for a static, CGO-free build.
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# Set the working directory inside the container.
WORKDIR /app

# Copy go.mod and go.sum files first to leverage Docker's layer caching.
COPY go.mod go.sum ./

# Download all dependencies.
RUN go mod download

# Copy the entire source code into the container.
COPY . .

# Build the Go application as a static binary.
# -ldflags="-w -s" strips debug information and symbols, reducing binary size.
# -tags netgo ensures static linking of network libraries.
# -o /api names the output binary.
RUN go build -ldflags="-w -s" -tags netgo -o /api ./cmd/api

# -------------------------------------
# Stage 2: Final Production Image
# -------------------------------------
# Use Google's Distroless static image. It contains the bare minimum for a static binary to run,
# including CA certificates and timezone data, but no shell or other utilities.
FROM gcr.io/distroless/static-debian11

# Set a working directory for consistency.
WORKDIR /app

# Create a dedicated, unprivileged user. Distroless images run as non-root by default,
# but we explicitly define the user for clarity and consistency.
USER 65532:65532

# Copy only the compiled binary from the builder stage into our new working directory.
COPY --from=builder /api /app/api

# Expose the port the application will run on. This should match the SERVER_PORT in your .env file.
EXPOSE 8080

# Define the command to run the application.
CMD ["/app/api"]