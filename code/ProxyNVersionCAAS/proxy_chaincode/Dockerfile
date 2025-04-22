# Step 1: Use the official Golang image as a base
FROM golang:1.22.2 AS builder

# Set the working directory inside the container
WORKDIR /workspace

# Copy the Go module files to the container
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code to the container
COPY . .

# Step 2: Build the Go executable
RUN CGO_ENABLED=0 GOOS=linux go build -o chaincode .

# Step 3: Create a smaller image for running the chaincode
FROM alpine:latest

# Install necessary dependencies for running Hyperledger Fabric chaincode
RUN apk add --no-cache \
	ca-certificates \
	bash \
	&& update-ca-certificates

# Set the working directory
WORKDIR /chaincode

# Copy the built binary from the builder image
COPY --from=builder /workspace/chaincode .

ENV GRPC_GO_LOG_SEVERITY_LEVEL=info
ENV GRPC_GO_LOG_VERBOSITY_LEVEL=99

# Ensure the binary is executable
RUN chmod +x /chaincode/chaincode

# Set the entry point to run the chaincode
ENTRYPOINT ["./chaincode"]

# Expose the port on which Hyperledger Fabric listens for chaincode invocations
EXPOSE 2051