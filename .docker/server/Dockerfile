FROM golang:1.22.5

WORKDIR /app

COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
RUN go mod download

COPY ./cmd/server ./cmd/server
COPY ./internal/server ./internal/server
COPY ./internal/common ./internal/common
COPY ./pkg ./pkg

RUN CGO_ENABLED=0 GOOS=linux go build -C ./cmd/server -o server -ldflags "-X main.buildVersion=dev -X 'main.buildDate=$(date)'"

CMD ["/app/cmd/server/server"]
