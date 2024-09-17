FROM golang:1.22.5

WORKDIR /app

COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
RUN go mod download

COPY ./cmd/agent ./cmd/agent
COPY ./internal/agent ./internal/agent
COPY ./internal/common ./internal/common
COPY ./pkg ./pkg

RUN CGO_ENABLED=0 GOOS=linux go build -C ./cmd/agent -o agent

CMD ["/app/cmd/agent/agent"]
