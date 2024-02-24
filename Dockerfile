

# Use a smaller base image for the server
FROM --platform=linux/amd64  golang:1.21-alpine3.19 AS builder

# Set the working directory
WORKDIR /build

# Copy the source files
COPY go/ ./
# downloading modules 
RUN go mod download

# Build the app
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o magicmirror ./cmd/main.go

# Run from scratch

FROM --platform=linux/amd64 scratch 

WORKDIR /app 

COPY --from=builder /build/magicmirror /app/magicmirror
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# should default to safe mode
ENV SAFE="true"

ENTRYPOINT ["/app/magicmirror"]

CMD []
