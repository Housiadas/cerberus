ARG GO_VERSION
# Build the Go Binary.
FROM golang:$GO_VERSION AS build_stage

WORKDIR /home
# Copy the source code into the container.
COPY ./ .
# Build the rest binary.
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -o cerberus-api cmd/rest/main.go

# Run the Go Binary in Alpine.
FROM alpine:3.19 AS application

WORKDIR /app
COPY --from=build_stage /home/cerberus-api .
ENTRYPOINT ["./cerberus-api"]
