FROM golang:1.23-alpine AS builder
WORKDIR /workspace
COPY go.work ./
COPY ./example/projects ./example/projects
RUN go build ./example/projects/backend-1

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /workspace/example/projects/backend-1 /app/backend
CMD ["/app/backend"]
