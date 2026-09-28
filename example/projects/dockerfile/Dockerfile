FROM golang:1.23-alpine AS shim-runtime
WORKDIR /workspace/shim
COPY pkg/go.mod pkg/go.sum /workspace/pkg/
COPY pkg /workspace/pkg
COPY shim/go.mod ./
COPY shim /workspace/shim
RUN cd /workspace/pkg && go mod download
RUN go mod tidy
RUN go build -o /shim ./cmd/shim

FROM golang:1.23-alpine
RUN apk add --no-cache ca-certificates
WORKDIR /workspace
COPY --from=shim-runtime /shim /shim
COPY go.work ./
COPY pkg ./pkg
COPY example ./example

RUN apk add curl

# Set APP_DIR to the project directory being deployed, for example
# /workspace/example/projects/backend-2. The shim owns port 9090 and
# launches the application on APP_PORT.
ENV APP_DIR=/workspace/example/projects \
    APP_PORT=8080 \
    PORT=9090 \
    GOWORK=off \
    SHIMNEY_CONFIG_PATH=/workspace/example/config.yaml

ENTRYPOINT ["/shim"]
